package srv

import (
	"fmt"
	"log"
	"net/http"

	godoo "github.com/mundacity/go-doo"
)

type SrvContext struct {
	config  godoo.ServerConfigVals
	handler *Handler
	Server  http.Server
}

func NewSrvContext() godoo.IServerContext {
	sc := &SrvContext{}
	return sc
}

func (s *SrvContext) SetupServerContext(cf godoo.ServerConfigVals) {

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

func (s *SrvContext) Serve() {
	log.Fatal(s.Server.ListenAndServe())
}
