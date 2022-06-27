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

// Helper function to set upper date bound
// when querying by date ranges - e.g. between
// 2022-05-01 and 2022-05-15
type SetUpperDateBound func() (bool, time.Time)

type UserQuery struct {
	Elem       UserQueryElement
	DateSetter SetUpperDateBound
}

// Defines methods used to interact with data storage
type IRepository interface {
	GetWhere(options []UserQuery, input TodoItem) ([]TodoItem, error)
	Add(itm *TodoItem) (int64, error)
	UpdateWhere(srchOptions, edtOptions []UserQuery, selector, newVals TodoItem) (int, error)
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
