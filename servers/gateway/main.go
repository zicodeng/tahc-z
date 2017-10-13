package main

import (
	"github.com/info344-a17/challenges-zicodeng/servers/gateway/handlers"
	"log"
	"net/http"
	"os"
)

//main is the main entry point for the server
func main() {
	// Read the ADDR environment variable to get the address
	// the server should listen on. If empty, default to ":80"
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":80"
	}

	// Create a new mux for the web server.
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("../../clients/summary")))

	// Tell the mux to call your handlers.SummaryHandler function
	// when the "/v1/summary" URL path is requested.
	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)

	// Start a web server listening on the address you read from
	// the environment variable, using the mux you created as
	// the root handler. Use log.Fatal() to report any errors
	// that occur when trying to start the web server.
	log.Printf("Server is listening at http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
