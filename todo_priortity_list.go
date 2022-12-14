package godoo

import (
	"container/heap"
	"errors"
)

// PriorityList is an implementation of ITodoCollection
type PriorityList struct {
	DateMode bool
	List     PriorityQueue
}

// Constructor for PriorityList()
func NewPriorityList() *PriorityList {
	q := NewPriorityQueue()
	pl := PriorityList{DateMode: false, List: *q}
	return &pl
}

// Add item to queue -
// ITodoCollection implementation
func (pl *PriorityList) Add(itm TodoItem) error {

	oldLen := pl.List.Len()
	_, exists := pl.List.Items[itm.Id]
	if exists {
		return &ItemIdAlreadyExistsError{}
	}

	heap.Push(&pl.List, itm)

	if pl.List.Len() == oldLen {
		return &ItemNotAddedToPriorityListError{}
	}

	return nil
}

// Delete item from queue -
// ITodoCollection implementation
func (pl *PriorityList) Delete(id int) error {

	_, exists := pl.List.Items[id]
	if !exists {
		return &ItemIdNotFoundError{}
	}

	delete(pl.List.Items, id)
	return nil
}

// Update existing queue item -
//ITodoCollection implementation
func (pl *PriorityList) Update(itm *TodoItem) error {

	oldItm, exists := pl.List.Items[itm.Id]
	if !exists {
		return &ItemIdNotFoundError{}
	}

	itm.index = oldItm.index
	pl.List.Items[itm.Id] = itm
	heap.Fix(&pl.List, itm.index)
	return nil
}

// Get next item from queue based on priority -
// ITodoCollection implementation
func (pl *PriorityList) GetNext() (*TodoItem, error) {

	if pl.List.Len() == 0 {
		return nil, errors.New("no items in list")
	}

	itm := heap.Pop(&pl.List)
	ret := itm.(*TodoItem)

	return ret, nil
}

func (pl *PriorityList) GetById(id int) (*TodoItem, error) {

	td, exists := pl.List.Items[id]
	if !exists {
		return nil, &ItemIdNotFoundError{}
	}

	return td, nil
}
