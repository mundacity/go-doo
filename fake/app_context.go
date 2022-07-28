package fake

import (
	"errors"
	"net/http"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/cli"
	lg "github.com/mundacity/quick-logger"
)

type App_Context struct {
	Config  godoo.ConfigVals
	cmdName string
}

func (a *App_Context) SetupCliContext(args []string) {
	a.Config = godoo.ConfigVals{}
	a.cmdName = args[0]
	a.Config.Args = args[1:]
	a.Config.MaxLen = 2000
	a.Config.IntDigits = 4
	a.Config.TagDelim = "*"
	a.Config.Instance = 0
	a.Config.DateLayout = "2006-01-02"

	lg.Logger = lg.NewDummyLogger()

	a.Config.Client = *http.DefaultClient
	a.Config.RemoteUrl = ""
	a.Config.Conn = ""
	a.Config.TodoRepo = RepoDud{}

}

func (a *App_Context) GetCommand() (godoo.ICommand, error) {
	var cmd godoo.ICommand
	var err error

	switch a.cmdName {
	case "add":
		cmd = cli.NewAddCommand(&a.Config)
	case "get":
		cmd = cli.NewGetCommand(&a.Config)
	case "edit":
		cmd = cli.NewEditCommand(&a.Config)
	default:
		return nil, errors.New("invalid command")
	}
	return cmd, err
}

func (a *App_Context) SetupFlagParser() {

}
