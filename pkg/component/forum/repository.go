package forum

import (
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/repository"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return Repository{db: db}
}

func (r *Repository) getForum(slug string) (*model.Forum, error) {
	forum := model.Forum{}
	err := r.db.Get(&forum, `select * from "`+consts.ForumTable+`" where slug = $1`, slug)
	if err != nil {
		return nil, repository.Error(err)
	}
	return &forum, nil
}

func (r *Repository) createForum(title, slug, user string) error {
	_, err := r.db.Exec(`insert into "`+consts.ForumTable+`" (title, slug, "user") values ($1, $2, $3)`, title, slug, user)
	return err
}

func (r *Repository) getThread(slug string) (*model.Thread, error) {
	thread := model.Thread{}
	err := r.db.Get(&thread, `select * from "`+consts.ThreadTable+`" where slug = $1`, slug)
	if err != nil {
		return nil, repository.Error(err)
	}
	return &thread, nil
}

func (r *Repository) createThread(forum string, thread threadCreate) (*model.Thread, error) {
	var id int
	err := r.db.
		QueryRow(
			`insert into "`+consts.ThreadTable+`"(title, "user", forum, message, slug, created) values ($1, $2, $3, $4, $5, $6) returning id`,
			thread.Title, thread.Author, forum, thread.Message, thread.Slug, thread.Created,
		).
		Scan(&id)
	if err != nil {
		return nil, err
	}
	created := model.Thread{}
	err = r.db.Get(&created, `select * from "`+consts.ThreadTable+`" where id = $1`, id)
	return &created, err
}
