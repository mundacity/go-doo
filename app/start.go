package app

import (
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/sqlite"
	lg "github.com/mundacity/quick-logger"
	"github.com/spf13/viper"
)

// Sets up and returns the operating context of the cli client application
// based on user configuration options. Also starts the logger.
func SetupCli(osArgs []string) (*AppContext, error) {

	app := AppContext{Args: osArgs}
	app.setCliContext()

	return &app, nil
}

// Sets up server context & logger in a similar way to SetupCli()
func SetSrvContext() (godoo.IRepository, bool) {

	SetConfigVals()
	cn := getConn()
	dl := viper.GetString("DATETIME_FORMAT")
	pl := viper.GetBool("MAINTAIN_PRIORITY_LIST")

	startLogger("srv application started")
	lg.Logger.Logf(lg.Info, "Conn: %v\n\tDateLayout: %v\n\tPriorityList: %v\n", cn, dl, pl)

	return getRepo(getDbKind(viper.GetString("DB_TYPE")), cn, dl, viper.GetInt("SERVER_PORT")), pl
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
	viper.SetDefault("MAINTAIN_PRIORITY_LIST", true)

	viper.SetConfigName("env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("")
	viper.ReadInConfig()
}

// Returns db path based on user configuration options
func getConn() string {
	testing := viper.GetBool("DEVELOPMENT")
	if testing {
		return viper.GetString("TESTING_CONN")
	} else {
		return viper.GetString("CONNECTION_STRING")
	}
}

// Starts a quick logger that can be used throughout the system
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

// Returns a specific db type. Only SQLite
// currently supported. Could add others but
// this is probably a case of YAGNI...
func getDbKind(k string) godoo.DbType {
	lg.Logger.Logf(lg.Info, "type of db = %v", k)

	switch k {
	case "sqlite":
		return godoo.Sqlite
	default:
		return godoo.Sqlite
	}
}

// Returns instantiated repo interface that
// is used when communicating with the database
func getRepo(dbKind godoo.DbType, connStr, dateLayout string, port int) godoo.IRepository {
	lg.Logger.Logf(lg.Info, "Port: %v", port)
	switch dbKind {
	case godoo.Sqlite:
		return sqlite.SetupRepo(connStr, dbKind, dateLayout, port)
	}

	lg.Logger.Log(lg.Warning, "repo wasn't set up properly")
	return nil
}
