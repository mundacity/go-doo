package srv

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/fake"
	lg "github.com/mundacity/quick-logger"
)

type test_request struct {
	method       string
	path         string
	expectedCode int
	name         string
}

type badJson struct {
	Name string `json:"name"`
	Val  int    `json:"val"`
}

type bad_json_request struct {
	json   badJson
	method string
	path   string
	name   string
	code   int
}

func getTestRequests() []test_request {
	return []test_request{{
		method:       http.MethodGet,
		path:         "/",
		expectedCode: http.StatusNotFound,
		name:         "get request to / expecting 404",
	},
	}
}

func getBadJsonTests() []bad_json_request {
	return []bad_json_request{{
		json:   badJson{Name: "dud", Val: 4},
		method: http.MethodPost,
		path:   "/add",
		code:   http.StatusBadRequest,
		name:   "adding with bad json",
	}, {
		json:   badJson{Name: "dud", Val: 4},
		method: http.MethodGet,
		path:   "/get",
		code:   http.StatusBadRequest,
		name:   "getting with bad json",
	}, {
		json:   badJson{Name: "dud", Val: 4},
		method: http.MethodPut,
		path:   "/edit",
		code:   http.StatusBadRequest,
		name:   "editing with bad json",
	},
	}
}

func TestBadJson(t *testing.T) {

	lg.Logger = lg.NewDummyLogger()
	tcs := getBadJsonTests()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			runBadJsonTest(t, tc)
		})
	}

}

func getSrvConfig() godoo.ServerConfigVals {
	c := godoo.ServerConfigVals{}
	c.Conn = ""
	c.DateFormat = "2006-01-02"
	c.PriorityList = godoo.NewPriorityList()
	c.RunPriorityList = true
	c.Repo = fake.RepoDud{}
	return c
}

func runBadJsonTest(t *testing.T, tc bad_json_request) {

	w := httptest.NewRecorder()

	f := FakeSrvContext{}
	c := getSrvConfig()
	f.SetupServerContext(c)

	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(tc.json)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(tc.method, tc.path, &b)
	f.handler.HandleRequests(w, req)
	//resp := fmt.Sprint(w.Body)

	if w.Code != tc.code {
		t.Errorf(">>>>FAIL: http status code mismatch: got %v, expecting %v", w.Code, tc.code)
	} else {
		t.Logf(">>>>PASS: http status code match: got %v, expecting %v", w.Code, tc.code)
	}
}

// func TestRunHandlerTests(t *testing.T) {

// 	lg.Logger = lg.NewDummyLogger()
// 	tcs := getTestRequests()

// 	for _, tc := range tcs {
// 		t.Run(tc.name, func(t *testing.T) {
// 			runHandlerTests(t, tc)
// 		})
// 	}
// }

// func runHandlerTests(t *testing.T, tc test_request) {

// 	w := httptest.NewRecorder()

// 	f := FakeSrvContext{}
// 	c := getSrvConfig()
// 	f.SetupServerContext(c)

// 	var b bytes.Buffer
// 	req, _ := http.NewRequest(tc.method, tc.path, &b)

// 	f.handler.HandleRequests(w, req)

// 	if w.Code != tc.expectedCode {
// 		t.Errorf(">>>>FAIL: http status code mismatch: got %v, expecting %v", w.Code, tc.expectedCode)
// 	} else {
// 		t.Logf(">>>>PASS: http status code match: got %v, expecting %v", w.Code, tc.expectedCode)
// 	}
// }
