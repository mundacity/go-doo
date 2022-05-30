package cli

import (
	"github.com/mundacity/go-doo/domain"
	"github.com/mundacity/go-doo/store"
)

func GetRepo(dbKind domain.DbType, connStr, dateLayout string) domain.IRepository {
	switch dbKind {
	case domain.Sqlite:
		return store.NewRepo(connStr, dbKind, dateLayout)
	}
	return nil
}
