package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	fp "github.com/mundacity/flag-parser"
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/app"
	lg "github.com/mundacity/quick-logger"
)

type EditCommand struct {
	appCtx            *app.AppContext
	parser            fp.FlagParser
	fs                *flag.FlagSet
	getNewVals        bool
	id                int
	body              string
	childOf           int
	deadline          string
	creationDate      string
	tagInput          string //add new, edit/delete existing
	appending         bool
	replacing         bool
	complete          bool
	newTag            string
	newBody           string
	newDeadline       string
	newParent         int
	newToggleComplete bool
	newPriority       priorityMode
}

// Sets up flag info & parser before returning a new edit comman
func NewEditCommand(ctx *app.AppContext) (*EditCommand, error) {
	eCmd := EditCommand{appCtx: ctx}
	lg.Logger.Log(lg.Info, "edit command created")

	eCmd.setupFlagSet()

	var err error
	if err = eCmd.setupFlagMapper(ctx.Args); err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("flag parser setup error: %v", err), runtime.Caller)
	}

	lg.Logger.Log(lg.Info, "flag parser successfully setup")
	return &eCmd, err
}

// Describes the flags and argument types associated with the command
func (eCmd *EditCommand) setupFlagSet() {

	eCmd.fs = flag.NewFlagSet("edit", flag.ContinueOnError)

	// selectors to determine which items to edit
	eCmd.fs.IntVar(&eCmd.id, strings.Trim(string(itmId), "-"), 0, "edit item by id")
	eCmd.fs.IntVar(&eCmd.childOf, strings.Trim(string(child), "-"), 0, "edit items by parentId - i.e. child of id passed")
	eCmd.fs.StringVar(&eCmd.deadline, strings.Trim(string(date), "-"), "", "edit items by deadline")
	eCmd.fs.StringVar(&eCmd.creationDate, strings.Trim(string(creation), "-"), "", "edit items by creation date")
	eCmd.fs.StringVar(&eCmd.body, strings.Trim(string(body), "-"), "", "edit items by body keyword")
	eCmd.fs.StringVar(&eCmd.tagInput, strings.Trim(string(tag), "-"), "", "edit items by tag")
	eCmd.fs.BoolVar(&eCmd.complete, strings.Trim(string(finished), "-"), false, "edit by completed items")

	// edit mode
	eCmd.fs.BoolVar(&eCmd.appending, strings.Trim(string(appendMode), "-"), false, "append new input to end of existing body/tag")
	eCmd.fs.BoolVar(&eCmd.replacing, strings.Trim(string(replaceMode), "-"), false, "replace existing body/tag with new input")

	// elements of item/s to edit
	eCmd.fs.BoolVar(&eCmd.newToggleComplete, strings.Trim(string(markComplete), "-"), false, "toggle item completion")
	eCmd.fs.StringVar(&eCmd.newTag, strings.Trim(string(changeTag), "-"), "", "change item/s tag")
	eCmd.fs.StringVar(&eCmd.newDeadline, strings.Trim(string(changedDeadline), "-"), "", "change item/s deadline")
	eCmd.fs.StringVar(&eCmd.newBody, strings.Trim(string(changeBody), "-"), "", "change item/s body")
	eCmd.fs.IntVar(&eCmd.newParent, strings.Trim(string(changeParent), "-"), 0, "change item/s parent id")
	eCmd.fs.StringVar((*string)(&eCmd.newPriority), strings.Trim(string(changeMode), "-"), "", "change item/s priority mode - low/medium/high")
}

// Pass canonical flags and user input to flag-parser package
func (eCmd *EditCommand) setupFlagMapper(userFlags []string) error {
	canonicalFlags, err := eCmd.getValidFlags()
	if err != nil {
		return err
	}

	eCmd.parser = *fp.NewFlagParser(canonicalFlags, userFlags, fp.WithNowAs(getNowString(), eCmd.appCtx.DateLayout))

	err = eCmd.parser.CheckInitialisation()
	if err != nil {
		return err
	}

	return nil
}

