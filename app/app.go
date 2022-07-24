package app

import (
	"fmt"
	"net/http"

	godoo "github.com/mundacity/go-doo"
	lg "github.com/mundacity/quick-logger"
	"github.com/spf13/viper"
)

// Denotes app instance's use of local, remote, or multiple storage options;
// determined via an environment variable
type InstanceType int

const (
	Local InstanceType = iota
	Remote
	Multiple // allows for simultaneous use of multiple storage options - all local, all remote, or a mix.
	// good excuse for experimenting with concurrency - go routines & channels
)

// AppContext encapsulates user inputs and is
// passed around to execute desired operations
type AppContext struct {
	Args       []string
	Client     http.Client
	TodoRepo   godoo.IRepository // todo: make a slice to implement multiple dbs
	Instance   InstanceType
	RemoteUrl  string
	DateLayout string
	conn       string
	MaxLen     int
	IntDigits  int
	TagDemlim  string
}

func (app *AppContext) setCliContext() {

	SetConfigVals()
	app.MaxLen = viper.GetInt("MAX_LENGTH")
	app.IntDigits = viper.GetInt("MAX_INT_DIGITS")
	app.TagDemlim = viper.GetString("TAG_DELIMITER")
	app.Instance = InstanceType(viper.GetInt("INSTANCE_TYPE"))
	app.DateLayout = viper.GetString("DATETIME_FORMAT")

	startLogger("cli application started...")
	tolog := []any{app.MaxLen, app.IntDigits, app.TagDemlim, app.Instance, app.DateLayout}
	s := "[MaxLen: %v, IntDigits: %v, TagDelim: %v, InstanceType: %v, DateLayout: %v]"

	if app.Instance != 0 {
		app.Client = *http.DefaultClient
		app.RemoteUrl = fmt.Sprintf("%v:%v", viper.GetString("BASE_URL"), viper.GetInt("SERVER_PORT"))

		tolog = append(tolog, app.RemoteUrl)
		s = s[:len(s)-1] + ", RemoteUrl: %v]"
		lg.Logger.Logf(lg.Info, s, tolog...)
		return
	}

	// only runs in local mode
	app.conn = getConn()
	app.TodoRepo = getRepo(getDbKind(viper.GetString("DB_TYPE")), app.conn, app.DateLayout, 0)

	tolog = append(tolog, app.conn)
	s += app.conn
	lg.Logger.Logf(lg.Info, s, tolog...)
}
