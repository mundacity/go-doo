package application

import (
	d "github.com/mundacity/go-doo/domain"
)

// SimpleList is an implementation of ITodoCollection
type SimpleList struct {
	List map[int]*d.TodoItem
}

func NewSimpleList() *SimpleList {
	sl := SimpleList{List: make(map[int]*d.TodoItem)}
	return &sl
}

func (sl *SimpleList) Add(itm d.TodoItem) error {
	_, exists := sl.List[itm.Id]
	if exists {
		return &d.ItemIdAlreadyExistsError{}
	}
	sl.List[itm.Id] = &itm
	return nil
}

func (sl *SimpleList) Delete(id int) error {

	_, exists := sl.List[id]
	if !exists {
		return &d.ItemIdNotFoundError{}
	}
	delete(sl.List, id)
	return nil
}

func (sl *SimpleList) Update(itm *d.TodoItem) error {

	_, exists := sl.List[itm.Id]
	if !exists {
		return &d.ItemIdNotFoundError{}
	}

	sl.List[itm.Id] = itm
	return nil
}

func (sl *SimpleList) GetNext() (*d.TodoItem, error) {
	t := d.NewTodoItem(d.WithPriorityLevel(d.None))
	return t, nil
}

func (sl *SimpleList) GetById(id int) (*d.TodoItem, error) {

	td, exists := sl.List[id]
	if !exists {
		return nil, &d.ItemIdNotFoundError{}
	}

	return td, nil
}
