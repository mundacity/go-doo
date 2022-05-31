package cli

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/mundacity/go-doo/domain"
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
	expSrchLst []domain.UserQueryElement
	expEdtLst  []domain.UserQueryElement
	expSrchItm domain.TodoItem
	expEdtItm  domain.TodoItem
}

func _getTestCasesForEditing() []edit_item_generation_test_case {
	return []edit_item_generation_test_case{{
		args:     []string{"-i", "4", "-B", "seems to be working"},
		expected: EditCommand{id: 4, newBody: "seems to be working"},
		err:      nil,
		name:     "find by id changed body no append/replace directive",
	}, {
		args:     []string{"-b", "edit", "command", "-F"},
		expected: EditCommand{body: "edit command", newlyComplete: true},
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

func getQueryBuildTestCases() []edit_query_build_test_case {
	return []edit_query_build_test_case{{
		input:      EditCommand{id: 4, newBody: "seems to be working"},
		name:       "id and new body",
		expSrchLst: []domain.UserQueryElement{domain.ById},
		expEdtLst:  []domain.UserQueryElement{domain.ByBody},
		expSrchItm: domain.TodoItem{Id: 4},
		expEdtItm:  domain.TodoItem{Body: "seems to be working"},
	}, {
		input:      EditCommand{body: "edit command", newlyComplete: true},
		name:       "body and complete",
		expSrchLst: []domain.UserQueryElement{domain.ByBody},
		expEdtLst:  []domain.UserQueryElement{domain.ByCompletion},
		expSrchItm: domain.TodoItem{Body: "edit command"},
		expEdtItm:  domain.TodoItem{IsComplete: true},
	}}
}

func TestEditFlargInput(t *testing.T) {
	tcs := _getTestCasesForEditing()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runEditTest(t, tc)
		})
	}
}

func TestQueryBuilding(t *testing.T) {
	tcs := getQueryBuildTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			runQueryBuildTests(t, tc)
		})
	}
}

func _runEditTest(t *testing.T, tc edit_item_generation_test_case) {

	//RunApp(tc.args, os.Stdout)

	app, _ := Init(tc.args)
	eCmd, _ := NewEditCommand(app)
	nowStr := returnNowString()
	eCmd.parser.NowMoment, _ = time.Parse(eCmd.appCtx.DateLayout, nowStr)
	eCmd.ParseFlags()

	same1, msg1 := compare(tc.expected, *eCmd)

	if same1 {
		t.Logf(">>>>PASS (finding): expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL (finding): %v", msg1)
	}
}

func runQueryBuildTests(t *testing.T, tc edit_query_build_test_case) {
	gotSrchLst, _ := tc.input.determineQueryType(domain.Get)
	gotEdtList, _ := tc.input.determineQueryType(domain.Update)

	gotSrchItm, _ := tc.input.GenerateTodoItem()
	tc.input.getNewVals = true
	gotEdtItm, _ := tc.input.GenerateTodoItem()

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

func compareQueryElemsLists(lst1, lst2 []domain.UserQueryElement) (bool, string) {
	if len(lst1) != len(lst2) {
		return false, "list length differs"
	}

	// TODO: find a better way of doing this - must be one
	var lst1Ints, lst2Ints []int

	for _, itm := range lst1 {
		lst1Ints = append(lst1Ints, int(itm))
	}

	for _, itm := range lst2 {
		lst2Ints = append(lst2Ints, int(itm))
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

func compareTdoItms(itm1, itm2 domain.TodoItem) (bool, string) {

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
		return false, "no deadline match"
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
	for t, _ := range itm1.Tags {
		itm1Tags = append(itm1Tags, t)
	}
	sort.Strings(itm1Tags)

	for t, _ := range itm2.Tags {
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
	if exp.newlyComplete != got.newlyComplete {
		return false, fmt.Sprintf("No match on newlyComplete. Expected '%v', got '%v'", exp.newlyComplete, got.newlyComplete)
	}
	return true, "all field values equal"
}
