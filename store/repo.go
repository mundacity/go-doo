package store

import (
	"context"
	"time"

	_ "github.com/mattn/go-sqlite3"
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/util"
)

func NewRepo(conn string, dbKind godoo.DbType, dateLayout string) *Repo {
	Db, _ := Init(conn, dbKind)
	r := Repo{db: Db, dl: dateLayout, kind: dbKind}
	return &r
}

func (r *Repo) Add(itm *godoo.TodoItem) (int64, error) {
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

	sql := getSql(godoo.Add, r.kind, items)

	res, err := tx.Exec(sql, itm.ParentId, util.StringFromDate(itm.CreationDate), d, itm.Body)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for t := range itm.Tags {
		sql := getSql(godoo.Add, r.kind, tags)
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

func (r *Repo) UpdateWhere(srchOptions, edtOptions []godoo.UserQuery, selector, newVals godoo.TodoItem) (int, error) {

	itmSql := getSql(godoo.Update, r.kind, items)
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

func (sr *Repo) GetWhere(options []godoo.UserQuery, input godoo.TodoItem) ([]godoo.TodoItem, error) {

	if len(options) == 0 {
		return sr.getAll()
	}

	mp := make(map[int]*godoo.TodoItem)
	whereLst := getWhereList(options, input)
	sql, vals := buildAndWhere(whereLst, getSql(godoo.Get, sr.kind, all)+" where ")

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

func (r *Repo) getAll() ([]godoo.TodoItem, error) {
	sql := getSql(godoo.Get, r.kind, all)
	mp := make(map[int]*godoo.TodoItem)

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

func getWhereList(options []godoo.UserQuery, input godoo.TodoItem) []where_map_entry {
	var lst []where_map_entry

	for _, opt := range options {
		if opt.Elem == godoo.ByAppending || opt.Elem == godoo.ByReplacement {
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

func getColAndVal(q godoo.UserQuery, input godoo.TodoItem) (string, any) {
	switch q.Elem {
	case godoo.ById:
		return "i.id", input.Id
	case godoo.ByChildId:
		return "", input.ChildItems
	case godoo.ByParentId:
		return "parentId", input.ParentId
	case godoo.ByTag:
		return "tag", getTagFromMap(input.Tags)
	case godoo.ByBody:
		return "body", input.Body
	case godoo.ByNextPriority:
		return "", nil // would follow a GetAll()
	case godoo.ByNextDate:
		return "", nil // same
	case godoo.ByDeadline:
		return "deadline", getDateRange(q, input)
	case godoo.ByCreationDate:
		return "creationDate", getDateRange(q, input)
	case godoo.ByCompletion:
		return "isComplete", input.IsComplete
	}
	return "", nil
}

func getDateRange(q godoo.UserQuery, itm godoo.TodoItem) []string {
	var ret []string
	var d time.Time
	if q.Elem == godoo.ByDeadline {
		d = itm.Deadline
	}
	if q.Elem == godoo.ByCreationDate {
		d = itm.CreationDate
	}

	if q.DateSetter == nil {
		ret = append(ret, util.StringFromDate(d))
		return ret
	}
	ok, lower := q.DateSetter()
	if !ok {
		ret = append(ret, util.StringFromDate(d))
		return ret
	}

	ret = append(ret, util.StringFromDate(d), util.StringFromDate(lower))
	return ret
}

func getTagFromMap(mp map[string]struct{}) string {
	var ret string
	for v := range mp {
		ret = v
		break // this is from terminal input so will only be one item
	}
	return ret
}
