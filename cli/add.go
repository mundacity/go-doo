package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	fp "github.com/mundacity/flag-parser"
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/app"
	"github.com/mundacity/go-doo/util"
)

// Lets user add new items
type AddCommand struct {
	appCtx       *app.AppContext
	parser       fp.FlagParser
	fs           *flag.FlagSet
	mode         priorityMode
	body         string // body of the item
	tagInput     string // tags with delimeter set by environment variable
	childOf      int    // child of the int argument
	parentOf     int    // parent of the int argument
	deadlineDate string
}

type priorityMode string

const (
	deadline priorityMode = "d"
	none     priorityMode = "n"
	low      priorityMode = "l"
	medium   priorityMode = "m"
	high     priorityMode = "h"
)

func (aCmd *AddCommand) GetValidFlags() ([]fp.FlagInfo, error) { // too tired to come up with anything more elegant - todo!
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

func NewAddCommand(ctx *app.AppContext) (*AddCommand, error) {
	addCmd := AddCommand{appCtx: ctx, fs: flag.NewFlagSet("add", flag.ContinueOnError)}

	addCmd.fs.StringVar((*string)(&addCmd.mode), strings.Trim(string(mode), "-"), string(none), "mode of operation: deadline, priority, none (default)")
	addCmd.fs.StringVar(&addCmd.body, strings.Trim(string(body), "-"), "", "main content of the todo item")
	addCmd.fs.StringVar(&addCmd.tagInput, strings.Trim(string(tag), "-"), "", "tag(s) added with/to the new item")
	addCmd.fs.IntVar(&addCmd.childOf, strings.Trim(string(child), "-"), 0, "make item a child of another item")
	addCmd.fs.IntVar(&addCmd.parentOf, strings.Trim(string(parent), "-"), 0, "make item a parent of another item")
	addCmd.fs.StringVar(&addCmd.deadlineDate, strings.Trim(string(deadline), "-"), "", "when item needs to be completed by")

	err := addCmd.SetupFlagMapper(ctx.Args)

	return &addCmd, err
}

func (aCmd *AddCommand) SetupFlagMapper(userFlags []string) error {
	canonicalFlags, err := aCmd.GetValidFlags()
	if err != nil {
		return err
	}

	aCmd.parser = *fp.NewFlagParser(canonicalFlags, userFlags, fp.WithNowAs(_getNowString(), aCmd.appCtx.DateLayout))

	err = aCmd.parser.CheckInitialisation()
	if err != nil {
		return err
	}

	return nil
}

func _getNowString() string {
	n := time.Now()
	return util.StringFromDate(n)
}

// ParseFlags implements method from ICommand interface
func (aCmd *AddCommand) ParseFlags() error {
	newArgs, err := aCmd.parser.ParseUserInput()

	if err != nil {
		return err
	}

	aCmd.appCtx.Args = newArgs
	return aCmd.fs.Parse(aCmd.appCtx.Args)
}

// Run implements method from ICommand interface
func (aCmd *AddCommand) Run(w io.Writer) error {

	td, _ := aCmd.GenerateTodoItem()

	if aCmd.appCtx.Instance == app.Remote {
		return aCmd.remoteAdd(w, td)
	}

	id, err := aCmd.appCtx.TodoRepo.Add(&td)
	if err != nil {
		return err
	}

	printMsg(int(id), w)

	return nil
}

func printMsg(id int, w io.Writer) {
	msg := fmt.Sprintf("Creation successful, ItemId: %v\n", id)
	w.Write([]byte(msg))
}

func (aCmd *AddCommand) remoteAdd(w io.Writer, td godoo.TodoItem) error {

	// --> very happy path; need to test
	baseUrl := "http://localhost:8080/add"

	body, err := json.Marshal(td)
	if err != nil {
		return err
	}

	rq, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	rq.Header.Set("content-type", "application/json")

	resp, err := aCmd.appCtx.Client.Do(rq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var i int64

	d := json.NewDecoder(resp.Body)
	d.Decode(&i)

	printMsg(int(i), w)

	return nil
}

func (aCmd *AddCommand) GenerateTodoItem() (godoo.TodoItem, error) {
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
