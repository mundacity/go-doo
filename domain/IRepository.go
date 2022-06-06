package domain

// Defines common behaviour of different collection types
type ITodoCollection interface {
	Add(itm TodoItem) error
	Delete(id int) error
	Update(itm *TodoItem) error
	GetById(id int) (*TodoItem, error)
	GetNext() (*TodoItem, error)
}

type QueryType int

const (
	Add QueryType = iota
	Get
	Update
	Delete
)

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

// Defines methods used to interact with data storage
type IRepository interface {
	GetWhere(options []UserQueryElement, input TodoItem) ([]TodoItem, error)
	Add(itm *TodoItem) (int64, error) // num of items stored/affected
	UpdateWhere(srchOptions, edtOptions []UserQueryElement, selector, newVals TodoItem) (int, error)
	// Delete(items ...int) error
}
