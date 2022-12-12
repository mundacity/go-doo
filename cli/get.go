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

type GetCommand struct {
	conf           *godoo.ConfigVals
	fs             *flag.FlagSet
	id             int
	next           bool   // default to priority, but can be changed by nextByDate flag
	tagInput       string // tags with delimeter set by environment variable
	bodyPhrase     string // key phrase within body
	childOf        int    // child of the int argument
	parentOf       int    // parent of the int argument
	deadlineDate   string
	creationDate   string
	getAll         bool
	complete       bool
	toggleComplete bool
	nextByDate     bool
}

// Returns new get command after setting up flag info and flag-parser
func NewGetCommand(c *godoo.ConfigVals) *GetCommand {
	gCmd := GetCommand{}
	gCmd.conf = c
	lg.Logger.Log(lg.Info, "get command created")

	gCmd.setupFlagSet()

	return &gCmd
}

// Describes the flags and argument types associated with the command
func (getCmd *GetCommand) setupFlagSet() {
	getCmd.fs = flag.NewFlagSet("get", flag.ContinueOnError)
	getCmd.fs.IntVar(&getCmd.id, strings.Trim(string(godoo.ItmId), "-"), 0, "search by item id")
	getCmd.fs.BoolVar(&getCmd.next, strings.Trim(string(godoo.Next), "-"), false, "get next item in priority list")
	getCmd.fs.BoolVar(&getCmd.nextByDate, strings.Trim(string(godoo.DateMode), "-"), false, "get next item by date priority")
	getCmd.fs.StringVar(&getCmd.deadlineDate, strings.Trim(string(godoo.Date), "-"), "", "date of existing item; if empty, modifies -n to return based on date instead of defaulting to priority")
	getCmd.fs.StringVar(&getCmd.creationDate, strings.Trim(string(godoo.Creation), "-"), "", "creation date of existing item")
	getCmd.fs.StringVar(&getCmd.tagInput, strings.Trim(string(godoo.Tag), "-"), "", "search by item tag")
	getCmd.fs.StringVar(&getCmd.bodyPhrase, strings.Trim(string(godoo.Body), "-"), "", "search by known phrase within body")

	getCmd.fs.IntVar(&getCmd.childOf, strings.Trim(string(godoo.Child), "-"), 0, "search based on parent Id; requested item is child of provided parent id")
	getCmd.fs.IntVar(&getCmd.parentOf, strings.Trim(string(godoo.Parent), "-"), 0, "search based on child Id; requested item is parent of provided child id")
	getCmd.fs.BoolVar(&getCmd.getAll, strings.Trim(string(godoo.All), "-"), false, "get all items")
	getCmd.fs.BoolVar(&getCmd.complete, strings.Trim(string(godoo.Finished), "-"), false, "search for completed items")
	getCmd.fs.BoolVar(&getCmd.toggleComplete, strings.Trim(string(godoo.MarkComplete), "-"), false, "search for unfinished items")

}

// ParseInput implements method from ICommand interface
func (g *GetCommand) ParseInput() error {
	newArgs, err := g.conf.Parser.ParseUserInput()

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("user input parsing error: %v", err), runtime.Caller)
		return err
	}

	g.conf.Args = newArgs
	lg.Logger.Log(lg.Info, "successfully parsed user input")
	return g.fs.Parse(g.conf.Args)
}

// Implements Run() method from ICommand interface
func (gCmd *GetCommand) Run(w io.Writer) error {

	input, _ := gCmd.BuildItemFromInput()

	var itms []godoo.TodoItem
	var err error

	qList, err := gCmd.DetermineQueryType(godoo.Get)
	if err != nil {
		return err
	}

	fullQry := godoo.FullUserQuery{QueryOptions: qList, QueryData: input}

	if gCmd.conf.Instance == godoo.Remote {
		return gCmd.remoteGet(w, fullQry)
	}

	itms, err = gCmd.conf.TodoRepo.GetWhere(fullQry)
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
func (gCmd *GetCommand) BuildItemFromInput() (godoo.TodoItem, error) {
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
		ret.CreationDate, _ = time.Parse(gCmd.conf.DateLayout, splt[0]) //only ever need first one
	}
	if len(gCmd.deadlineDate) > 0 {
		splt := strings.Split(gCmd.deadlineDate, ":")
		ret.Deadline, _ = time.Parse(gCmd.conf.DateLayout, splt[0])
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
func (gCmd *GetCommand) DetermineQueryType(qType godoo.QueryType) ([]godoo.UserQueryOption, error) {
	var ret []godoo.UserQueryOption

	if gCmd.next { //todo: add flag for getting by date priority
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByNextPriority})
		return ret, nil // no further params needed/allowed
	}

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

	// by string
	if gCmd.tagInput != "" {
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByTag})
	}
	if gCmd.bodyPhrase != "" {
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByBody})
	}

	// by times
	if len(gCmd.deadlineDate) > 0 {
		d := getUpperDateBound(gCmd.deadlineDate, gCmd.conf.DateLayout)
		ret = append(ret, godoo.UserQueryOption{Elem: godoo.ByDeadline, UpperBoundDate: d})

	}
	if gCmd.creationDate != "" {
		d := getUpperDateBound(gCmd.creationDate, gCmd.conf.DateLayout)
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
	baseUrl := gCmd.conf.RemoteUrl + "/get"

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

	//key, _ := auth.GetPublicKey(gCmd.conf.SrvPublicKeyPath)

	rq.Header.Set("Token", gCmd.conf.JwtString)

	// getting response
	resp, err := sendRequest(rq, &gCmd.conf.Client)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("error receiving response: %v", err), runtime.Caller)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		url := gCmd.conf.RemoteUrl + "/authenticate"
		req, _ := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))

		jwt, err := authenticateUser(gCmd.conf.SrvPublicKeyPath, &gCmd.conf.Client, req)
		if err != nil {
			return err
		}
		gCmd.conf.JwtString = jwt
		return &ReAuthenticationRequired{}
	}

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
