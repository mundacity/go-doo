package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

func returnSqliteDb(path string, isNewDb bool) *sql.DB {

	ret, _ := sql.Open("sqlite3", path)
	if isNewDb {
		tx, _ := ret.BeginTx(context.Background(), nil)
		tx.Exec("CREATE TABLE items (id integer primary key autoincrement, parentId integer, creationDate text not null, deadline text not null, body text not null, isComplete boolean default false not null);")
		tx.Exec("CREATE TABLE tags (id integer primary key autoincrement, itemId integer, tag text not null);")
		tx.Commit()
	}
	return ret
}

func GetInsert(tbl int) string {
	if tbl == 0 {
		return "insert into items (parentId, creationDate, deadline, body) values (?, ?, ?, ?)"
	} else if tbl == 1 {
		return "INSERT INTO tags (itemId, tag) VALUES (?, ?)"
	}
	return ""
}

func GetSelect(tbl int) string {
	// table doesn't matter atm
	return "select i.id, parentId, creationDate, deadline, body, isComplete, ifnull(tag, '') tag " +
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
