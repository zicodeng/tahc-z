package main

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/handlers"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/attempts"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/resetcodes"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/sessions"
	"github.com/streadway/amqp"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"os"
	"regexp"
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
	pubsub := redisClient.Subscribe("microservices")

	serviceList := handlers.NewServiceList()
	go listenForServices(pubsub, serviceList)
	// Remove crashed microservices.
	go removeCrashedServices(serviceList)

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

	notifier := handlers.NewNotifier()
	mux.Handle("/v1/ws", ctx.NewWebSocketsHandler(notifier))
	mqAddr := os.Getenv("MQADDR")
	if len(mqAddr) == 0 {
		log.Fatal("Please set the MQADDR variable to the address of your MQ server")
	}
	go listenToMQ(mqAddr, notifier)

	// Hard-code the network addresses where our microservice instances
	// are listening into environment variables the gateway reads at startup.
	// This is an easy way to get started,
	// but it doesn't make it easy to dynamically add
	// new instances of an existing microservice, or entirely new microservices.

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

type receivedService struct {
	Name        string
	PathPattern string
	Address     string
	Heartbeat   int
}

// Constantly listen for "Microservices" Redis channel.
func listenForServices(pubsub *redis.PubSub, serviceList *handlers.ServiceList) {
	log.Println("Listening for microservices")
	for {
		msg, err := receivePubSubMessage(pubsub)
		// If there is still an error receiving message even after retries,
		// return this function.
		if err != nil {
			log.Println(err)
			return
		}
		svc := &receivedService{}
		err = json.Unmarshal([]byte(msg.Payload), svc)
		if err != nil {
			log.Printf("Error unmarshalling received microservice JSON to struct: %v", err)
		}
		serviceList.Mx.Lock()
		_, hasSvc := serviceList.Services[svc.Name]
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
				log.Printf("Microservice %s: new instance found\n", svc.Name)
				serviceList.Services[svc.Name].Instances[svc.Address] = handlers.NewServiceInstance(svc.Address, time.Now())
			}

		} else {
			// If this microservice is not in our list,
			// create a new instance of that microservice
			// and add to the list.
			log.Printf("New microservice %s found\n", svc.Name)
			instances := make(map[string]*handlers.ServiceInstance)
			instances[svc.Address] = handlers.NewServiceInstance(svc.Address, time.Now())
			serviceList.Services[svc.Name] = handlers.NewService(svc.Name, regexp.MustCompile(svc.PathPattern), svc.Heartbeat, instances)
		}
		serviceList.Mx.Unlock()
	}
}

var maxReceiveMessageRetries = 5

// If there is an error receiving Redis Pub/Sub messages,
// that's probably because the Redis server is no longer reachable.
// If that's the case, try to receive the message again for a max number of retries.
func receivePubSubMessage(pubsub *redis.PubSub) (*redis.Message, error) {
	var msg *redis.Message
	var err error
	for i := 0; i < maxReceiveMessageRetries; i++ {
		// pubsub.ReceiveMessage() will block until there is a message to receive.
		msg, err = pubsub.ReceiveMessage()
		if err == nil {
			return msg, nil
		}
		log.Printf("Error receiving message from Redis Pub/Sub: %s", err)
		log.Printf("Will try again in %d seconds", i*2)
		time.Sleep(time.Duration(i*2) * time.Second)
	}
	return nil, err
}

// Periodically looks for service instances
// for which we haven't received a heartbeat in a while,
// and remove those instances from your list
func removeCrashedServices(serviceList *handlers.ServiceList) {
	for {
		time.Sleep(time.Second * 10)

		serviceList.Mx.Lock()
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
		serviceList.Mx.Unlock()
	}
}

const maxConnRetries = 5
const qName = "testQ"

func listenToMQ(addr string, notifier *handlers.Notifier) {
	conn, err := connectToMQ(addr)
	if err != nil {
		log.Fatalf("error connecting to MQ server: %s", err)
	}
	log.Printf("connected to MQ server")
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("error opening channel: %v", err)
	}
	log.Println("created MQ channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(qName, false, false, false, false, nil)
	if err != nil {
		log.Fatalf("error declaring queue: %v", err)
	}
	log.Println("declared MQ queue")
	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("error listening to queue: %v", err)
	}
	log.Println("listening for new MQ messages...")
	for msg := range messages {
		// log.Printf("new message id %s received from MQ", string(msg.Body))
		// Load messages received from RabbitMQ's eventQ channel to
		// notifier's eventQ channel, so that messages will be
		// broadcasted to all clients throught websocket.
		notifier.Notify(msg.Body)
	}
}

func connectToMQ(addr string) (*amqp.Connection, error) {
	mqURL := "amqp://" + addr
	var conn *amqp.Connection
	var err error
	for i := 1; i <= maxConnRetries; i++ {
		conn, err = amqp.Dial(mqURL)
		if err == nil {
			return conn, nil
		}
		log.Printf("error connecting to MQ server at %s: %s", mqURL, err)
		log.Printf("will attempt another connection in %d seconds", i*2)
		time.Sleep(time.Duration(i*2) * time.Second)
	}
	return nil, err
}
