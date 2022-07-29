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

	godoo "github.com/mundacity/go-doo"
	lg "github.com/mundacity/quick-logger"
)

type EditCommand struct {
	conf              *godoo.ConfigVals
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
func NewEditCommand(conf *godoo.ConfigVals) *EditCommand {
	eCmd := EditCommand{}
	eCmd.conf = conf
	lg.Logger.Log(lg.Info, "edit command created")

	eCmd.setupFlagSet()

	return &eCmd
}

// Describes the flags and argument types associated with the command
func (eCmd *EditCommand) setupFlagSet() {

	eCmd.fs = flag.NewFlagSet("edit", flag.ContinueOnError)

	// selectors to determine which items to edit
	eCmd.fs.IntVar(&eCmd.id, strings.Trim(string(godoo.ItmId), "-"), 0, "edit item by id")
	eCmd.fs.IntVar(&eCmd.childOf, strings.Trim(string(godoo.Child), "-"), 0, "edit items by parentId - i.e. child of id passed")
	eCmd.fs.StringVar(&eCmd.deadline, strings.Trim(string(godoo.Date), "-"), "", "edit items by deadline")
	eCmd.fs.StringVar(&eCmd.creationDate, strings.Trim(string(godoo.Creation), "-"), "", "edit items by creation date")
	eCmd.fs.StringVar(&eCmd.body, strings.Trim(string(godoo.Body), "-"), "", "edit items by body keyword")
	eCmd.fs.StringVar(&eCmd.tagInput, strings.Trim(string(godoo.Tag), "-"), "", "edit items by tag")
	eCmd.fs.BoolVar(&eCmd.complete, strings.Trim(string(godoo.Finished), "-"), false, "edit by completed items")

	// edit mode
	eCmd.fs.BoolVar(&eCmd.appending, strings.Trim(string(godoo.AppendMode), "-"), false, "append new input to end of existing body/tag")
	eCmd.fs.BoolVar(&eCmd.replacing, strings.Trim(string(godoo.ReplaceMode), "-"), false, "replace existing body/tag with new input")

	// elements of item/s to edit
	eCmd.fs.BoolVar(&eCmd.newToggleComplete, strings.Trim(string(godoo.MarkComplete), "-"), false, "toggle item completion")
	eCmd.fs.StringVar(&eCmd.newTag, strings.Trim(string(godoo.ChangeTag), "-"), "", "change item/s tag")
	eCmd.fs.StringVar(&eCmd.newDeadline, strings.Trim(string(godoo.ChangedDeadline), "-"), "", "change item/s deadline")
	eCmd.fs.StringVar(&eCmd.newBody, strings.Trim(string(godoo.ChangeBody), "-"), "", "change item/s body")
	eCmd.fs.IntVar(&eCmd.newParent, strings.Trim(string(godoo.ChangeParent), "-"), 0, "change item/s parent id")
	eCmd.fs.StringVar((*string)(&eCmd.newPriority), strings.Trim(string(godoo.ChangeMode), "-"), "", "change item/s priority mode - low/medium/high")
}

// ParseInput implements method from ICommand interface
func (eCmd *EditCommand) ParseInput() error {
	newArgs, err := eCmd.conf.Parser.ParseUserInput()

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("user input parsing error: %v", err), runtime.Caller)
		return err
	}

	eCmd.conf.Args = newArgs
	lg.Logger.Log(lg.Info, "successfully parsed user input")
	return eCmd.fs.Parse(eCmd.conf.Args)
}

// Implements ICommand Run() method
func (eCmd *EditCommand) Run(w io.Writer) error {

	eCmd.getAdditionalInput()
	srchQryLst, err := eCmd.DetermineQueryType(godoo.Get)
	if err != nil {
		return err
	}
	edtQryLst, err := eCmd.DetermineQueryType(godoo.Update)
	if err != nil {
		return err
	}

	toEdit, err := eCmd.BuildItemFromInput()
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error while interpreting user input: %v", err), runtime.Caller)
		return err
	}

	eCmd.getNewVals = true
	newVals, err := eCmd.BuildItemFromInput()
	eCmd.getNewVals = false
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error while interpreting user input: %v", err), runtime.Caller)
		return err
	}

	srchFq := godoo.FullUserQuery{QueryOptions: srchQryLst, QueryData: toEdit}
	edtFq := godoo.FullUserQuery{QueryOptions: edtQryLst, QueryData: newVals}

	if eCmd.conf.Instance == godoo.Remote {
		return eCmd.remoteEdit(w, srchFq, edtFq)
	}

	num, err := eCmd.conf.TodoRepo.UpdateWhere(srchFq, edtFq)
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
func (eCmd *EditCommand) BuildItemFromInput() (godoo.TodoItem, error) {
	ret := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))

	if !eCmd.getNewVals { // searching
		ret.Id = eCmd.id
		if eCmd.childOf != 0 {
			ret.ParentId = eCmd.childOf
			ret.IsChild = true
		}
		if eCmd.creationDate != "" {
			splt := strings.Split(eCmd.creationDate, ":")
			ret.CreationDate, _ = time.Parse(eCmd.conf.DateLayout, splt[0]) //whether range or not, only ever going to need first one
		}
		if eCmd.deadline != "" {
			splt := strings.Split(eCmd.deadline, ":")
			ret.Deadline, _ = time.Parse(eCmd.conf.DateLayout, splt[0])
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
			ret.Deadline, _ = time.Parse(eCmd.conf.DateLayout, eCmd.newDeadline)
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
	case "n":
		return godoo.None, nil
	default:
		return godoo.None, &InvalidArgumentError{}
	}
}

// In remote mode, coordinates request & response to/from remote server.
func (eCmd *EditCommand) remoteEdit(w io.Writer, srchFq, edtFq godoo.FullUserQuery) error {
	// --> very happy path; need to test
	baseUrl := eCmd.conf.RemoteUrl + "/edit"

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

	resp, err := eCmd.conf.Client.Do(rq)
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
func (eCmd *EditCommand) DetermineQueryType(qType godoo.QueryType) ([]godoo.UserQueryOption, error) {
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
			d := getUpperDateBound(eCmd.deadline, eCmd.conf.DateLayout)
			ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByDeadline, UpperBoundDate: d})
		}
		if eCmd.creationDate != "" {
			d := getUpperDateBound(eCmd.creationDate, eCmd.conf.DateLayout)
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
