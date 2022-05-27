package application

import d "github.com/mundacity/go-doo/domain"

// ItemsList is a container for specific implementations of ITodoCollection
type ItemsList struct {
	List  d.ITodoCollection
	LType ListType
}

type ListOption func(lst *ItemsList)
type ListType int

const (
	Simple ListType = iota
	Queue
)

func WithListType(lt ListType) ListOption {
	return func(lst *ItemsList) {
		lst.LType = lt
		if lt == Simple {
			lst.List = new(SimpleList)
		} else if lt == Queue {
			lst.List = NewPriorityList()
		}
	}
}

func NewItemsList(o ListOption) *ItemsList {
	tq := new(ItemsList)
	o(tq)
	return tq
}
