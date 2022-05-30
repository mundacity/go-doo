package util

import (
	"testing"
	"time"
)

type string_from_slice_test_case struct {
	input    []string
	expected string
	name     string
}

type string_from_date_test_case struct {
	input    time.Time
	expected string
	name     string
}

func getSliceToStringTestCases() []string_from_slice_test_case {
	return []string_from_slice_test_case{{
		input:    []string{"this", "is", "a", "slice", "of", "strings"},
		expected: "this is a slice of strings",
		name:     "simple conversion",
	}, {
		input:    []string{" ", "this", "is", "a", "slice", "of", "strings"},
		expected: "this is a slice of strings",
		name:     "leading space",
	}, {
		input:    []string{" ", "this", "is", "a", "slice", "of", "strings", " "},
		expected: "this is a slice of strings",
		name:     "leading and trailing space",
	}, {
		input:    []string{" ", "this", "is", "a", " ", "slice", "of", "strings", " "},
		expected: "this is a   slice of strings",
		name:     "space in middle == 3 spaces between",
	}}
}

func getDateToStringTestCases() []string_from_date_test_case {

	tm1 := time.Date(2022, 5, 30, 0, 0, 0, 0, time.Local)
	tm2 := time.Date(2022, 5, 1, 0, 0, 0, 0, time.Local)
	tm3 := time.Date(2022, 10, 1, 0, 0, 0, 0, time.Local)
	tm4 := time.Date(2022, 10, 10, 0, 0, 0, 0, time.Local)

	return []string_from_date_test_case{{
		input:    tm1,
		expected: "2022-05-30",
		name:     "month below 10",
	}, {
		input:    tm2,
		expected: "2022-05-01",
		name:     "month & day below 10",
	}, {
		input:    tm3,
		expected: "2022-10-01",
		name:     "day below 10",
	}, {
		input:    tm4,
		expected: "2022-10-10",
		name:     "month & day 10 or over",
	}}
}

func TestStringFromSlice(t *testing.T) {
	tcs := getSliceToStringTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			runSliceToString(t, tc)
		})
	}
}

func runSliceToString(t *testing.T, tc string_from_slice_test_case) {
	got := StringFromSlice(tc.input)

	if got == tc.expected {
		t.Logf(">>>>PASS: expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL: expected '%v', got '%v'", tc.expected, got)
	}
}

func TestStringFromDate(t *testing.T) {
	tcs := getDateToStringTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			runDateToString(t, tc)
		})
	}
}

func runDateToString(t *testing.T, tc string_from_date_test_case) {
	got := StringFromDate(tc.input)

	if got == tc.expected {
		t.Logf(">>>>PASS: expected and got are equal")
	} else {
		t.Errorf(">>>>FAIL: expected '%v', got '%v'", tc.expected, got)
	}
}
