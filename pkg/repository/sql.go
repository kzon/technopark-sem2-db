package repository

import (
	"database/sql"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
)

func Error(err error) error {
	switch err {
	case sql.ErrNoRows:
		return consts.ErrNotFound
	default:
		return err
	}
}
