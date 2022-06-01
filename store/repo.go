package store

import (
	"context"

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

func (r *Repo) GetAll() ([]domain.TodoItem, error) {

	mp := make(map[int]*domain.TodoItem)

	sql := getSql(domain.Get, r.kind, all)

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

func (r *Repo) UpdateWhere(srchOptions, edtOptions []domain.UserQueryElement, selector, newVals domain.TodoItem) (int, error) {

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

func (r *Repo) GetById(id int) (domain.TodoItem, error) {
	var dud domain.TodoItem
	mp := make(map[int]*domain.TodoItem)
	sql := getSql(domain.Get, r.kind, all) + " where i.id = ?" //  TODO: move this... better would be to get rid of id-specific method & just use GetWhere()

	all, err := r.db.Query(sql, id)
	if err != nil {
		return dud, err
	}

	lst, err := r.processQuery(all, mp)
	if err != nil {
		return dud, err
	}

	return lst[0], nil
}

func (sr *Repo) GetWhere(options []domain.UserQueryElement, input domain.TodoItem) ([]domain.TodoItem, error) {

	// len(options) always > 0
	// 	--> domain.ByCompletion added in GetCommand if otherwise empty

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

func getWhereList(options []domain.UserQueryElement, input domain.TodoItem) []where_map_entry {
	var lst []where_map_entry

	for _, opt := range options {
		if opt == domain.ByAppending || opt == domain.ByReplacement {
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

func getColAndVal(q domain.UserQueryElement, input domain.TodoItem) (string, any) {
	switch q {
	case domain.ById:
		return "id", input.Id
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
		return "deadline", util.StringFromDate(input.Deadline)
	case domain.ByCreationDate:
		return "creationDate", util.StringFromDate(input.CreationDate)
	case domain.ByCompletion:
		return "isComplete", input.IsComplete
	}
	return "", nil
}

func getTagFromMap(mp map[string]struct{}) string {
	var ret string
	for v := range mp {
		ret = v
		break // this is from terminal input so will only be one item
	}
	return ret
}
