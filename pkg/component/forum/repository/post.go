package repository

import (
	"fmt"
	forumModel "github.com/kzon/technopark-sem2-db/pkg/component/forum/model"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/repository"
	"strconv"
	"strings"
	"time"
)

const (
	SortFlat       = "flat"
	SortTree       = "tree"
	SortParentTree = "parent_tree"

	pathDelim = "."
)

type pageParams struct {
	limit int
	since *int
	sort  string
	desc  bool
}

func (r *Repository) GetPostByID(id int) (*model.Post, error) {
	return r.getPost("id=$1", id)
}

func (r *Repository) getPost(filter string, params ...interface{}) (*model.Post, error) {
	return r.getPostFields("*", filter, params...)
}

func (r *Repository) getPostFields(fields, filter string, params ...interface{}) (*model.Post, error) {
	p := model.Post{}
	err := r.db.Get(&p, "select "+fields+" from post where "+filter, params...)
	if err != nil {
		return nil, repository.Error(err)
	}
	return &p, nil
}

func (r *Repository) getPosts(orderBy []string, limit int, filter string, params ...interface{}) (model.Posts, error) {
	query := fmt.Sprintf(`select * from post where %s order by %s`, filter, strings.Join(orderBy, ","))
	if limit > 0 {
		query += fmt.Sprintf(" limit %d", limit)
	}
	posts := make(model.Posts, 0, limit)
	err := r.db.Select(&posts, query, params...)
	return posts, err
}

func (r *Repository) CreatePosts(posts []forumModel.PostCreate, thread *model.Thread) (model.Posts, error) {
	forum, err := r.GetForumSlug(thread.Forum)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	result := make(model.Posts, 0, len(posts))
	for _, post := range posts {
		created, err := r.createPost(forum, thread, post, now)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
	}
	return result, nil
}

func (r *Repository) createPost(forum *model.Forum, thread *model.Thread, post forumModel.PostCreate, created time.Time) (*model.Post, error) {
	id, err := r.createPostInTx(forum, thread, post, created)
	if err != nil {
		return nil, err
	}
	return r.GetPostByID(id)
}

func (r *Repository) createPostInTx(forum *model.Forum, thread *model.Thread, post forumModel.PostCreate, created time.Time) (id int, err error) {
	path, err := r.getPostPath(post)
	if err != nil {
		return
	}
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	err = tx.
		QueryRow(
			`insert into post (thread, forum, parent, path, author, message, created) values ($1, $2, $3, $4, $5, $6, $7) returning id`,
			thread.ID, thread.Forum, post.Parent, path, post.Author, post.Message, created,
		).
		Scan(&id)
	if err != nil {
		tx.Rollback()
		return
	}
	_, err = tx.Exec(`update forum set posts = posts + 1 where slug = $1`, forum.Slug)
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

func (r *Repository) UpdatePost(id int, message string) (*model.Post, error) {
	if message != "" {
		_, err := r.db.Exec(
			`update post set "message" = $1, "isEdited" = true where id = $2 and "message" <> $1`,
			message, id,
		)
		if err != nil {
			return nil, err
		}
	}
	return r.GetPostByID(id)
}

func (r *Repository) getPostPath(post forumModel.PostCreate) (string, error) {
	if post.Parent == 0 {
		return "", nil
	}
	parent, err := r.getPostFields("path", "id=$1", post.Parent)
	if err != nil {
		return "", err
	}
	if parent.Path == "" {
		return strconv.Itoa(post.Parent), nil
	}
	return parent.Path + pathDelim + strconv.Itoa(post.Parent), nil
}