// Describes valid flag info for flag-parser
func (eCmd *EditCommand) getValidFlags() ([]fp.FlagInfo, error) {
	var ret []fp.FlagInfo

	maxIntDigits := eCmd.appCtx.IntDigits
	lenMax := eCmd.appCtx.MaxLen

	f1 := fp.FlagInfo{FlagName: string(body), FlagType: fp.Str, MaxLen: lenMax}
	f2 := fp.FlagInfo{FlagName: string(itmId), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f3 := fp.FlagInfo{FlagName: string(date), FlagType: fp.DateTime, MaxLen: 21, AllowDateRange: true}
	f4 := fp.FlagInfo{FlagName: string(tag), FlagType: fp.Str, MaxLen: lenMax}
	f5 := fp.FlagInfo{FlagName: string(child), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f6 := fp.FlagInfo{FlagName: string(creation), FlagType: fp.DateTime, MaxLen: 21, AllowDateRange: true}
	f14 := fp.FlagInfo{FlagName: string(finished), FlagType: fp.Boolean, Standalone: true}

	f7 := fp.FlagInfo{FlagName: string(appendMode), FlagType: fp.Boolean, Standalone: true}
	f8 := fp.FlagInfo{FlagName: string(replaceMode), FlagType: fp.Boolean, Standalone: true}

	f9 := fp.FlagInfo{FlagName: string(changeBody), FlagType: fp.Str, MaxLen: lenMax}
	f10 := fp.FlagInfo{FlagName: string(changeTag), FlagType: fp.Str, MaxLen: lenMax}
	f11 := fp.FlagInfo{FlagName: string(changeParent), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f12 := fp.FlagInfo{FlagName: string(changedDeadline), FlagType: fp.DateTime, MaxLen: 20}
	f13 := fp.FlagInfo{FlagName: string(markComplete), FlagType: fp.Boolean, Standalone: true}
	f15 := fp.FlagInfo{FlagName: string(changeMode), FlagType: fp.Str, MaxLen: 1}

	ret = append(ret, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15)
	return ret, nil
}

// ParseInput implements method from ICommand interface
func (eCmd *EditCommand) ParseInput() error {
	newArgs, err := eCmd.parser.ParseUserInput()

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("user input parsing error: %v", err), runtime.Caller)
		return err
	}

	eCmd.appCtx.Args = newArgs
	lg.Logger.Log(lg.Info, "successfully parsed user input")
	return eCmd.fs.Parse(eCmd.appCtx.Args)
}

// Implements ICommand Run() method
func (eCmd *EditCommand) Run(w io.Writer) error {

	eCmd.getAdditionalInput()
	srchQryLst, err := eCmd.determineQueryType(godoo.Get)
	if err != nil {
		return err
	}
	edtQryLst, err := eCmd.determineQueryType(godoo.Update)
	if err != nil {
		return err
	}

	toEdit, err := eCmd.setupTodoItemBasedOnUserInput()
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error while interpreting user input: %v", err), runtime.Caller)
		return err
	}

	eCmd.getNewVals = true
	newVals, err := eCmd.setupTodoItemBasedOnUserInput()
	eCmd.getNewVals = false
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error while interpreting user input: %v", err), runtime.Caller)
		return err
	}

	srchFq := godoo.FullUserQuery{QueryOptions: srchQryLst, QueryData: toEdit}
	edtFq := godoo.FullUserQuery{QueryOptions: edtQryLst, QueryData: newVals}

	if eCmd.appCtx.Instance == app.Remote {
		return eCmd.remoteEdit(w, srchFq, edtFq)
	}

	num, err := eCmd.appCtx.TodoRepo.UpdateWhere(srchFq, edtFq)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("failed to edit item: %v", err), runtime.Caller)
		return err
	}

	printEditMessage(num, w)
	lg.Logger.Log(lg.Info, "local item successfully edited")
	return nil
}

// Checks whether user replacing or appending to existing item bodies
func (eCmd *EditCommand) getAdditionalInput() error {
	if len(eCmd.newBody) > 0 || len(eCmd.newTag) > 0 {
		if !eCmd.appending && !eCmd.replacing {
			// get user input to figure what they want
			fmt.Print("\nNo edit mode specified. Choose append (a), replace (r). Any other key to cancel...\n")

			lg.Logger.Log(lg.Info, "user asked for additional input")

			rdr := bufio.NewReader(os.Stdin)
			choice, _, err := rdr.ReadRune()
			if err != nil {
				lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error receiving additional user input: %v", err), runtime.Caller)
				fmt.Printf("Error occurred: %v, cancelling operation.", err)
			}

			if choice == 'r' {
				eCmd.replacing = true
			} else if choice == 'a' {
				eCmd.appending = true
			} else {
				lg.Logger.Logf(lg.Warning, "invalid additional user input: %v", choice)
				return errors.New("cancelling operation")
			}
			lg.Logger.Logf(lg.Info, "user choice: %v", choice)
		}
	}
	return nil
}

