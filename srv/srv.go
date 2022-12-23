package srv

import (
	"fmt"
	"log"
	"net/http"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/auth"
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

	authConf := auth.NewAuthConfig(s.config.KeyPath, s.config.UserPasswordHash, s.config.ExpirationLimit)

	mux := http.NewServeMux()
	mux.HandleFunc("/test", s.handler.TestHandler)
	mux.HandleFunc("/add", auth.ValidateJwt(s.config.KeyPath, s.handler.HandleRequests))
	mux.HandleFunc("/get", auth.ValidateJwt(s.config.KeyPath, s.handler.HandleRequests))
	mux.HandleFunc("/edit", auth.ValidateJwt(s.config.KeyPath, s.handler.HandleRequests))
	mux.HandleFunc("/authenticate", auth.Authenticate(authConf, s.handler.AuthHandler))

	add := fmt.Sprintf(":%v", s.config.Port)
	s.Server = http.Server{
		Addr:    add,
		Handler: mux,
	}
}

// TODO: reconfigure to use middleware in idiomatic style
func (s *SrvContext) Serve() {
	log.Fatal(s.Server.ListenAndServe())
}
