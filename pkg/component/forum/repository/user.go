package repository

import (
	"database/sql"
	"fmt"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/repository"
	"strings"
)

func (r *Repository) GetUserByID(id int) (*model.User, error) {
	user, err := r.getUserFields("*", "id=$1", id)
	if err == sql.ErrNoRows {
		return nil, consts.ErrNotFound
	}
	return user, err
}

func (r *Repository) GetUserByNickname(nickname string) (*model.User, error) {
	if _, ok := r.getCachedUser(nickname); !ok {
		return nil, consts.ErrNotFound
	}
	user, err := r.getUserFields("*", "nickname=$1", nickname)
	if err != nil {
		return nil, repository.Error(err)
	}
	return user, nil
}

func (r *Repository) getUserIDByNickname(nickname string) (int, error) {
	user, err := r.getUserFields("id", "nickname=$1", nickname)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (r *Repository) getUserFields(fields, filter string, params ...interface{}) (*model.User, error) {
	query := "select " + fields + ` from "user"`
	if filter != "" {
		query += " where " + filter
	}
	var user model.User
	err := r.db.Get(&user, query, params...)
	return &user, err
}

func (r *Repository) getCachedUser(nickname string) (string, bool) {
	r.userCacheMutex.RLock()
	userNickname, ok := r.userCache[strings.ToLower(nickname)]
	r.userCacheMutex.RUnlock()
	return userNickname, ok
}

func (r *Repository) cacheUser(nickname string) {
	r.userCacheMutex.Lock()
	r.userCache[strings.ToLower(nickname)] = nickname
	r.userCacheMutex.Unlock()
}

//GetUserNickname returns real user nickname (case insensitive search)
func (r *Repository) GetUserNickname(nickname string) (string, error) {
	if nickname, ok := r.getCachedUser(nickname); ok {
		return nickname, nil
	}
	user, err := r.getUserFields("nickname", "nickname=$1", nickname)
	if err != nil {
		return "", repository.Error(err)
	}
	return user.Nickname, nil
}

func (r *Repository) getUserByEmail(email string) (*model.User, error) {
	user, err := r.getUserFields("*", "email=$1", email)
	if err != nil {
		return nil, repository.Error(err)
	}
	return user, nil
}

func (r *Repository) GetUsersByNicknameOrEmail(nickname, email string) ([]*model.User, error) {
	var users []*model.User
	err := r.db.Select(&users,
		`select * from "user" where nickname = $1 or email = $2`,
		nickname, email,
	)
	if err != nil {
		return nil, repository.Error(err)
	}
	return users, nil
}

func (r *Repository) CreateUser(nickname, email, fullname, about string) (*model.User, error) {
	var id int
	err := r.db.QueryRow(
		`insert into "user" (nickname, email, fullname, about) values ($1, $2, $3, $4) returning id`,
		nickname, email, fullname, about,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	r.cacheUser(nickname)
	return r.GetUserByID(id)
}

func (r *Repository) UpdateUserByNickname(nickname, email, fullname, about string) error {
	userByEmail, err := r.getUserByEmail(email)
	if err != nil && err != consts.ErrNotFound {
		return err
	}
	if userByEmail != nil && userByEmail.Nickname != nickname {
		return fmt.Errorf("%w: user with this email already exists", consts.ErrConflict)
	}
	result, err := r.db.Exec(
		`update "user" set email=$1, fullname=$2, about=$3 where nickname=$4`,
		email, fullname, about, nickname,
	)
	if err != nil {
		return repository.Error(err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return consts.ErrNotFound
	}
	return nil
}
