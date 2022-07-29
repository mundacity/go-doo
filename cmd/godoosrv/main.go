package main

import (
	"github.com/mundacity/go-doo/app"
	"github.com/mundacity/go-doo/srv"
)

func main() {

	cf := app.GetSrvConfig()
	ct := srv.NewSrvContext()
	ct.SetupServerContext(cf)

	ct.Serve()
}
