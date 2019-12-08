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

func (u *Usecase) createForum(title, slug, nickname string) (*model.Forum, error) {
	user, err := u.userRepo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}

	existingForum, err := u.forumRepo.getForumBySlug(slug)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existingForum != nil {
		return existingForum, fmt.Errorf("%w: forum with this slug already exists", consts.ErrConflict)
	}

	return u.forumRepo.createForum(title, slug, user.Nickname)
}

func (u *Usecase) createThread(forumSlug string, thread threadCreate) (*model.Thread, error) {
	if _, err := u.userRepo.GetUserByNickname(thread.Author); err != nil {
		return nil, err
	}
	forum, err := u.forumRepo.getForumBySlug(forumSlug)
	if err != nil {
		return nil, err
	}

	if thread.Slug != "" {
		existing, err := u.forumRepo.getThreadBySlug(thread.Slug)
		if err != nil && err != consts.ErrNotFound {
			return nil, err
		}
		if existing != nil {
			return existing, fmt.Errorf("%w: thread with this slug already exists", consts.ErrConflict)
		}
	}

	if thread.Created == "" {
		thread.Created = time.Now().Format(time.RFC3339)
	}

	return u.forumRepo.createThread(forum, thread)
}

func (u *Usecase) createPosts(threadSlugOrID string, posts []postCreate) ([]*model.Post, error) {
	thread, err := u.forumRepo.getThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return nil, err
	}
	return u.forumRepo.createPosts(thread, posts)
}

func (u *Usecase) getForum(slug string) (*model.Forum, error) {
	return u.forumRepo.getForumBySlug(slug)
}

func (u *Usecase) getForumThreads(forum, since string, limit int, desc bool) ([]*model.Thread, error) {
	if _, err := u.forumRepo.getForumBySlug(forum); err != nil {
		return nil, err
	}
	var threads []*model.Thread
	var err error
	if since == "" {
		threads, err = u.forumRepo.getForumThreads(forum, limit, desc)
	} else {
		threads, err = u.forumRepo.getForumThreadsSince(forum, since, limit, desc)
	}
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (u *Usecase) voteForThread(threadSlugOrID string, vote vote) (thread *model.Thread, err error) {
	thread, err = u.forumRepo.getThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return
	}
	user, err := u.userRepo.GetUserByNickname(vote.Nickname)
	if err != nil {
		return
	}
	newVotes, err := u.forumRepo.addThreadVote(thread, user.Nickname, vote.Voice)
	thread.Votes = newVotes
	return
}

func (u *Usecase) getThread(threadSlugOrID string) (*model.Thread, error) {
	return u.forumRepo.getThreadBySlugOrID(threadSlugOrID)
}

func (u *Usecase) getThreadPosts(threadSlugOrID string, limit int, since, sort string, desc bool) ([]*model.Post, error) {
	thread, err := u.forumRepo.getThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return nil, err
	}
	return u.forumRepo.getThreadPosts(thread.ID, limit, since, sort, desc)
}
