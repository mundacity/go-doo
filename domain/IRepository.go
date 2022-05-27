package domain

// Defines common behaviour of different collection types
type ITodoCollection interface {
	Add(itm TodoItem) error
	Delete(id int) error
	Update(itm *TodoItem) error
	GetById(id int) (*TodoItem, error)
	GetNext() (*TodoItem, error)
}

type GetQueryType int

const (
	ById GetQueryType = iota
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
	GetAll() ([]TodoItem, error)
	GetById(id int) (TodoItem, error)
	GetWhere(options []GetQueryType, input TodoItem) ([]TodoItem, error)
	Add(itm *TodoItem) (int64, error) // num of items stored/affected
	UpdateWhere(srchOptions, edtOptions []GetQueryType, selector, newVals TodoItem) (int, error)
	// Delete(items ...int) error
}
