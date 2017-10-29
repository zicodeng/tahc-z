package main

import (
	"github.com/go-redis/redis"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/handlers"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/models/users"
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/sessions"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"os"
	"time"
)

//main is the main entry point for the server
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

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	redisStore := sessions.NewRedisStore(redisClient, time.Hour)

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

	mongoStore := users.NewMongoStore(mongoSession, "info_344", "users")

	// Initialize HandlerContext.
	ctx := handlers.NewHandlerContext(sessionKey, redisStore, mongoStore)

	// Create a new mux for the web server.
	mux := http.NewServeMux()

	// Tell the mux to call your handlers.SummaryHandler function
	// when the "/v1/summary" URL path is requested.
	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)

	mux.HandleFunc("/v1/users", ctx.UsersHandler)
	mux.HandleFunc("/v1/users/me", ctx.UsersMeHandler)

	mux.HandleFunc("/v1/sessions", ctx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/mine", ctx.SessionsMineHandler)

	// Wraps mux inside CORSHandler.
	corsMux := handlers.NewCORSHandler(mux)

	// Start a web server listening on the address you read from
	// the environment variable, using the mux you created as
	// the root handler. Use log.Fatal() to report any errors
	// that occur when trying to start the web server.
	log.Printf("Server is listening at https://%s\n", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlscert, tlskey, corsMux))
}
