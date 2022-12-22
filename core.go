package godoo

import (
	"io"
	"net/http"
	"time"
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

type IServerContext interface {
	SetupServerContext(conf ServerConfigVals)
	Serve()
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
	Args             []string
	Client           http.Client
	TodoRepo         IRepository
	Instance         InstanceType
	RemoteUrl        string
	DateLayout       string
	NowString        string
	Conn             string
	MaxLen           int
	IntDigits        int
	TagDelim         string
	Parser           IFlagParser
	SrvPublicKeyPath string
	JwtString        string
}

type ServerConfigVals struct {
	Repo             IRepository
	Conn             string
	DateFormat       string
	PriorityList     *PriorityList
	RunPriorityList  bool
	Port             int
	KeyPath          string
	UserPasswordHash string
	ExpirationLimit  int
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

// Differnt kinds of supported RDBMS
type DbType string

const (
	Sqlite DbType = "sqlite3"
)

// Enum for basic CRUD operations.
// Used to switch on to get relevant SQL.
type QueryType int

const (
	Add QueryType = iota
	Get
	Edit
	Delete
)

func (q QueryType) String() string {
	switch q {
	case Add:
		return "add"
	case Get:
		return "get"
	case Edit:
		return "edit"
	case Delete:
		return "delete"
	default:
		return ""
	}
}

// Enums describing the different item
// attributes that the user can query & edit
type UserQueryElement int

const (
	ById UserQueryElement = iota
	ByChildId
	ByParentId
	ByTag
	ByBody
	ByNextPriority
	ByNextDate
	ByDeadline
	ByCreationDate
	ByReplacement
	ByAppending
	ByCompletion
)

// Wrapper for a single UserQueryElement and
// a SetUpperDateBound function to allow for
// date range searching
type UserQueryOption struct {
	Elem           UserQueryElement `json:"elem"`
	UpperBoundDate time.Time        `json:"upperBound"`
}

// Single query object to combine query options
// alongside the relevant query data.
// Ex. ById is the query option and '8' is the
// query data
type FullUserQuery struct {
	QueryOptions []UserQueryOption `json:"qryOpts"`
	QueryData    TodoItem          `json:"qryData"`
}

// Defines methods used to interact with data storage
type IRepository interface {
	GetAll() ([]TodoItem, error)
	GetWhere(query FullUserQuery) ([]TodoItem, error)
	Add(itm *TodoItem) (int64, error)
	UpdateWhere(srchQry, edtQry FullUserQuery) (int, error)
	// Delete(items ...int) error
}

// Defines common behaviour of different collection types
type ITodoCollection interface {
	Add(itm TodoItem) error
	Delete(id int) error
	Update(itm *TodoItem) error
	GetById(id int) (*TodoItem, error)
	GetNext() (*TodoItem, error)
}
