package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"sync"
)

type Repository struct {
	db *sqlx.DB

	userCache      map[string]string
	userCacheMutex sync.RWMutex
}

func NewRepository(db *sqlx.DB) Repository {
	return Repository{
		db:        db,
		userCache: make(map[string]string),
	}
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
