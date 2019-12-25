package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	apiModel "github.com/kzon/technopark-sem2-db/pkg/api/model"
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
)

var zeroPathStud = strings.Repeat("0", maxIDLength)

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

func (r *Repository) getPostsByIDs(ids []int) (model.Posts, error) {
	posts := make(model.Posts, 0)
	query, args, err := sqlx.In(`select * from post where id in (?) order by id`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&posts, query, args...)
	return posts, err
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

func (r *Repository) CreatePosts(posts []*apiModel.PostCreate, thread *model.Thread) (model.Posts, error) {
	forum, err := r.GetForumSlug(thread.Forum)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	result := make(model.Posts, 0, len(posts))
	for _, chunk := range r.chunkPosts(posts) {
		created, err := r.createPostsInTx(forum, thread, chunk, now)
		if err != nil {
			return nil, err
		}
		result = append(result, created...)
	}
	return result, nil
}

func (r *Repository) chunkPosts(posts []*apiModel.PostCreate) [][]*apiModel.PostCreate {
	chunked := make([][]*apiModel.PostCreate, 0)
	chunkSize := 200
	for i := 0; i < len(posts); i += chunkSize {
		end := i + chunkSize
		if end > len(posts) {
			end = len(posts)
		}
		chunked = append(chunked, posts[i:end])
	}
	return chunked
}

func (r *Repository) createPostsInTx(forum *model.Forum, thread *model.Thread, posts []*apiModel.PostCreate, created time.Time) (model.Posts, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	insertedIds, err := r.bulkCreatePosts(tx, forum, thread, posts, created)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = r.incForumPostsCount(tx, forum.Slug, len(posts))
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = r.fillPostsPath(tx, insertedIds, posts)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return r.getPostsByIDs(insertedIds)
}

func (r *Repository) bulkCreatePosts(tx *sqlx.Tx, forum *model.Forum, thread *model.Thread, posts []*apiModel.PostCreate, created time.Time) ([]int, error) {
	columns := 6
	placeholders := make([]string, 0, len(posts))
	args := make([]interface{}, 0, len(posts)*columns)
	i := 0
	for _, post := range posts {
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d)",
			i*columns+1, i*columns+2, i*columns+3, i*columns+4, i*columns+5, i*columns+6,
		))
		args = append(args, []interface{}{thread.ID, thread.Forum, post.Parent, post.Author, post.Message, created}...)
		i++
	}
	query := fmt.Sprintf(
		"insert into post (thread, forum, parent, author, message, created) values %s returning id",
		strings.Join(placeholders, ","),
	)
	ids := make([]int, 0)
	err := tx.Select(&ids, query, args...)
	return ids, err
}

func (r *Repository) incForumPostsCount(tx *sqlx.Tx, forum string, newCount int) error {
	_, err := tx.Exec(`update forum set posts = posts + $1 where slug = $2`, newCount, forum)
	return err
}

func (r *Repository) fillPostsPath(tx *sqlx.Tx, ids []int, posts []*apiModel.PostCreate) error {
	for i, id := range ids {
		post := posts[i]
		path, err := r.getPostPath(id, post.Parent)
		if err != nil {
			return err
		}
		err = r.updatePostPath(tx, id, path)
		if err != nil {
			return err
		}
	}
	return nil
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

func (r *Repository) UpdatePostMessage(id int, message string) (*model.Post, error) {
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

func (r *Repository) updatePostPath(tx *sqlx.Tx, id int, path string) error {
	_, err := tx.Exec(`update post set path = $1 where id = $2`, path, id)
	return err
}
