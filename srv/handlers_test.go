package srv

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/fake"
	lg "github.com/mundacity/quick-logger"
)

type test_request struct {
	method       string
	path         string
	expectedCode int
	name         string
	caseId       int
}

func getTestRequests() []test_request {
	return []test_request{{
		method:       http.MethodGet,
		path:         "/get",
		expectedCode: http.StatusBadRequest,
		name:         "malformed JSON body to /get",
		caseId:       0,
	}, {
		method:       http.MethodPost,
		path:         "/add",
		expectedCode: http.StatusBadRequest,
		name:         "malformed JSON body to /add",
		caseId:       0,
	}, {
		method:       http.MethodPut,
		path:         "/edit",
		expectedCode: http.StatusBadRequest,
		name:         "malformed JSON body to /edit",
		caseId:       0,
	}, {
		method:       http.MethodGet,
		path:         "/get",
		expectedCode: http.StatusOK,
		name:         "valid query by id request to /get",
		caseId:       2,
	}, {
		method:       http.MethodPost,
		path:         "/add",
		expectedCode: http.StatusBadRequest,
		name:         "invalid POST to /add - no creationDate",
		caseId:       1,
	}, {
		method:       http.MethodPost,
		path:         "/add",
		expectedCode: http.StatusOK,
		name:         "valid POST to /add",
		caseId:       3,
	}, {
		method:       http.MethodPut,
		path:         "/edit",
		expectedCode: http.StatusOK,
		name:         "valid query by id request to /edit",
		caseId:       4,
	}, {
		method:       http.MethodPut,
		path:         "/edit",
		expectedCode: http.StatusBadRequest,
		name:         "invalid query to /edit - missing query portion",
		caseId:       5,
	},
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

func TestRunHandlerTests(t *testing.T) {

	lg.Logger = lg.NewDummyLogger()
	tcs := getTestRequests()

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			runHandlerTests(t, tc)
		})
	}
}

func runHandlerTests(t *testing.T, tc test_request) {

	w := httptest.NewRecorder()

	f := FakeSrvContext{}
	c := getSrvConfig()
	f.SetupServerContext(c)

	body, _ := json.Marshal(getTestBody(tc.caseId))

	rq, _ := http.NewRequest(tc.method, tc.path, bytes.NewBuffer(body))

	f.handler.HandleRequests(w, rq)

	if w.Code != tc.expectedCode {
		t.Errorf(">>>>FAIL: http status code mismatch: got %v, expecting %v", w.Code, tc.expectedCode)
	} else {
		t.Logf(">>>>PASS: http status code match: got %v, expecting %v", w.Code, tc.expectedCode)
	}
}

func getTestBody(id int) any {
	switch id {
	case 1:
		td := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
		td.Id = 5
		return td // no creation date ==> bad request
	case 3:
		td := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
		td.Id = 5
		td.CreationDate = time.Now()
		return td
	case 2:
		td := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
		td.Id = 8
		qry := []godoo.UserQueryOption{}
		qry = append(qry, godoo.UserQueryOption{Elem: godoo.ById})
		fq := godoo.FullUserQuery{QueryOptions: qry, QueryData: *td}
		return fq
	case 4:
		var retSch, retEdt []godoo.UserQueryOption
		retSch = append(retSch, godoo.UserQueryOption{Elem: godoo.ById})
		retEdt = append(retEdt, godoo.UserQueryOption{Elem: godoo.ById}, godoo.UserQueryOption{Elem: godoo.ByReplacement})

		toEdit := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
		toEdit.Id = 5
		toEdit.CreationDate = time.Now()

		newItm := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
		newItm.Body = "this is the new body"

		srchFq := godoo.FullUserQuery{QueryOptions: retSch, QueryData: *toEdit}
		edtFq := godoo.FullUserQuery{QueryOptions: retEdt, QueryData: *newItm}

		return []godoo.FullUserQuery{srchFq, edtFq}
	case 5:
		var retSch []godoo.UserQueryOption
		retSch = append(retSch, godoo.UserQueryOption{Elem: godoo.ById})

		toEdit := godoo.NewTodoItem(godoo.WithPriorityLevel(godoo.None))
		toEdit.Id = 5
		toEdit.CreationDate = time.Now()

		srchFq := godoo.FullUserQuery{QueryOptions: retSch, QueryData: *toEdit}
		return []godoo.FullUserQuery{srchFq}
	default:
		return false
	}
}
