package handlers

import (
	"encoding/json"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/sessions"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"sync"
	"time"
)

// ServiceList contains a list of services.
type ServiceList struct {
	Services map[string]*service
	mx       sync.RWMutex
}

// NewServiceList creates a new ServiceList.
func NewServiceList() *ServiceList {
	return &ServiceList{
		Services: make(map[string]*service),
	}
}

// service represents any microservice our gateway
// will be received from Redis "microservice" channel.
type service struct {
	Name              string
	PathPatternRegexp *regexp.Regexp
	Heartbeat         int // The microservice's normal heartbeat.
	// The key of the Instances map is this instance's unique address.
	Instances map[string]*serviceInstance
	proxy     *httputil.ReverseProxy
}

// newService creates a new microservice.
func newService(name string, pathPatternRegexp *regexp.Regexp, heartbeat int, instances map[string]*serviceInstance) *service {
	addrs := []string{}
	for addr := range instances {
		addrs = append(addrs, addr)
	}
	proxy := newServiceProxy(addrs)
	return &service{name, pathPatternRegexp, heartbeat, instances, proxy}
}

// serviceInstance is an instance of a given microservice.
// A microservice might have multiple instances for balancing loads.
type serviceInstance struct {
	Address       string
	LastHeartbeat time.Time
}

// newServiceInstance creates a new microservice instance.
func newServiceInstance(addr string, lastHeartbeat time.Time) *serviceInstance {
	return &serviceInstance{addr, lastHeartbeat}
}

// ReceivedService represents microservice information received from Redis Pub/Sub.
type ReceivedService struct {
	Name        string
	PathPattern string
	Address     string
	Heartbeat   int
}

// Register either registers a new microservice if it doesn't exist,
// or register a new microservice instance if that microservice already exists in the list.
func (serviceList *ServiceList) Register(receivedSvc *ReceivedService) {
	serviceList.mx.Lock()
	svc, hasSvc := serviceList.Services[receivedSvc.Name]
	// If this microservice is already in our list...
	if hasSvc {
		// Check if this specific microservice instance exists in our list by its unique address...
		instance, hasInstance := svc.Instances[receivedSvc.Address]
		if hasInstance {
			// If this microservice instance is in our list,
			// update its lastHeartbeat time field.
			instance.LastHeartbeat = time.Now()
		} else {
			// If not, add this instance to our list.
			log.Printf("Microservice %s: new instance found\n", receivedSvc.Name)
			svc.Instances[receivedSvc.Address] = newServiceInstance(receivedSvc.Address, time.Now())
		}

	} else {
		// If this microservice is not in our list,
		// create a new instance of that microservice
		// and add to the list.
		log.Printf("New microservice %s found\n", receivedSvc.Name)
		instances := make(map[string]*serviceInstance)
		instances[receivedSvc.Address] = newServiceInstance(receivedSvc.Address, time.Now())
		serviceList.Services[receivedSvc.Name] = newService(
			receivedSvc.Name,
			regexp.MustCompile(receivedSvc.PathPattern),
			receivedSvc.Heartbeat,
			instances,
		)
	}
	serviceList.mx.Unlock()
}

// Remove either removes a dangling microservice if it does not have any active instance running,
// or remove a crashed microservice instance.
func (serviceList *ServiceList) Remove() {
	serviceList.mx.Lock()
	for svcName := range serviceList.Services {
		svc := serviceList.Services[svcName]
		for addr, instance := range svc.Instances {
			if time.Now().Sub(instance.LastHeartbeat).Seconds() > float64(svc.Heartbeat)+10 {
				log.Printf("Microservice %s: crashed instance removed", svcName)
				// Remove the crashed microservice instance from the service list.
				delete(svc.Instances, addr)
				// Remove the entire microservice from the service list
				// if it has no instance running.
				if len(svc.Instances) == 0 {
					log.Printf("Dangling microservice %s removed\n", svcName)
					delete(serviceList.Services, svcName)
				}
			}
		}
	}
	serviceList.mx.Unlock()
}

// DSDHandler is a dynamic service discovery middleware handler
// that checks the requested resource path
// against the pathPattern properties of the Services field.
type DSDHandler struct {
	Handler     http.Handler
	ServiceList *ServiceList
	Context     *HandlerContext
}

// NewDSDHandler wraps another handler into DSDHandler.
func NewDSDHandler(handlerToWrap http.Handler, serviceList *ServiceList, ctx *HandlerContext) *DSDHandler {
	return &DSDHandler{handlerToWrap, serviceList, ctx}
}

// ServeHTTP is a method of DSDHandler.
// Now our DSDHandler is a http.Handler.
func (dsdh *DSDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Validate the user.
	user := dsdh.getCurrentUser(r)
	if user != nil {
		userJSON, err := json.Marshal(user)
		if err != nil {
			log.Printf("error marshaling user: %v", err)
		}
		r.Header.Add("X-User", string(userJSON))
	} else {
		// If there is no user found,
		// explicitly remove X-User header to
		// prevent a hacker who tries to sneak in
		// by setting a fake X-User header in the request.
		r.Header.Del("X-User")
	}

	// Use the received microservice path pattern
	// to determine which microservice should this requset
	// be forwarded to.
	dsdh.ServiceList.mx.RLock()
	defer dsdh.ServiceList.mx.RUnlock()
	for _, svc := range dsdh.ServiceList.Services {
		pattern := svc.PathPatternRegexp
		if pattern.MatchString(r.URL.Path) {
			svc.proxy.ServeHTTP(w, r)
			// Return this function if we find a match,
			// and request is routed to our microservice.
			return
		}
	}

	// If no match is not found,
	// it means this request should not be forwarded to any microservices,
	// just call our real handler to handle it.
	dsdh.Handler.ServeHTTP(w, r)
}

// newServiceProxy forwards relevant requests to microservices based on resource path.
// The microservices should have corresponding handlers that can handle those requests.
func newServiceProxy(addrs []string) *httputil.ReverseProxy {
	i := 0
	mutex := sync.Mutex{}
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			mutex.Lock()
			r.URL.Host = addrs[i%len(addrs)]
			i++
			mutex.Unlock()
			r.URL.Scheme = "http"
		},
	}
}

func (dsdh *DSDHandler) getCurrentUser(r *http.Request) *users.User {
	sessionState := &SessionState{}
	_, err := sessions.GetState(r, dsdh.Context.SigningKey, dsdh.Context.SessionStore, sessionState)
	if err != nil {
		return nil
	}
	return sessionState.User
}
