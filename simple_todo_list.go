package godoo

// SimpleList is an implementation of ITodoCollection
type SimpleList struct {
	List map[int]*TodoItem
}

func NewSimpleList() *SimpleList {
	sl := SimpleList{List: make(map[int]*TodoItem)}
	return &sl
}

func (sl *SimpleList) Add(itm TodoItem) error {
	_, exists := sl.List[itm.Id]
	if exists {
		return &ItemIdAlreadyExistsError{}
	}
	sl.List[itm.Id] = &itm
	return nil
}

func (sl *SimpleList) Delete(id int) error {

	_, exists := sl.List[id]
	if !exists {
		return &ItemIdNotFoundError{}
	}
	delete(sl.List, id)
	return nil
}

func (sl *SimpleList) Update(itm *TodoItem) error {

	_, exists := sl.List[itm.Id]
	if !exists {
		return &ItemIdNotFoundError{}
	}

	sl.List[itm.Id] = itm
	return nil
}

func (sl *SimpleList) GetNext() (*TodoItem, error) {
	t := NewTodoItem(WithPriorityLevel(None))
	return t, nil
}

func (sl *SimpleList) GetById(id int) (*TodoItem, error) {

	td, exists := sl.List[id]
	if !exists {
		return nil, &ItemIdNotFoundError{}
	}

	return td, nil
}
