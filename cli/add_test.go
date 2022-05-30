package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mundacity/go-doo/domain"
	"github.com/mundacity/go-doo/util"
)

type add_test_case struct {
	args     []string
	expected domain.TodoItem
	err      error
	name     string
	envVal   int
}

func _getTestCases() []add_test_case {
	return []add_test_case{{
		args:     []string{"-b", "this", "is", "a", "body", "-", "with", "a", "dash"},
		expected: domain.TodoItem{Body: "this is a body - with a dash", CreationDate: time.Now(), Priority: domain.None},
		err:      nil,
		name:     "body only",
		envVal:   0,
	}, {
		args:     []string{"-t", "tag", "w", "spc", "-c", "9", "-b", "body", "with", "spaces", "-m", "h"},
		expected: domain.TodoItem{Body: "body with spaces", ParentId: 9, Priority: domain.High, Tags: _getTagMap("tag w spc", "*")},
		err:      nil,
		name:     "multiple flags & args",
		envVal:   0,
	}, {
		args:     []string{"-c", "7", "body", "with no body flag", "and spaces", "-m", "l"},
		expected: domain.TodoItem{Body: "body with no body flag and spaces", ParentId: 7, Priority: domain.Low},
		err:      nil,
		name:     "no body tag",
		envVal:   0,
	}, {
		args:     []string{"this is a spaceful body", "-d", "-1y", "1m", "2d"},
		expected: domain.TodoItem{Body: "this is a spaceful body", Priority: domain.DateBased, Deadline: time.Date(2021, 04, 16, 0, 0, 0, 0, time.UTC)},
		err:      nil,
		name:     "deadline test",
		envVal:   0,
	}, {
		args:     []string{"I'm", "including", "an apostrophe", "-d", "-1y", "1m", "2d"},
		expected: domain.TodoItem{Body: "I'm including an apostrophe", Priority: domain.DateBased, Deadline: time.Date(2021, 04, 16, 0, 0, 0, 0, time.UTC)},
		err:      nil,
		name:     "apostrophe test",
		envVal:   0,
	}, {
		args:     []string{"flagless", "body", "with", "spaces", "-t", "tag1*tag2*tag3"},
		expected: domain.TodoItem{Body: "flagless body with spaces", Priority: domain.None, Tags: _getTagMap("tag1*tag2*tag3", "*")},
		err:      nil,
		name:     "apostrophe test",
		envVal:   0,
	}}
}

func _getTagMap(input, delim string) map[string]struct{} {
	out := strings.Split(input, delim)
	lgth := len(out)
	mp := make(map[string]struct{}, lgth)
	for _, t := range out {
		mp[t] = struct{}{}
	}
	return mp
}

func TestItemGeneration(t *testing.T) {

	tcs := _getTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_runAddTest(t, tc)
		})
	}
}

func _runAddTest(t *testing.T, tc add_test_case) {

	//_quickTest(tc)

	app, _ := Init(tc.args)
	addCmd, _ := NewAddCommand(app)
	nowStr := returnNowString()
	addCmd.parser.NowMoment, _ = time.Parse(addCmd.appCtx.DateLayout, nowStr)

	addCmd.ParseFlags()
	td, _ := addCmd.GenerateTodoItem()

	same, msg := compareTestResults(tc.expected, td)

	if same {
		t.Logf(">>>>PASS: expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL: %v", msg)
	}

}

func compareTestResults(expected, got domain.TodoItem) (bool, string) {

	if expected.Body != got.Body {
		return false, fmt.Sprintf("body doesn't match. Expected '%v', got '%v'", expected.Body, got.Body)
	}
	if expected.Deadline != got.Deadline {
		return false, fmt.Sprintf("deadline doesn't match. Expected '%v', got '%v'", expected.Deadline, got.Deadline)
	}
	if expected.ParentId != got.ParentId {
		return false, fmt.Sprintf("parentId doesn't match. Expected '%v', got '%v'", expected.ParentId, got.ParentId)
	}
	if expected.Priority != got.Priority {
		return false, fmt.Sprintf("priority doesn't match. Expected '%v', got '%v'", expected.Priority, got.Priority)
	}
	if len(expected.Tags) != len(got.Tags) {
		return false, fmt.Sprintf("len doesn't match. Expected '%v', got '%v'", len(expected.Tags), len(got.Tags))
	}
	if expected.Id != got.Id {
		return false, fmt.Sprintf("id doesn't match. Expected '%v', got '%v'", expected.Id, got.Id)
	}

	for s := range expected.Tags {

		_, got_exists := got.Tags[s]

		if !got_exists {
			return false, fmt.Sprintf("key missing: '%v'", s)
		}
	}
	return true, ""
}

func quickTest(tc add_test_case) {
	RunApp(tc.args, os.Stdout)
}

func returnNowString() string {
	n := time.Date(2022, 03, 14, 0, 0, 0, 0, time.UTC)
	return util.StringFromDate(n)
}
