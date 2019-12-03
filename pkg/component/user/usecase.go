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

func (u *Usecase) getUserByNickname(nickname string) (*userOutput, error) {
	user, err := u.repo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}
	output := u.outputUser(user)
	return &output, nil
}

func (u *Usecase) createUser(nickname, email, fullname, about string) ([]userOutput, error) {
	existing, err := u.repo.getUsersByNicknameOrEmail(nickname, email)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return u.outputUsers(existing), consts.ErrConflict
	}
	err = u.repo.createUser(nickname, email, fullname, about)
	if err != nil {
		return nil, err
	}
	return []userOutput{{
		Email:    email,
		Fullname: fullname,
		About:    about,
		Nickname: nickname,
	}}, nil
}

func (u *Usecase) updateUser(nickname, email, fullname, about string) (*userOutput, error) {
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
	updatedUser, err := u.repo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}
	result := u.outputUser(updatedUser)
	return &result, nil
}

func (u *Usecase) outputUsers(users []*model.User) []userOutput {
	usersToOutput := make([]userOutput, 0, len(users))
	for _, user := range users {
		usersToOutput = append(usersToOutput, u.outputUser(user))
	}
	return usersToOutput
}

func (u *Usecase) outputUser(user *model.User) userOutput {
	return userOutput{
		Email:    user.Email,
		Fullname: user.Fullname,
		About:    user.About,
		Nickname: user.Nickname,
	}
}
