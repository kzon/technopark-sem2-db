package repository

import (
	"fmt"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"strings"
)

func (r *Repository) GetThreadPosts(thread, limit int, since *int, sort string, desc bool) (model.Posts, error) {
	switch sort {
	case SortFlat, "":
		return r.getThreadPostsFlat(thread, limit, since, desc)
	case SortTree:
		return r.getThreadPostsTree(thread, limit, since, desc)
	case SortParentTree:
		return r.getThreadPostsParentTree(thread, limit, since, desc)
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
	conditions := []string{"thread = $1"}
	params := []interface{}{thread}
	if since != nil {
		sinceCond, err := r.getSinceCondition(since, desc)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, sinceCond)
	}

	orderBy := []string{"path " + r.getOrder(desc)}
	filter := strings.Join(conditions, " and ")
	return r.getPosts(orderBy, limit, filter, params...)
}

func (r *Repository) getThreadPostsParentTree(thread, limit int, since *int, desc bool) (model.Posts, error) {
	conditions := []string{"parent=0", "thread=$1"}

	if since != nil {
		var operator = ">"
		if desc {
			operator = "<"
		}
		sincePost, err := r.getPostFields("path", "id=$1", *since)
		if err != nil {
			return nil, err
		}
		sinceCond := fmt.Sprintf("path %s '%s'", operator, r.getRootPath(sincePost.Path))
		conditions = append(conditions, sinceCond)
	}

	filter := strings.Join(conditions, " and ")
	var parents model.Posts
	err := r.db.Select(&parents, fmt.Sprintf(
		`select * from post where %s order by id %s limit %d`, filter, r.getOrder(desc), limit),
		thread,
	)
	if err != nil {
		return nil, err
	}
	posts := make(model.Posts, 0)
	for _, parent := range parents {
		var childs model.Posts
		err := r.db.Select(&childs, fmt.Sprintf(
			`select * from post where substring(path,1,7) = '%s' and parent<>0 order by path`, r.padPostID(parent.ID),
		))
		if err != nil {
			return nil, err
		}
		posts = append(posts, parent)
		posts = append(posts, childs...)
	}
	return posts, nil
}

func (r *Repository) getSinceCondition(since *int, desc bool) (string, error) {
	var operator = ">"
	if desc {
		operator = "<"
	}
	sincePost, err := r.getPostFields("path", "id=$1", *since)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("path %s '%s'", operator, sincePost.Path), nil
}

func (r *Repository) getRootPath(path string) string {
	root := strings.Split(path, pathDelim)[0]
	return root + strings.Repeat(pathDelim+zeroPathStud, maxTreeLevel-1)
}
