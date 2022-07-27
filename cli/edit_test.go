package cli

import (
	"fmt"
	"sort"
	"testing"
	"time"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/app"
	"github.com/mundacity/go-doo/util"
)

type edit_item_generation_test_case struct {
	args     []string
	expected EditCommand
	err      error
	name     string
}

type edit_query_build_test_case struct {
	input      EditCommand
	name       string
	expSrchLst []godoo.UserQueryElement
	expEdtLst  []godoo.UserQueryElement
	expSrchItm godoo.TodoItem
	expEdtItm  godoo.TodoItem
	args       []string //only needed when testing dates
}

func _getTestCasesForEditing() []edit_item_generation_test_case {
	return []edit_item_generation_test_case{{
		args:     []string{"-i", "18", "-D", "1d"},
		expected: EditCommand{id: 18, newDeadline: "2022-03-15"},
		err:      nil,
		name:     "find by id toggle completion",
	}, {
		args:     []string{"-i", "3", "-F"},
		expected: EditCommand{id: 3, newToggleComplete: true},
		err:      nil,
		name:     "find by id toggle completion",
	}, {
		args:     []string{"-i", "4", "-B", "seems to be working"},
		expected: EditCommand{id: 4, newBody: "seems to be working"},
		err:      nil,
		name:     "find by id changed body no append/replace directive",
	}, {
		args:     []string{"-b", "edit", "command", "-F"},
		expected: EditCommand{body: "edit command", newToggleComplete: true},
		err:      nil,
		name:     "find by body key phrase mark complete",
	}, {
		args:     []string{"-i", "15", "--replace", "-B", "cleaned", "out", "by", "edit", "command"},
		expected: EditCommand{id: 15, newBody: "cleaned out by edit command", replacing: true},
		err:      nil,
		name:     "find by id edit body with replace directive",
	}, {
		args:     []string{"-b", "multiple", "-B", "appended", "to", "end", "of", "body", "by", "edit", "command", "--append"},
		expected: EditCommand{body: "multiple", newBody: "appended to end of body by edit command", appending: true},
		err:      nil,
		name:     "find by body edit body with append directive",
	}}
}

func getEditQueryBuildTestCases() []edit_query_build_test_case {
	return []edit_query_build_test_case{{
		input:      EditCommand{id: 3, newParent: 1},
		name:       "id and new parent",
		expSrchLst: []godoo.UserQueryElement{godoo.ById},
		expEdtLst:  []godoo.UserQueryElement{godoo.ByParentId},
		expSrchItm: *getTodoItm([]any{3, nil, nil, nil, nil, false}),
		expEdtItm:  *getTodoItm([]any{nil, 1, nil, nil, nil, false}),
	}, {
		input:      EditCommand{id: 4, newBody: "seems to be working"},
		name:       "id and new body",
		expSrchLst: []godoo.UserQueryElement{godoo.ById},
		expEdtLst:  []godoo.UserQueryElement{godoo.ByBody},
		expSrchItm: godoo.TodoItem{Id: 4},
		expEdtItm:  godoo.TodoItem{Body: "seems to be working"},
	}, {
		input:      EditCommand{body: "edit command", newToggleComplete: true},
		name:       "body and complete",
		expSrchLst: []godoo.UserQueryElement{godoo.ByBody},
		expEdtLst:  []godoo.UserQueryElement{godoo.ByCompletion},
		expSrchItm: godoo.TodoItem{Body: "edit command"},
		expEdtItm:  godoo.TodoItem{IsComplete: true},
	}, {
		input:      EditCommand{id: 15, replacing: true, newBody: "cleaned out by edit command"},
		name:       "id - new body replaced",
		expSrchLst: []godoo.UserQueryElement{godoo.ById},
		expEdtLst:  []godoo.UserQueryElement{godoo.ByBody, godoo.ByReplacement},
		expSrchItm: godoo.TodoItem{Id: 15},
		expEdtItm:  godoo.TodoItem{Body: "cleaned out by edit command"},
	}, {
		input:      EditCommand{body: "multiple", appending: true, newBody: "cleaned out by edit command"},
		name:       "body - new body appended",
		expSrchLst: []godoo.UserQueryElement{godoo.ByBody},
		expEdtLst:  []godoo.UserQueryElement{godoo.ByBody, godoo.ByAppending},
		expSrchItm: godoo.TodoItem{Body: "multiple"},
		expEdtItm:  godoo.TodoItem{Body: "cleaned out by edit command"},
	}, {
		input:      EditCommand{body: "multiple", tagInput: "dev", childOf: 4, appending: true, newBody: "cleaned out by edit command", newToggleComplete: true},
		name:       "body, tag, child - new body appended marked complete",
		expSrchLst: []godoo.UserQueryElement{godoo.ByBody, godoo.ByTag, godoo.ByParentId},
		expEdtLst:  []godoo.UserQueryElement{godoo.ByBody, godoo.ByAppending, godoo.ByCompletion},
		expSrchItm: *getTodoItm([]any{nil, 4, "multiple", "dev", nil, false}),
		expEdtItm:  *getTodoItm([]any{nil, nil, "cleaned out by edit command", nil, nil, true}),
	}}
}

