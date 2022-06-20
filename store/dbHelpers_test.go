package store

import (
	"fmt"
	"testing"
	"time"

	"github.com/mundacity/go-doo/domain"
)

type update_data_assembling_test_case struct {
	sql      string
	srchOpts []domain.UserQuery
	edtOpts  []domain.UserQuery
	slctr    domain.TodoItem
	newData  domain.TodoItem
	expSql   string
	expVals  []any
	name     string
}

func getUpdateAssemblingTestCases() []update_data_assembling_test_case {

	return []update_data_assembling_test_case{{
		sql:      getSql(domain.Update, domain.Sqlite, items),
		srchOpts: []domain.UserQuery{{Elem: domain.ByDeadline, DateSetter: getResAndDate(true, "2022-01-18")}, {Elem: domain.ByCreationDate, DateSetter: getResAndDate(true, "2022-01-03")}, {Elem: domain.ByBody}},
		slctr:    domain.TodoItem{Deadline: parseDate("2022-01-10"), CreationDate: parseDate("2021-12-23"), Body: "z start"},
		edtOpts:  []domain.UserQuery{{Elem: domain.ByDeadline}, {Elem: domain.ByBody}, {Elem: domain.ByAppending}},
		newData:  domain.TodoItem{Deadline: parseDate("2022-02-02"), Body: " dud"},
		expSql:   "update items as i set deadline = ?, body = body || ? where deadline between ? and ? and creationDate between ? and ? and body like ?",
		expVals:  []any{"2022-02-02", " dud", "2022-01-10", "2022-01-18", "2021-12-23", "2022-01-03", "%z start%"},
		name:     "add one day to deadline by id",
	}, {
		sql:      getSql(domain.Update, domain.Sqlite, items),
		srchOpts: []domain.UserQuery{{Elem: domain.ById}},
		slctr:    domain.TodoItem{Id: 18},
		edtOpts:  []domain.UserQuery{{Elem: domain.ByDeadline}},
		newData:  domain.TodoItem{Deadline: parseDate("2022-06-02")},
		expSql:   "update items as i set deadline = ? where i.id = ?",
		expVals:  []any{"2022-06-02", 18},
		name:     "add one day to deadline by id",
	}, {
		sql:      getSql(domain.Update, domain.Sqlite, items),
		srchOpts: []domain.UserQuery{{Elem: domain.ByCreationDate}, {Elem: domain.ByDeadline, DateSetter: getResAndDate(true, "2022-06-22")}},
		slctr:    domain.TodoItem{Deadline: parseDate("2022-06-10"), CreationDate: parseDate("2022-06-01")},
		edtOpts:  []domain.UserQuery{{Elem: domain.ByCompletion}},
		newData:  domain.TodoItem{IsComplete: true},
		expSql:   "update items as i set isComplete = not isComplete where creationDate = ? and deadline between ? and ?",
		expVals:  []any{"2022-06-01", "2022-06-10", "2022-06-22"},
		name:     "toggle completion search on set creationDate and deadline range",
	}, {
		sql:      getSql(domain.Update, domain.Sqlite, items),
		srchOpts: []domain.UserQuery{{Elem: domain.ByCreationDate, DateSetter: getResAndDate(true, "2022-06-05")}, {Elem: domain.ByDeadline, DateSetter: getResAndDate(true, "2022-06-22")}},
		slctr:    domain.TodoItem{Deadline: parseDate("2022-06-10"), CreationDate: parseDate("2022-06-01")},
		edtOpts:  []domain.UserQuery{{Elem: domain.ByCompletion}},
		newData:  domain.TodoItem{IsComplete: true},
		expSql:   "update items as i set isComplete = not isComplete where creationDate between ? and ? and deadline between ? and ?",
		expVals:  []any{"2022-06-01", "2022-06-05", "2022-06-10", "2022-06-22"},
		name:     "toggle completion search on creationDate range and deadline range",
	}, {
		sql:      getSql(domain.Update, domain.Sqlite, items),
		srchOpts: []domain.UserQuery{{Elem: domain.ByCreationDate}, {Elem: domain.ByDeadline}},
		slctr:    domain.TodoItem{Deadline: parseDate("2022-06-10"), CreationDate: parseDate("2022-06-01")},
		edtOpts:  []domain.UserQuery{{Elem: domain.ByCompletion}},
		newData:  domain.TodoItem{IsComplete: true},
		expSql:   "update items as i set isComplete = not isComplete where creationDate = ? and deadline = ?",
		expVals:  []any{"2022-06-01", "2022-06-10"},
		name:     "toggle completion search set creationDate and set deadline",
	}, {
		sql:      getSql(domain.Update, domain.Sqlite, items),
		srchOpts: []domain.UserQuery{{Elem: domain.ByCreationDate}, {Elem: domain.ByDeadline}, {Elem: domain.ByBody}},
		slctr:    domain.TodoItem{Deadline: parseDate("2022-06-10"), CreationDate: parseDate("2022-06-01"), Body: "key phrase"},
		edtOpts:  []domain.UserQuery{{Elem: domain.ByBody}, {Elem: domain.ByAppending}},
		newData:  domain.TodoItem{Body: "new body"},
		expSql:   "update items as i set body = body || ? where creationDate = ? and deadline = ? and body like ?",
		expVals:  []any{"new body", "2022-06-01", "2022-06-10", "%key phrase%"},
		name:     "append to body search set creationDate set deadline and body",
	}, {
		sql:      getSql(domain.Update, domain.Sqlite, items),
		srchOpts: []domain.UserQuery{{Elem: domain.ByBody}, {Elem: domain.ByParentId}},
		slctr:    domain.TodoItem{Body: "key phrase", ParentId: 8},
		edtOpts:  []domain.UserQuery{{Elem: domain.ByBody}, {Elem: domain.ByAppending}, {Elem: domain.ByParentId}},
		newData:  domain.TodoItem{Body: "new body", ParentId: 12, IsChild: true},
		expSql:   "update items as i set body = body || ?, parentId = ? where body like ? and parentId = ?",
		expVals:  []any{"new body", 12, "%key phrase%", 8},
		name:     "append to body change parent search body and parent",
	}, {
		sql:      getSql(domain.Update, domain.Sqlite, items),
		srchOpts: []domain.UserQuery{{Elem: domain.ByBody}, {Elem: domain.ByParentId}},
		slctr:    domain.TodoItem{Body: "key phrase", ParentId: 8},
		edtOpts:  []domain.UserQuery{{Elem: domain.ByBody}, {Elem: domain.ByReplacement}, {Elem: domain.ByParentId}},
		newData:  domain.TodoItem{Body: "new body", ParentId: 12, IsChild: true},
		expSql:   "update items as i set body = ?, parentId = ? where body like ? and parentId = ?",
		expVals:  []any{"new body", 12, "%key phrase%", 8},
		name:     "replace body change parent search body and parent",
	}}
}

func parseDate(dStr string) time.Time {
	ret, _ := time.Parse("2006-01-02", dStr)
	return ret
}

func getResAndDate(yesNo bool, dStr string) domain.SetUpperDateBound {
	if yesNo {
		dt, _ := time.Parse("2006-01-02", dStr)
		return func() (bool, time.Time) {
			{
				return true, dt
			}
		}
	}
	return func() (bool, time.Time) {
		{
			return false, time.Now()
		}
	}
}

func getInMemDb() *Repo {
	return NewRepo("", domain.Sqlite, "2006-01-02")
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

	sql, pairs := r.assembleUpdateData(tc.sql, tc.srchOpts, tc.edtOpts, tc.slctr, tc.newData)

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
