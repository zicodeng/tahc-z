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
	Services map[string]*Service
	Mx       sync.RWMutex
}

// NewServiceList creates a new ServiceList.
func NewServiceList() *ServiceList {
	return &ServiceList{
		Services: make(map[string]*Service),
	}
}

// Service represents any microservice our gateway
// will be received from Redis "microservice" channel.
type Service struct {
	Name        string
	PathPattern string
	Heartbeat   int // The microservice's normal heartbeat.
	// The key of the Instances map is this instance's unique address.
	Instances map[string]*ServiceInstance
}

// NewService creates a new microservice.
func NewService(name string, pathPattern string, heartbeat int, instances map[string]*ServiceInstance) *Service {
	return &Service{name, pathPattern, heartbeat, instances}
}

// ServiceInstance is an instance of a given microservice.
// A microservice might have multiple instances for balancing loads.
type ServiceInstance struct {
	Address       string
	LastHeartbeat time.Time
}

// NewServiceInstance creates a new microservice instance.
func NewServiceInstance(addr string, lastHeartbeat time.Time) *ServiceInstance {
	return &ServiceInstance{addr, lastHeartbeat}
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
	// Use the received microservice path pattern
	// to determine which microservice should this requset
	// be forwarded to.
	dsdh.ServiceList.Mx.RLock()
	defer dsdh.ServiceList.Mx.RUnlock()
	for _, svc := range dsdh.ServiceList.Services {
		pattern := svc.PathPattern
		re := regexp.MustCompile(pattern)
		if re.MatchString(r.URL.Path) {
			addrs := []string{}
			for addr := range svc.Instances {
				addrs = append(addrs, addr)
			}
			proxy := dsdh.newServiceProxy(addrs)
			proxy.ServeHTTP(w, r)
		}
	}

	// If no match is not found,
	// it means this request should not be forwarded to any microservices,
	// just call our real handler to handle it.
	dsdh.Handler.ServeHTTP(w, r)
}

// newServiceProxy forwards relevant requests to microservices based on resource path.
// The microservices should have corresponding handlers that can handle those requests.
func (dsdh *DSDHandler) newServiceProxy(addrs []string) *httputil.ReverseProxy {
	i := 0
	mutex := sync.Mutex{}
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			user := dsdh.getCurrentUser(r)
			if user != nil {
				userJSON, err := json.Marshal(user)
				if err != nil {
					log.Printf("error marshaling user: %v", err)
				}
				r.Header.Add("X-User", string(userJSON))
			}
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