// ***TESTING ONLY*** order significant: id, parentid, body, tag, deadline, isComplete
func getTodoItm(data []any) *godoo.TodoItem {

	// hideous - just for testing!
	itm := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))

	if data[0] != nil {
		itm.Id = data[0].(int)
	}
	if data[1] != nil {
		itm.ParentId = data[1].(int)
		itm.IsChild = true
	}
	if data[2] != nil {
		itm.Body = data[2].(string)
	}
	if data[3] != nil && len(data[3].(string)) > 0 {
		itm.Tags[data[3].(string)] = struct{}{}
	}
	if data[4] != nil {
		itm.Deadline = data[4].(time.Time)
	}
	itm.IsComplete = data[5].(bool)

	return itm
}

func TestEditFlargInput(t *testing.T) {
	tcs := _getTestCasesForEditing()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runEditTest(t, tc)
		})
	}
}

func TestEditQueryBuilding(t *testing.T) {
	tcs := getEditQueryBuildTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			runQueryBuildTests(t, tc)
		})
	}
}

func _runEditTest(t *testing.T, tc edit_item_generation_test_case) {

	//s := []string{"edit", "-i", "51", "-B", "added via http", "--append"}
	//RunCli(s, os.Stdout)

	app, _ := app.SetupCli(tc.args)
	eCmd, _ := NewEditCommand(app)
	nowStr := returnNowString()
	eCmd.parser.NowMoment, _ = time.Parse(eCmd.appCtx.DateLayout, nowStr)
	eCmd.ParseInput()

	same1, msg1 := compare(tc.expected, *eCmd)

	if same1 {
		t.Logf(">>>>PASS (finding): expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL (finding): %v", msg1)
	}
}

func runQueryBuildTests(t *testing.T, tc edit_query_build_test_case) {
	gotSrchLst, _ := tc.input.determineQueryType(godoo.Get)
	gotEdtList, _ := tc.input.determineQueryType(godoo.Update)

	gotSrchItm, _ := tc.input.setupTodoItemBasedOnUserInput()
	tc.input.getNewVals = true
	gotEdtItm, _ := tc.input.setupTodoItemBasedOnUserInput()

	same, msg := compareTdoItms(tc.expSrchItm, gotSrchItm)
	if same {
		t.Logf(">>>>PASS (srchItm): expected & got are equal")
	} else {
		t.Errorf(">>>>FAIL (srchItm): expected & got are not equal --> '%v'", msg)
	}

	same, msg = compareTdoItms(tc.expEdtItm, gotEdtItm)
	if same {
		t.Logf(">>>>PASS (edtItm): expected & got are equal")
	} else {
		t.Errorf(">>>>FAIL (edtItm): expected & got are not equal --> '%v'", msg)
	}

	same, msg = compareQueryElemsLists(tc.expSrchLst, gotSrchLst)
	if same {
		t.Logf(">>>>PASS (srchLst): expected & got are equal")
	} else {
		t.Errorf(">>>>FAIL (srchLst): expected & got are not equal --> '%v'", msg)
	}

	same, msg = compareQueryElemsLists(tc.expEdtLst, gotEdtList)
	if same {
		t.Logf(">>>>PASS (edtLst): expected & got are equal")
	} else {
		t.Errorf(">>>>FAIL (edtLst): expected & got are not equal --> '%v'", msg)
	}
}

