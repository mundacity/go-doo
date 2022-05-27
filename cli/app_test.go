package cli

// import (
// 	"os"
// 	"strconv"
// 	"testing"
// )

// type init_test_case struct {
// 	name     string
// 	envVal   int
// 	expected error
// }

// type process_args_test_case struct {
// 	name          string
// 	args          []string
// 	expectedBools []bool
// 	expectedError error
// 	validity      bool
// }

// func _getTestCasesForInit() []init_test_case {
// 	return []init_test_case{{
// 		name:     "val = 0 (valid)",
// 		envVal:   0,
// 		expected: nil,
// 	}, {
// 		name:     "val = 1 (valid)",
// 		envVal:   1,
// 		expected: nil,
// 	}, {
// 		name:     "val = -1 (invalid)",
// 		envVal:   -1,
// 		expected: &InstanceTypeNotRecognised{},
// 	}, {
// 		name:     "val = 2 (valid)",
// 		envVal:   2,
// 		expected: nil,
// 	},
// 	}
// }

// func _getTestCasesForProcessArgs() []process_args_test_case {
// 	return []process_args_test_case{{
// 		name:          "adding only",
// 		args:          []string{"-add"},
// 		expectedBools: []bool{true, false, false, false},
// 		expectedError: nil,
// 		validity:      true,
// 	}, {
// 		name:          "getting only",
// 		args:          []string{"-get"},
// 		expectedBools: []bool{false, true, false, false},
// 		expectedError: nil,
// 		validity:      true,
// 	}, {
// 		name:          "editing only",
// 		args:          []string{"-edit"},
// 		expectedBools: []bool{false, false, true, false},
// 		expectedError: nil,
// 		validity:      true,
// 	}, {
// 		name:          "deleting only",
// 		args:          []string{"-del"},
// 		expectedBools: []bool{false, false, false, true},
// 		expectedError: nil,
// 		validity:      true,
// 	}, {
// 		name:          "two args",
// 		args:          []string{"-add", "-del"},
// 		expectedBools: []bool{true, false, false, true},
// 		expectedError: nil,
// 		validity:      false,
// 	}, {
// 		name:          "three args",
// 		args:          []string{"-add", "-edit", "-del"},
// 		expectedBools: []bool{true, false, true, true},
// 		expectedError: nil,
// 		validity:      false,
// 	}, {
// 		name:          "four args",
// 		args:          []string{"-add", "-get", "-edit", "-del"},
// 		expectedBools: []bool{true, true, true, true},
// 		expectedError: nil,
// 		validity:      false,
// 	}, {
// 		name:          "four args reverse order",
// 		args:          []string{"-del", "-edit", "-get", "-add"},
// 		expectedBools: []bool{true, true, true, true},
// 		expectedError: nil,
// 		validity:      false,
// 	},
// 	}
// }

// func TestInitMethod(t *testing.T) {
// 	oldEnv := os.Getenv("INSTANCE_TYPE")
// 	tcs := _getTestCasesForInit()

// 	for _, tc := range tcs {
// 		t.Run(tc.name, func(t *testing.T) {
// 			_testEnvInput(t, tc)
// 		})
// 	}

// 	os.Setenv("INSTANCE_TYPE", oldEnv)
// }

// func _testEnvInput(t *testing.T, tc init_test_case) {
// 	s := strconv.Itoa(tc.envVal)
// 	os.Setenv("INSTANCE_TYPE", s)

// 	_, err := Init(nil) // init only assigns args to appContext so nil's fine
// 	if tc.expected != err {
// 		t.Errorf(">>>>FAILED: expected '%v', got '%v'", tc.expected, err)
// 	} else {
// 		t.Logf(">>>>PASSED: expected '%v', got '%v'", tc.expected, err)
// 	}
// }

// func TestArgProcessing(t *testing.T) {
// 	tcs := _getTestCasesForProcessArgs()

// 	for _, tc := range tcs {
// 		t.Run(tc.name, func(t *testing.T) {
// 			_testArgProcessingErrors(t, tc)
// 		})
// 		t.Run(tc.name+"_bool check", func(t *testing.T) {
// 			_testArgProcessingBools(t, tc)
// 		})
// 		t.Run(tc.name+"_options validity check", func(t *testing.T) {
// 			_testBoolSliceValidity(t, tc)
// 		})
// 	}
// }

// func _testArgProcessingErrors(t *testing.T, tc process_args_test_case) {
// 	app := appContext{args: tc.args}
// 	_, err := processArgs(&app)

// 	if tc.expectedError != err {
// 		t.Errorf(">>>>FAILED: expected '%v', got '%v'", tc.expectedError, err)
// 	} else {
// 		t.Logf(">>>>PASSED: expected '%v', got '%v'", tc.expectedError, err)
// 	}
// }

// func _testArgProcessingBools(t *testing.T, tc process_args_test_case) {
// 	app := appContext{args: tc.args[0:]}
// 	bls, _ := processArgs(&app)
// 	pass := true
// 	var actual []bool

// 	for i, b := range bls {
// 		actual = append(actual, *b)
// 		if *b != tc.expectedBools[i] {
// 			pass = false
// 		}
// 	}

// 	if pass {
// 		t.Logf(">>>>PASSED: expected '%v', got '%v'", tc.expectedBools, actual)
// 	} else {
// 		t.Errorf(">>>>FAILED: expected '%v', got '%v'", tc.expectedBools, actual)
// 	}
// }

// func _testBoolSliceValidity(t *testing.T, tc process_args_test_case) {
// 	app := appContext{args: tc.args[0:]}
// 	bls, _ := processArgs(&app)
// 	actual := inputIsValid(bls)

// 	if tc.validity == actual {
// 		t.Logf(">>>>PASSED: expected '%v', got '%v'", tc.validity, actual)
// 	} else {
// 		t.Errorf(">>>>FAILED: expected '%v', got '%v'", tc.validity, actual)
// 	}
// }
