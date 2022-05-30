package store

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/mundacity/go-doo/domain"
)

func Init(path string, dbKind domain.DbType) (*sql.DB, error) {

	var newDb bool
	if _, err := os.Stat(path); err != nil {
		newDb = true
		os.Create(path)
	}

	switch dbKind {
	case domain.Sqlite:
		return returnSqliteDb(path, newDb), nil
	}

	return nil, errors.New("db initialisation error")
}

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
