package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	godoo "github.com/mundacity/go-doo"
	lg "github.com/mundacity/quick-logger"
)

// AddCommand implements the ICommand interface and lets the user add new items to storage
type AddCommand struct {
	conf         *godoo.ConfigVals
	args         []string //in conf but easier to deal with
	fs           *flag.FlagSet
	mode         priorityMode
	body         string //body of the item
	tagInput     string //tags with delimeter set by environment variable
	childOf      int    //child of the int argument
	parentOf     int    //parent of the int argument
	deadlineDate string
}

// Returns a new AddCommand, but also sets up the flagset and parser
func NewAddCommand(config *godoo.ConfigVals) *AddCommand {
	addCmd := AddCommand{}
	addCmd.args = config.Args
	addCmd.conf = config
	lg.Logger.Log(lg.Info, "add command created")

	addCmd.setupFlagSet()

	return &addCmd
}

// ParseInput implements method from ICommand interface
func (aCmd *AddCommand) ParseInput() error {
	newArgs, err := aCmd.conf.Parser.ParseUserInput()

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("user input parsing error: %v", err), runtime.Caller)
		return err
	}

	aCmd.args = newArgs
	lg.Logger.Log(lg.Info, "successfully parsed user input")
	return aCmd.fs.Parse(aCmd.args)
}

// Run implements method from ICommand interface
func (aCmd *AddCommand) Run(w io.Writer) error {

	td, _ := aCmd.BuildItemFromInput()

	if aCmd.conf.Instance == godoo.Remote {
		return aCmd.remoteAdd(w, td)
	}

	id, err := aCmd.conf.TodoRepo.Add(&td)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("failed to add item: %v", err), runtime.Caller)
		return err
	}

	printAddMessage(int(id), w)
	lg.Logger.Log(lg.Info, "local item successfully added")

	return nil
}

// Populates a godoo.TodoItem with user-supplied data for transer to database
func (aCmd *AddCommand) BuildItemFromInput() (godoo.TodoItem, error) {

	var td godoo.TodoItem

	if len(aCmd.deadlineDate) > 0 {
		aCmd.mode = deadline
	}

	td, err := getItemWithPriority(*aCmd)
	if err != nil {
		return td, err
	}

	td.Body = aCmd.body
	td.CreationDate, _ = time.Parse(aCmd.conf.DateLayout, aCmd.conf.DateLayout)
	td.ParentId = aCmd.childOf

	parseTagInput(&td, aCmd.tagInput, aCmd.conf.TagDelim)
	return td, nil
}

func (aCmd *AddCommand) CheckConfig() *godoo.ConfigVals {
	return aCmd.conf
}

func (aCmd *AddCommand) remoteAdd(w io.Writer, td godoo.TodoItem) error {

	baseUrl := aCmd.conf.RemoteUrl + "/add"

	body, err := json.Marshal(td)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("json marshalling error: %v", err), runtime.Caller)
		return err
	}

	rq, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer(body))
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("request generation error: %v", err), runtime.Caller)
		return err
	}

	resp, err := remoteRun(rq, aCmd)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var i int64

	d := json.NewDecoder(resp.Body)
	d.DisallowUnknownFields()
	if err = d.Decode(&i); err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("json decoding error: %v", err), runtime.Caller)
		return err
	}

	printAddMessage(int(i), w)
	lg.Logger.Log(lg.Info, "remote item successfully added")

	return nil
}

// Describes the flags and argument types associated with the command
func (aCmd *AddCommand) setupFlagSet() {

	aCmd.fs = flag.NewFlagSet("add", flag.ContinueOnError)

	aCmd.fs.StringVar((*string)(&aCmd.mode), strings.Trim(string(godoo.Mode), "-"), string(none), "mode of operation: deadline, priority, none (default)")
	aCmd.fs.StringVar(&aCmd.body, strings.Trim(string(godoo.Body), "-"), "", "main content of the todo item")
	aCmd.fs.StringVar(&aCmd.tagInput, strings.Trim(string(godoo.Tag), "-"), "", "tag(s) added with/to the new item")
	aCmd.fs.IntVar(&aCmd.childOf, strings.Trim(string(godoo.Child), "-"), 0, "make item a child of another item")
	aCmd.fs.IntVar(&aCmd.parentOf, strings.Trim(string(godoo.Parent), "-"), 0, "make item a parent of another item")
	aCmd.fs.StringVar(&aCmd.deadlineDate, strings.Trim(string(godoo.Date), "-"), "", "when item needs to be completed by")
}

// helper to parse delimited tag input;
// requires <td> tag map to be initialised (e.g. via constructor func)
func parseTagInput(td *godoo.TodoItem, input, delim string) {
	if len(input) > 0 {
		tgs := strings.Split(input, delim)
		for _, t := range tgs {
			td.Tags[t] = struct{}{}
		}
	}
}

// returns a new ToDoItem based on priority level
func getItemWithPriority(aCmd AddCommand) (godoo.TodoItem, error) {

	switch aCmd.mode {
	case low:
		return *godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.Low)), nil
	case medium:
		return *godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.Medium)), nil
	case high:
		return *godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.High)), nil
	case deadline:
		return *godoo.NewTodoItem(godoo.WithDateBasedPriority(aCmd.deadlineDate, aCmd.conf.DateLayout)), nil
	case none:
		return *godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None)), nil
	default:
		return godoo.TodoItem{}, &InvalidArgumentError{}
	}
}
