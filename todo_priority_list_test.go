package godoo

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type test_item_info struct {
	id       int
	name     string
	priority PriorityLevel
	deadline time.Time
}

func TestPriorityListCreation(t *testing.T) {
	pl := NewPriorityList()

	if pl.List.Items == nil {
		t.Errorf("%v%v", getText(false), "priority list not created")
	} else {
		t.Logf("%v%v", getText(true), "priority list created successfully")
	}
}

func TestAddToPriorityList(t *testing.T) {
	pl := NewPriorityList()
	td := NewTodoItem(WithPriorityLevel(None))

	err := pl.Add(*td)
	if err != nil {
		t.Errorf("%v%v", getText(false), err)
	} else {
		t.Logf("%v%v", getText(true), "item successfully added")
	}
}

func TestPush(t *testing.T) {
	pl := NewPriorityList()
	itms := getTodoSliceWitPriorityRating()

	for idx, itm := range itms {
		t.Run(fmt.Sprintf("Test push item %v", idx), func(t *testing.T) {
			test_push(t, itm, pl)
		})
	}
}

func test_push(t *testing.T, itm *TodoItem, pl *PriorityList) {

	currLen := len(pl.List.Items)
	err := pl.Add(*itm)

	if err == nil && len(pl.List.Items) == currLen+1 {
		t.Logf("%v%v", getText(true), "push to list successful")
	} else {
		t.Errorf("%v%v", getText(false), "push to list NOT successful")
	}
}

func TestPop(t *testing.T) {
	pl := NewPriorityList()
	itms := getTodoSliceWitPriorityRating()

	for _, itm := range itms {
		pl.Add(*itm)
	}

	res, err := pl.GetNext()
	if err == nil && res.Id == 22 {
		t.Logf("%v%v", getText(true), "highest priority item popped")
	} else {
		t.Errorf("%v%v", getText(false), fmt.Sprintf("highest priority item not popped; expected id = 22, got id = %v", res.Id))
	}
}

func getTodoSliceWitPriorityRating() []*TodoItem {
	var tds []*TodoItem
	itm1 := NewTodoItem(WithPriorityLevel(None))
	itm1.Id = 11
	itm2 := NewTodoItem(WithPriorityLevel(High))
	itm2.Id = 22
	itm3 := NewTodoItem(WithPriorityLevel(Low))
	itm3.Id = 33
	itm4 := NewTodoItem(WithPriorityLevel(Low))
	itm4.Id = 44
	tds = append(tds, itm1, itm2, itm3, itm4)

	return tds
}

func TestPoppingMultiple(t *testing.T) {
	itms := getTestPoppingItems()
	pl := NewPriorityList()

	for _, itm := range itms {
		td := NewTodoItem(WithPriorityLevel(itm.priority))
		td.Id = itm.id
		td.Body = itm.name
		pl.Add(*td)
	}

	expct_poss_1 := "10,11,9,7,8"
	expct_poss_2 := "10,9,11,7,8"

	actual := ""

	for pl.List.Len() > 0 {
		itm, _ := pl.GetNext()
		actual += fmt.Sprintf("%v,", fmt.Sprint(itm.Id))
	}
	actual = strings.Trim(actual, ",")

	if actual == expct_poss_1 || actual == expct_poss_2 {
		t.Logf("%v%v", getText(true), fmt.Sprintf("expected '%v', got '%v' or '%v'", actual, expct_poss_1, expct_poss_2))
	} else {
		t.Errorf("%v%v", getText(false), fmt.Sprintf("expected '%v', got '%v' or '%v'", actual, expct_poss_1, expct_poss_2))
	}
}

func getTestPoppingItems() []test_item_info {

	dl := "2006-01-02"
	dateStr := "2022-03-14"

	t, _ := time.Parse(dl, dateStr)
	itms := []test_item_info{{
		id:       7,
		name:     "id = 7, priority low",
		priority: Low,
		deadline: t.Add(time.Hour * 24),
	}, {
		id:       8,
		name:     "id = 8, priority none",
		priority: None,
		deadline: t,
	}, {
		id:       10,
		name:     "id = 10, priority high",
		priority: High,
		deadline: t,
	}, {
		id:       9,
		name:     "id = 9, priority medium",
		priority: Medium,
		deadline: t.Add(time.Hour * 24),
	}, {
		id:       11,
		name:     "id = 11, priority medium",
		priority: Medium,
		deadline: t.Add(time.Hour * 48),
	},
	}
	return itms
}

