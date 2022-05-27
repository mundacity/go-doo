package domain

type NegativeParentIdError struct{}

func (e *NegativeParentIdError) Error() string {
	return "supplied ParentId less than zero"
}

type ItemIdNotFoundError struct{}

func (i *ItemIdNotFoundError) Error() string {
	return "supplied id does not exist"
}

type TagAlreadyExistsError struct{}

func (e *TagAlreadyExistsError) Error() string {
	return "supplied tag already present"
}

type ItemIdAlreadyExistsError struct{}

func (e *ItemIdAlreadyExistsError) Error() string {
	return "id already in list"
}

type ItemNotAddedToPriorityListError struct{}

func (e *ItemNotAddedToPriorityListError) Error() string {
	return "item not pushed to heap"
}
