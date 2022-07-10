package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	path := os.Getenv("DATA_FILE_PATH")
	if len(path) == 0 {
		path = ".\\data.json"
	}

	// store.Setup(path)

	mux := http.NewServeMux()
	// mux.HandleFunc("/healthcheck", handlers.HealthCheckHandler)
	// mux.HandleFunc("/", handlers.SecretHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