func TestDeleteNonExistantItem(t *testing.T) {
	pl := NewPriorityList()
	td := NewTodoItem(WithPriorityLevel(None))
	td.Id = 1
	pl.Add(*td)

	err := pl.Delete(2)
	if err != nil {
		t.Logf("%v%v", getText(true), fmt.Sprintf("expected 'supplied id not found' error, got '%v'", err))
	} else {
		t.Errorf("%v%v", getText(false), fmt.Sprintf("expected 'supplied id not found' error, got '%v'", nil))
	}
}

func TestValidDelete(t *testing.T) {
	pl := NewPriorityList()
	td := NewTodoItem(WithPriorityLevel(None))
	td.Id = 1
	pl.Add(*td)

	err := pl.Delete(1)
	if err != nil {
		t.Errorf("%v%v", getText(false), fmt.Sprintf("expected no error, got '%v'", err))
	} else {
		t.Logf("%v%v", getText(true), fmt.Sprintf("expected no error, got '%v'", err))
	}
}

func TestValidHighPriorityAdditionAndSubsequentPop(t *testing.T) {
	itms := getTestPoppingItems()
	pl := NewPriorityList()

	for _, itm := range itms {
		td := NewTodoItem(WithPriorityLevel(itm.priority))
		td.Id = itm.id
		td.Body = itm.name
		pl.Add(*td)
	}

	_, _ = pl.GetNext() // get rid of highest priority; now got low, none, medium, medium priority in list

	td2 := NewTodoItem(WithPriorityLevel(High))
	td2.Id = 777
	td3 := NewTodoItem(WithPriorityLevel(Low))
	td3.Id = 888
	pl.Add(*td2)
	pl.Add(*td3)

	popped, _ := pl.GetNext()
	if popped.Id == 777 {
		t.Logf("%v%v", getText(true), fmt.Sprintf("expected id == 777, got %v", popped.Id))
	} else {
		t.Errorf("%v%v", getText(false), fmt.Sprintf("expected id == 777, got %v", popped.Id))
	}
}

func TestValidUpdate(t *testing.T) {
	itms := getTestPoppingItems()
	pl := NewPriorityList()

	for _, itm := range itms {
		td := NewTodoItem(WithPriorityLevel(itm.priority))
		td.Id = itm.id
		td.Body = itm.name
		pl.Add(*td)
	}

	_, _ = pl.GetNext() // get rid of highest priority; now got low, none, medium, medium priority in list

	td := NewTodoItem(WithPriorityLevel(None))
	td.Id = 1000
	pl.Add(*td)

	newPriority := High
	td.Priority = newPriority
	pl.Update(td)

	itm, _ := pl.GetNext()

	if itm.Priority != newPriority {
		t.Errorf("%v%v", getText(false), fmt.Sprintf("expected priority = '%v', got '%v'", newPriority, itm.Priority))
	} else {
		t.Logf("%v%v", getText(true), fmt.Sprintf("expected priority = '%v', got '%v'", newPriority, itm.Priority))
	}
}

func TestInvalidUpdate(t *testing.T) {
	itms := getTestPoppingItems()
	pl := NewPriorityList()

	for _, itm := range itms {
		td := NewTodoItem(WithPriorityLevel(itm.priority))
		td.Id = itm.id
		td.Body = itm.name
		pl.Add(*td)
	}

	td, _ := pl.GetById(itms[0].id)
	tdVal := *td
	tdVal.Body = "holy moly"
	tdVal.Id = 89883
	err := pl.Update(&tdVal)

	if err != nil {
		t.Logf("%v%v", getText(true), fmt.Sprintf("expected 'supplied id does not exist', got '%v'", err))
	} else {
		t.Errorf("%v%v", getText(false), fmt.Sprintf("expected 'supplied id does not exist', got '%v'", err))
	}
}

func TestDateModePriority(t *testing.T) {

	// outcome: highest priority first; if 2 or more of
	// equal priority, then assess by date & return the
	// item with the closest deadline

	itms := getTestPoppingItems()
	pl := NewPriorityList()

	for _, itm := range itms {
		td := NewTodoItem(WithPriorityLevel(itm.priority))
		td.Id = itm.id
		td.Body = itm.name
		td.Deadline = itm.deadline
		pl.Add(*td)
	}

	expIdOrder := []int{10, 9, 11, 7, 8}
	gotOrder := []int{}

	for range expIdOrder {
		td, _ := pl.GetNext()
		gotOrder = append(gotOrder, td.Id)
	}

	for i := range expIdOrder {
		if expIdOrder[i] != gotOrder[i] {
			t.Errorf("%v%v", getText(false), fmt.Sprintf("slice order not the same: exp = %v, got = %v", expIdOrder[i], gotOrder[i]))
			return
		}
	}

	t.Log("slice order as expected")

}
