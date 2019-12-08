package forum

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/model"
	repo "github.com/kzon/technopark-sem2-db/pkg/repository"
	"strconv"
	"time"
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

func (r *Repository) getPosts(orderBy string, desc bool, limit int, filter string, params ...interface{}) ([]*model.Post, error) {
	var posts []*model.Post
	sort := ""
	if desc {
		sort = "desc"
	}
	err := r.db.Select(&posts,
		fmt.Sprintf(`select * from post where %s order by %s %s limit %d`, filter, orderBy, sort, limit),
		params...,
	)
	return posts, err
}

func (r *Repository) getThreadPosts(thread, limit int, since, sort string, desc bool) ([]*model.Post, error) {
	switch sort {
	case "", "flat":
		return r.getThreadPostsFlat(thread, limit, since, desc)
	case "tree":
		return r.getThreadPostsTree(thread, 0, limit, since, desc)
	case "parent_tree":
		return r.getThreadPostsParentTree(thread, limit, since, desc)
	}
	return nil, fmt.Errorf("%w: unknown sort method '%s'", consts.ErrNotFound, sort)
}

func (r *Repository) getThreadPostsFlat(thread, limit int, since string, desc bool) ([]*model.Post, error) {
	posts := make([]*model.Post, 0)
	sort := ""
	if desc {
		sort = "desc"
	}
	var err error
	if since == "" {
		err = r.db.Select(&posts,
			fmt.Sprintf(`select * from post where thread = $1 order by created %s, id %s limit $2`, sort, sort),
			thread, limit,
		)
	} else {
		err = r.db.Select(&posts,
			fmt.Sprintf(`select * from post where thread = $1 and id > $2 order by created %s, id %s limit $3`, sort, sort),
			thread, since, limit,
		)
	}
	return posts, err
}

func (r *Repository) getPostsByParent(thread, parent, limit int, since string, desc bool) ([]*model.Post, error) {
	filter := "thread = $1 and parent = $2"
	params := []interface{}{thread, parent}
	if since != "" {
		filter += " and id > $3"
		params = append(params, since)
	}
	return r.getPosts("created, id", desc, limit, filter, params...)
}

func (r *Repository) getThreadPostsTree(thread, parent, limit int, since string, desc bool) ([]*model.Post, error) {
	rootPosts, err := r.getPostsByParent(thread, parent, limit, since, desc)
	if err != nil {
		return nil, err
	}
	if len(rootPosts) == 0 {
		return rootPosts, nil
	}
	result := make([]*model.Post, 0, limit/2)
	count := 0
	for _, rootPost := range rootPosts {
		childPosts, err := r.getThreadPostsTree(thread, rootPost.ID, limit-count-1, since, desc)
		if err != nil {
			return nil, err
		}
		if desc {
			result = append(result, childPosts...)
			result = append(result, rootPost)
		} else {
			result = append(result, rootPost)
			result = append(result, childPosts...)
		}
		count += len(childPosts) + 1
		if count == limit {
			break
		}
	}
	return result, nil
}

func (r *Repository) getThreadPostsParentTree(thread, limit int, since string, desc bool) ([]*model.Post, error) {
	sort := ""
	if desc {
		sort = "desc"
	}
	var rootPosts []*model.Post
	err := r.db.Select(&rootPosts,
		fmt.Sprintf(`select * from post where thread = $1 and parent = 0 order by created %s, id %s limit $2`, sort, sort),
		thread, limit,
	)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Post, 0)
	for _, rootPost := range rootPosts {
		var childPosts []*model.Post
		err := r.db.Select(&childPosts,
			`select * from post where thread = $1 and parent = $2 order by created, id`,
			thread, rootPost.ID,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, rootPost)
		result = append(result, childPosts...)
	}
	return result, nil
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

func (r *Repository) createPosts(thread *model.Thread, posts []postCreate) ([]*model.Post, error) {
	if !r.postsParentsExists(posts) {
		return nil, fmt.Errorf("%w: post parent do not exists", consts.ErrConflict)
	}
	now := time.Now()
	result := make([]*model.Post, 0, len(posts))
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
