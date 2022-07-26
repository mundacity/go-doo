package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mundacity/go-doo/app"
	"github.com/mundacity/go-doo/util"
)

// Sets out the methods implemented by commands that the user can execute
type ICommand interface {
	ParseInput() error
	Run(io.Writer) error
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

	err = cmd.ParseInput()
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

// if user is using a date range, get the upper bound of that range
func getUpperDateBound(dateText string, dateLayout string) time.Time {
	splt := splitDates(dateText)
	var d time.Time

	if len(splt) > 1 {
		d, _ = time.Parse(dateLayout, splt[1])
	}

	return d
}

func splitDates(s string) []string {
	return strings.Split(s, ":")
}

// Helper function used by commands when setting the 'now' time for flag-parser
func getNowString() string {
	n := time.Now()
	return util.StringFromDate(n)
}
