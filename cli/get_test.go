package cli

import (
	"fmt"
	"testing"
	"time"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/app"
)

type get_test_case struct {
	args     []string
	expected GetCommand
	err      error
	name     string
}

type get_query_build_test_case struct {
	input      GetCommand
	name       string
	expSrchLst []godoo.UserQueryElement
	expSrchItm godoo.TodoItem
}

func _getTestCasesForGetting() []get_test_case {
	return []get_test_case{{
		args:     []string{"-a"},
		expected: GetCommand{getAll: true},
		err:      nil,
		name:     "get all",
	}, {
		args:     []string{"-i", "23"},
		expected: GetCommand{id: 23},
		err:      nil,
		name:     "get by id",
	}, {
		args:     []string{"-c", "9"},
		expected: GetCommand{childOf: 9},
		err:      nil,
		name:     "get by child id",
	}, {
		args:     []string{"-F", "-d", "2022-06-01:2022-06-18"},
		expected: GetCommand{complete: false, deadlineDate: "2022-06-01:2022-06-18"},
		err:      nil,
		name:     "get incomplete with literal deadline range (maxLen be at least 21)",
	}}
}

func getGetQueryBuildTestCases() []get_query_build_test_case {
	return []get_query_build_test_case{{
		input:      GetCommand{getAll: true},
		name:       "get all",
		expSrchLst: []godoo.UserQueryElement{},
		expSrchItm: *getTodoItm([]any{nil, nil, nil, nil, nil, false}),
	}, {
		input:      GetCommand{getAll: true, toggleComplete: true},
		name:       "get all incomplete",
		expSrchLst: []godoo.UserQueryElement{godoo.ByCompletion},
		expSrchItm: *getTodoItm([]any{nil, nil, nil, nil, nil, false}),
	}, {
		input:      GetCommand{getAll: true, complete: true},
		name:       "get all complete",
		expSrchLst: []godoo.UserQueryElement{godoo.ByCompletion},
		expSrchItm: *getTodoItm([]any{nil, nil, nil, nil, nil, true}),
	}, {
		input:      GetCommand{bodyPhrase: "edit command", childOf: 99, tagInput: "test"},
		name:       "body child tag",
		expSrchLst: []godoo.UserQueryElement{godoo.ByBody, godoo.ByParentId, godoo.ByTag},
		expSrchItm: *getTodoItm([]any{nil, 99, "edit command", "test", nil, false}),
	}, {
		input:      GetCommand{id: 15},
		name:       "id",
		expSrchLst: []godoo.UserQueryElement{godoo.ById},
		expSrchItm: *getTodoItm([]any{15, nil, nil, nil, nil, false}),
	}, {
		input:      GetCommand{bodyPhrase: "multiple", complete: true, childOf: 8},
		name:       "body complete child",
		expSrchLst: []godoo.UserQueryElement{godoo.ByBody, godoo.ByCompletion, godoo.ByParentId},
		expSrchItm: *getTodoItm([]any{nil, 8, "multiple", nil, nil, true}),
	}, {
		input:      GetCommand{complete: true},
		name:       "completion",
		expSrchLst: []godoo.UserQueryElement{godoo.ByCompletion},
		expSrchItm: *getTodoItm([]any{nil, nil, nil, nil, nil, true}),
	}}
}

func TestGetting(t *testing.T) {

	tcs := _getTestCasesForGetting()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runGetTest(t, tc)
		})
	}
}

func _runGetTest(t *testing.T, tc get_test_case) {

	app, _ := app.SetupCli(tc.args)
	gCmd, _ := NewGetCommand(app)
	nowStr := returnNowString()
	gCmd.parser.NowMoment, _ = time.Parse(gCmd.appCtx.DateLayout, nowStr)

	gCmd.ParseInput()

	same, msg := compareResults(tc.expected, *gCmd)

	if same {
		t.Logf(">>>>PASS: expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL: %v", msg)
	}
}

func TestGetQueryBuilding(t *testing.T) {
	tcs := getGetQueryBuildTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			runGetQueryBuildTests(t, tc)
		})
	}
}

func runGetQueryBuildTests(t *testing.T, tc get_query_build_test_case) {
	gotSrchLst, _ := tc.input.determineQueryType()
	gotSrchItm, _ := tc.input.interpretUserInput()

	same, msg := compareTdoItms(tc.expSrchItm, gotSrchItm)
	if same {
		t.Logf(">>>>PASS (srchItm): expected & got are equal")
	} else {
		t.Errorf(">>>>FAIL (srchItm): expected & got are not equal --> '%v'", msg)
	}

	same, msg = compareQueryElemsLists(tc.expSrchLst, gotSrchLst)
	if same {
		t.Logf(">>>>PASS (srchLst): expected & got are equal")
	} else {
		t.Errorf(">>>>FAIL (srchLst): expected & got are not equal --> '%v'", msg)
	}
}

func compareResults(exp, got GetCommand) (bool, string) {
	if exp.id != got.id {
		return false, fmt.Sprintf("No match on id. Expected '%v', got '%v'", exp.id, got.id)
	}
	if exp.next != got.next {
		return false, fmt.Sprintf("No match on next. Expected '%v', got '%v'", exp.next, got.next)
	}
	if exp.tagInput != got.tagInput {
		return false, fmt.Sprintf("No match on tagInput. Expected '%v', got '%v'", exp.tagInput, got.tagInput)
	}
	if exp.bodyPhrase != got.bodyPhrase {
		return false, fmt.Sprintf("No match on bodyPhrase. Expected '%v', got '%v'", exp.bodyPhrase, got.bodyPhrase)
	}
	if exp.childOf != got.childOf {
		return false, fmt.Sprintf("No match on childOf. Expected '%v', got '%v'", exp.childOf, got.childOf)
	}
	if exp.parentOf != got.parentOf {
		return false, fmt.Sprintf("No match on parentOf. Expected '%v', got '%v'", exp.parentOf, got.parentOf)
	}
	if exp.deadlineDate != got.deadlineDate {
		return false, fmt.Sprintf("No match on deadlineDate. Expected '%v', got '%v'", exp.deadlineDate, got.deadlineDate)
	}
	if exp.creationDate != got.creationDate {
		return false, fmt.Sprintf("No match on creationDate. Expected '%v', got '%v'", exp.creationDate, got.creationDate)
	}
	if exp.getAll != got.getAll {
		return false, fmt.Sprintf("No match on getAll. Expected '%v', got '%v'", exp.getAll, got.getAll)
	}
	if exp.complete != got.complete {
		return false, fmt.Sprintf("No match on complete. Expected '%v', got '%v'", exp.complete, got.complete)
	}
	return true, "all field values matching"
}
