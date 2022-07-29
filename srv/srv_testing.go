package srv

import (
	"fmt"
	"net/http"

	godoo "github.com/mundacity/go-doo"
)

type FakeSrvContext struct {
	config  godoo.ServerConfigVals
	handler *Handler
	Server  http.Server
}

func (s *FakeSrvContext) SetupServerContext(cf godoo.ServerConfigVals) {

	s.config = cf
	s.handler = NewHandler(cf)

	mux := http.NewServeMux()
	mux.HandleFunc("/test", s.handler.TestHandler)
	mux.HandleFunc("/add", s.handler.HandleRequests)
	mux.HandleFunc("/get", s.handler.HandleRequests)
	mux.HandleFunc("/edit", s.handler.HandleRequests)

	add := fmt.Sprintf(":%v", s.config.Port)
	s.Server = http.Server{
		Addr:    add,
		Handler: mux,
	}
}

func (s *FakeSrvContext) Serve() {
	//log.Fatal(s.Server.ListenAndServe())
}
