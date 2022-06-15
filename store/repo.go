package store

import (
	"context"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mundacity/go-doo/domain"
	"github.com/mundacity/go-doo/util"
)

func NewRepo(conn string, dbKind domain.DbType, dateLayout string) *Repo {
	Db, _ := Init(conn, dbKind)
	r := Repo{db: Db, dl: dateLayout, kind: dbKind}
	return &r
}

func (r *Repo) Add(itm *domain.TodoItem) (int64, error) {
	var d string
	if itm.Deadline.IsZero() {
		d = ""
	} else {
		d = util.StringFromDate(itm.Deadline)
	}

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	sql := getSql(domain.Add, r.kind, items)

	res, err := tx.Exec(sql, itm.ParentId, util.StringFromDate(itm.CreationDate), d, itm.Body)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for t := range itm.Tags {
		sql := getSql(domain.Add, r.kind, tags)
		_, err := tx.Exec(sql, id, t)
		if err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repo) UpdateWhere(srchOptions, edtOptions []domain.UserQuery, selector, newVals domain.TodoItem) (int, error) {

	itmSql := getSql(domain.Update, r.kind, items)
	//tagSql := "update tags set " // will need to do these separately

	itmSql, data := r.assembleUpdateData(itmSql, srchOptions, edtOptions, selector, newVals)

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(itmSql, data...)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rows), nil
}

func (sr *Repo) GetWhere(options []domain.UserQuery, input domain.TodoItem) ([]domain.TodoItem, error) {

	if len(options) == 0 {
		return sr.getAll()
	}

	mp := make(map[int]*domain.TodoItem)
	whereLst := getWhereList(options, input)
	sql, vals := buildAndWhere(whereLst, getSql(domain.Get, sr.kind, all)+" where ")

	all, err := sr.db.Query(sql, vals...)
	if err != nil {
		return nil, err
	}

	ret, err := sr.processQuery(all, mp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *Repo) getAll() ([]domain.TodoItem, error) {
	sql := getSql(domain.Get, r.kind, all)
	mp := make(map[int]*domain.TodoItem)

	all, err := r.db.Query(sql)
	if err != nil {
		return nil, err
	}

	ret, err := r.processQuery(all, mp)
	if err != nil {
		return nil, err
	}
	return ret, nil

}

func getWhereList(options []domain.UserQuery, input domain.TodoItem) []where_map_entry {
	var lst []where_map_entry

	for _, opt := range options {
		if opt.Elem == domain.ByAppending || opt.Elem == domain.ByReplacement {
			// query modifiers; not query types/options
			continue
		}
		col, val := getColAndVal(opt, input)
		if col != "" {
			lst = append(lst, where_map_entry{col, val})
		}
	}
	return lst
}

func getColAndVal(q domain.UserQuery, input domain.TodoItem) (string, any) {
	switch q.Elem {
	case domain.ById:
		return "i.id", input.Id
	case domain.ByChildId:
		return "", input.ChildItems
	case domain.ByParentId:
		return "parentId", input.ParentId
	case domain.ByTag:
		return "tag", getTagFromMap(input.Tags)
	case domain.ByBody:
		return "body", input.Body
	case domain.ByNextPriority:
		return "", nil // would follow a GetAll()
	case domain.ByNextDate:
		return "", nil // same
	case domain.ByDeadline:
		return "deadline", getDateString(q, input)
	case domain.ByCreationDate:
		return "creationDate", util.StringFromDate(input.CreationDate)
	case domain.ByCompletion:
		return "isComplete", input.IsComplete
	}
	return "", nil
}

func getDateString(q domain.UserQuery, itm domain.TodoItem) string {

	var d time.Time
	if q.Elem == domain.ByDeadline {
		d = itm.Deadline
	}
	if q.Elem == domain.ByCreationDate {
		d = itm.CreationDate
	}

	ok, lower := q.DateSetter()
	if !ok {
		return util.StringFromDate(d)
	}

	return fmt.Sprintf("between '%v' and '%v'", d, util.StringFromDate(lower))
}

func getTagFromMap(mp map[string]struct{}) string {
	var ret string
	for v := range mp {
		ret = v
		break // this is from terminal input so will only be one item
	}
	return ret
}