func compareQueryElemsLists(lst1 []godoo.UserQueryElement, lst2 []godoo.UserQueryOption) (bool, string) {
	if len(lst1) != len(lst2) {
		return false, "list length differs"
	}

	// TODO: find a better way of doing this - must be one
	var lst1Ints, lst2Ints []int

	for _, itm := range lst1 {
		lst1Ints = append(lst1Ints, int(itm))
	}

	for _, itm := range lst2 {
		lst2Ints = append(lst2Ints, int(itm.Elem))
	}

	sort.Ints(lst1Ints)
	sort.Ints(lst2Ints)

	for k, v := range lst1Ints {
		if v != lst2Ints[k] {
			return false, "mismatch on items"
		}
	}

	return true, "full match"
}

func compareTdoItms(itm1, itm2 godoo.TodoItem) (bool, string) {

	if itm1.Id != itm2.Id {
		return false, "no id match"
	}
	if itm1.ParentId != itm2.ParentId {
		return false, "no parent id match"
	}
	if itm1.IsChild != itm2.IsChild {
		return false, "no isChild match"
	}
	if itm1.CreationDate != itm2.CreationDate {
		return false, "no creationDate match"
	}
	if itm1.Deadline != itm2.Deadline {
		t1 := util.StringFromDate(itm1.Deadline)
		t2 := util.StringFromDate(itm2.Deadline)
		return false, fmt.Sprintf("no deadline match - %v vs. %v", t1, t2)
	}
	if itm1.Priority != itm2.Priority {
		return false, "no priority match"
	}
	if itm1.Body != itm2.Body {
		return false, "no body match"
	}
	if itm1.IsComplete != itm2.IsComplete {
		return false, "no isComplete match"
	}
	if len(itm1.ChildItems) != len(itm2.ChildItems) {
		return false, "no match on length of childItems"
	}
	// don't think I need to check each child item (yet)
	if len(itm1.Tags) != len(itm2.Tags) {
		return false, "no match on length of tags"
	}

	var itm1Tags, itm2Tags []string
	for t := range itm1.Tags {
		itm1Tags = append(itm1Tags, t)
	}
	sort.Strings(itm1Tags)

	for t := range itm2.Tags {
		itm2Tags = append(itm2Tags, t)
	}
	sort.Strings(itm2Tags)

	for i, t := range itm1Tags {
		if t != itm1Tags[i] {
			return false, "tag mismatch"
		}
	}
	return true, "full match"
}

func compare(exp, got EditCommand) (bool, string) {
	if exp.appending != got.appending {
		return false, fmt.Sprintf("No match on appending mode. Expected '%v', got '%v'", exp.appending, got.appending)
	}
	if exp.replacing != got.replacing {
		return false, fmt.Sprintf("No match on replacing mode. Expected '%v', got '%v'", exp.replacing, got.replacing)
	}
	if exp.id != got.id {
		return false, fmt.Sprintf("No match on id. Expected '%v', got '%v'", exp.id, got.id)
	}
	if exp.body != got.body {
		return false, fmt.Sprintf("No match on body. Expected '%v', got '%v'", exp.body, got.body)
	}
	if exp.childOf != got.childOf {
		return false, fmt.Sprintf("No match on childOf. Expected '%v', got '%v'", exp.childOf, got.childOf)
	}
	if exp.deadline != got.deadline {
		return false, fmt.Sprintf("No match on deadline. Expected '%v', got '%v'", exp.deadline, got.deadline)
	}
	if exp.tagInput != got.tagInput {
		return false, fmt.Sprintf("No match on tagInput. Expected '%v', got '%v'", exp.tagInput, got.tagInput)
	}
	if exp.complete != got.complete {
		return false, fmt.Sprintf("No match on complete. Expected '%v', got '%v'", exp.complete, got.complete)
	}
	if exp.newBody != got.newBody {
		return false, fmt.Sprintf("No match on newBody. Expected '%v', got '%v'", exp.newBody, got.newBody)
	}
	if exp.newDeadline != got.newDeadline {
		return false, fmt.Sprintf("No match on newDeadline. Expected '%v', got '%v'", exp.newDeadline, got.newDeadline)
	}
	if exp.newParent != got.newParent {
		return false, fmt.Sprintf("No match on newParent. Expected '%v', got '%v'", exp.newParent, got.newParent)
	}
	if exp.newTag != got.newTag {
		return false, fmt.Sprintf("No match on newTag. Expected '%v', got '%v'", exp.newTag, got.newTag)
	}
	if exp.newToggleComplete != got.newToggleComplete {
		return false, fmt.Sprintf("No match on newlyComplete. Expected '%v', got '%v'", exp.newToggleComplete, got.newToggleComplete)
	}
	return true, "all field values equal"
}