// Populates a godoo.TodoItem with user-supplied data to pass
// to database for querying/editing
func (eCmd *EditCommand) setupTodoItemBasedOnUserInput() (godoo.TodoItem, error) {
	ret := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))

	if !eCmd.getNewVals { // searching
		ret.Id = eCmd.id
		if eCmd.childOf != 0 {
			ret.ParentId = eCmd.childOf
			ret.IsChild = true
		}
		if eCmd.creationDate != "" {
			splt := strings.Split(eCmd.creationDate, ":")
			ret.CreationDate, _ = time.Parse(eCmd.appCtx.DateLayout, splt[0]) //whether range or not, only ever going to need first one
		}
		if eCmd.deadline != "" {
			splt := strings.Split(eCmd.deadline, ":")
			ret.Deadline, _ = time.Parse(eCmd.appCtx.DateLayout, splt[0])
		}
		if eCmd.body != "" {
			ret.Body = eCmd.body
		}
		if eCmd.tagInput != "" {
			ret.Tags[eCmd.tagInput] = struct{}{}
		}

		ret.IsComplete = eCmd.complete

	} else {
		if eCmd.newParent != 0 {
			ret.ParentId = eCmd.newParent
			ret.IsChild = true
		}
		if eCmd.newDeadline != "" {
			ret.Deadline, _ = time.Parse(eCmd.appCtx.DateLayout, eCmd.newDeadline)
		}
		if eCmd.newBody != "" {
			if eCmd.appending {
				ret.Body = " " + eCmd.newBody // 'old bodynew body' vs. 'old body new body'
			}
			ret.Body = eCmd.newBody
		}
		if eCmd.newTag != "" {
			ret.Tags[eCmd.newTag] = struct{}{}
		}
		if eCmd.newToggleComplete {
			ret.IsComplete = true
		}
		if len(string(eCmd.newPriority)) > 0 {
			p, err := convertPriority(string(eCmd.newPriority))
			if err != nil {
				lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("priority conversion error: %v", err), runtime.Caller)
				return *ret, err
			}

			ret.Priority = p
		}
	}

	return *ret, nil
}

func convertPriority(s string) (godoo.PriorityLevel, error) {
	sl := strings.ToLower(s)
	switch sl {
	case "l":
		return godoo.Low, nil
	case "m":
		return godoo.Medium, nil
	case "h":
		return godoo.High, nil
	default:
		return godoo.None, &InvalidArgumentError{}
	}
}

// In remote mode, coordinates request & response to/from remote server.
func (eCmd *EditCommand) remoteEdit(w io.Writer, srchFq, edtFq godoo.FullUserQuery) error {
	// --> very happy path; need to test
	baseUrl := eCmd.appCtx.RemoteUrl + "/edit"

	s := []godoo.FullUserQuery{srchFq, edtFq}

	body, err := json.Marshal(s)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("json marshalling error: %v", err), runtime.Caller)
		return err
	}

	rq, err := http.NewRequest("PUT", baseUrl, bytes.NewBuffer(body))
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("request generation error: %v", err), runtime.Caller)
		return err
	}
	rq.Header.Set("content-type", "application/json")

	resp, err := eCmd.appCtx.Client.Do(rq)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error receiving response: %v", err), runtime.Caller)
		return err
	}
	defer resp.Body.Close()

	var i int

	d := json.NewDecoder(resp.Body)

	if err = d.Decode(&i); err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("json decoding error: %v", err), runtime.Caller)
	}

	printEditMessage(i, w)
	lg.Logger.Logf(lg.Info, "%v remote item/s successfully edited", i)

	return nil
}

// Interprets user input to determine intentions in both the search and edit
// portions of input. If no edit options provided, returns error.
func (eCmd *EditCommand) determineQueryType(qType godoo.QueryType) ([]godoo.UserQueryOption, error) {
	var ret []godoo.UserQueryOption

	switch qType {
	case godoo.Get:
		// by id numbers
		if eCmd.id != 0 {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ById})
		}
		if eCmd.childOf != 0 {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByParentId})
		}

		// by string
		if eCmd.tagInput != "" {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByTag})
		}
		if eCmd.body != "" {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByBody})
		}

		// by times
		if eCmd.deadline != "" {
			d := getUpperDateBound(eCmd.deadline, eCmd.appCtx.DateLayout)
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByDeadline, UpperBoundDate: d})
		}
		if eCmd.creationDate != "" {
			d := getUpperDateBound(eCmd.creationDate, eCmd.appCtx.DateLayout)
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByCreationDate, UpperBoundDate: d})
		}
		if eCmd.complete {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByCompletion})
		}

		lg.Logger.QuickFmtLog(lg.Info, "query options (getting): ", ", ", ret)

	case godoo.Update:
		if eCmd.newParent != 0 {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByParentId})
		}

		// by string
		if eCmd.newTag != "" {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByTag})
		}
		if eCmd.newBody != "" {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByBody})
		}

		// by times
		if eCmd.newDeadline != "" {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByDeadline})
		}
		if eCmd.appending {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByAppending})
		}
		if eCmd.replacing {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByReplacement})
		}
		if eCmd.newToggleComplete {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByCompletion})
		}
		if len(string(eCmd.newPriority)) > 0 {
			//ret.Priority = converPriority(string(eCmd.newPriority))
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByNextPriority})
		}

		if len(ret) == 0 {
			lg.Logger.LogWithCallerInfo(lg.Error, "no edit options provided", runtime.Caller)
			return ret, &NoEditInstructionsError{}
		}

		lg.Logger.QuickFmtLog(lg.Info, "query options (editing): ", ", ", ret)
	}

	return ret, nil
}
