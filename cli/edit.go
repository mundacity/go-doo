package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	fp "github.com/mundacity/flag-parser"
	"github.com/mundacity/go-doo/domain"
)

type EditCommand struct {
	appCtx        *AppContext
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

func NewEditCommand(ctx *AppContext) (*EditCommand, error) {
	eCmd := EditCommand{appCtx: ctx, fs: flag.NewFlagSet("edit", flag.ContinueOnError)}

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

	err := eCmd.SetupFlagMapper(ctx.args)
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

	maxIntDigits := eCmd.appCtx.intDigits
	lenMax := eCmd.appCtx.maxLen

	f1 := fp.FlagInfo{FlagName: string(body), FlagType: fp.Str, MaxLen: lenMax}
	f2 := fp.FlagInfo{FlagName: string(itmId), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f3 := fp.FlagInfo{FlagName: string(date), FlagType: fp.DateTime, MaxLen: 20}
	f4 := fp.FlagInfo{FlagName: string(tag), FlagType: fp.Str, MaxLen: lenMax}
	f5 := fp.FlagInfo{FlagName: string(child), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f6 := fp.FlagInfo{FlagName: string(creation), FlagType: fp.DateTime, MaxLen: 20}
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

	eCmd.appCtx.args = newArgs
	return eCmd.fs.Parse(eCmd.appCtx.args)
}

func (eCmd *EditCommand) GenerateTodoItem() (domain.TodoItem, error) {
	ret := domain.NewTodoItem(domain.WithPriorityLevel(domain.None))

	if !eCmd.getNewVals { // searching
		ret.Id = eCmd.id
		if eCmd.childOf != 0 {
			ret.ParentId = eCmd.childOf
			ret.IsChild = true
		}
		if eCmd.creationDate != "" {
			ret.CreationDate, _ = time.Parse(eCmd.appCtx.DateLayout, eCmd.creationDate)
		}
		if eCmd.deadline != "" {
			ret.Deadline, _ = time.Parse(eCmd.appCtx.DateLayout, eCmd.deadline)
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
			rdr := bufio.NewReader(os.Stdin)
			choice, _, err := rdr.ReadRune()
			if err != nil {
				fmt.Printf("Error occurred: %v, cancelling operation.", err)
			}

			if choice == 'r' {
				eCmd.replacing = true
			} else if choice == 'a' {
				eCmd.appending = true
			} else {
				return errors.New("cancelling operation")
			}
		}
	}
	return nil
}

func (eCmd *EditCommand) Run(w io.Writer) error {

	eCmd.getAdditionalInput()
	srchQryLst, err := eCmd.determineQueryType(domain.Get)
	if err != nil {
		return err
	}
	edtQryLst, err := eCmd.determineQueryType(domain.Update)
	if err != nil {
		return err
	}

	toEdit, _ := eCmd.GenerateTodoItem()
	eCmd.getNewVals = true
	newVals, _ := eCmd.GenerateTodoItem()
	eCmd.getNewVals = false

	num, err := eCmd.appCtx.todoRepo.UpdateWhere(srchQryLst, edtQryLst, toEdit, newVals)
	if err != nil {
		return err
	}

	s := ""
	if num == 0 || num > 1 {
		s = "s"
	}
	msg := fmt.Sprintf("--> Edited %v item%v\n", num, s)
	w.Write([]byte(msg))

	return nil
}

func (eCmd *EditCommand) determineQueryType(qType domain.QueryType) ([]domain.UserQuery, error) {
	var ret []domain.UserQuery

	switch qType {
	case domain.Get:
		// by id numbers
		if eCmd.id != 0 {
			ret = append(ret, domain.UserQuery{Elem: domain.ById})
		}
		if eCmd.childOf != 0 {
			ret = append(ret, domain.UserQuery{Elem: domain.ByParentId})
		}

		// by string
		if eCmd.tagInput != "" {
			ret = append(ret, domain.UserQuery{Elem: domain.ByTag})
		}
		if eCmd.body != "" {
			ret = append(ret, domain.UserQuery{Elem: domain.ByBody})
		}

		// by times
		if eCmd.deadline != "" {
			ret = append(ret, domain.UserQuery{Elem: domain.ByDeadline})
		}
		if eCmd.creationDate != "" {
			ret = append(ret, domain.UserQuery{Elem: domain.ByCreationDate})
		}
		if eCmd.complete {
			ret = append(ret, domain.UserQuery{Elem: domain.ByCompletion})
		}
	case domain.Update:
		if eCmd.newParent != 0 {
			ret = append(ret, domain.UserQuery{Elem: domain.ByParentId})
		}

		// by string
		if eCmd.newTag != "" {
			ret = append(ret, domain.UserQuery{Elem: domain.ByTag})
		}
		if eCmd.newBody != "" {
			ret = append(ret, domain.UserQuery{Elem: domain.ByBody})
		}

		// by times
		if eCmd.newDeadline != "" {
			ret = append(ret, domain.UserQuery{Elem: domain.ByDeadline})
		}
		if eCmd.appending {
			ret = append(ret, domain.UserQuery{Elem: domain.ByAppending})
		}
		if eCmd.replacing {
			ret = append(ret, domain.UserQuery{Elem: domain.ByReplacement})
		}
		if eCmd.newlyComplete {
			ret = append(ret, domain.UserQuery{Elem: domain.ByCompletion})
		}
	}

	return ret, nil
}
