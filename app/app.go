package app

import (
	"net/http"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/sqlite"
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

// Set default configuration values and read from env file
func SetConfigVals() {
	viper.SetDefault("MAX_LENGTH", 2000)
	viper.SetDefault("MAX_INT_DIGITS", 4)
	viper.SetDefault("TAG_DELIMITER", "*")
	viper.SetDefault("DATETIME_FORMAT", "2006-01-02")
	viper.SetDefault("INSTANCE_TYPE", 0)
	viper.SetDefault("DB_TYPE", "sqlite")

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
	if app.Instance != 0 {
		app.Client = *http.DefaultClient
	}
	app.conn = getConn()

	app.DateLayout = viper.GetString("DATETIME_FORMAT")
	app.TodoRepo = getRepo(getDbKind(viper.GetString("DB_TYPE")), app.conn, app.DateLayout)
}

func SetSrvContext() {
	SetConfigVals()
	cn := getConn()
	dl := viper.GetString("DATETIME_FORMAT")
	getRepo(getDbKind(viper.GetString("DB_TYPE")), cn, dl)
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
	switch k {
	case "sqlite":
		return godoo.Sqlite
	default:
		return godoo.Sqlite
	}
}

func getRepo(dbKind godoo.DbType, connStr, dateLayout string) godoo.IRepository {
	switch dbKind {
	case godoo.Sqlite:
		return sqlite.SetupRepo(connStr, dbKind, dateLayout)
	}
	return nil
}
