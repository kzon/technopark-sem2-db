package forum

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	repo "github.com/kzon/technopark-sem2-db/pkg/repository"
	"github.com/kzon/technopark-sem2-db/pkg/util"
	"sort"
	"strconv"
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

func (r *Repository) getForumThreads(forum string, limit int, desc bool) (model.Threads, error) {
	query := "select * from thread where forum = $1 order by created"
	if desc {
		query += " desc"
	}
	query += " limit " + strconv.Itoa(limit)
	var threads model.Threads
	err := r.db.Select(&threads, query, forum)
	return threads, err
}

func (r *Repository) getForumThreadsSince(forum, since string, limit int, desc bool) (model.Threads, error) {
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
	threads := make(model.Threads, 0)
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

func (r *Repository) getPosts(orderBy []string, limit int, filter string, params ...interface{}) (model.Posts, error) {
	query := fmt.Sprintf(`select * from post where %s order by %s`, filter, strings.Join(orderBy, ","))
	if limit > 0 {
		query += fmt.Sprintf(" limit %d", limit)
	}
	posts := make(model.Posts, 0, limit)
	err := r.db.Select(&posts, query, params...)
	return posts, err
}

func (r *Repository) getThreadPosts(thread, limit int, since *int, sort string, desc bool) (model.Posts, error) {
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

func (r *Repository) updateThread(threadSlugOrID string, message, title string) (*model.Thread, error) {
	thread, err := r.getThreadBySlugOrID(threadSlugOrID)
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

func (r *Repository) createPosts(thread *model.Thread, posts []postCreate) (model.Posts, error) {
	if !r.postsParentsExists(posts) {
		return nil, fmt.Errorf("%w: post parent do not exists", consts.ErrConflict)
	}
	now := time.Now()
	result := make(model.Posts, 0, len(posts))
	for _, post := range posts {
		created, err := r.createPost(thread, post, now)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
	}
	return result, nil
}

func (r *Repository) createPost(thread *model.Thread, post postCreate, created time.Time) (*model.Post, error) {
	var id int
	err := r.db.
		QueryRow(
			`insert into post (thread, forum, parent, author, message, created) values ($1, $2, $3, $4, $5, $6) returning id`,
			thread.ID, thread.Forum, post.Parent, post.Author, post.Message, created,
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
