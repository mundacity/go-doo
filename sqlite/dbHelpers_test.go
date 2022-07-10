package sqlite

import (
	"fmt"
	"testing"
	"time"

	godoo "github.com/mundacity/go-doo"
)

type update_data_assembling_test_case struct {
	sql      string
	srchOpts []godoo.UserQueryOption
	edtOpts  []godoo.UserQueryOption
	slctr    godoo.TodoItem
	newData  godoo.TodoItem
	expSql   string
	expVals  []any
	name     string
}

func getUpdateAssemblingTestCases() []update_data_assembling_test_case {

	return []update_data_assembling_test_case{{
		sql:      getSql(godoo.Update, godoo.Sqlite, items),
		srchOpts: []godoo.UserQueryOption{{Elem: godoo.ByDeadline, UpperBoundDate: convertToUpperBound("2022-01-18")}, {Elem: godoo.ByCreationDate, UpperBoundDate: convertToUpperBound("2022-01-03")}, {Elem: godoo.ByBody}},
		slctr:    godoo.TodoItem{Deadline: parseDate("2022-01-10"), CreationDate: parseDate("2021-12-23"), Body: "z start"},
		edtOpts:  []godoo.UserQueryOption{{Elem: godoo.ByDeadline}, {Elem: godoo.ByBody}, {Elem: godoo.ByAppending}},
		newData:  godoo.TodoItem{Deadline: parseDate("2022-02-02"), Body: " dud"},
		expSql:   "update items as i set deadline = ?, body = body || ? where deadline between ? and ? and creationDate between ? and ? and body like ?",
		expVals:  []any{"2022-02-02", " dud", "2022-01-10", "2022-01-18", "2021-12-23", "2022-01-03", "%z start%"},
		name:     "add one day to deadline by id",
	}, {
		sql:      getSql(godoo.Update, godoo.Sqlite, items),
		srchOpts: []godoo.UserQueryOption{{Elem: godoo.ById}},
		slctr:    godoo.TodoItem{Id: 18},
		edtOpts:  []godoo.UserQueryOption{{Elem: godoo.ByDeadline}},
		newData:  godoo.TodoItem{Deadline: parseDate("2022-06-02")},
		expSql:   "update items as i set deadline = ? where i.id = ?",
		expVals:  []any{"2022-06-02", 18},
		name:     "add one day to deadline by id",
	}, {
		sql:      getSql(godoo.Update, godoo.Sqlite, items),
		srchOpts: []godoo.UserQueryOption{{Elem: godoo.ByCreationDate}, {Elem: godoo.ByDeadline, UpperBoundDate: convertToUpperBound("2022-06-22")}},
		slctr:    godoo.TodoItem{Deadline: parseDate("2022-06-10"), CreationDate: parseDate("2022-06-01")},
		edtOpts:  []godoo.UserQueryOption{{Elem: godoo.ByCompletion}},
		newData:  godoo.TodoItem{IsComplete: true},
		expSql:   "update items as i set isComplete = not isComplete where creationDate = ? and deadline between ? and ?",
		expVals:  []any{"2022-06-01", "2022-06-10", "2022-06-22"},
		name:     "toggle completion search on set creationDate and deadline range",
	}, {
		sql:      getSql(godoo.Update, godoo.Sqlite, items),
		srchOpts: []godoo.UserQueryOption{{Elem: godoo.ByCreationDate, UpperBoundDate: convertToUpperBound("2022-06-05")}, {Elem: godoo.ByDeadline, UpperBoundDate: convertToUpperBound("2022-06-22")}},
		slctr:    godoo.TodoItem{Deadline: parseDate("2022-06-10"), CreationDate: parseDate("2022-06-01")},
		edtOpts:  []godoo.UserQueryOption{{Elem: godoo.ByCompletion}},
		newData:  godoo.TodoItem{IsComplete: true},
		expSql:   "update items as i set isComplete = not isComplete where creationDate between ? and ? and deadline between ? and ?",
		expVals:  []any{"2022-06-01", "2022-06-05", "2022-06-10", "2022-06-22"},
		name:     "toggle completion search on creationDate range and deadline range",
	}, {
		sql:      getSql(godoo.Update, godoo.Sqlite, items),
		srchOpts: []godoo.UserQueryOption{{Elem: godoo.ByCreationDate}, {Elem: godoo.ByDeadline}},
		slctr:    godoo.TodoItem{Deadline: parseDate("2022-06-10"), CreationDate: parseDate("2022-06-01")},
		edtOpts:  []godoo.UserQueryOption{{Elem: godoo.ByCompletion}},
		newData:  godoo.TodoItem{IsComplete: true},
		expSql:   "update items as i set isComplete = not isComplete where creationDate = ? and deadline = ?",
		expVals:  []any{"2022-06-01", "2022-06-10"},
		name:     "toggle completion search set creationDate and set deadline",
	}, {
		sql:      getSql(godoo.Update, godoo.Sqlite, items),
		srchOpts: []godoo.UserQueryOption{{Elem: godoo.ByCreationDate}, {Elem: godoo.ByDeadline}, {Elem: godoo.ByBody}},
		slctr:    godoo.TodoItem{Deadline: parseDate("2022-06-10"), CreationDate: parseDate("2022-06-01"), Body: "key phrase"},
		edtOpts:  []godoo.UserQueryOption{{Elem: godoo.ByBody}, {Elem: godoo.ByAppending}},
		newData:  godoo.TodoItem{Body: "new body"},
		expSql:   "update items as i set body = body || ? where creationDate = ? and deadline = ? and body like ?",
		expVals:  []any{"new body", "2022-06-01", "2022-06-10", "%key phrase%"},
		name:     "append to body search set creationDate set deadline and body",
	}, {
		sql:      getSql(godoo.Update, godoo.Sqlite, items),
		srchOpts: []godoo.UserQueryOption{{Elem: godoo.ByBody}, {Elem: godoo.ByParentId}},
		slctr:    godoo.TodoItem{Body: "key phrase", ParentId: 8},
		edtOpts:  []godoo.UserQueryOption{{Elem: godoo.ByBody}, {Elem: godoo.ByAppending}, {Elem: godoo.ByParentId}},
		newData:  godoo.TodoItem{Body: "new body", ParentId: 12, IsChild: true},
		expSql:   "update items as i set body = body || ?, parentId = ? where body like ? and parentId = ?",
		expVals:  []any{"new body", 12, "%key phrase%", 8},
		name:     "append to body change parent search body and parent",
	}, {
		sql:      getSql(godoo.Update, godoo.Sqlite, items),
		srchOpts: []godoo.UserQueryOption{{Elem: godoo.ByBody}, {Elem: godoo.ByParentId}},
		slctr:    godoo.TodoItem{Body: "key phrase", ParentId: 8},
		edtOpts:  []godoo.UserQueryOption{{Elem: godoo.ByBody}, {Elem: godoo.ByReplacement}, {Elem: godoo.ByParentId}},
		newData:  godoo.TodoItem{Body: "new body", ParentId: 12, IsChild: true},
		expSql:   "update items as i set body = ?, parentId = ? where body like ? and parentId = ?",
		expVals:  []any{"new body", 12, "%key phrase%", 8},
		name:     "replace body change parent search body and parent",
	}}
}

