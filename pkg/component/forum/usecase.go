package forum

import (
	"fmt"
	forumModel "github.com/kzon/technopark-sem2-db/pkg/component/forum/model"
	"github.com/kzon/technopark-sem2-db/pkg/component/forum/repository"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"time"
)

type Usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) Usecase {
	return Usecase{repo: repo}
}

func (u *Usecase) getUserByNickname(nickname string) (*model.User, error) {
	return u.repo.GetUserByNickname(nickname)
}

func (u *Usecase) createUser(nickname, email, fullname, about string) ([]*model.User, error) {
	existing, err := u.repo.GetUsersByNicknameOrEmail(nickname, email)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return existing, consts.ErrConflict
	}
	user, err := u.repo.CreateUser(nickname, email, fullname, about)
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
	if err := u.repo.UpdateUserByNickname(nickname, email, fullname, about); err != nil {
		return nil, err
	}
	return u.repo.GetUserByNickname(nickname)
}

func (u *Usecase) createForum(title, slug, nickname string) (*model.Forum, error) {
	userNickname, err := u.repo.GetUserNickname(nickname)
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
	if _, err := u.repo.GetUserNickname(thread.Author); err != nil {
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

func (u *Usecase) createPosts(threadSlugOrID string, posts []*forumModel.PostCreate) ([]*forumModel.PostOutput, error) {
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
	if _, err := u.repo.GetUserNickname(post.Author); err != nil {
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
	forum, err := u.repo.GetForumBySlug(slug)
	if err != nil {
		return nil, err
	}
	if forum.Posts == 0 {
		forum.Posts, err = u.repo.FillForumPostsCount(forum.ID)
		if err != nil {
			return nil, err
		}
	}
	return forum, nil
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
	userNickname, err := u.repo.GetUserNickname(vote.Nickname)
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

func (u *Usecase) getThreadPosts(threadSlugOrID string, limit int, since *int, sort string, desc bool) ([]*forumModel.PostOutput, error) {
	thread, err := u.repo.GetThreadFieldsBySlugOrID("id", threadSlugOrID)
	if err != nil {
		return nil, err
	}
	return u.repo.GetThreadPosts(thread.ID, limit, since, sort, desc)
}

type postDetails struct {
	Post   *forumModel.PostOutput
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
			details.Author, err = u.repo.GetUserByNickname(post.Author)
		case "forum":
			details.Forum, err = u.repo.GetForumBySlug(post.Forum)
			if details.Forum.Posts == 0 {
				details.Forum.Posts, err = u.repo.FillForumPostsCount(details.Forum.ID)
				if err != nil {
					return nil, err
				}
			}
		case "thread":
			details.Thread, err = u.repo.GetThreadByID(post.Thread)
		}
		if err != nil {
			return nil, err
		}
	}
	return &details, nil
}

func (u *Usecase) updatePost(id int, message string) (*forumModel.PostOutput, error) {
	return u.repo.UpdatePost(id, message)
}
