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
	// the server should listen on. If empty, default to ":443".
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":443"
	}

	// Path to TLS public certificate.
	tlscert := os.Getenv("TLSCERT")
	// Path to the associated private key.
	tlskey := os.Getenv("TLSKEY")
	if len(tlskey) == 0 || len(tlscert) == 0 {
		log.Fatal("Please set TLSCERT and TLSKEY environment variables")
	}

	// Create a new mux for the web server.
	mux := http.NewServeMux()

	// Tell the mux to call your handlers.SummaryHandler function
	// when the "/v1/summary" URL path is requested.
	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)

	// Start a web server listening on the address you read from
	// the environment variable, using the mux you created as
	// the root handler. Use log.Fatal() to report any errors
	// that occur when trying to start the web server.
	log.Printf("Server is listening at https://%s\n", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlscert, tlskey, mux))
}
