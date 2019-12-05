package forum

import (
	"github.com/jmoiron/sqlx"
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

func (r *Repository) getThread(filter string, params ...interface{}) (*model.Thread, error) {
	thread := model.Thread{}
	err := r.db.Get(&thread, "select * from thread where "+filter, params...)
	if err != nil {
		return nil, repo.Error(err)
	}
	return &thread, nil
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

func (r *Repository) createThread(forum string, thread threadCreate) (*model.Thread, error) {
	var id int
	err := r.db.
		QueryRow(
			`insert into thread (title, "user", forum, message, slug, created) values ($1, $2, $3, $4, $5, $6) returning id`,
			thread.Title, thread.Author, forum, thread.Message, thread.Slug, thread.Created,
		).
		Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.getThreadByID(id)
}
