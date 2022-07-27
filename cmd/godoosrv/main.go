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

	r, pl := app.SetSrvContext()
	h := srv.NewHandler(pl, r)

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
