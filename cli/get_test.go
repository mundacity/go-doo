package cli

import (
	"fmt"
	"testing"
	"time"
)

type get_test_case struct {
	args     []string
	expected GetCommand
	err      error
	name     string
}

func _getTestCasesForGetting() []get_test_case {
	return []get_test_case{{
		args:     []string{"-a"},
		expected: GetCommand{getAll: true, deadlineDate: "."},
		err:      nil,
		name:     "get all",
	}, {
		args:     []string{"-i", "23"},
		expected: GetCommand{id: 23, deadlineDate: "."},
		err:      nil,
		name:     "get by id",
	}, {
		args:     []string{"-c", "9"},
		expected: GetCommand{childOf: 9, deadlineDate: "."},
		err:      nil,
		name:     "get by child id",
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

	//_quickTest(tc)

	app, _ := Init(tc.args)
	gCmd, _ := NewGetCommand(app)
	nowStr := returnNowString()
	gCmd.parser.NowMoment, _ = time.Parse(gCmd.appCtx.DateLayout, nowStr)

	gCmd.ParseFlags()

	same, msg := compareResults(tc.expected, *gCmd)

	if same {
		t.Logf(">>>>PASS: expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL: %v", msg)
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
