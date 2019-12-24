package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
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

	pathDelim    = "."
	maxIDLength  = 7
	maxTreeLevel = 5

	postsChunkSize = 50
)

var zeroPathStud = strings.Repeat("0", maxIDLength)

func (r *Repository) GetPostByID(id int) (*forumModel.PostOutput, error) {
	post, err := r.getPost("id=$1", id)
	if err != nil {
		return nil, err
	}
	author, err := r.GetUserByID(post.Author)
	if err != nil {
		return nil, err
	}
	forum, err := r.GetForumByID(post.Forum)
	if err != nil {
		return nil, err
	}
	result := forumModel.PostOutput{
		ID:       post.ID,
		Parent:   post.Parent,
		Author:   author.Nickname,
		Forum:    forum.Slug,
		Thread:   post.Thread,
		Message:  post.Message,
		IsEdited: post.IsEdited,
		Created:  post.Created,
	}
	return &result, nil
}

func (r *Repository) GetPostThreadByID(id int) (*model.Post, error) {
	return r.getPostFields("thread", "id=$1", id)
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

func (r *Repository) getPostsByIDs(ids []int, forum *model.Forum, thread *model.Thread) ([]*forumModel.PostOutput, error) {
	posts := make(model.Posts, 0, len(ids))
	query, args, err := sqlx.In(`select * from post where id in (?) order by id`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&posts, query, args...)
	result := make([]*forumModel.PostOutput, 0, len(posts))
	for _, post := range posts {
		author, err := r.GetUserByID(post.Author)
		if err != nil {
			return nil, err
		}
		result = append(result, &forumModel.PostOutput{
			ID:       post.ID,
			Parent:   post.Parent,
			Author:   author.Nickname,
			Forum:    forum.Slug,
			Thread:   thread.ID,
			Message:  post.Message,
			IsEdited: post.IsEdited,
			Created:  post.Created,
		})
	}
	return result, err
}

func (r *Repository) getPosts(orderBy []string, limit int, filter string, params ...interface{}) ([]*forumModel.PostOutput, error) {
	query := fmt.Sprintf(`select post.id,parent,thread,message,created,
		forum.slug as forum,"user".nickname as author 
		from post 
		join forum on forum.id=post.forum
		join "user" on "user".id=post.author
		where %s order by %s`,
		filter, strings.Join(orderBy, ","),
	)
	if limit > 0 {
		query += fmt.Sprintf(" limit %d", limit)
	}
	posts := make([]*forumModel.PostOutput, 0, limit)
	err := r.db.Select(&posts, query, params...)
	return posts, err
}

func (r *Repository) CreatePosts(posts []*forumModel.PostCreate, thread *model.Thread) ([]*forumModel.PostOutput, error) {
	forum, err := r.GetForumBySlug(thread.Forum)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	result := make([]*forumModel.PostOutput, 0, len(posts))
	for _, chunk := range r.chunkPosts(posts) {
		created, err := r.createPostsInTx(forum, thread, chunk, now)
		if err != nil {
			return nil, err
		}
		result = append(result, created...)
	}
	return result, nil
}

func (r *Repository) chunkPosts(posts []*forumModel.PostCreate) [][]*forumModel.PostCreate {
	chunked := make([][]*forumModel.PostCreate, 0)
	for i := 0; i < len(posts); i += postsChunkSize {
		end := i + postsChunkSize
		if end > len(posts) {
			end = len(posts)
		}
		chunked = append(chunked, posts[i:end])
	}
	return chunked
}

func (r *Repository) createPostsInTx(forum *model.Forum, thread *model.Thread, posts []*forumModel.PostCreate, created time.Time) ([]*forumModel.PostOutput, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	insertedIDs, err := r.bulkCreatePosts(tx, forum, thread, posts, created)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return r.getPostsByIDs(insertedIDs, forum, thread)
}

func (r *Repository) bulkCreatePosts(tx *sqlx.Tx, forum *model.Forum, thread *model.Thread, posts []*forumModel.PostCreate, created time.Time) ([]int, error) {
	columns := 8
	placeholders := make([]string, 0, len(posts))
	args := make([]interface{}, 0, len(posts)*columns)
	i := 0
	createdIDs := make([]int, len(posts))
	for _, post := range posts {
		id, err := r.getNextPostID()
		if err != nil {
			return nil, err
		}

		path, err := r.getPostPath(id, post.Parent)
		if err != nil {
			return nil, err
		}

		authorID, err := r.getUserIDByNickname(post.Author)
		if err != nil {
			return nil, err
		}

		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*columns+1, i*columns+2, i*columns+3, i*columns+4, i*columns+5, i*columns+6, i*columns+7, i*columns+8,
		))
		args = append(args, id, thread.ID, forum.ID, post.Parent, path, authorID, post.Message, created)
		createdIDs = append(createdIDs, id)
		i++
	}
	query := fmt.Sprintf(
		"insert into post (id, thread, forum, parent, path, author, message, created) values %s",
		strings.Join(placeholders, ","),
	)
	_, err := tx.Exec(query, args...)
	return createdIDs, err
}

func (r *Repository) getPostPath(id, parentID int) (string, error) {
	var base string
	if parentID == 0 {
		base = r.getZeroPostPath()
	} else {
		parent, err := r.getPostFields("path", "id=$1", parentID)
		if err != nil {
			return "", err
		}
		base = parent.Path
	}
	path := strings.Replace(base, zeroPathStud, r.padPostID(id), 1)
	return path, nil
}

func (r *Repository) getZeroPostPath() string {
	path := zeroPathStud
	for i := 0; i < maxTreeLevel-1; i++ {
		path += pathDelim + zeroPathStud
	}
	return path
}

func (r *Repository) padPostID(id int) string {
	return fmt.Sprintf("%0"+strconv.Itoa(maxIDLength)+"d", id)
}

func (r *Repository) getNextPostID() (int, error) {
	var id int
	err := r.db.Get(&id, `select nextval(pg_get_serial_sequence('post', 'id'))`)
	return id, err
}

func (r *Repository) UpdatePost(id int, message string) (*forumModel.PostOutput, error) {
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
