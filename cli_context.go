package godoo

import (
	"io"
	"net/http"
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

// passed to different commands to run cli
type ICliContext interface {
	SetupCliContext(args []string)
	SetupFlagParser()
	GetCommand() (ICommand, error)
}

// Sets out the methods implemented by commands that the user can execute
type ICommand interface {
	ParseInput() error
	Run(io.Writer) error
	BuildItemFromInput() (TodoItem, error)
}

type IQueryingCommand interface {
	DetermineQueryType(qType QueryType) ([]UserQueryOption, error)
}

type IFlagParser interface {
	ParseUserInput() ([]string, error)
}

type ConfigVals struct {
	Args       []string
	Client     http.Client
	TodoRepo   IRepository
	Instance   InstanceType
	RemoteUrl  string
	DateLayout string
	NowString  string
	Conn       string
	MaxLen     int
	IntDigits  int
	TagDelim   string
	Parser     IFlagParser
}

// Flags used throughout the system
type CMD_FLAG string

const (
	All             CMD_FLAG = "-a"
	Body            CMD_FLAG = "-b"
	Child           CMD_FLAG = "-c"
	Date            CMD_FLAG = "-d"
	Creation        CMD_FLAG = "-e" // e for existence!
	Finished        CMD_FLAG = "-f" // item complete
	ItmId           CMD_FLAG = "-i"
	Mode            CMD_FLAG = "-m"
	Next            CMD_FLAG = "-n"
	Parent          CMD_FLAG = "-p"
	Tag             CMD_FLAG = "-t"
	ChangeBody      CMD_FLAG = "-B" //append or replace
	ChangeParent    CMD_FLAG = "-C"
	ChangedDeadline CMD_FLAG = "-D"
	MarkComplete    CMD_FLAG = "-F"
	ChangeMode      CMD_FLAG = "-M"
	ChangeTag       CMD_FLAG = "-T" //append, replace, or remove
	AppendMode      CMD_FLAG = "--append"
	ReplaceMode     CMD_FLAG = "--replace"
	// Modifies the behaviour of the -n flag (next) in get command.
	// Instead of next by priority, it's next by date.
	DateMode CMD_FLAG = "--date"
)
