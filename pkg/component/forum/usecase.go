package forum

import (
	"fmt"
	forumModel "github.com/kzon/technopark-sem2-db/pkg/component/forum/model"
	"github.com/kzon/technopark-sem2-db/pkg/component/forum/repository"
	userComponent "github.com/kzon/technopark-sem2-db/pkg/component/user"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"time"
)

type Usecase struct {
	forumRepo repository.Repository
	userRepo  userComponent.Repository
}

func NewUsecase(forumRepo repository.Repository, userRepo userComponent.Repository) Usecase {
	return Usecase{forumRepo: forumRepo, userRepo: userRepo}
}

func (u *Usecase) createForum(title, slug, nickname string) (*model.Forum, error) {
	user, err := u.userRepo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}

	existingForum, err := u.forumRepo.GetForumBySlug(slug)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existingForum != nil {
		return existingForum, fmt.Errorf("%w: forum with this slug already exists", consts.ErrConflict)
	}

	return u.forumRepo.CreateForum(title, slug, user.Nickname)
}

func (u *Usecase) createThread(forumSlug string, thread forumModel.ThreadCreate) (*model.Thread, error) {
	if _, err := u.userRepo.GetUserByNickname(thread.Author); err != nil {
		return nil, err
	}
	forum, err := u.forumRepo.GetForumBySlug(forumSlug)
	if err != nil {
		return nil, err
	}

	if thread.Slug != "" {
		existing, err := u.forumRepo.GetThreadBySlug(thread.Slug)
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

	return u.forumRepo.CreateThread(forum, thread)
}

func (u *Usecase) updateThread(threadSlugOrID string, message, title string) (*model.Thread, error) {
	return u.forumRepo.UpdateThread(threadSlugOrID, message, title)
}

func (u *Usecase) createPosts(threadSlugOrID string, posts []forumModel.PostCreate) (model.Posts, error) {
	return u.forumRepo.CreatePosts(threadSlugOrID, posts)
}

func (u *Usecase) getForum(slug string) (*model.Forum, error) {
	return u.forumRepo.GetForumBySlug(slug)
}

func (u *Usecase) getForumThreads(forum, since string, limit int, desc bool) (model.Threads, error) {
	if _, err := u.forumRepo.GetForumBySlug(forum); err != nil {
		return nil, err
	}
	var threads model.Threads
	var err error
	if since == "" {
		threads, err = u.forumRepo.GetForumThreads(forum, limit, desc)
	} else {
		threads, err = u.forumRepo.GetForumThreadsSince(forum, since, limit, desc)
	}
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (u *Usecase) getForumUsers(forum, since string, limit int, desc bool) (model.Users, error) {
	return u.forumRepo.GetForumUsers(forum, since, limit, desc)
}

func (u *Usecase) voteForThread(threadSlugOrID string, vote forumModel.Vote) (thread *model.Thread, err error) {
	thread, err = u.forumRepo.GetThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return
	}
	user, err := u.userRepo.GetUserByNickname(vote.Nickname)
	if err != nil {
		return
	}
	newVotes, err := u.forumRepo.AddThreadVote(thread, user.Nickname, vote.Voice)
	thread.Votes = newVotes
	return
}

func (u *Usecase) getThread(threadSlugOrID string) (*model.Thread, error) {
	return u.forumRepo.GetThreadBySlugOrID(threadSlugOrID)
}

func (u *Usecase) getThreadPosts(threadSlugOrID string, limit int, since *int, sort string, desc bool) (model.Posts, error) {
	thread, err := u.forumRepo.GetThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return nil, err
	}
	return u.forumRepo.GetThreadPosts(thread.ID, limit, since, sort, desc)
}
