package repository

import (
	"github.com/kzon/technopark-sem2-db/pkg/model"
	"github.com/kzon/technopark-sem2-db/pkg/util"
	"sort"
)

func (r *Repository) getThreadPostsTreeOld(thread int, params pageParams) (model.Posts, error) {
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
