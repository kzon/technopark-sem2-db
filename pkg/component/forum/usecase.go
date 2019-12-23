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
	repo     repository.Repository
	userRepo userComponent.Repository
}

func NewUsecase(forumRepo repository.Repository, userRepo userComponent.Repository) Usecase {
	return Usecase{repo: forumRepo, userRepo: userRepo}
}

func (u *Usecase) createForum(title, slug, nickname string) (*model.Forum, error) {
	userNickname, err := u.userRepo.GetUserNickname(nickname)
	if err != nil {
		return nil, err
	}

	existingForum, err := u.repo.GetForumBySlug(slug)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existingForum != nil {
		return existingForum, fmt.Errorf("%w: forum with this slug already exists", consts.ErrConflict)
	}

	return u.repo.CreateForum(title, slug, userNickname)
}

func (u *Usecase) createThread(forumSlug string, thread forumModel.ThreadCreate) (*model.Thread, error) {
	if _, err := u.userRepo.GetUserNickname(thread.Author); err != nil {
		return nil, err
	}
	forum, err := u.repo.GetForumSlug(forumSlug)
	if err != nil {
		return nil, err
	}

	if thread.Slug != "" {
		existing, err := u.repo.GetThreadBySlug(thread.Slug)
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

	return u.repo.CreateThread(forum, thread)
}

func (u *Usecase) updateThread(threadSlugOrID string, message, title string) (*model.Thread, error) {
	return u.repo.UpdateThread(threadSlugOrID, message, title)
}

func (u *Usecase) createPosts(threadSlugOrID string, posts []*forumModel.PostCreate) (model.Posts, error) {
	thread, err := u.repo.GetThreadFieldsBySlugOrID("id, forum", threadSlugOrID)
	if err != nil {
		return nil, err
	}
	if err := u.checkPostsCreate(posts, thread.ID); err != nil {
		return nil, err
	}
	return u.repo.CreatePosts(posts, thread)
}

func (u *Usecase) checkPostsCreate(posts []*forumModel.PostCreate, threadID int) error {
	for _, post := range posts {
		if err := u.checkPostCreate(post, threadID); err != nil {
			return err
		}
	}
	return nil
}

func (u *Usecase) checkPostCreate(post *forumModel.PostCreate, threadID int) error {
	if _, err := u.userRepo.GetUserNickname(post.Author); err != nil {
		return err
	}
	if post.Parent != 0 {
		parent, err := u.repo.GetPostThreadByID(post.Parent)
		if err == consts.ErrNotFound {
			return fmt.Errorf("%w: post parent do not exists", consts.ErrConflict)
		}
		if err != nil {
			return err
		}
		if parent.Thread != threadID {
			return fmt.Errorf("%w: parent post was created in another thread", consts.ErrConflict)
		}
	}
	return nil
}

func (u *Usecase) getForum(slug string) (*model.Forum, error) {
	if err := u.repo.FillForumPostsCount(slug); err != nil {
		return nil, err
	}
	return u.repo.GetForumBySlug(slug)
}

func (u *Usecase) getForumThreads(forum, since string, limit int, desc bool) (model.Threads, error) {
	if _, err := u.repo.GetForumSlug(forum); err != nil {
		return nil, err
	}
	var threads model.Threads
	var err error
	if since == "" {
		threads, err = u.repo.GetForumThreads(forum, limit, desc)
	} else {
		threads, err = u.repo.GetForumThreadsSince(forum, since, limit, desc)
	}
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (u *Usecase) getForumUsers(forum, since string, limit int, desc bool) (model.Users, error) {
	return u.repo.GetForumUsers(forum, since, limit, desc)
}

func (u *Usecase) voteForThread(threadSlugOrID string, vote forumModel.Vote) (*model.Thread, error) {
	thread, err := u.repo.GetThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return nil, err
	}
	userNickname, err := u.userRepo.GetUserNickname(vote.Nickname)
	if err != nil {
		return nil, err
	}
	newVotes, err := u.repo.AddThreadVote(thread, userNickname, vote.Voice)
	thread.Votes = newVotes
	return thread, err
}

func (u *Usecase) getThread(threadSlugOrID string) (*model.Thread, error) {
	return u.repo.GetThreadBySlugOrID(threadSlugOrID)
}

func (u *Usecase) getThreadPosts(threadSlugOrID string, limit int, since *int, sort string, desc bool) (model.Posts, error) {
	thread, err := u.repo.GetThreadFieldsBySlugOrID("id", threadSlugOrID)
	if err != nil {
		return nil, err
	}
	return u.repo.GetThreadPosts(thread.ID, limit, since, sort, desc)
}

type postDetails struct {
	Post   *model.Post
	Author *model.User
	Forum  *model.Forum
	Thread *model.Thread
}

func (u *Usecase) getPostDetails(id int, related []string) (*postDetails, error) {
	post, err := u.repo.GetPostByID(id)
	if err != nil {
		return nil, err
	}
	details := postDetails{Post: post}
	for _, r := range related {
		switch r {
		case "user":
			details.Author, err = u.userRepo.GetUserByNickname(post.Author)
		case "forum":
			if err := u.repo.FillForumPostsCount(post.Forum); err != nil {
				return nil, err
			}
			details.Forum, err = u.repo.GetForumBySlug(post.Forum)
		case "thread":
			details.Thread, err = u.repo.GetThreadByID(post.Thread)
		}
		if err != nil {
			return nil, err
		}
	}
	return &details, nil
}

func (u *Usecase) updatePost(id int, message string) (*model.Post, error) {
	return u.repo.UpdatePostMessage(id, message)
}
