package cli

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/sqlite"
	"github.com/spf13/viper"
)

// Denotes app instance's use of local, remote, or multiple storage options;
// determined via an environment variable
type InstanceType int

const (
	local InstanceType = iota
	remote
	multiple // allows for simultaneous use of multiple storage options - all local, all remote, or a mix.
	// good excuse for experimenting with concurrency - go routines & channels
)

// Flags used throughout the system
type CMD_FLAG string

const (
	all             CMD_FLAG = "-a"
	body            CMD_FLAG = "-b"
	child           CMD_FLAG = "-c"
	date            CMD_FLAG = "-d"
	creation        CMD_FLAG = "-e" // e for existence!
	finished        CMD_FLAG = "-f" // item complete
	itmId           CMD_FLAG = "-i"
	mode            CMD_FLAG = "-m"
	next            CMD_FLAG = "-n"
	parent          CMD_FLAG = "-p"
	tag             CMD_FLAG = "-t"
	changeBody      CMD_FLAG = "-B" // append or replace
	changeParent    CMD_FLAG = "-C"
	changedDeadline CMD_FLAG = "-D"
	markComplete    CMD_FLAG = "-F"
	changeMode      CMD_FLAG = "-M"
	changeTag       CMD_FLAG = "-T" // append, replace, or remove
	appendMode      CMD_FLAG = "--append"
	replaceMode     CMD_FLAG = "--replace"
)

// AppContext encapsulates user inputs and is
// passed around to execute desired operations
type AppContext struct {
	args       []string
	client     http.Client
	todoRepo   godoo.IRepository // todo: make a slice to implement multiple dbs
	instance   InstanceType
	DateLayout string
	conn       string
	maxLen     int
	intDigits  int
	tagDemlim  string
}

// Init sets up the appContext to be used in
// properly executing user commands
func Init(osArgs []string) (*AppContext, error) {

	app := AppContext{args: osArgs}
	app.config()
	return &app, nil
}

func RunApp(osArgs []string, w io.Writer) int {

	app, err := Init(osArgs)
	if err != nil {
		fmt.Printf("%v", err)
		return 2
	}

	cmd, err := _getBasicCommand(app)
	if err != nil {
		fmt.Printf("%v", err)
		return 2
	}

	err = cmd.ParseFlags()
	if err != nil {
		fmt.Printf("error: '%v'", err)
		return 2
	}

	err = cmd.Run(w)
	if err != nil {
		fmt.Printf("error: '%v'", err)
		return 2
	}

	return 0
}

func (app *AppContext) config() {

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

	app.maxLen = viper.GetInt("MAX_LENGTH")
	app.intDigits = viper.GetInt("MAX_INT_DIGITS")
	app.tagDemlim = viper.GetString("TAG_DELIMITER")
	app.instance = InstanceType(viper.GetInt("INSTANCE_TYPE"))

	testing := viper.GetBool("DEVELOPMENT")
	if testing {
		app.conn = viper.GetString("TESTING_CONN")
	} else {
		app.conn = viper.GetString("CONNECTION_STRING")
	}

	app.DateLayout = viper.GetString("DATETIME_FORMAT")
	app.todoRepo = GetRepo(getDbKind(viper.GetString("DB_TYPE")), app.conn, app.DateLayout)
}

func _getBasicCommand(ctx *AppContext) (ICommand, error) {
	arg1 := ctx.args[0]
	ctx.args = ctx.args[1:]
	var cmd ICommand
	var err error

	switch arg1 {
	case "add":
		cmd, err = NewAddCommand(ctx)
	case "get":
		cmd, err = NewGetCommand(ctx)
	case "edit":
		cmd, err = NewEditCommand(ctx)
	default:
		return nil, errors.New("invalid command")
	}
	return cmd, err
}

func getDbKind(k string) godoo.DbType {
	switch k {
	case "sqlite":
		return godoo.Sqlite
	default:
		return godoo.Sqlite
	}
}

func GetRepo(dbKind godoo.DbType, connStr, dateLayout string) godoo.IRepository {
	switch dbKind {
	case godoo.Sqlite:
		return sqlite.NewRepo(connStr, dbKind, dateLayout)
	}
	return nil
}
