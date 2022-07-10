package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mundacity/go-doo/app"
)

func main() {
	path := os.Getenv("DATA_FILE_PATH")
	if len(path) == 0 {
		path = ".\\data.json"
	}

	app.SetSrvContext()

	mux := http.NewServeMux()
	// mux.HandleFunc("/healthcheck", handlers.HealthCheckHandler)
	// mux.HandleFunc("/", handlers.SecretHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
