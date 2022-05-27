package cli

import (
	"testing"
	"time"

	"github.com/mundacity/go-doo/domain"
)

func _getTestCasesForEditing() []test_case {
	return []test_case{{
		args:     []string{"edit", "-i", "4", "-B", "seems to be working"},
		expected: domain.TodoItem{Body: "this is a body - with a dash", CreationDate: time.Now(), Priority: domain.None},
		err:      nil,
		name:     "edit body no append or replace directive",
		envVal:   0,
	}, {
		args:     []string{"edit", "-b", "edit", "command", "-F"},
		expected: domain.TodoItem{Body: "this is a body - with a dash", CreationDate: time.Now(), Priority: domain.None},
		err:      nil,
		name:     "body only",
		envVal:   0,
	}, {
		args:     []string{"edit", "-i", "15", "--replace", "-B", "cleaned", "out", "by", "edit", "command"},
		expected: domain.TodoItem{Body: "this is a body - with a dash", CreationDate: time.Now(), Priority: domain.None},
		err:      nil,
		name:     "body only",
		envVal:   0,
	}, {
		args:     []string{"edit", "-b", "multiple", "-B", "appended", "to", "end", "of", "body", "by", "edit", "command", "--append"},
		expected: domain.TodoItem{Body: "this is a body - with a dash", CreationDate: time.Now(), Priority: domain.None},
		err:      nil,
		name:     "body only",
		envVal:   0,
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

func _runEditTest(t *testing.T, tc test_case) {

	_quickTest(tc)

	app, _ := Init(tc.args)
	eCmd, _ := NewEditCommand(app)
	nowStr := returnNowString()
	eCmd.parser.NowMoment, _ = time.Parse(eCmd.appCtx.DateLayout, nowStr)

	eCmd.ParseFlags()
	td, _ := eCmd.GenerateTodoItem()

	same, msg := _compare(tc.expected, td)

	if same {
		t.Logf(">>>>PASS: expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL: %v", msg)
	}
}
