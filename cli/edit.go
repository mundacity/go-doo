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
	lg "github.com/mundacity/go-doo/logging"
)

type EditCommand struct {
	appCtx        *app.AppContext
	parser        fp.FlagParser
	fs            *flag.FlagSet
	getNewVals    bool
	id            int
	body          string
	childOf       int
	deadline      string
	creationDate  string
	tagInput      string // add new, edit/delete existing
	appending     bool
	replacing     bool
	complete      bool
	newTag        string
	newBody       string
	newDeadline   string
	newParent     int
	newlyComplete bool

	//mode priorityMode TODO
}

func NewEditCommand(ctx *app.AppContext) (*EditCommand, error) {
	eCmd := EditCommand{appCtx: ctx, fs: flag.NewFlagSet("edit", flag.ContinueOnError)}
	lg.Logger.Log(lg.Info, "edit command created")

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
	eCmd.fs.BoolVar(&eCmd.newlyComplete, strings.Trim(string(markComplete), "-"), false, "toggle item completion")
	eCmd.fs.StringVar(&eCmd.newTag, strings.Trim(string(changeTag), "-"), "", "change item/s tag")
	eCmd.fs.StringVar(&eCmd.newDeadline, strings.Trim(string(changedDeadline), "-"), "", "change item/s deadline")
	eCmd.fs.StringVar(&eCmd.newBody, strings.Trim(string(changeBody), "-"), "", "change item/s body")
	eCmd.fs.IntVar(&eCmd.newParent, strings.Trim(string(changeParent), "-"), 0, "change item/s parent id")

	var err error
	if err = eCmd.SetupFlagMapper(ctx.Args); err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("flag parser setup error: %v", err), runtime.Caller)
	}

	lg.Logger.Log(lg.Info, "flag parser successfully setup")
	return &eCmd, err
}

func (eCmd *EditCommand) SetupFlagMapper(userFlags []string) error {
	canonicalFlags, err := eCmd.GetValidFlags()
	if err != nil {
		return err
	}

	eCmd.parser = *fp.NewFlagParser(canonicalFlags, userFlags, fp.WithNowAs(_getNowString(), eCmd.appCtx.DateLayout))

	err = eCmd.parser.CheckInitialisation()
	if err != nil {
		return err
	}

	return nil
}

func (eCmd *EditCommand) GetValidFlags() ([]fp.FlagInfo, error) {
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

	ret = append(ret, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14)
	return ret, nil
}

// ParseFlags implements method from ICommand interface
func (eCmd *EditCommand) ParseFlags() error {
	newArgs, err := eCmd.parser.ParseUserInput()

	if err != nil {
		return err
	}

	eCmd.appCtx.Args = newArgs
	return eCmd.fs.Parse(eCmd.appCtx.Args)
}

func (eCmd *EditCommand) GenerateTodoItem() (godoo.TodoItem, error) {
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
		if eCmd.newlyComplete {
			ret.IsComplete = true
		}
	}

	return *ret, nil
}

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

	toEdit, _ := eCmd.GenerateTodoItem()
	eCmd.getNewVals = true
	newVals, _ := eCmd.GenerateTodoItem()
	eCmd.getNewVals = false

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

	getMsg(num, w)
	lg.Logger.Log(lg.Info, "local item successfully edited")
	return nil
}

func getMsg(num int, w io.Writer) {
	s := ""
	if num == 0 || num > 1 {
		s = "s"
	}
	msg := fmt.Sprintf("--> Edited %v item%v\n", num, s)
	w.Write([]byte(msg))
}

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

	getMsg(i, w)
	lg.Logger.Logf(lg.Info, "%v remote item/s successfully edited", i)

	return nil
}

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
		if eCmd.newlyComplete {
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByCompletion})
		}

		lg.Logger.QuickFmtLog(lg.Info, "query options (editing): ", ", ", ret)
	}

	return ret, nil
}
