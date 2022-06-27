package main_test

import (
	"fmt"
	"testing"

	godoo "github.com/mundacity/go-doo"
)

type test_case struct {
	Id   int
	Item godoo.TodoItem
	Name string
}

func TestValidListCreateion(t *testing.T) {
	lst := godoo.NewSimpleList()

	if lst.List != nil {
		t.Log("\n\t>>>>PASSED: SimpleList created")
		return
	}
	t.Error("\n\t>>>>FAILED: SimpleList NOT created")
}

func TestAddToList(t *testing.T) {
	sl := godoo.NewSimpleList()
	itm := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
	itm.Id = 1

	err := sl.Add(*itm)

	if len(sl.List) == 1 && err == nil {
		t.Logf("%v%v", getText(true), "item added successfully")
		return
	}
	t.Errorf("%v%v", getText(false), fmt.Sprintf("expected len = 1, but got len = %v", len(sl.List)))
}

func getText(result bool) string {
	t := "\n\t>>>>"

	if result {
		return t + "PASSED: "
	}
	return t + "FAILED: "
}

func getListTestCases() []test_case {
	lst := []int{1, 3, 17, 889, 10471}
	return getSlice(lst)
}

func getSlice(lst []int) []test_case {
	var tcLst []test_case
	for i := range lst {
		td := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
		tc := test_case{Id: i, Item: *td, Name: fmt.Sprintf("body from nil to: %v", i)}
		tcLst = append(tcLst, tc)
	}
	return tcLst
}

func TestEditingItmes(t *testing.T) {
	tcs := getListTestCases()

	for _, tc := range tcs {
		s := fmt.Sprintf("new body = %v", tc.Id)
		t.Run(tc.Name, func(t *testing.T) {
			testEdit(t, &tc.Item, s)
		})
	}
}

func testEdit(t *testing.T, itm *godoo.TodoItem, newBody string) {
	old := itm.Body
	itm.Body = newBody

	if itm.Body == old {
		t.Errorf("%v%v", getText(false), fmt.Sprintf("expecting '%v', got '%v'", newBody, itm.Body))
		return
	}
	t.Logf("%v%v", getText(true), fmt.Sprintf("expecting '%v', got '%v'", newBody, itm.Body))
}

func TestAddExistingId(t *testing.T) {
	tcs := getListTestCases()
	existingId := tcs[0].Id

	pl := godoo.NewPriorityList()
	for _, itm := range tcs {
		td := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
		td.Id = itm.Id
		pl.Add(*td)
	}

	td2 := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
	td2.Id = existingId

	err := pl.Add(*td2)

	if err != nil {
		t.Logf("%v%v", getText(true), "item already exists; not added")
	} else {
		t.Errorf("%v%v", getText(false), "item with duplicate id successfully added")
	}
}
