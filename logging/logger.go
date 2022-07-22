package logging

import (
	"fmt"
	"log"
	"os"
)

type LogLevel string

const (
	Info    LogLevel = "INFO: "
	Warning LogLevel = "WARNING: "
	Error   LogLevel = "ERROR: "
)

// Describes logger functions
type ILogger interface {
	// Basic: log level and plain string msg to log
	Log(lv LogLevel, msg string)
	// Pass formatted strings with corresponding args
	Logf(lv LogLevel, formatted string, args ...any)
	// Pass in a slice to quickly format a lot of data
	QuickFmtLog(lv LogLevel, initialText, delim string, args ...any)
	// The func arg accepts runtime.Caller to print the calling location. Note missing parentheses.
	// This assumes that you don't always need/want the full file path - e.g. only if there's an error.
	// In the standard implementation, the location is prepended to the initialText
	// arg and Log(lv, initialText) is then called
	LogWithCallerInfo(lv LogLevel, initialText string, f func(int) (pc uintptr, file string, line int, ok bool))
}

// Basic implementation of the ILogger interface
type AppLogger struct {
	Logger *log.Logger
	depth  int
}

// Logger that can be used throughout the system if you want
var Logger ILogger

// Sets up a new logger. If path arg is empty, it will throw
// a fatal error. The depth arg denotes how many directories to include
// when logging calling locations. For example, 2 would print /app/app.go
func New(path string, depth int) *AppLogger {

	fl, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("could not initialise logger; attempted path '%v'", path)
	}

	return &AppLogger{Logger: log.New(fl, "", log.Ldate|log.Ltime), depth: depth}
}

func (lg *AppLogger) Logf(lv LogLevel, formatted string, args ...any) {
	lg.Logger.SetPrefix(string(lv))
	lg.Logger.Printf(formatted, args...)
}

func (lg *AppLogger) Log(lv LogLevel, msg string) {
	lg.Logger.SetPrefix(string(lv))
	lg.Logger.Print(msg)
}

func (lg *AppLogger) QuickFmtLog(lv LogLevel, initialText, delim string, args ...any) {
	sep := ""
	str := initialText

	for _, v := range args {
		str += fmt.Sprintf("%v%v", sep, v)
		if len(sep) == 0 {
			sep = delim
		}
	}

	lg.Logger.SetPrefix(string(lv))
	lg.Logger.Print(str)
}

func (lg *AppLogger) LogWithCallerInfo(lv LogLevel, msg string, f func(int) (uintptr, string, int, bool)) {
	_, s, i, _ := f(2)
	var out string
	c := 0

	for i := len(s) - 1; i > 0; i-- {
		if s[i] == '/' {
			c++
			if c == lg.depth {
				out = s[i:]
				break
			}
		}
	}

	newMsg := fmt.Sprintf("%v:%v: %v", out, i, msg)
	lg.Log(lv, newMsg)
}
