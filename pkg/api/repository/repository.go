package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/api/repository/cache"
)

type Repository struct {
	db    *sqlx.DB
	users cache.UserCache
}

func NewRepository(db *sqlx.DB) Repository {
	return Repository{db: db, users: cache.NewUserCache()}
}

func (r *Repository) getOrder(desc bool) string {
	if desc {
		return " desc"
	}
	return ""
}
func (r *Repository) getLimit(limit int) string {
	if limit > 0 {
		return fmt.Sprintf(" limit %d", limit)
	}
	return ""
}

func (r *Repository) getSinceOperator(desc bool) string {
	if desc {
		return "<"
	}
	return ">"
}
