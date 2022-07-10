package godoo

import "time"

// Types relating to repositories
//
// Describes repo functions & the types used to interact with them

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
	Update
	Delete
)

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
