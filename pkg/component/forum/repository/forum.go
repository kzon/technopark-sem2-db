package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/repository"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return Repository{db: db}
}

func (r *Repository) GetForumByID(id int) (*model.Forum, error) {
	return r.getForum("*", "id=$1", id)
}

func (r *Repository) GetForumBySlug(slug string) (*model.Forum, error) {
	return r.getForum("*", "slug=$1", slug)
}

func (r *Repository) GetForumSlug(slug string) (*model.Forum, error) {
	return r.getForum("slug", "slug=$1", slug)
}

func (r *Repository) getForum(fields, filter string, params ...interface{}) (*model.Forum, error) {
	f := model.Forum{}
	err := r.db.Get(&f, `select `+fields+` from forum where `+filter, params...)
	if err != nil {
		return nil, repository.Error(err)
	}
	return &f, nil
}

func (r *Repository) CreateForum(title, slug, user string) (*model.Forum, error) {
	var id int
	err := r.db.
		QueryRow(`insert into forum (title, slug, "user") values ($1, $2, $3) returning id`, title, slug, user).
		Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetForumByID(id)
}

func (r *Repository) getOrder(desc bool) string {
	if desc {
		return "desc"
	}
	return "asc"
}
