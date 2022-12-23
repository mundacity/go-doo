package srv

import (
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
}

func (s *FakeSrvContext) Serve() {}
