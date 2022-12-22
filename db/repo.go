package db

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/db/sqlite"
	"github.com/mundacity/go-doo/util"
)

// Encapsulates everything that isn't specific to a given db vendor
//
// Vendor-specific logic etc located in separate subfolders

// Basic type to encapsulate the various IRepository methods
type Repo struct {
	db   *sql.DB
	dl   string
	kind godoo.DbType
	Port int
	Mtx  sync.Mutex
}

// Repo used by the app
var AppRepo Repo

// Enum to decribe which db Table/s to work with
type table int

const (
	items table = iota
	tags
	all
)

// Helps when scanning using sql.Rows.Scan
type temp_item struct {
	id           int
	parentId     int
	creationDate string
	deadline     string
	body         string
	isComplete   bool
	tag          string
	priority     int
}

// Field & value pairing to allow for composite where clauses
type where_map_entry struct {
	columnName string
	colValue   any
}

func SetupRepo(conn string, dbKind godoo.DbType, dateLayout string, port int) *Repo {
	Db := setup(conn)
	AppRepo = Repo{db: Db, dl: dateLayout, kind: dbKind, Port: port}
	return &AppRepo
}

func setup(path string) *sql.DB {
	var newDb bool
	if _, err := os.Stat(path); err != nil {
		newDb = true
		os.Create(path)
	}

	return sqlite.ReturnSqliteDb(path, newDb)
}

func (r *Repo) Add(itm *godoo.TodoItem) (int64, error) {
	var d string
	if itm.Deadline.IsZero() {
		d = ""
	} else {
		d = util.StringFromDate(itm.Deadline)
	}

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	sql := getSql(godoo.Add, r.kind, items)

	res, err := tx.Exec(sql, itm.ParentId, util.StringFromDate(itm.CreationDate), d, itm.Body, int(itm.Priority))
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

func (r *Repo) UpdateWhere(srchQry, edtQry godoo.FullUserQuery) (int, error) {

	itmSql := getSql(godoo.Edit, r.kind, items)
	//tagSql := "update tags set " // will need to do these separately

	itmSql, data := r.assembleUpdateData(itmSql, srchQry, edtQry)

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

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

func (r *Repo) GetWhere(qry godoo.FullUserQuery) ([]godoo.TodoItem, error) {

	if len(qry.QueryOptions) == 0 {
		return r.GetAll()
	}

	mp := make(map[int]*godoo.TodoItem)
	whereLst := getWhereList(qry)
	sql, vals := buildAndWhere(whereLst, getSql(godoo.Get, r.kind, all)+" where ")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	all, err := r.db.Query(sql, vals...)
	if err != nil {
		return nil, err
	}

	ret, err := r.processQuery(all, mp)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *Repo) GetAll() ([]godoo.TodoItem, error) {
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

func getWhereList(qry godoo.FullUserQuery) []where_map_entry {
	var lst []where_map_entry

	for _, opt := range qry.QueryOptions {
		if opt.Elem == godoo.ByAppending || opt.Elem == godoo.ByReplacement {
			// query modifiers; not query types/options
			continue
		}
		col, val := getColAndVal(opt, qry.QueryData)
		if col != "" {
			lst = append(lst, where_map_entry{col, val})
		}
	}
	return lst
}

func getColAndVal(q godoo.UserQueryOption, input godoo.TodoItem) (string, any) {
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
		return "priority", int(input.Priority) //probably just for updates
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

func getDateRange(q godoo.UserQueryOption, itm godoo.TodoItem) []string {
	var ret []string
	var d time.Time
	if q.Elem == godoo.ByDeadline {
		d = itm.Deadline
	}
	if q.Elem == godoo.ByCreationDate {
		d = itm.CreationDate
	}

	if q.UpperBoundDate.IsZero() {
		ret = append(ret, util.StringFromDate(d))
		return ret
	}

	ret = append(ret, util.StringFromDate(d), util.StringFromDate(q.UpperBoundDate))
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
