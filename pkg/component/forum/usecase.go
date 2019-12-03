package forum

import (
	"fmt"
	userComponent "github.com/kzon/technopark-sem2-db/pkg/component/user"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"time"
)

type Usecase struct {
	forumRepo Repository
	userRepo  userComponent.Repository
}

func NewUsecase(forumRepo Repository, userRepo userComponent.Repository) Usecase {
	return Usecase{forumRepo: forumRepo, userRepo: userRepo}
}

func (u *Usecase) createForum(title, slug, nickname string) (*forumOutput, error) {
	user, err := u.userRepo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}

	existingForum, err := u.forumRepo.getForum(slug)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existingForum != nil {
		result, err := u.outputForum(existingForum)
		if err != nil {
			return nil, err
		}
		return result, fmt.Errorf("%w: forum with this slug already exists", consts.ErrConflict)
	}

	err = u.forumRepo.createForum(title, slug, user.Nickname)
	if err != nil {
		return nil, err
	}
	return &forumOutput{
		Title:   title,
		User:    user.Nickname,
		Slug:    slug,
		Posts:   0,
		Threads: 0,
	}, nil
}

func (u *Usecase) getForum(slug string) (*forumOutput, error) {
	forum, err := u.forumRepo.getForum(slug)
	if err != nil {
		return nil, err
	}
	return u.outputForum(forum)
}

func (u *Usecase) outputForum(forum *model.Forum) (*forumOutput, error) {
	if forum == nil {
		return nil, nil
	}
	return &forumOutput{
		Title:   forum.Title,
		User:    forum.User,
		Slug:    forum.Slug,
		Posts:   forum.Posts,
		Threads: forum.Threads,
	}, nil
}

func (u *Usecase) createThread(forum string, thread threadCreate) (*threadOutput, error) {
	if _, err := u.userRepo.GetUserByNickname(thread.Author); err != nil {
		return nil, err
	}
	if _, err := u.forumRepo.getForum(forum); err != nil {
		return nil, err
	}

	if thread.Slug != "" {
		existing, err := u.forumRepo.getThread(thread.Slug)
		if err != nil && err != consts.ErrNotFound {
			return nil, err
		}
		if existing != nil {
			return u.outputThread(existing), fmt.Errorf("%w: thread with this slug already exists", consts.ErrConflict)
		}
	}

	if thread.Created == "" {
		thread.Created = time.Now().Format(time.RFC3339)
	}

	created, err := u.forumRepo.createThread(forum, thread)
	return u.outputThread(created), err
}

func (u *Usecase) outputThread(thread *model.Thread) *threadOutput {
	if thread == nil {
		return nil
	}
	return &threadOutput{
		Author:  thread.User,
		Created: thread.Created,
		Forum:   thread.Forum,
		ID:      thread.ID,
		Message: thread.Message,
		Slug:    thread.Slug,
		Title:   thread.Title,
		Votes:   thread.Votes,
	}
}
