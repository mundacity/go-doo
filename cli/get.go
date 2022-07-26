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

type GetCommand struct {
	appCtx         *app.AppContext
	parser         fp.FlagParser
	fs             *flag.FlagSet
	id             int
	next           bool   // default to priority, but can be changed to date with -d flag
	tagInput       string // tags with delimeter set by environment variable
	bodyPhrase     string // key phrase within body
	childOf        int    // child of the int argument
	parentOf       int    // parent of the int argument
	deadlineDate   string
	creationDate   string
	getAll         bool
	complete       bool
	toggleComplete bool
}

// Returns new get command after setting up flag info and flag-parser
func NewGetCommand(ctx *app.AppContext) (*GetCommand, error) {
	getCmd := GetCommand{appCtx: ctx}
	lg.Logger.Log(lg.Info, "get command created")

	getCmd.setupFlagSet()

	var err error
	if err = getCmd.setupFlagMapper(ctx.Args); err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("flag parser setup error: %v", err), runtime.Caller)
	}

	lg.Logger.Log(lg.Info, "flag parser successfully setup")
	return &getCmd, err
}

// Describes the flags and argument types associated with the command
func (getCmd *GetCommand) setupFlagSet() {
	getCmd.fs = flag.NewFlagSet("get", flag.ContinueOnError)
	getCmd.fs.IntVar(&getCmd.id, strings.Trim(string(itmId), "-"), 0, "search by item id")
	getCmd.fs.BoolVar(&getCmd.next, strings.Trim(string(next), "-"), false, "get next item")

	/* Deadline -d:
	 	+ default = single fullstop; ignore if default
		+ if called with no value (i.e. empty '-d' flag), modifies -n flag to return by date instead of defaulting to priority
		+ if not default but has value (i.e. called with a non-nil string), assumed to be a date to search for
	*/

	getCmd.fs.StringVar(&getCmd.deadlineDate, strings.Trim(string(date), "-"), ".", "date of existing item; if empty, modifies -n to return based on date instead of defaulting to priority")
	getCmd.fs.StringVar(&getCmd.creationDate, strings.Trim(string(creation), "-"), "", "creation date of existing item")
	getCmd.fs.StringVar(&getCmd.tagInput, strings.Trim(string(tag), "-"), "", "search by item tag")
	getCmd.fs.StringVar(&getCmd.bodyPhrase, strings.Trim(string(body), "-"), "", "search by known phrase within body")

	getCmd.fs.IntVar(&getCmd.childOf, strings.Trim(string(child), "-"), 0, "search based on parent Id; requested item is child of provided parent id")
	getCmd.fs.IntVar(&getCmd.parentOf, strings.Trim(string(parent), "-"), 0, "search based on child Id; requested item is parent of provided child id")
	getCmd.fs.BoolVar(&getCmd.getAll, strings.Trim(string(all), "-"), false, "get all items")
	getCmd.fs.BoolVar(&getCmd.complete, strings.Trim(string(finished), "-"), false, "search for completed items")
	getCmd.fs.BoolVar(&getCmd.toggleComplete, strings.Trim(string(markComplete), "-"), false, "search for unfinished items")

}

// Pass canonical flags and user input to flag-parser package
func (getCmd *GetCommand) setupFlagMapper(userFlags []string) error {
	canonicalFlags, err := getCmd.getValidFlags()
	if err != nil {
		return err
	}

	getCmd.parser = *fp.NewFlagParser(canonicalFlags, userFlags, fp.WithNowAs(getNowString(), getCmd.appCtx.DateLayout))

	err = getCmd.parser.CheckInitialisation()
	if err != nil {
		return err
	}

	return nil
}

// Describes valid flag info for flag-parser
func (getCmd *GetCommand) getValidFlags() ([]fp.FlagInfo, error) {
	var ret []fp.FlagInfo

	maxIntDigits := getCmd.appCtx.IntDigits

	lenMax := getCmd.appCtx.MaxLen

	f8 := fp.FlagInfo{FlagName: string(body), FlagType: fp.Str, MaxLen: lenMax}
	f2 := fp.FlagInfo{FlagName: string(itmId), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f3 := fp.FlagInfo{FlagName: string(next), FlagType: fp.Boolean, Standalone: true} // TODO: implement
	f4 := fp.FlagInfo{FlagName: string(date), FlagType: fp.DateTime, MaxLen: 21, AllowDateRange: true}
	f5 := fp.FlagInfo{FlagName: string(tag), FlagType: fp.Str, MaxLen: lenMax}
	f6 := fp.FlagInfo{FlagName: string(child), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f7 := fp.FlagInfo{FlagName: string(parent), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f9 := fp.FlagInfo{FlagName: string(creation), FlagType: fp.DateTime, MaxLen: 21, AllowDateRange: true}
	f10 := fp.FlagInfo{FlagName: string(all), FlagType: fp.Boolean, Standalone: true}
	f11 := fp.FlagInfo{FlagName: string(finished), FlagType: fp.Boolean, Standalone: true}
	f12 := fp.FlagInfo{FlagName: string(markComplete), FlagType: fp.Boolean, Standalone: true}

	ret = append(ret, f8, f2, f3, f4, f5, f6, f7, f9, f10, f11, f12)
	return ret, nil
}

// ParseInput implements method from ICommand interface
func (getCmd *GetCommand) ParseInput() error {
	newArgs, err := getCmd.parser.ParseUserInput()

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("user input parsing error: %v", err), runtime.Caller)
		return err
	}

	getCmd.appCtx.Args = newArgs
	lg.Logger.Log(lg.Info, "successfully parsed user input")
	return getCmd.fs.Parse(getCmd.appCtx.Args)
}

// Implements Run() method from ICommand interface
func (gCmd *GetCommand) Run(w io.Writer) error {

	input, _ := gCmd.interpretUserInput()

	var itms []godoo.TodoItem
	var err error

	qList, err := gCmd.determineQueryType()
	if err != nil {
		return err
	}

	fullQry := godoo.FullUserQuery{QueryOptions: qList, QueryData: input}

	if gCmd.appCtx.Instance == app.Remote {
		return gCmd.remoteGet(w, fullQry)
	}

	itms, err = gCmd.appCtx.TodoRepo.GetWhere(fullQry)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("failed to get item: %v", err), runtime.Caller)
		return err
	}

	msg := getOutputGenerationFunc(itms)
	w.Write([]byte(msg()))
	lg.Logger.Log(lg.Info, "local item successfully retrieved")

	return nil
}

