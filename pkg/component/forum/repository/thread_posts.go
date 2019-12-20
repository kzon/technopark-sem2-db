package repository

import (
	"fmt"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
)

func (r *Repository) GetThreadPosts(thread, limit int, since *int, sort string, desc bool) (model.Posts, error) {
	switch sort {
	case SortFlat, "":
		return r.getThreadPostsFlat(thread, limit, since, desc)
	case SortTree:
		return r.getThreadPostsTree(thread, limit, since, desc)
	case SortParentTree:
		params := pageParams{
			limit: limit,
			since: since,
			sort:  sort,
			desc:  desc,
		}
		return r.getThreadPostsTreeOld(thread, params)
	}
	return nil, fmt.Errorf("%w: unknown sort method '%s'", consts.ErrNotFound, sort)
}

func (r *Repository) getThreadPostsFlat(thread, limit int, since *int, desc bool) (model.Posts, error) {
	order := "asc"
	if desc {
		order = "desc"
	}
	orderBy := []string{"created " + order, "id " + order}
	filter := "thread = $1"
	params := []interface{}{thread}
	if since != nil {
		if desc {
			filter += " and id < $2"
		} else {
			filter += " and id > $2"
		}
		params = append(params, *since)
	}
	return r.getPosts(orderBy, limit, filter, params...)
}

func (r *Repository) getThreadPostsTree(thread, limit int, since *int, desc bool) (model.Posts, error) {
	order := "asc"
	if desc {
		order = "desc"
	}
	orderBy := []string{"path " + order, "created " + order, "id " + order}
	filter := "thread = $1"
	params := []interface{}{thread}
	if since != nil {
		var operator = ">"
		if desc {
			operator = "<"
		}
		sincePost, err := r.getPostFields("path", "id=$1", *since)
		if err != nil {
			return nil, err
		}
		filter += fmt.Sprintf(" and path %s '%s'", operator, sincePost.Path)
	}
	return r.getPosts(orderBy, limit, filter, params...)
}
