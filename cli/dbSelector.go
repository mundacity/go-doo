package cli

import (
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/store"
)

func GetRepo(dbKind godoo.DbType, connStr, dateLayout string) godoo.IRepository {
	switch dbKind {
	case godoo.Sqlite:
		return store.NewRepo(connStr, dbKind, dateLayout)
	}
	return nil
}
