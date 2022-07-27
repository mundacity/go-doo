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

	fp "github.com/mundacity/flag-parser"
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/app"
	lg "github.com/mundacity/quick-logger"
)

// AddCommand implements the ICommand interface and lets the user add new items to storage
type AddCommand struct {
	appCtx       *app.AppContext
	parser       fp.FlagParser
	fs           *flag.FlagSet
	mode         priorityMode
	body         string //body of the item
	tagInput     string //tags with delimeter set by environment variable
	childOf      int    //child of the int argument
	parentOf     int    //parent of the int argument
	deadlineDate string
}

// Returns a new AddCommand, but also sets up the flagset and parser
func NewAddCommand(ctx *app.AppContext) (*AddCommand, error) {
	addCmd := AddCommand{appCtx: ctx}
	lg.Logger.Log(lg.Info, "add command created")

	addCmd.setupFlagSet()

	err := addCmd.setupFlagMapper(ctx.Args)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("flag parser setup error: %v", err), runtime.Caller)
		return &addCmd, err
	}

	lg.Logger.Log(lg.Info, "flag parser successfully setup")
	return &addCmd, err
}

// Describes the flags and argument types associated with the command
func (aCmd *AddCommand) setupFlagSet() {

	aCmd.fs = flag.NewFlagSet("add", flag.ContinueOnError)

	aCmd.fs.StringVar((*string)(&aCmd.mode), strings.Trim(string(mode), "-"), string(none), "mode of operation: deadline, priority, none (default)")
	aCmd.fs.StringVar(&aCmd.body, strings.Trim(string(body), "-"), "", "main content of the todo item")
	aCmd.fs.StringVar(&aCmd.tagInput, strings.Trim(string(tag), "-"), "", "tag(s) added with/to the new item")
	aCmd.fs.IntVar(&aCmd.childOf, strings.Trim(string(child), "-"), 0, "make item a child of another item")
	aCmd.fs.IntVar(&aCmd.parentOf, strings.Trim(string(parent), "-"), 0, "make item a parent of another item")
	aCmd.fs.StringVar(&aCmd.deadlineDate, strings.Trim(string(deadline), "-"), "", "when item needs to be completed by")
}

// Pass canonical flags and user input to flag-parser package
func (aCmd *AddCommand) setupFlagMapper(userFlags []string) error {
	canonicalFlags, err := aCmd.getValidFlags()
	if err != nil {
		return err
	}

	aCmd.parser = *fp.NewFlagParser(canonicalFlags, userFlags, fp.WithNowAs(getNowString(), aCmd.appCtx.DateLayout))

	err = aCmd.parser.CheckInitialisation()
	if err != nil {
		return err
	}

	return nil
}

// Describes valid flag info for flag-parser
func (aCmd *AddCommand) getValidFlags() ([]fp.FlagInfo, error) {
	var ret []fp.FlagInfo

	lenMax := aCmd.appCtx.MaxLen
	maxIntDigits := aCmd.appCtx.IntDigits

	f2 := fp.FlagInfo{FlagName: string(body), FlagType: fp.Str, MaxLen: lenMax}
	f3 := fp.FlagInfo{FlagName: string(mode), FlagType: fp.Str, MaxLen: 1}
	f4 := fp.FlagInfo{FlagName: string(tag), FlagType: fp.Str, MaxLen: lenMax}
	f5 := fp.FlagInfo{FlagName: string(child), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f6 := fp.FlagInfo{FlagName: string(parent), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f7 := fp.FlagInfo{FlagName: string(date), FlagType: fp.DateTime, MaxLen: 20}

	ret = append(ret, f2, f3, f4, f5, f6, f7)
	return ret, nil
}

// ParseInput implements method from ICommand interface
func (aCmd *AddCommand) ParseInput() error {
	newArgs, err := aCmd.parser.ParseUserInput()

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("user input parsing error: %v", err), runtime.Caller)
		return err
	}

	aCmd.appCtx.Args = newArgs
	lg.Logger.Log(lg.Info, "successfully parsed user input")
	return aCmd.fs.Parse(aCmd.appCtx.Args)
}

// Run implements method from ICommand interface
func (aCmd *AddCommand) Run(w io.Writer) error {

	td, _ := aCmd.setUpItemFromUserInput()

	if aCmd.appCtx.Instance == app.Remote {
		return aCmd.remoteAdd(w, td)
	}

	id, err := aCmd.appCtx.TodoRepo.Add(&td)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("failed to add item: %v", err), runtime.Caller)
		return err
	}

	printAddMessage(int(id), w)
	lg.Logger.Log(lg.Info, "local item successfully added")

	return nil
}

// Populates a godoo.TodoItem with user-supplied data for transer to database
func (aCmd *AddCommand) setUpItemFromUserInput() (godoo.TodoItem, error) {
	var td godoo.TodoItem

	if len(aCmd.deadlineDate) > 0 {
		aCmd.mode = deadline
	}

	switch aCmd.mode {
	case low:
		td = *godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.Low))
	case medium:
		td = *godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.Medium))
	case high:
		td = *godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.High))
	case deadline:
		td = *godoo.NewTodoItem(godoo.WithDateBasedPriority(aCmd.deadlineDate, aCmd.appCtx.DateLayout))
		d, _ := time.Parse(aCmd.appCtx.DateLayout, aCmd.deadlineDate)
		td.Deadline = d
	default:
		td = *godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
	}

	td.Body = aCmd.body
	td.CreationDate = time.Now()
	td.ParentId = aCmd.childOf

	parseTagInput(&td, aCmd.tagInput, aCmd.appCtx.TagDemlim)
	return td, nil
}

func (aCmd *AddCommand) remoteAdd(w io.Writer, td godoo.TodoItem) error {

	// --> very happy path; need to test
	baseUrl := aCmd.appCtx.RemoteUrl + "/add"

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
	rq.Header.Set("content-type", "application/json")

	resp, err := aCmd.appCtx.Client.Do(rq)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error receiving response: %v", err), runtime.Caller)
		return err
	}
	defer resp.Body.Close()

	var i int64

	d := json.NewDecoder(resp.Body)

	if err = d.Decode(&i); err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("json decoding error: %v", err), runtime.Caller)
	}

	printAddMessage(int(i), w)
	lg.Logger.Log(lg.Info, "remote item successfully added")

	return nil
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
