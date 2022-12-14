package godoo

import (
	"errors"
	"time"
)

// Function used in TodoItem initialisation to set PriorityLevel
type PriorityOption func(itm *TodoItem)

type PriorityLevel int

const (
	None PriorityLevel = iota
	Low
	Medium
	High
	DateBased
)

func (p PriorityLevel) String() string {
	switch p {

	case Low:
		return "low"
	case Medium:
		return "medium"
	case High:
		return "high"
	case DateBased:
		return "deadline"
	default:
		return "none"
	}
}

type TodoItem struct {
	Id           int                 `json:"itemId"`
	ParentId     int                 `json:"parentId"`
	IsChild      bool                `json:"isChild"`
	CreationDate time.Time           `json:"creationDate"`
	Deadline     time.Time           `json:"deadlineDate"`
	Priority     PriorityLevel       `json:"priority"`
	Body         string              `json:"itemText"`
	IsComplete   bool                `json:"isComplete"`
	ChildItems   map[int]struct{}    `json:"children"` // map of TodoItem.id with empty struct
	Tags         map[string]struct{} `json:"tags"`
	Index        int                 // for the implementation of a priority list
}

// NewTodoItem constructor initialises maps
func NewTodoItem(p PriorityOption) *TodoItem {
	itm := &TodoItem{ChildItems: make(map[int]struct{}), Tags: make(map[string]struct{})}
	p(itm)
	return itm
}

func WithPriorityLevel(p PriorityLevel) PriorityOption {
	return func(itm *TodoItem) {
		itm.Priority = p
	}
}

func WithDateBasedPriority(date, dateFormat string) PriorityOption {
	return func(itm *TodoItem) {
		itm.Priority = DateBased
		dl, _ := time.Parse(dateFormat, date)
		itm.Deadline = dl
	}
}

func (itm *TodoItem) SetParent(parentId int) error {
	switch {
	case parentId == 0: // a reset
		itm.ParentId = 0
		itm.IsChild = false
		return nil
	case parentId < 0:
		return &NegativeParentIdError{}
	case parentId > 0:
		itm.ParentId = parentId
		itm.IsChild = true
		return nil
	}
	return errors.New("supplied ParentId invalid")
}

func (itm *TodoItem) RemoveParent(parent *TodoItem) error {
	parent.RemoveChildItem(itm.Id)
	return itm.SetParent(0)
}

func (itm *TodoItem) AddChildItem(childId int) {
	itm.ChildItems[childId] = struct{}{}
}

func (itm *TodoItem) RemoveChildItem(childId int) error {
	_, exists := itm.ChildItems[childId]

	if !exists {
		return &ItemIdNotFoundError{}
	}
	delete(itm.ChildItems, childId)
	return nil
}

func (itm *TodoItem) AddTag(t string) error {
	_, exists := itm.Tags[t]
	if exists {
		return &TagAlreadyExistsError{}
	}

	itm.Tags[t] = struct{}{}
	return nil
}

type TagAlreadyExistsError struct{}

func (e *TagAlreadyExistsError) Error() string {
	return "supplied tag already present"
}

type NegativeParentIdError struct{}

func (e *NegativeParentIdError) Error() string {
	return "supplied ParentId less than zero"
}
