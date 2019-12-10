package repository

import (
	"fmt"
	forumModel "github.com/kzon/technopark-sem2-db/pkg/component/forum/model"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/repository"
	"github.com/kzon/technopark-sem2-db/pkg/util"
	"sort"
	"strings"
	"time"
)

const (
	SortFlat       = "flat"
	SortTree       = "tree"
	SortParentTree = "parent_tree"
)

type pageParams struct {
	limit int
	since *int
	sort  string
	desc  bool
}

func (r *Repository) GetPostByID(id int) (*model.Post, error) {
	return r.getPost("id=$1", id)
}

func (r *Repository) getPost(filter string, params ...interface{}) (*model.Post, error) {
	p := model.Post{}
	err := r.db.Get(&p, "select * from post where "+filter, params...)
	if err != nil {
		return nil, repository.Error(err)
	}
	return &p, nil
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

func (r *Repository) GetThreadPosts(thread, limit int, since *int, sort string, desc bool) (model.Posts, error) {
	switch sort {
	case SortFlat, "":
		return r.getThreadPostsFlat(thread, limit, since, desc)
	case SortTree, SortParentTree:
		params := pageParams{
			limit: limit,
			since: since,
			sort:  sort,
			desc:  desc,
		}
		return r.getThreadPostsTree(thread, params)
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

func (r *Repository) getThreadPostsTree(thread int, params pageParams) (model.Posts, error) {
	order := "asc"
	if params.desc && params.sort != SortParentTree {
		order = "desc"
	}
	posts, err := r.getPosts([]string{"created " + order, "id " + order}, 0, "thread = $1", thread)
	if err != nil {
		return nil, err
	}

	postsByParent, minParent := r.indexPostsByParent(posts)

	if params.desc && params.sort == SortParentTree {
		postsByParent[minParent] = r.sortPostsDesc(postsByParent[minParent])
	}
	tree := r.getPostsTree(postsByParent, minParent, params)
	if params.since == nil {
		return tree, nil
	}
	return r.filterPostsTree(tree, minParent, params), nil
}

func (r *Repository) indexPostsByParent(posts model.Posts) (map[int]model.Posts, int) {
	postsByParent := make(map[int]model.Posts)
	minParent := util.MaxInt
	for _, post := range posts {
		parent := post.Parent
		if _, ok := postsByParent[parent]; !ok {
			postsByParent[parent] = make(model.Posts, 0, 2)
			if parent < minParent {
				minParent = parent
			}
		}
		postsByParent[parent] = append(postsByParent[parent], post)
	}
	return postsByParent, minParent
}

type postSortDesc model.Posts

func (p postSortDesc) Len() int {
	return len(p)
}

func (p postSortDesc) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p postSortDesc) Less(i, j int) bool {
	p1, p2 := p[i], p[j]
	if p1.Created == p2.Created {
		return p1.ID > p2.ID
	}
	return p1.Created > p2.Created
}

func (r *Repository) sortPostsDesc(posts model.Posts) model.Posts {
	sort.Sort(postSortDesc(posts))
	return posts
}

func (r *Repository) getPostsTree(postsByParent map[int]model.Posts, parent int, params pageParams) model.Posts {
	result := make(model.Posts, 0)
	for i, parentPost := range postsByParent[parent] {
		childParams := params
		if params.since == nil && params.sort != SortParentTree {
			childParams.limit -= len(result) + 1
			if params.desc && childParams.limit == 0 {
				childParams.limit = 1
			}
		}
		childPosts := r.getPostsTree(postsByParent, parentPost.ID, childParams)

		switch params.sort {
		case SortParentTree:
			result = append(result, parentPost)
			result = append(result, childPosts...)
		default:
			posts := make(model.Posts, 0)
			if params.desc {
				posts = append(posts, childPosts...)
				posts = append(posts, parentPost)
			} else {
				posts = append(posts, parentPost)
				posts = append(posts, childPosts...)
			}
			if len(result)+len(posts) > params.limit && params.since == nil {
				result = append(result, posts[0:params.limit-len(result)]...)
			} else {
				result = append(result, posts...)
			}
		}

		if r.postsTreeLimitExceeded(i, len(result), params) && params.since == nil {
			break
		}
	}
	return result
}

func (r *Repository) postsTreeLimitExceeded(i int, l int, params pageParams) bool {
	if params.sort == SortParentTree {
		return i == params.limit-1
	}
	return l == params.limit
}

func (r *Repository) filterPostsTree(tree model.Posts, minParent int, params pageParams) model.Posts {
	filtered := make(model.Posts, 0, params.limit)
	count := 0
	for i := range tree {
		if count != 0 || (i > 0 && tree[i-1].ID == *params.since) {
			filtered = append(filtered, tree[i])
			if params.sort != SortParentTree || tree[i].Parent == minParent {
				count++
			}
		}
		if count == params.limit {
			break
		}
	}
	return filtered
}

func (r *Repository) CreatePosts(threadSlugOrID string, posts []forumModel.PostCreate) (model.Posts, error) {
	t, err := r.GetThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return nil, err
	}
	f, err := r.GetForumBySlug(t.Forum)
	if err != nil {
		return nil, err
	}
	if !r.postsParentsExists(posts) {
		return nil, fmt.Errorf("%w: post parent do not exists", consts.ErrConflict)
	}
	now := time.Now()
	result := make(model.Posts, 0, len(posts))
	for _, p := range posts {
		created, err := r.createPost(f, t, p, now)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
	}
	return result, nil
}

func (r *Repository) createPost(forum *model.Forum, thread *model.Thread, post forumModel.PostCreate, created time.Time) (*model.Post, error) {
	id, err := r.createPostInTx(forum, thread, post, created)
	if err != nil {
		return nil, err
	}
	return r.GetPostByID(id)
}

func (r *Repository) createPostInTx(forum *model.Forum, thread *model.Thread, post forumModel.PostCreate, created time.Time) (id int, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	err = tx.
		QueryRow(
			`insert into post (thread, forum, parent, author, message, created) values ($1, $2, $3, $4, $5, $6) returning id`,
			thread.ID, thread.Forum, post.Parent, post.Author, post.Message, created,
		).
		Scan(&id)
	if err != nil {
		tx.Rollback()
		return
	}
	_, err = tx.Exec(`update forum set posts = posts + 1 where slug = $1`, forum.Slug)
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

func (r *Repository) postsParentsExists(posts []forumModel.PostCreate) bool {
	for _, post := range posts {
		if post.Parent == 0 {
			continue
		}
		if _, err := r.GetPostByID(post.Parent); err != nil {
			return false
		}
	}
	return true
}

func (r *Repository) UpdatePost(id int, message string) (*model.Post, error) {
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
