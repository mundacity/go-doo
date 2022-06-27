package godoo

// ItemsList is a container for specific implementations of ITodoCollection
type ItemsList struct {
	List  ITodoCollection
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

type ItemIdNotFoundError struct{}

func (i *ItemIdNotFoundError) Error() string {
	return "supplied id does not exist"
}

type ItemIdAlreadyExistsError struct{}

func (e *ItemIdAlreadyExistsError) Error() string {
	return "id already in list"
}

type ItemNotAddedToPriorityListError struct{}

func (e *ItemNotAddedToPriorityListError) Error() string {
	return "item not pushed to heap"
}