// Populates a godoo.TodoItem with user-supplied data to query database
func (gCmd *GetCommand) interpretUserInput() (godoo.TodoItem, error) {
	ret := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))

	ret.Id = gCmd.id
	if gCmd.childOf != 0 {
		ret.ParentId = gCmd.childOf
		ret.IsChild = true
	}

	if gCmd.parentOf != 0 {
		ret.ChildItems[gCmd.parentOf] = struct{}{}
	}

	if gCmd.creationDate != "" {
		splt := strings.Split(gCmd.creationDate, ":")
		ret.CreationDate, _ = time.Parse(gCmd.appCtx.DateLayout, splt[0]) //only ever need first one
	}
	if gCmd.deadlineDate != "" && gCmd.deadlineDate != "." {
		splt := strings.Split(gCmd.deadlineDate, ":")
		ret.Deadline, _ = time.Parse(gCmd.appCtx.DateLayout, splt[0])
	}
	if gCmd.bodyPhrase != "" {
		ret.Body = gCmd.bodyPhrase
	}
	if gCmd.tagInput != "" {
		ret.Tags[gCmd.tagInput] = struct{}{}
	}
	if gCmd.complete {
		ret.IsComplete = true
	} else if gCmd.toggleComplete {
		ret.IsComplete = false
	}
	return *ret, nil
}

// Determines the different elements by which the user is searching for an item
func (gCmd *GetCommand) determineQueryType() ([]godoo.UserQueryOption, error) {
	var ret []godoo.UserQueryOption

	// by id numbers
	if gCmd.id != 0 {
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ById})
		return ret, nil // no further search params needed; id unique
	}
	if gCmd.parentOf != 0 {
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByChildId})
		return ret, nil // no further search params needed; only 1 parent possible
	}
	if gCmd.childOf != 0 {
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByParentId})
	}

	if gCmd.next {
		if gCmd.deadlineDate == "" {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByNextDate})
		} else if gCmd.deadlineDate == "." {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByNextPriority})
		}
	}

	// by string
	if gCmd.tagInput != "" {
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByTag})
	}
	if gCmd.bodyPhrase != "" {
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByBody})
	}

	// by times
	if gCmd.deadlineDate != "" {
		if gCmd.deadlineDate != "." {
			d := getUpperDateBound(gCmd.deadlineDate, gCmd.appCtx.DateLayout)
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByDeadline, UpperBoundDate: d})
		}
	}
	if gCmd.creationDate != "" {
		d := getUpperDateBound(gCmd.creationDate, gCmd.appCtx.DateLayout)
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByCreationDate, UpperBoundDate: d})
	}
	if gCmd.complete || gCmd.toggleComplete {
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByCompletion})
	}

	lg.Logger.QuickFmtLog(lg.Info, "query options (getting): ", ", ", ret)
	return ret, nil
}

// Coordinates request/response in remote mode
func (gCmd *GetCommand) remoteGet(w io.Writer, fq godoo.FullUserQuery) error {

	// --> very happy path; need to test
	baseUrl := gCmd.appCtx.RemoteUrl + "/get"

	// building request
	body, err := json.Marshal(fq)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("json marshalling error: %v", err), runtime.Caller)
		return err
	}

	rq, err := http.NewRequest("GET", baseUrl, bytes.NewBuffer(body))
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("request generation error: %v", err), runtime.Caller)
		return err
	}
	rq.Header.Set("content-type", "application/json")

	// getting response
	resp, err := gCmd.appCtx.Client.Do(rq)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error receiving response: %v", err), runtime.Caller)
		return err
	}
	defer resp.Body.Close()

	var itms []godoo.TodoItem
	d := json.NewDecoder(resp.Body)
	var err2 error

	if err2 = d.Decode(&itms); err2 != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("json decoding error: %v", err2), runtime.Caller)
	}

	// printing to console
	msg := getOutputGenerationFunc(itms)
	w.Write([]byte(msg()))

	lg.Logger.Logf(lg.Info, "successfully retrieved %v item/s", len(itms))

	return nil
}
