package forum

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	repo "github.com/kzon/technopark-sem2-db/pkg/repository"
	"strconv"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return Repository{db: db}
}

func (r *Repository) getForumByID(id int) (*model.Forum, error) {
	return r.getForum("id=$1", id)
}

func (r *Repository) getForumBySlug(slug string) (*model.Forum, error) {
	return r.getForum("slug=$1", slug)
}

func (r *Repository) getForum(filter string, params ...interface{}) (*model.Forum, error) {
	forum := model.Forum{}
	err := r.db.Get(&forum, `select * from forum where `+filter, params...)
	if err != nil {
		return nil, repo.Error(err)
	}
	return &forum, nil
}

func (r *Repository) getForumThreads(forum string, limit int, desc bool) ([]*model.Thread, error) {
	query := "select * from thread where forum = $1 order by created"
	if desc {
		query += " desc"
	}
	query += " limit " + strconv.Itoa(limit)
	var threads []*model.Thread
	err := r.db.Select(&threads, query, forum)
	return threads, err
}

func (r *Repository) getForumThreadsSince(forum, since string, limit int, desc bool) ([]*model.Thread, error) {
	query := "select * from thread where forum = $1"
	if desc {
		query += " and created <= $2"
	} else {
		query += " and created >= $2"
	}
	query += " order by created"
	if desc {
		query += " desc"
	}
	query += " limit " + strconv.Itoa(limit)
	threads := make([]*model.Thread, 0)
	err := r.db.Select(&threads, query, forum, since)
	return threads, err
}

func (r *Repository) getThreadByID(id int) (*model.Thread, error) {
	return r.getThread("id=$1", id)
}

func (r *Repository) getThreadBySlug(slug string) (*model.Thread, error) {
	return r.getThread("slug=$1", slug)
}

func (r *Repository) getThreadBySlugOrID(slugOrID string) (*model.Thread, error) {
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		return r.getThread("slug=$1", slugOrID)
	}
	return r.getThread("id=$1", id)
}

func (r *Repository) getThread(filter string, params ...interface{}) (*model.Thread, error) {
	thread := model.Thread{}
	err := r.db.Get(&thread, "select * from thread where "+filter, params...)
	if err != nil {
		return nil, repo.Error(err)
	}
	return &thread, nil
}

func (r *Repository) getPostByID(id int) (*model.Post, error) {
	return r.getPost("id=$1", id)
}

func (r *Repository) getPost(filter string, params ...interface{}) (*model.Post, error) {
	post := model.Post{}
	err := r.db.Get(&post, "select * from post where "+filter, params...)
	if err != nil {
		return nil, repo.Error(err)
	}
	return &post, nil
}

func (r *Repository) createForum(title, slug, user string) (*model.Forum, error) {
	var id int
	err := r.db.
		QueryRow(`insert into forum (title, slug, "user") values ($1, $2, $3) returning id`, title, slug, user).
		Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.getForumByID(id)
}

func (r *Repository) createThread(forum *model.Forum, thread threadCreate) (*model.Thread, error) {
	var id int
	err := r.db.
		QueryRow(
			`insert into thread (title, author, forum, message, slug, created) values ($1, $2, $3, $4, $5, $6) returning id`,
			thread.Title, thread.Author, forum.Slug, thread.Message, thread.Slug, thread.Created,
		).
		Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.getThreadByID(id)
}

func (r *Repository) createPosts(thread *model.Thread, posts []postCreate) ([]*model.Post, error) {
	if !r.postsParentsExists(posts) {
		return nil, fmt.Errorf("%w: post parent do not exists", consts.ErrConflict)
	}
	result := make([]*model.Post, 0, len(posts))
	for _, post := range posts {
		created, err := r.createPost(thread, post)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
	}
	return result, nil
}

func (r *Repository) createPost(thread *model.Thread, post postCreate) (*model.Post, error) {
	var id int
	err := r.db.
		QueryRow(
			`insert into post (thread, forum, parent, author, message) values ($1, $2, $3, $4, $5) returning id`,
			thread.ID, thread.Forum, post.Parent, post.Author, post.Message,
		).
		Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.getPostByID(id)
}

func (r *Repository) postsParentsExists(posts []postCreate) bool {
	for _, post := range posts {
		if post.Parent == 0 {
			continue
		}
		if _, err := r.getPostByID(post.Parent); err != nil {
			return false
		}
	}
	return true
}

func (r *Repository) getVoice(nickname string, threadID int) (int, error) {
	var voice int
	err := r.db.Get(&voice, `select voice from vote where nickname = $1 and thread = $2`, nickname, threadID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return voice, err
}

func (r *Repository) addThreadVote(thread *model.Thread, nickname string, voice int) (newVotes int, err error) {
	oldVoice, err := r.getVoice(nickname, thread.ID)
	if err != nil {
		return
	}
	if oldVoice == voice {
		return thread.Votes, nil
	}
	newVoice := voice - oldVoice
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	err = tx.
		QueryRow(`update thread set votes = votes + $1 where id = $2 returning votes`, newVoice, thread.ID).
		Scan(&newVotes)
	if err != nil {
		tx.Rollback()
		return
	}
	if _, err = tx.Exec(`delete from vote where thread = $1 and nickname = $2`, thread.ID, nickname); err != nil {
		tx.Rollback()
		return
	}
	if _, err = tx.Exec(`insert into vote (thread, nickname, voice) values ($1, $2, $3)`, thread.ID, nickname, voice); err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}
