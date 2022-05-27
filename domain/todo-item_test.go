package domain

import (
	"fmt"
	"testing"
)

type testType struct {
	id       int
	testName string
}

type testOption int

const (
	Standard testOption = 0
	Chld     testOption = 1
)

func TestRemoveParent_IsChildShouldBeFalseParentIdShouldBeZero(t *testing.T) {

	itm := NewTodoItem(WithPriorityLevel(None))
	itm.Id = 7 // dud value to test reset
	parent := NewTodoItem(WithPriorityLevel(High))
	parent.Id = 10
	err := itm.RemoveParent(parent)

	if err != nil {
		t.Errorf("\n\t>>>>FAILED: error is %v, expected nil", err)
	} else {
		t.Logf("\n\t>>>>PASSED: error is %v, expected nil", err)
	}

	if itm.ParentId != 0 {
		t.Errorf("\n\t>>>>FAILED: itm.ParentId == %v, expected 0", itm.ParentId)
	} else {
		t.Logf("\n\t>>>>PASSED: itm.ParentId == %v, expected 0", itm.ParentId)
	}

	if itm.IsChild {
		t.Errorf("\n\t>>>>FAILED: itm.IsChild is %v, expected false", itm.IsChild)
	} else {
		t.Logf("\n\t>>>>PASSED: itm.IsChild is %v, expected false", itm.IsChild)
	}
}

func TestErrorsForInvalidParentIds(t *testing.T) {
	ids := []int{-1, -58, -1700}
	tc := getTestCases(ids, Standard)

	item := new(TodoItem)

	for _, c := range tc {
		t.Run(c.testName, func(t *testing.T) {
			setInvalidParentIdAndCheckErrors(c.id, item, t)
		})
	}
}

func setInvalidParentIdAndCheckErrors(id int, itm *TodoItem, t *testing.T) {
	err := itm.SetParent(id)
	if err == err.(*NegativeParentIdError) {
		t.Logf("\n\t>>>>PASSED: error is '%v', expected NegativeParentIdError", err)
		return
	}
	t.Errorf("\n\t>>>>FAILED: error is '%v', expected NegativeParentIdError", err)
}

func TestIsChildWhenParentIdSet_ValidId(t *testing.T) {

	idList := []int{1, 72, 350, 1234, 200005, 36, 7004, 18, 999, 100000}
	tc := getTestCases(idList, Standard)

	item := NewTodoItem(WithPriorityLevel(None))

	for _, c := range tc {
		t.Run(c.testName, func(t *testing.T) {
			setIdAndCheckIfIsChildBoolTrue(c.id, item, t)
		})
	}
}

func setIdAndCheckIfIsChildBoolTrue(id int, item *TodoItem, t *testing.T) {
	item.SetParent(id)
	if item.IsChild {
		t.Logf("\n\t>>>>PASSED: item.IsChild is %v, expected %v", item.IsChild, true)
		return
	}
	t.Errorf("\n\t>>>>FAILED: item.IsChild is %v, expected %v", item.IsChild, true)
}

func getTestCases(ids []int, o testOption) []testType {
	var ret []testType
	for _, id := range ids {
		testCase := testType{id: id, testName: getName(id, o)}
		ret = append(ret, testCase)
	}
	return ret
}

func getName(id int, o testOption) string {

	switch o {
	default:
		return fmt.Sprintf("set id = %v", id)
	case Chld:
		return fmt.Sprintf("add childItem with id = %v", id)
	}

}

func TestAddChildItem(t *testing.T) {
	idList := []int{11, 27, 35, 151, 30005, 306, 7004, 18}
	tc := getTestCases(idList, Chld)
	itm := NewTodoItem(WithPriorityLevel(None))
	itm.Id = 1

	for _, c := range tc {
		t.Run(c.testName, func(t *testing.T) {
			addChildIdToChildItemsSlice(t, itm, c.id)
		})
	}
}

func addChildIdToChildItemsSlice(t *testing.T, itm *TodoItem, childId int) {
	itmCount := len(itm.ChildItems)
	itm.AddChildItem(childId)
	expected := itmCount + 1
	actual := len(itm.ChildItems)

	if actual == expected {
		t.Logf("\n\t>>>>PASSED: itm.ChildItems count is %v, expected %v", actual, expected)
	} else {
		t.Errorf("\n\t>>>>FAILED: itm.ChildItems count is %v, expected %v", actual, expected)
	}

	_, exists := itm.ChildItems[childId]
	if exists {
		t.Logf("\n\t>>>>PASSED: childItem '%v' is in map", childId)
	} else {
		t.Errorf("\n\t>>>>FAILED: childItem '%v' is not in map", childId)
	}
}

func TestRemoveChildItem(t *testing.T) {
	itm := NewTodoItem(WithPriorityLevel(None))
	itm.Id = 1
	valid := 17
	invalid := 7

	itm.AddChildItem(valid)

	err := itm.RemoveChildItem(invalid)
	if err != err.(*ItemIdNotFoundError) {
		t.Errorf("\n\t>>>>FAILED: deleted non-existant childItem, err is '%v' expected 'ItemNotFoundError'", err)
	} else {
		t.Logf("\n\t>>>>PASSED: couldn't delete non-existant childItem, err is '%v' expected 'ItemNotFoundError'", err)
	}

	err = itm.RemoveChildItem(valid)
	if err != nil {
		t.Errorf("\n\t>>>>FAILED: didn't delete childItem, err is '%v' expected 'nil'", err)
	} else {
		t.Logf("\n\t>>>>PASSED: deleted childItem, err is '%v' expected 'nil'", err)
	}

}
