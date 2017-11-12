package main

import (
	"github.com/info344-a17/challenges-zicodeng/servers/summary/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":80"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)

	log.Printf("Server is listening at http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
