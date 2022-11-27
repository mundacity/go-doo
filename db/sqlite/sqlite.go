package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

// Vendor-specific SQL

func ReturnSqliteDb(path string, isNewDb bool) *sql.DB {

	ret, _ := sql.Open("sqlite3", path)
	if isNewDb {
		tx, _ := ret.BeginTx(context.Background(), nil)
		tx.Exec("CREATE TABLE items (id integer primary key autoincrement, " +
			"parentId integer, " +
			"creationDate text not null, " +
			"deadline text not null, " +
			"body text not null, " +
			"isComplete boolean default false not null, " +
			"priority integer default 0 not null);")
		tx.Exec("CREATE TABLE tags (id integer primary key autoincrement, " +
			"itemId integer, " +
			"tag text not null);")
		tx.Commit()
	}
	return ret
}

func GetInsert(tbl int) string {
	if tbl == 0 {
		return "insert into items (parentId, creationDate, deadline, body, priority) values (?, ?, ?, ?, ?)"
	} else if tbl == 1 {
		return "INSERT INTO tags (itemId, tag) VALUES (?, ?)"
	}
	return ""
}

func GetSelect(tbl int) string {
	// table doesn't matter atm
	return "select i.id, parentId, creationDate, deadline, body, isComplete, ifnull(tag, '') tag, priority " +
		"from items i left join tags t " +
		"on i.id = t.itemId"
}

func GetUpdate(tbl int) string {
	switch tbl {
	case 0:
		return "update items as i set "
	case 1:
		return "update tags set "
	case 2:
		return "update items i inner join tags t on i.id = t.itemId set "
	}
	return ""
}

func GetBodyAndWhereVal(existingVal any) any {
	return fmt.Sprintf("%%%v%%", existingVal)
}

func BuildBodySqlStr(andStr, colName string) string {
	return fmt.Sprintf("%v%v like ?", andStr, colName)
}

func GenerateUpdatePairs(colName, comma, sqlBase string, colVal any, appending, replacing bool) (string, []any) {

	var vals []any

	if colName == "body" {
		if appending {
			sqlBase += fmt.Sprintf("%v = %v || ?%v", colName, colName, comma)
		} else if replacing {
			sqlBase += fmt.Sprintf("%v = ?%v", colName, comma)
		}
		vals = append(vals, colVal)
		return sqlBase, vals
	}
	if colName == "isComplete" {
		sqlBase += fmt.Sprintf("%v = not %v%v", colName, colName, comma)
		return sqlBase, vals
	}
	//only ever 1 value for an update
	if colName == "creationDate" || colName == "deadline" {
		vs := colVal.([]string)
		colVal = vs[0]
	}

	sqlBase += fmt.Sprintf("%v = ?%v", colName, comma)
	vals = append(vals, colVal)

	return sqlBase, vals
}

func GenerateWhereClause(colName, sqlBase, andStr string, colVal any) (string, []any) {

	vals := make([]any, 1)

	if colName == "body" {
		colVal = fmt.Sprintf("%%%v%%", colVal) //like '%fishy%'
		sqlBase += fmt.Sprintf("%v%v like ?", andStr, colName)
		vals[0] = colVal

		return sqlBase, vals
	}
	if colName == "creationDate" || colName == "deadline" {

		vs := colVal.([]string)
		if len(vs) > 1 { //it's a range search
			sqlBase += fmt.Sprintf("%v%v between ? and ?", andStr, colName)
			vals[0] = vs[0]
			vals = append(vals, vs[1])

			return sqlBase, vals
		} else {
			colVal = vs[0]
		}
	}
	sqlBase += fmt.Sprintf("%v%v = ?", andStr, colName)
	vals[0] = colVal

	return sqlBase, vals

}
