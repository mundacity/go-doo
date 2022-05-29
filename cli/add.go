package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	fp "github.com/mundacity/flag-parser"
	"github.com/mundacity/go-doo/domain"
	"github.com/mundacity/go-doo/util"
)

// Lets user add new items
type AddCommand struct {
	appCtx       *AppContext
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

	lenMax := aCmd.appCtx.maxLen
	maxIntDigits := aCmd.appCtx.intDigits

	f2 := fp.FlagInfo{FlagName: string(body), FlagType: fp.Str, MaxLen: lenMax}
	f3 := fp.FlagInfo{FlagName: string(mode), FlagType: fp.Str, MaxLen: 1}
	f4 := fp.FlagInfo{FlagName: string(tag), FlagType: fp.Str, MaxLen: lenMax}
	f5 := fp.FlagInfo{FlagName: string(child), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f6 := fp.FlagInfo{FlagName: string(parent), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f7 := fp.FlagInfo{FlagName: string(date), FlagType: fp.DateTime, MaxLen: 20}

	ret = append(ret, f2, f3, f4, f5, f6, f7)
	return ret, nil
}

func NewAddCommand(ctx *AppContext) (*AddCommand, error) {
	addCmd := AddCommand{appCtx: ctx, fs: flag.NewFlagSet("add", flag.ContinueOnError)}

	addCmd.fs.StringVar((*string)(&addCmd.mode), strings.Trim(string(mode), "-"), string(none), "mode of operation: deadline, priority, none (default)")
	addCmd.fs.StringVar(&addCmd.body, strings.Trim(string(body), "-"), "", "main content of the todo item")
	addCmd.fs.StringVar(&addCmd.tagInput, strings.Trim(string(tag), "-"), "", "tag(s) added with/to the new item")
	addCmd.fs.IntVar(&addCmd.childOf, strings.Trim(string(child), "-"), 0, "make item a child of another item")
	addCmd.fs.IntVar(&addCmd.parentOf, strings.Trim(string(parent), "-"), 0, "make item a parent of another item")
	addCmd.fs.StringVar(&addCmd.deadlineDate, strings.Trim(string(deadline), "-"), "", "when item needs to be completed by")

	err := addCmd.SetupFlagMapper(ctx.args)

	return &addCmd, err
}

func (aCmd *AddCommand) SetupFlagMapper(userFlags []string) error {
	canonicalFlags, err := aCmd.GetValidFlags()
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

func getNowString() string {
	n := time.Now()
	return util.StringFromDate(n)
}

// ParseFlags implements method from ICommand interface
func (aCmd *AddCommand) ParseFlags() error {
	newArgs, err := aCmd.parser.ParseUserInput()

	if err != nil {
		return err
	}

	aCmd.appCtx.args = newArgs
	return aCmd.fs.Parse(aCmd.appCtx.args)
}

// Run implements method from ICommand interface
func (aCmd *AddCommand) Run(w io.Writer) error {
	td, _ := aCmd.GenerateTodoItem()

	id, err := aCmd.appCtx.todoRepo.Add(&td)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Creation successful, ItemId: %v\n", id)
	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	return nil
}

func (aCmd *AddCommand) GenerateTodoItem() (domain.TodoItem, error) {
	var td domain.TodoItem

	if len(aCmd.deadlineDate) > 0 {
		aCmd.mode = deadline
	}

	switch aCmd.mode {
	case low:
		td = *domain.NewTodoItem(domain.WithPriorityLevel(domain.Low))
	case medium:
		td = *domain.NewTodoItem(domain.WithPriorityLevel(domain.Medium))
	case high:
		td = *domain.NewTodoItem(domain.WithPriorityLevel(domain.High))
	case deadline:
		td = *domain.NewTodoItem(domain.WithDateBasedPriority(aCmd.deadlineDate))
		d, _ := time.Parse(aCmd.appCtx.DateLayout, aCmd.deadlineDate)
		td.Deadline = d
	default:
		td = *domain.NewTodoItem(domain.WithPriorityLevel(domain.None))
	}

	td.Body = aCmd.body
	td.CreationDate = time.Now()
	td.ParentId = aCmd.childOf

	parseTagInput(&td, aCmd.tagInput, aCmd.appCtx.tagDemlim)
	return td, nil
}

// helper to parse delimited tag input;
// requires <td> tag map to be initialised (e.g. via constructor func)
func parseTagInput(td *domain.TodoItem, input, delim string) {
	if len(input) > 0 {
		tgs := strings.Split(input, delim)
		for _, t := range tgs {
			td.Tags[t] = struct{}{}
		}
	}
}
