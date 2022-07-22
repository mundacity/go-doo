package app

import (
	"fmt"
	"net/http"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/sqlite"
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

// init sets up the appContext to be used in
// properly executing user commands
func SetupCli(osArgs []string) (*AppContext, error) {

	app := AppContext{Args: osArgs}
	app.setCliContext()

	return &app, nil
}

func startLogger(msg string) {
	enable := viper.GetBool("ENABLE_LOGGING")

	if !enable {
		lg.Logger = lg.NewDummyLogger()
		return
	}

	logPath := viper.GetString("LOG_FILE_PATH")
	lg.Logger = lg.New(logPath, 2)
	lg.Logger.Log(lg.Info, msg)
}

// Set default configuration values and read from env file
func SetConfigVals() {
	viper.SetDefault("MAX_LENGTH", 2000)
	viper.SetDefault("MAX_INT_DIGITS", 4)
	viper.SetDefault("TAG_DELIMITER", "*")
	viper.SetDefault("DATETIME_FORMAT", "2006-01-02")
	viper.SetDefault("INSTANCE_TYPE", 0)
	viper.SetDefault("DB_TYPE", "sqlite")
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("BASE_URL", "http://localhost")
	viper.SetDefault("LOG_FILE_PATH", "godoo-logs")

	viper.SetConfigName("env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("C:\\fe\\")
	viper.ReadInConfig()
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

	app.conn = getConn()
	app.TodoRepo = getRepo(getDbKind(viper.GetString("DB_TYPE")), app.conn, app.DateLayout, 0)

	tolog = append(tolog, app.conn)
	s += app.conn
	lg.Logger.Logf(lg.Info, s, tolog...)
}

func SetSrvContext() godoo.IRepository {

	SetConfigVals()
	cn := getConn()
	dl := viper.GetString("DATETIME_FORMAT")

	startLogger("srv application started")
	lg.Logger.Logf(lg.Info, "Conn: %v\n\tDateLayout: %v\n", cn, dl)

	return getRepo(getDbKind(viper.GetString("DB_TYPE")), cn, dl, viper.GetInt("SERVER_PORT"))
}

func getConn() string {
	testing := viper.GetBool("DEVELOPMENT")
	if testing {
		return viper.GetString("TESTING_CONN")
	} else {
		return viper.GetString("CONNECTION_STRING")
	}
}

func getDbKind(k string) godoo.DbType {
	lg.Logger.Logf(lg.Info, "type of db = %v", k)

	switch k {
	case "sqlite":
		return godoo.Sqlite
	default:
		return godoo.Sqlite
	}
}

func getRepo(dbKind godoo.DbType, connStr, dateLayout string, port int) godoo.IRepository {
	lg.Logger.Logf(lg.Info, "Port: %v", port)
	switch dbKind {
	case godoo.Sqlite:
		return sqlite.SetupRepo(connStr, dbKind, dateLayout, port)
	}

	lg.Logger.Log(lg.Warning, "repo wasn't set up properly")
	return nil
}
