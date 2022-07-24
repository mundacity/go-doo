package cli

import (
	"errors"
	"fmt"
	"io"

	fp "github.com/mundacity/flag-parser"
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/app"
)

type ICommand interface {
	ParseFlags() error
	Run(io.Writer) error
	GetValidFlags() ([]fp.FlagInfo, error)
	SetupFlagMapper(userInput []string) error
	GenerateTodoItem() (godoo.TodoItem, error)
}

// Flags used throughout the system
type CMD_FLAG string

const (
	all             CMD_FLAG = "-a"
	body            CMD_FLAG = "-b"
	child           CMD_FLAG = "-c"
	date            CMD_FLAG = "-d"
	creation        CMD_FLAG = "-e" // e for existence!
	finished        CMD_FLAG = "-f" // item complete
	itmId           CMD_FLAG = "-i"
	mode            CMD_FLAG = "-m"
	next            CMD_FLAG = "-n"
	parent          CMD_FLAG = "-p"
	tag             CMD_FLAG = "-t"
	changeBody      CMD_FLAG = "-B" //append or replace
	changeParent    CMD_FLAG = "-C"
	changedDeadline CMD_FLAG = "-D"
	markComplete    CMD_FLAG = "-F"
	changeMode      CMD_FLAG = "-M"
	changeTag       CMD_FLAG = "-T" //append, replace, or remove
	appendMode      CMD_FLAG = "--append"
	replaceMode     CMD_FLAG = "--replace"
)

// RunCli is the main entry point of the cli client application.
// It passes initial setup off to the 'app' package and then
// passes execution off to the relevant command
func RunCli(osArgs []string, w io.Writer) int {

	app, err := app.SetupCli(osArgs)
	if err != nil {
		fmt.Printf("%v", err)
		return 2
	}

	cmd, err := getBasicCommand(app)
	if err != nil {
		fmt.Printf("%v", err)
		return 2
	}

	err = cmd.ParseFlags()
	if err != nil {
		fmt.Printf("error: '%v'", err)
		return 2
	}

	err = cmd.Run(w)
	if err != nil {
		fmt.Printf("error: '%v'", err)
		return 2
	}

	return 0
}

func getBasicCommand(ctx *app.AppContext) (ICommand, error) {
	arg1 := ctx.Args[0]
	ctx.Args = ctx.Args[1:]
	var cmd ICommand
	var err error

	switch arg1 {
	case "add":
		cmd, err = NewAddCommand(ctx)
	case "get":
		cmd, err = NewGetCommand(ctx)
	case "edit":
		cmd, err = NewEditCommand(ctx)
	default:
		return nil, errors.New("invalid command")
	}
	return cmd, err
}
