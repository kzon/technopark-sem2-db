package user

import (
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
)

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return Usecase{repo: repo}
}

func (u *Usecase) getUserByNickname(nickname string) (*model.User, error) {
	return u.repo.GetUserByNickname(nickname)
}

func (u *Usecase) createUser(nickname, email, fullname, about string) ([]*model.User, error) {
	existing, err := u.repo.getUsersByNicknameOrEmail(nickname, email)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return existing, consts.ErrConflict
	}
	user, err := u.repo.createUser(nickname, email, fullname, about)
	return []*model.User{user}, err
}

func (u *Usecase) updateUser(nickname, email, fullname, about string) (*model.User, error) {
	userToUpdate, err := u.repo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}
	if email == "" {
		email = userToUpdate.Email
	}
	if fullname == "" {
		fullname = userToUpdate.Fullname
	}
	if about == "" {
		about = userToUpdate.About
	}
	if err := u.repo.updateUserByNickname(nickname, email, fullname, about); err != nil {
		return nil, err
	}
	return u.repo.GetUserByNickname(nickname)
}
