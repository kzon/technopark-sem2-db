package repository

import (
	"fmt"
	model2 "github.com/kzon/technopark-sem2-db/pkg/component/forum/model"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/repository"
	"strconv"
)

func (r *Repository) GetForumThreads(forum string, limit int, desc bool) (model.Threads, error) {
	query := fmt.Sprintf(
		"select * from thread where forum = $1 order by created %s limit $2",
		r.getOrder(desc),
	)
	var threads model.Threads
	err := r.db.Select(&threads, query, forum, limit)
	return threads, err
}

func (r *Repository) GetForumThreadsSince(forum, since string, limit int, desc bool) (model.Threads, error) {
	createdCond := ">="
	if desc {
		createdCond = "<="
	}
	query := fmt.Sprintf(
		"select * from thread where forum = $1 and created %s $2 order by created %s limit $3",
		createdCond, r.getOrder(desc),
	)
	threads := make(model.Threads, 0)
	err := r.db.Select(&threads, query, forum, since, limit)
	return threads, err
}

func (r *Repository) GetThreadByID(id int) (*model.Thread, error) {
	return r.getThread("*", "id=$1", id)
}

func (r *Repository) GetThreadBySlug(slug string) (*model.Thread, error) {
	return r.getThread("*", "slug=$1", slug)
}

func (r *Repository) GetThreadBySlugOrID(slugOrID string) (*model.Thread, error) {
	return r.GetThreadFieldsBySlugOrID("*", slugOrID)
}

func (r *Repository) GetThreadFieldsBySlugOrID(fields, slugOrID string) (*model.Thread, error) {
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		return r.getThread(fields, "slug=$1", slugOrID)
	}
	return r.getThread(fields, "id=$1", id)
}

func (r *Repository) getThread(fields, filter string, params ...interface{}) (*model.Thread, error) {
	t := model.Thread{}
	err := r.db.Get(&t, "select "+fields+" from thread where "+filter, params...)
	if err != nil {
		return nil, repository.Error(err)
	}
	return &t, nil
}

func (r *Repository) CreateThread(forum *model.Forum, thread model2.ThreadCreate) (*model.Thread, error) {
	id, err := r.createThreadInTx(forum, thread)
	if err != nil {
		return nil, err
	}
	return r.GetThreadByID(id)
}

func (r *Repository) createThreadInTx(forum *model.Forum, thread model2.ThreadCreate) (id int, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	err = tx.
		QueryRow(
			`insert into thread (title, author, forum, message, slug, created) values ($1, $2, $3, $4, $5, $6) returning id`,
			thread.Title, thread.Author, forum.Slug, thread.Message, thread.Slug, thread.Created,
		).
		Scan(&id)
	if err != nil {
		tx.Rollback()
		return
	}
	_, err = tx.Exec(`update forum set threads = threads + 1 where slug = $1`, forum.Slug)
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

func (r *Repository) UpdateThread(threadSlugOrID string, message, title string) (*model.Thread, error) {
	thread, err := r.GetThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return nil, err
	}
	if message != "" {
		thread.Message = message
	}
	if title != "" {
		thread.Title = title
	}
	_, err = r.db.Exec(
		`update thread set "message" = $1, title = $2 where id = $3`,
		thread.Message, thread.Title, thread.ID,
	)
	return thread, err
}
