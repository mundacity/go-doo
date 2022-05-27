package cli

import (
	"io"

	fp "github.com/mundacity/flag-parser"
	"github.com/mundacity/go-doo/domain"
)

type ICommand interface {
	ParseFlags() error
	Run(io.Writer) error
	GetValidFlags() ([]fp.FlagInfo, error)
	SetupFlagMapper(userInput []string) error
	GenerateTodoItem() (domain.TodoItem, error)
}
