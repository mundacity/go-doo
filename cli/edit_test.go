package cli

import (
	"fmt"
	"testing"
	"time"
)

type edit_test_case struct {
	args     []string
	expected EditCommand
	err      error
	name     string
}

func _getTestCasesForEditing() []edit_test_case {
	return []edit_test_case{{
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

func TestEdit(t *testing.T) {
	tcs := _getTestCasesForEditing()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runEditTest(t, tc)
		})
	}
}

func _runEditTest(t *testing.T, tc edit_test_case) {

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
