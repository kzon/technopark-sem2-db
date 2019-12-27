package repository

import (
	"database/sql"
	"fmt"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/repository"
)

func (r *Repository) GetUserByNickname(nickname string) (*model.User, error) {
	user := model.User{}
	err := r.db.Get(&user, `select * from "user" where nickname = $1`, nickname)
	if err != nil {
		return nil, repository.Error(err)
	}
	return &user, nil
}

func (r *Repository) GetUserNickname(nickname string) (string, error) {
	userNick, err := r.users.GetNickCaseInsensitive(nickname)
	if err == nil {
		return userNick, nil
	}
	user := model.User{}
	err = r.db.Get(&user, `select id,nickname from "user" where nickname = $1`, nickname)
	if err != nil {
		return "", repository.Error(err)
	}
	r.users.Add(user.ID, user.Nickname)
	return user.Nickname, nil
}

func (r *Repository) getUserByID(userID int) (*model.User, error) {
	user := model.User{}
	err := r.db.Get(&user, `select * from "user" where id = $1`, userID)
	if err == sql.ErrNoRows {
		return nil, consts.ErrNotFound
	}
	return &user, err
}

func (r *Repository) getUserByEmail(email string) (*model.User, error) {
	user := model.User{}
	err := r.db.Get(&user, `select * from "user" where email = $1`, email)
	if err != nil {
		return nil, repository.Error(err)
	}
	return &user, nil
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
	r.users.Add(id, nickname)
	return r.getUserByID(id)
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
