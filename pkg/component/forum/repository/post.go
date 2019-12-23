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

type postPath struct {
	ID   int
	Path string
}

func (r *Repository) GetPostByID(id int) (*model.Post, error) {
	return r.getPost("id=$1", id)
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

func (r *Repository) CreatePosts(posts []*forumModel.PostCreate, thread *model.Thread) (model.Posts, error) {
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

func (r *Repository) createPostsInTx(forum *model.Forum, thread *model.Thread, posts []*forumModel.PostCreate, created time.Time) (model.Posts, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	insertedIds, err := r.bulkCreatePosts(tx, forum, thread, posts, created)
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

func (r *Repository) bulkCreatePosts(tx *sqlx.Tx, forum *model.Forum, thread *model.Thread, posts []*forumModel.PostCreate, created time.Time) ([]int, error) {
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

func (r *Repository) fillPostsPath(tx *sqlx.Tx, ids []int, posts []*forumModel.PostCreate) error {
	paths := make([]postPath, 0, len(posts))
	for i, id := range ids {
		post := posts[i]
		path, err := r.getPostPath(id, post.Parent)
		if err != nil {
			return err
		}
		paths = append(paths, postPath{ID: id, Path: path})
	}
	return r.updatePostsPath(tx, paths)
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

func (r *Repository) updatePostsPath(tx *sqlx.Tx, paths []postPath) error {
	if _, err := tx.Exec(`create temporary table if not exists post_path (id int, path text)`); err != nil {
		return err
	}
	columns := 2
	placeholders := make([]string, 0, len(paths))
	args := make([]interface{}, 0, len(paths)*columns)
	i := 0
	for _, p := range paths {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*columns+1, i*columns+2))
		args = append(args, p.ID)
		args = append(args, p.Path)
		i++
	}
	query := "insert into post_path (id, path) values " + strings.Join(placeholders, ",")
	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}
	_, err := tx.Exec(`update post set path = post_path.path from post_path where post.id = post_path.id`)
	return err
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
