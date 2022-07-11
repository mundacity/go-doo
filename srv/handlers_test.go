package srv_test

import (
	"net/http"
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
		method: http.MethodPost,
		path:   "/edit",
		code:   http.StatusBadRequest,
		name:   "editing with bad json",
	},
	}
}

// func TestBadJson(t *testing.T) {

// 	tcs := getBadJsonTests()
// 	for _, tc := range tcs {
// 		t.Run(tc.name, func(t *testing.T) {
// 			runBadJsonTest(t, tc)
// 		})
// 	}

// }

// func runBadJsonTest(t *testing.T, tc bad_json_request) {
// 	w := httptest.NewRecorder()
// 	h := srv.Handler{Repo: fake.Repo{}}

// 	var b bytes.Buffer
// 	err := json.NewEncoder(&b).Encode(tc.json)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	req, _ := http.NewRequest(tc.method, tc.path, &b)
// 	h.HandleRequests(w, req)
// 	resp := fmt.Sprint(w.Body)

// 	if resp == "" {

// 	}
// }

// func TestRunHandlerTests(t *testing.T) {

// 	tcs := getTestRequests()

// 	for _, tc := range tcs {
// 		t.Run(tc.name, func(t *testing.T) {
// 			runHandlerTests(t, tc)
// 		})
// 	}
// }

// func runHandlerTests(t *testing.T, tc test_request) {
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest(http.MethodGet, tc.path, nil)
// 	srv.HandleRequests(w, req)
// }
