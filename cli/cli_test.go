package cli

import (
	"errors"
	"net/http"
	"time"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/util"
	lg "github.com/mundacity/quick-logger"
)

var fakeArgs []string

type FakeAppContext struct {
	Config  godoo.ConfigVals
	cmdName string
}

type FakeParser struct {
}

func (a *FakeAppContext) SetupCliContext(args []string) {
	a.Config = godoo.ConfigVals{}
	a.cmdName = args[0]
	a.Config.Args = args[1:]

	fakeArgs = args[1:]

	a.Config.MaxLen = 2000
	a.Config.IntDigits = 4
	a.Config.TagDelim = "*"
	a.Config.Instance = 0
	a.Config.DateLayout = "2006-01-02"

	n := time.Date(2022, 03, 14, 0, 0, 0, 0, time.UTC)
	a.Config.NowString = util.StringFromDate(n)

	a.SetupFlagParser()
	lg.Logger = lg.NewDummyLogger()

	a.Config.Client = *http.DefaultClient
	a.Config.RemoteUrl = ""
	a.Config.Conn = ""
}

func (a *FakeAppContext) GetCommand() (godoo.ICommand, error) {
	var cmd godoo.ICommand
	var err error

	switch a.cmdName {
	case "add":
		cmd = NewAddCommand(&a.Config)
	case "get":
		cmd = NewGetCommand(&a.Config)
	case "edit":
		cmd = NewEditCommand(&a.Config)
	default:
		return nil, errors.New("invalid command")
	}
	return cmd, err
}

func (a *FakeAppContext) SetupFlagParser() {
	a.Config.Parser = &FakeParser{}
}

func (fp *FakeParser) ParseUserInput() ([]string, error) {
	// not testing actual parsing so just return args
	return fakeArgs, nil
}
