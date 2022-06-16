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

type GetCommand struct {
	appCtx         *AppContext
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

func NewGetCommand(ctx *AppContext) (*GetCommand, error) {
	getCmd := GetCommand{appCtx: ctx, fs: flag.NewFlagSet("get", flag.ContinueOnError)}

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

	err := getCmd.SetupFlagMapper(ctx.args)

	return &getCmd, err
}

func (getCmd *GetCommand) SetupFlagMapper(userFlags []string) error {
	canonicalFlags, err := getCmd.GetValidFlags()
	if err != nil {
		return err
	}

	getCmd.parser = *fp.NewFlagParser(canonicalFlags, userFlags, fp.WithNowAs(_getNowString(), getCmd.appCtx.DateLayout))

	err = getCmd.parser.CheckInitialisation()
	if err != nil {
		return err
	}

	return nil
}

func (getCmd *GetCommand) GetValidFlags() ([]fp.FlagInfo, error) {
	var ret []fp.FlagInfo

	maxIntDigits := getCmd.appCtx.intDigits

	lenMax := getCmd.appCtx.maxLen

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

// ParseFlags implements method from ICommand interface
func (getCmd *GetCommand) ParseFlags() error {
	newArgs, err := getCmd.parser.ParseUserInput()

	if err != nil {
		return err
	}

	getCmd.appCtx.args = newArgs
	return getCmd.fs.Parse(getCmd.appCtx.args)
}

func (gCmd *GetCommand) GenerateTodoItem() (domain.TodoItem, error) {
	ret := domain.NewTodoItem(domain.WithPriorityLevel(domain.None))

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

func (gCmd *GetCommand) Run(w io.Writer) error {

	input, _ := gCmd.GenerateTodoItem()

	var itms []domain.TodoItem
	var err error

	qList, err := gCmd.determineQueryType()
	if err != nil {
		return err
	}

	itms, err = gCmd.appCtx.todoRepo.GetWhere(qList, input)
	if err != nil {
		return err
	}

	msg := gCmd.getFunc(itms)
	w.Write([]byte(msg()))

	return nil
}

func (gCmd *GetCommand) getFunc(itms []domain.TodoItem) func() string {
	f := func() string {
		var str string
		for _, itm := range itms {
			str += gCmd.buildOutput(itm) + "\n"
		}
		c := len(itms)
		s := ""
		if c == 0 || c > 1 {
			s = "s"
		}
		str += fmt.Sprintf("--> Returned %v item%v\n", c, s)
		return str
	}
	return f
}

func (gCmd *GetCommand) determineQueryType() ([]domain.UserQuery, error) {
	var ret []domain.UserQuery

	// by id numbers
	if gCmd.id != 0 {
		ret = append(ret, domain.UserQuery{Elem: domain.ById})
		return ret, nil // no further search params needed; id unique
	}
	if gCmd.parentOf != 0 {
		ret = append(ret, domain.UserQuery{Elem: domain.ByChildId})
		return ret, nil // no further search params needed; only 1 parent possible
	}
	if gCmd.childOf != 0 {
		ret = append(ret, domain.UserQuery{Elem: domain.ByParentId})
	}

	if gCmd.next {
		if gCmd.deadlineDate == "" {
			ret = append(ret, domain.UserQuery{Elem: domain.ByNextDate})
		} else if gCmd.deadlineDate == "." {
			ret = append(ret, domain.UserQuery{Elem: domain.ByNextPriority})
		}
	}

	// by string
	if gCmd.tagInput != "" {
		ret = append(ret, domain.UserQuery{Elem: domain.ByTag})
	}
	if gCmd.bodyPhrase != "" {
		ret = append(ret, domain.UserQuery{Elem: domain.ByBody})
	}

	// by times
	if gCmd.deadlineDate != "" {
		if gCmd.deadlineDate != "." {
			f := getDateBoundFunc(gCmd.deadlineDate, gCmd.appCtx.DateLayout)
			ret = append(ret, domain.UserQuery{Elem: domain.ByDeadline, DateSetter: f})
		}
	}
	if gCmd.creationDate != "" {
		f := getDateBoundFunc(gCmd.creationDate, gCmd.appCtx.DateLayout)
		ret = append(ret, domain.UserQuery{Elem: domain.ByCreationDate, DateSetter: f})
	}
	if gCmd.complete || gCmd.toggleComplete {
		ret = append(ret, domain.UserQuery{Elem: domain.ByCompletion})
	}

	return ret, nil
}

func getDateBoundFunc(dateText string, dateLayout string) domain.SetUpperDateBound {
	splt := splitDates(dateText)
	var f domain.SetUpperDateBound

	if len(splt) > 1 {
		f = func() (bool, time.Time) {
			d, _ := time.Parse(dateLayout, splt[1])
			return true, d
		}
	} else {
		f = func() (bool, time.Time) {
			return false, time.Now()
		}
	}
	return f
}

func splitDates(s string) []string {
	return strings.Split(s, ":")
}

func (gCmd *GetCommand) buildOutput(itm domain.TodoItem) string {
	var retStr string
	tagOut := getTagOutput(itm.Tags)
	deadline := "n/a"
	if !itm.Deadline.IsZero() {
		deadline = util.StringFromDate(itm.Deadline)
	}
	done := Red + "Not done" + Reset
	if itm.IsComplete {
		done = Green + "Done" + Reset
	}
	retStr += fmt.Sprintf(Yellow+"-- Id:"+Reset+" [%v][%v]\n\t"+Cyan+"- Created:"+Reset+"  %v     "+Cyan+"ParentId:"+Reset+" %v     "+Cyan+"Priority:"+Reset+" %v\n\t"+Cyan+"- Deadline:"+Reset+" %v\n\t"+Cyan+"- Tags:"+Reset+"     %v\n\t"+Cyan+"- Body:"+Reset+"     %v\n", itm.Id, done, util.StringFromDate(itm.CreationDate), itm.ParentId, itm.Priority, deadline, tagOut, itm.Body)
	return retStr
}

func getTagOutput(mp map[string]struct{}) string {
	var ret string
	sep := "; "

	for v := range mp {
		if len(v) > 0 {
			ret += v + sep
		}
	}
	return strings.TrimSuffix(ret, "; ")
}
