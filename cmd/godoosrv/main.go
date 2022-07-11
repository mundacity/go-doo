package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mundacity/go-doo/app"
	"github.com/mundacity/go-doo/sqlite"
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

	add := fmt.Sprintf(":%v", sqlite.AppRepo.Port)
	server := http.Server{
		Addr:    add,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
