package main

import (
	"os"

	"github.com/mundacity/go-doo/cli"
)

func main() {
	os.Exit(cli.RunCli(os.Args[1:], os.Stdout)) // first arg is app name
}
