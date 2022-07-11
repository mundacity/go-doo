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
	r := app.SetSrvContext()
	// method/s to set up the full todoList to allow for priority queue etc
	h := srv.Handler{Repo: r}

	mux := http.NewServeMux()
	mux.HandleFunc("/test", h.TestHandler)
	mux.HandleFunc("/add", h.HandleRequests)
	mux.HandleFunc("/get", h.HandleRequests)
	mux.HandleFunc("/edit", h.HandleRequests)

	add := fmt.Sprintf(":%v", sqlite.AppRepo.Port)
	server := http.Server{
		Addr:    add,
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
