package repository

import (
	"fmt"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/repository"
)

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

func (r *Repository) FillForumPostsCount(forum string) error {
	var count int
	if err := r.db.Get(&count, `select count(*) from post where forum=$1`, forum); err != nil {
		return err
	}
	_, err := r.db.Exec(`update forum set posts=$1`, count)
	return err
}

func (r *Repository) GetForumUsers(forumSlug, since string, limit int, desc bool) (model.Users, error) {
	forum, err := r.GetForumSlug(forumSlug)
	if err != nil {
		return nil, err
	}
	needFill, err := r.needFillForumUsers(forum.Slug)
	if err != nil {
		return nil, err
	}
	if needFill {
		if err := r.fillForumUsers(forum.Slug); err != nil {
			return nil, err
		}
	}

	sinceFilter := ""
	if since != "" {
		if desc {
			sinceFilter = "and nickname < $2"
		} else {
			sinceFilter = "and nickname > $2"
		}
	}
	limitExpr := ""
	if limit > 0 {
		limitExpr = fmt.Sprintf("limit %d", limit)
	}
	query := fmt.Sprintf(
		`select "user".* from "user" join forum_user on nickname = forum_user.user
				where forum = $1 %s order by nickname %s %s`,
		sinceFilter, r.getOrder(desc), limitExpr,
	)
	users := make(model.Users, 0)
	if since == "" {
		err = r.db.Select(&users, query, forum.Slug)
	} else {
		err = r.db.Select(&users, query, forum.Slug, since)
	}
	return users, err
}

func (r *Repository) needFillForumUsers(forum string) (bool, error) {
	var result bool
	err := r.db.Get(&result, `select not exists(select forum from forum_user where forum = $1)`, forum)
	return result, err
}

func (r *Repository) fillForumUsers(forum string) error {
	nicknames := make([]string, 0)
	err := r.db.Select(&nicknames,
		`select nickname from "user" where (
				exists(select id from post where author = nickname and forum = $1) or
				exists(select id from thread where author = nickname and forum = $1)
			)`, forum)
	if err != nil {
		return err
	}
	for _, n := range nicknames {
		_, err := r.db.Exec(`insert into forum_user (forum, "user") values ($1, $2)`, forum, n)
		if err != nil {
			return err
		}
	}
	return nil
}
