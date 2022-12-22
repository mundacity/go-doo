package db

import (
	"database/sql"
	"time"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/db/sqlite"
)

func getSql(qType godoo.QueryType, dbKind godoo.DbType, tbl table) string {
	switch qType {
	case godoo.Add:
		return getInsertSql(dbKind, tbl)
	case godoo.Get:
		return getSelectSql(dbKind, tbl)
	case godoo.Edit:
		return getBaseUpdateSql(dbKind, tbl)
	default:
		return ""
	}
}

func getInsertSql(db godoo.DbType, tbl table) string {
	switch db {
	case godoo.Sqlite:
		return sqlite.GetInsert(int(tbl))
	}
	return ""
}

func getSelectSql(db godoo.DbType, tbl table) string {
	// table doesn't matter atm
	switch db {
	case godoo.Sqlite:
		return sqlite.GetSelect(int(tbl))
	}
	return ""
}

func getBaseUpdateSql(db godoo.DbType, tbl table) string {
	switch db {
	case godoo.Sqlite:
		return sqlite.GetUpdate(int(tbl))
	}
	return ""
}

func (sr *Repo) processQuery(all *sql.Rows, mp map[int]*godoo.TodoItem) ([]godoo.TodoItem, error) {
	var ret []godoo.TodoItem

	defer all.Close()
	for all.Next() {
		// read row into temp item
		var itm temp_item
		if err := all.Scan(&itm.id, &itm.parentId, &itm.creationDate, &itm.deadline, &itm.body, &itm.isComplete, &itm.tag, &itm.priority); err != nil {
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
	ret.Priority = godoo.PriorityLevel(tmp.priority)

	return ret
}

func (r *Repo) assembleUpdateData(sql string, srchQry, edtQry godoo.FullUserQuery) (string, []any) {

	updateLst := getWhereList(edtQry) // to generate 'a-h' in 'update items set a=b, c=d, e=f, g=h where x'
	whereLst := getWhereList(srchQry) // to generate 'x' in above

	sql, pairs := r.buildUpdatePairs(updateLst, sql, edtQry)
	sql, vals := buildAndWhere(whereLst, sql+"where ")

	pairs = append(pairs, vals...)
	return sql, pairs
}

func (r *Repo) buildUpdatePairs(input []where_map_entry, sqlBase string, qry godoo.FullUserQuery) (string, []any) {
	var appending, replacing bool

	for _, o := range qry.QueryOptions {
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

		s, v := sqlite.GenerateUpdatePairs(itm.columnName, comma, sqlBase, itm.colValue, appending, replacing)
		sqlBase = s
		vals = append(vals, v...)
	}
	return sqlBase, vals
}

func buildAndWhere(input []where_map_entry, sqlBase string) (string, []any) {
	andStr := ""
	vals := make([]any, 1)

	for i, w := range input {
		if i > 0 {
			andStr = " and "
		}

		s, v := sqlite.GenerateWhereClause(w.columnName, sqlBase, andStr, w.colValue)
		sqlBase = s
		vals = append(vals, v...)
	}

	return sqlBase, vals[1:]
}
