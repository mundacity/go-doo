package application

import (
	"container/heap"
	"errors"

	d "github.com/mundacity/go-doo/domain"
)

// PriorityList is an implementation of ITodoCollection
type PriorityList struct {
	DateMode bool
	List     d.PriorityQueue
}

// Constructor for PriorityList()
func NewPriorityList() *PriorityList {
	q := d.NewPriorityQueue()
	pl := PriorityList{DateMode: false, List: *q}
	return &pl
}

// Add item to queue -
// ITodoCollection implementation
func (pl *PriorityList) Add(itm d.TodoItem) error {

	oldLen := pl.List.Len()
	_, exists := pl.List.Items[itm.Id]
	if exists {
		return &d.ItemIdAlreadyExistsError{}
	}

	heap.Push(&pl.List, itm)

	if pl.List.Len() == oldLen {
		return &d.ItemNotAddedToPriorityListError{}
	}

	return nil
}

// Delete item from queue -
// ITodoCollection implementation
func (pl *PriorityList) Delete(id int) error {

	_, exists := pl.List.Items[id]
	if !exists {
		return &d.ItemIdNotFoundError{}
	}

	delete(pl.List.Items, id)
	return nil
}

// Update existing queue item -
//ITodoCollection implementation
func (pl *PriorityList) Update(itm *d.TodoItem) error {

	oldItm, exists := pl.List.Items[itm.Id]
	if !exists {
		return &d.ItemIdNotFoundError{}
	}

	itm.Index = oldItm.Index
	pl.List.Items[itm.Id] = itm
	heap.Fix(&pl.List, itm.Index)
	return nil
}

// Get next item from queue based on priority -
// ITodoCollection implementation
func (pl *PriorityList) GetNext() (*d.TodoItem, error) {

	if pl.List.Len() == 0 {
		return nil, errors.New("no items in list")
	}

	itm := heap.Pop(&pl.List)
	ret := itm.(*d.TodoItem)

	return ret, nil
}

func (pl *PriorityList) GetById(id int) (*d.TodoItem, error) {

	td, exists := pl.List.Items[id]
	if !exists {
		return nil, &d.ItemIdNotFoundError{}
	}

	return td, nil
}
