package cli

import (
	"fmt"
	"strings"
	"testing"
	"time"

	godoo "github.com/mundacity/go-doo"
)

type add_test_case struct {
	args     []string
	expected godoo.TodoItem
	err      error
	name     string
	envVal   int
}

func _getTestCases() []add_test_case {
	return []add_test_case{{
		args:     []string{"add", "-b", "this is a body - with a dash"},
		expected: godoo.TodoItem{Body: "this is a body - with a dash", CreationDate: time.Now(), Priority: godoo.None},
		err:      nil,
		name:     "body only",
		envVal:   0,
	}, {
		args:     []string{"add", "-t", "tag w spc", "-c", "9", "-b", "body with spaces", "-m", "h"},
		expected: godoo.TodoItem{Body: "body with spaces", ParentId: 9, Priority: godoo.High, Tags: _getTagMap("tag w spc", "*")},
		err:      nil,
		name:     "multiple flags & args",
		envVal:   0,
	}, {
		args:     []string{"add", "-c", "7", "-b", "body with no body flag and spaces", "-m", "l"},
		expected: godoo.TodoItem{Body: "body with no body flag and spaces", ParentId: 7, Priority: godoo.Low},
		err:      nil,
		name:     "no body tag",
		envVal:   0,
	}, {
		args:     []string{"add", "-b", "this is a spaceful body", "-d", "2021-04-16"},
		expected: godoo.TodoItem{Body: "this is a spaceful body", Priority: godoo.DateBased, Deadline: time.Date(2021, 04, 16, 0, 0, 0, 0, time.UTC)},
		err:      nil,
		name:     "deadline test",
		envVal:   0,
	}, {
		args:     []string{"add", "-b", "I'm including an apostrophe", "-d", "2021-04-16"},
		expected: godoo.TodoItem{Body: "I'm including an apostrophe", Priority: godoo.DateBased, Deadline: time.Date(2021, 04, 16, 0, 0, 0, 0, time.UTC)},
		err:      nil,
		name:     "apostrophe test",
		envVal:   0,
	}, {
		args:     []string{"add", "-b", "flagless body with spaces", "-t", "tag1*tag2*tag3"},
		expected: godoo.TodoItem{Body: "flagless body with spaces", Priority: godoo.None, Tags: _getTagMap("tag1*tag2*tag3", "*")},
		err:      nil,
		name:     "apostrophe test",
		envVal:   0,
	}, {
		args:     []string{"add", "-b", "flagless body with spaces", "-t", "tag1*tag2*tag3", "-m", "h"},
		expected: godoo.TodoItem{Body: "flagless body with spaces", Priority: godoo.High, Tags: _getTagMap("tag1*tag2*tag3", "*")},
		err:      nil,
		name:     "high priority item",
		envVal:   0,
	}, {
		args:     []string{"add", "-b", "new item body", "-m", "z"},
		expected: godoo.TodoItem{Body: "new item body", Priority: godoo.None},
		err:      &InvalidArgumentError{},
		name:     "invalid priority arg",
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

	CliContext = &FakeAppContext{}
	CliContext.SetupCliContext(tc.args)
	cmd, _ := CliContext.GetCommand()

	cmd.ParseInput()
	td, err := cmd.BuildItemFromInput()

	if err != nil {
		if tc.err == err {
			t.Logf(">>>>PASS: expected [%v] & got [%v]", tc.err, err)
			return
		} else {
			t.Errorf(">>>>FAIL: expected [%v] & got [%v]", tc.err, err)
			return
		}
	}

	same, msg := compareTestResults(tc.expected, td)

	if same {
		t.Logf(">>>>PASS: expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL: %v", msg)
	}

}

func compareTestResults(expected, got godoo.TodoItem) (bool, string) {

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

// func returnNowString() string {
// 	n := time.Date(2022, 03, 14, 0, 0, 0, 0, time.UTC)
// 	return util.StringFromDate(n)
// }
