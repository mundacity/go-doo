package cli

import (
	"io"

	fp "github.com/mundacity/flag-parser"
	godoo "github.com/mundacity/go-doo"
)

type ICommand interface {
	ParseFlags() error
	Run(io.Writer) error
	GetValidFlags() ([]fp.FlagInfo, error)
	SetupFlagMapper(userInput []string) error
	GenerateTodoItem() (godoo.TodoItem, error)
}
