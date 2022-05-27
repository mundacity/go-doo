package services

import (
	"github.com/mundacity/go-doo/domain"
	"github.com/mundacity/go-doo/store"
)

func GetRepo(dbKind store.DbType, connStr, dateLayout string) domain.IRepository {
	switch dbKind {
	case store.Sqlite:
		return store.NewRepo(connStr, dbKind, dateLayout)
	}
	return nil
}
