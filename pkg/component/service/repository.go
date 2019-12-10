package service

import "github.com/jmoiron/sqlx"

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return Repository{db: db}
}

func (r *Repository) countForums() (count int, err error) {
	return r.count("forum")
}

func (r *Repository) countPosts() (count int, err error) {
	return r.count("post")
}

func (r *Repository) countThreads() (count int, err error) {
	return r.count("thread")
}

func (r *Repository) countUsers() (count int, err error) {
	return r.count("user")
}

func (r *Repository) count(table string) (count int, err error) {
	err = r.db.Get(&count, `select count(*) from "`+table+`"`)
	return
}

func (r *Repository) clear() error {
	_, err := r.db.Exec(`truncate thread, post, forum, "user", vote`)
	return err
}
