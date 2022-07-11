package fake

import (
	godoo "github.com/mundacity/go-doo"
)

type Repo struct{}

func (m Repo) GetWhere(query godoo.FullUserQuery) ([]godoo.TodoItem, error) {

	//dl := "2006-01-02"
	//d1, _ := time.Parse(dl, "2021-08-03")
	//d2, _ := time.Parse(dl, "2022-10-29")

	return []godoo.TodoItem{
		{Id: 1}, //, ParentId: 2, IsChild: true, CreationDate: d1, Deadline: d2, Priority: godoo.None, Body: "mock body", IsComplete: false},
	}, nil

}

func (m Repo) Add(itm *godoo.TodoItem) (int64, error) {
	return 1, nil
}

func (m Repo) UpdateWhere(srchQry, edtQry godoo.FullUserQuery) (int, error) {
	return 3, nil
}
