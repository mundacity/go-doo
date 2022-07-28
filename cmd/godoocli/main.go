package main

import (
	"os"

	"github.com/mundacity/go-doo/app"
)

func main() {
	os.Exit(app.RunCliApp(os.Args[1:], os.Stdout)) // first arg is app name
}
