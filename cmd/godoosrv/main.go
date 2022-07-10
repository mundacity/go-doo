package main

import (
	"log"
	"net/http"

	"github.com/mundacity/go-doo/app"
	"github.com/mundacity/go-doo/srv"
)

func main() {
	app.SetSrvContext()
	// method/s to set up the full todoList to allow for priority queue etc

	mux := http.NewServeMux()
	mux.HandleFunc("/test", srv.TestHandler)
	mux.HandleFunc("/add", srv.AddHandler)
	mux.HandleFunc("/get", srv.GetHandler)
	mux.HandleFunc("/edit", srv.EditHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
