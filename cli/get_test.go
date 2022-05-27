package cli

import (
	"testing"
	"time"

	"github.com/mundacity/go-doo/domain"
)

func _getTestCasesForGetting() []test_case {
	return []test_case{{
		args:     []string{"get", "-a"},
		expected: domain.TodoItem{Body: "this is a body - with a dash", CreationDate: time.Now(), Priority: domain.None},
		err:      nil,
		name:     "body only",
		envVal:   0,
	}, {
		args:     []string{"get", "-i", "23"},
		expected: domain.TodoItem{Body: "body with spaces", ParentId: 9, Priority: domain.High, Tags: _getTagMap("tag w spc", ";")},
		err:      nil,
		name:     "multiple flags & args",
		envVal:   0,
	}, {
		args:     []string{"get", "-c", "9"},
		expected: domain.TodoItem{Body: "body with spaces", ParentId: 9, Priority: domain.High, Tags: _getTagMap("tag w spc", ";")},
		err:      nil,
		name:     "multiple flags & args",
		envVal:   0,
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

func _runGetTest(t *testing.T, tc test_case) {

	_quickTest(tc)

	app, _ := Init(tc.args)
	gCmd, _ := NewGetCommand(app)
	nowStr := returnNowString()
	gCmd.parser.NowMoment, _ = time.Parse(gCmd.appCtx.DateLayout, nowStr)

	gCmd.ParseFlags()
	td, _ := gCmd.GenerateTodoItem()

	same, msg := _compare(tc.expected, td)

	if same {
		t.Logf(">>>>PASS: expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL: %v", msg)
	}
}
