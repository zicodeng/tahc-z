package main

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/info344-a17/challenges-zicodeng/servers/summary/handlers"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":80"
	}

	// Set up Redis connection.
	redisAddr := os.Getenv("REDISADDR")
	if len(redisAddr) == 0 {
		redisAddr = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	go publishService(addr, redisClient)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)

	log.Printf("Server is listening at http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// summaryService contains information about this microservice.
type summaryService struct {
	Name        string
	PathPattern string
	Address     string
}

// publishes information about this microservice to Redis Pub/Sub.
func publishService(addr string, redisClient *redis.Client) {
	sumSvc := &summaryService{
		Name:        "summary",
		PathPattern: "/v1/summary",
		Address:     addr,
	}

	j, err := json.Marshal(sumSvc)
	if nil != err {
		log.Printf("error marshalling struct to JSON: %v\n", err)
	}

	for _ = range time.Tick(time.Second * 10) {
		redisClient.Publish("microservices", j)
	}
}
