package main

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/handlers"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/attempts"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/resetcodes"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/sessions"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"os"
	"time"
)

// main is the main entry point for the server.
func main() {
	// Read the ADDR environment variable to get the address
	// the server should listen on. If empty, default to ":443".
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = "localhost:443"
	}

	// Path to TLS public certificate.
	tlscert := os.Getenv("TLSCERT")
	// Path to the associated private key.
	tlskey := os.Getenv("TLSKEY")
	if len(tlskey) == 0 || len(tlscert) == 0 {
		log.Fatal("Please set TLSCERT and TLSKEY environment variables")
	}

	// sessionKey is the signing key for SessionID.
	sessionKey := os.Getenv("SESSIONKEY")
	if len(sessionKey) == 0 {
		log.Fatal("Please set SESSIONKEY environment variable")
	}

	// Set up Redis connection.
	redisAddr := os.Getenv("REDISADDR")
	if len(redisAddr) == 0 {
		redisAddr = "localhost:6379"
	}

	// Shared Redis client.
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Listening for microservices.
	serviceList := handlers.NewServiceList()
	go listenForServices(redisClient, serviceList)

	// Redis store for storing SessionState.
	sessionStore := sessions.NewRedisStore(redisClient, time.Hour)

	// Redis store for storing Attempt.
	attemptStore := attempts.NewRedisStore(redisClient)

	// Redis store for storing ResetCode.
	resetCodeStore := resetcodes.NewRedisStore(redisClient, resetcodes.CodeDuration)

	// Set up MongoDB connection.
	dbAddr := os.Getenv("DBADDR")
	if len(dbAddr) == 0 {
		dbAddr = "localhost:27017"
	}

	// Create a Mongo session.
	mongoSession, err := mgo.Dial(dbAddr)
	if err != nil {
		log.Fatalf("error dialing mongo: %v", err)
	}

	userStore := users.NewMongoStore(mongoSession, "info_344", "users")

	// Loading existing users into Trie at start-up.
	trie := userStore.Index()

	// Initialize HandlerContext.
	ctx := handlers.NewHandlerContext(sessionKey, trie, sessionStore, userStore, attemptStore, resetCodeStore)

	// Messaging microservice addresses.
	// msgAddrs := os.Getenv("MESSAGESVCADDR")
	// if len(msgAddrs) == 0 {
	// 	log.Fatal("Please set MESSAGESVCADDR environment variables")
	// }
	// msgAddrSlice := strings.Split(msgAddrs, ",")

	// // Summary microservice addresses.
	// sumAddrs := os.Getenv("SUMMARYSVCADDR")
	// if len(sumAddrs) == 0 {
	// 	log.Fatal("Please set SUMMARYSVCADDR environment variables")
	// }
	// sumAddrSlice := strings.Split(sumAddrs, ",")

	// Create a new mux for the web server.
	mux := http.NewServeMux()

	// Gateway
	mux.HandleFunc("/v1/users", ctx.UsersHandler)
	mux.HandleFunc("/v1/users/me", ctx.UsersMeHandler)

	mux.HandleFunc("/v1/sessions", ctx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/mine", ctx.SessionsMineHandler)

	mux.HandleFunc("/v1/resetcodes", ctx.ResetCodesHandler)
	mux.HandleFunc("/v1/passwords", ctx.ResetPasswordHandler)

	// Messaging microservice.
	// mux.Handle("/v1/channels", newServiceProxy(msgAddrSlice, ctx))
	// mux.Handle("/v1/channels/", newServiceProxy(msgAddrSlice, ctx))
	// mux.Handle("/v1/messages/", newServiceProxy(msgAddrSlice, ctx))

	// Summary microservice.
	// mux.Handle("/v1/summary", newServiceProxy(sumAddrSlice, ctx))

	// Chained middlewares.
	// Wraps mux inside DSDHandler.
	dsdMux := handlers.NewDSDHandler(mux, serviceList, ctx)
	// Wraps mux inside CORSHandler.
	corsMux := handlers.NewCORSHandler(dsdMux)

	// Start a web server listening on the address you read from
	// the environment variable, using the mux you created as
	// the root handler. Use log.Fatal() to report any errors
	// that occur when trying to start the web server.
	log.Printf("Server is listening at https://%s\n", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlscert, tlskey, corsMux))
}

// Constantly listen for "microservices" Redis channel.
func listenForServices(redisClient *redis.Client, serviceList *handlers.ServiceList) {
	pubsub := redisClient.Subscribe("microservices")
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Println(err)
		}
		svc := &receivedService{}
		err = json.Unmarshal([]byte(msg.Payload), svc)
		if err != nil {
			log.Printf("error unmarshalling received microservice JSON to struct: %v", err)
		}

		// Add the received microservice instance into our service list.
		_, hasSvc := serviceList.Services[svc.Name]
		//  add mutex here
		// If this microservice is already in our list...
		if hasSvc {
			// Check if this specific microservice instance exists in our list by its unique address...
			_, hasInstance := serviceList.Services[svc.Name].Instances[svc.Address]
			if hasInstance {
				// If this microservice instance is in our list,
				// update its lastHeartbeat time field.
				serviceList.Services[svc.Name].Instances[svc.Address].LastHeartbeat = time.Now()
			} else {
				// If not, add this instance to our list.
				serviceList.Services[svc.Name].Instances[svc.Address] = handlers.NewServiceInstance(svc.Address, time.Now())
			}

		} else {
			// If this microservice is not in our list,
			// create a new instance of that microservice
			// and add to the list.
			instances := make(map[string]*handlers.ServiceInstance)
			instances[svc.Address] = handlers.NewServiceInstance(svc.Address, time.Now())
			serviceList.Services[svc.Name] = handlers.NewService(svc.Name, svc.PathPattern, instances)
		}
	}
}

type receivedService struct {
	Name        string
	PathPattern string
	Address     string
}
