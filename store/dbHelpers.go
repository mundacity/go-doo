package store

import (
	"database/sql"
	"fmt"
	"time"

	godoo "github.com/mundacity/go-doo"
)

type table int

const (
	items table = iota
	tags
	all
)

// Encapsulates the various IRepository methods
// for a sqlite database
type Repo struct {
	db   *sql.DB
	dl   string
	kind godoo.DbType
}

// Helps when scanning
type temp_item struct {
	id           int
	parentId     int
	creationDate string
	deadline     string
	body         string
	isComplete   bool
	tag          string
}

// Field & value pairing to allow for composite where clauses
type where_map_entry struct {
	columnName string
	colValue   any
}

// Get TodoItem from temp_item
func (r *Repo) tempConversion(tmp temp_item) godoo.TodoItem {
	var ret godoo.TodoItem
	ret.Tags = make(map[string]struct{})

	ret.Id = tmp.id
	ret.ParentId = tmp.parentId
	ret.CreationDate, _ = time.Parse(r.dl, tmp.creationDate)
	ret.Deadline, _ = time.Parse(r.dl, tmp.deadline)
	ret.Body = tmp.body
	ret.IsComplete = tmp.isComplete
	ret.Tags[tmp.tag] = struct{}{}

	return ret
}

func getSql(qType godoo.QueryType, dbKind godoo.DbType, tbl table) string {
	switch qType {
	case godoo.Add:
		return getInsertSql(dbKind, tbl)
	case godoo.Get:
		return getSelectSql(dbKind, tbl)
	case godoo.Update:
		return getBaseUpdateSql(dbKind, tbl)
	default:
		return ""
	}
}

func getInsertSql(db godoo.DbType, tbl table) string {
	switch db {
	case godoo.Sqlite:
		if tbl == items {
			return "insert into items (parentId, creationDate, deadline, body) values (?, ?, ?, ?)"
		} else if tbl == tags {
			return "INSERT INTO tags (itemId, tag) VALUES (?, ?)"
		}
	}
	return ""
}

func getSelectSql(db godoo.DbType, tbl table) string {
	// table doesn't matter atm
	switch db {
	case godoo.Sqlite:
		return "select i.id, parentId, creationDate, deadline, body, isComplete, ifnull(tag, '') tag " +
			"from items i left join tags t " +
			"on i.id = t.itemId"
	}
	return ""
}

func getBaseUpdateSql(db godoo.DbType, tbl table) string {
	switch tbl {
	case items:
		return "update items as i set "
	case tags:
		return "update tags set "
	case all:
		return "update items i inner join tags t on i.id = t.itemId set "
	}
	return ""
}

func (sr *Repo) processQuery(all *sql.Rows, mp map[int]*godoo.TodoItem) ([]godoo.TodoItem, error) {
	var ret []godoo.TodoItem

	defer all.Close()
	for all.Next() {
		// read row into temp item
		var itm temp_item
		if err := all.Scan(&itm.id, &itm.parentId, &itm.creationDate, &itm.deadline, &itm.body, &itm.isComplete, &itm.tag); err != nil {
			return nil, err
		}

		// handle tags - only want one item per i.id number
		conv := sr.tempConversion(itm)
		td, exists := mp[conv.Id]
		if exists {
			td.Tags[itm.tag] = struct{}{}
		} else {
			mp[conv.Id] = &conv
		}

	}
	if err := all.Err(); err != nil {
		return nil, err
	}

	// convert to slice
	for _, v := range mp {
		ret = append(ret, *v)
	}
	return ret, nil
}

func (r *Repo) assembleUpdateData(sql string,
	srchOptions, edtOptions []godoo.UserQuery,
	selector, newVals godoo.TodoItem) (string, []any) {

	updateLst := getWhereList(edtOptions, newVals)  // to generate 'a-h' in 'update items set a=b, c=d, e=f, g=h where x'
	whereLst := getWhereList(srchOptions, selector) // to generate 'x' in above

	sql, pairs := buildUpdatePairs(updateLst, sql, edtOptions)
	sql, vals := buildAndWhere(whereLst, sql+"where ")

	pairs = append(pairs, vals...)
	return sql, pairs
}

func buildAndWhere(input []where_map_entry, sqlBase string) (string, []any) {
	andStr := ""
	vals := make([]any, len(input))

	offset := 0
	for i, w := range input { //like '%fishy%'
		if i > 0 {
			andStr = " and "
		}

		if w.columnName == "body" {
			w.colValue = fmt.Sprintf("%%%v%%", w.colValue)
			sqlBase += fmt.Sprintf("%v%v like ?", andStr, w.columnName)
			vals[i+offset] = w.colValue
			continue
		}
		if w.columnName == "creationDate" || w.columnName == "deadline" {

			vs := w.colValue.([]string)
			if len(vs) > 1 { //it's a range search
				sqlBase += fmt.Sprintf("%v%v between ? and ?", andStr, w.columnName)
				vals[i+offset] = vs[0]
				offset++
				vals = append(vals, nil)
				vals[i+offset] = vs[1]
				continue
			} else {
				w.colValue = vs[0]
			}
		}
		sqlBase += fmt.Sprintf("%v%v = ?", andStr, w.columnName)
		vals[i+offset] = w.colValue
	}
	return sqlBase, vals
}

func buildUpdatePairs(input []where_map_entry, sqlBase string, options []godoo.UserQuery) (string, []any) {
	var appending, replacing bool

	for _, o := range options {
		if o.Elem == godoo.ByAppending {
			appending = true
			break
		}
		if o.Elem == godoo.ByReplacement {
			replacing = true
		}
	}

	comma := ", "
	var vals []any
	for i, itm := range input {
		if i == len(input)-1 {
			comma = " "
		}

		if itm.columnName == "body" {
			if appending {
				sqlBase += fmt.Sprintf("%v = %v || ?%v", itm.columnName, itm.columnName, comma)
			} else if replacing {
				sqlBase += fmt.Sprintf("%v = ?%v", itm.columnName, comma)
			}
			vals = append(vals, itm.colValue)
			continue
		}
		if itm.columnName == "isComplete" {
			sqlBase += fmt.Sprintf("%v = not %v%v", itm.columnName, itm.columnName, comma)
			continue
		}
		//only ever 1 value for an update
		if itm.columnName == "creationDate" || itm.columnName == "deadline" {
			vs := itm.colValue.([]string)
			itm.colValue = vs[0]
		}

		sqlBase += fmt.Sprintf("%v = ?%v", itm.columnName, comma)
		vals = append(vals, itm.colValue)
	}
	return sqlBase, vals
}