func parseDate(dStr string) time.Time {
	ret, _ := time.Parse("2006-01-02", dStr)
	return ret
}

func convertToUpperBound(dStr string) time.Time {
	dt, _ := time.Parse("2006-01-02", dStr)
	return dt
}

func getInMemDb() *Repo {
	return SetupRepo("", godoo.Sqlite, "2006-01-02")
}

func TestUpdateAssembling(t *testing.T) {
	r := getInMemDb()
	tcs := getUpdateAssemblingTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			runAssembleTest(t, tc, r)
		})
	}
}

func runAssembleTest(t *testing.T, tc update_data_assembling_test_case, r *Repo) {

	srchFq := godoo.FullUserQuery{QueryOptions: tc.srchOpts, QueryData: tc.slctr}
	edtFq := godoo.FullUserQuery{QueryOptions: tc.edtOpts, QueryData: tc.newData}

	sql, pairs := r.assembleUpdateData(tc.sql, srchFq, edtFq)

	if sql == tc.expSql {
		t.Logf(">>>>PASSED: expected & got sql are equal")
	} else {
		t.Errorf(">>>>FAILED:\nExp: '%v'\nGot: '%v'\n", tc.expSql, sql)
	}

	theSame, msg := areTheSame(pairs, tc.expVals)

	if theSame {
		t.Logf(">>>>PASSED: %v", msg)
	} else {
		t.Errorf(">>>>FAILED: %v", msg)
	}
}

func areTheSame(got, exp []any) (bool, string) {

	if len(got) != len(exp) {
		return false, fmt.Sprintf("slices unequal length: 'got': %v and 'exp': %v", len(got), len(exp))
	}

	msg := "slices fully match"
	res := true

	for i, v := range got {
		if v != exp[i] {
			res = false
			if msg == "slices fully match" {
				msg = "items mismatched:"
			}
			msg += fmt.Sprintf(" item[%v]: 'got'{%v} vs 'exp'{%v},", i, v, exp[i])
		}
	}
	return res, msg
}
