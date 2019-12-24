package forum

import (
	"errors"
	forumModel "github.com/kzon/technopark-sem2-db/pkg/component/forum/model"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/delivery"
	"github.com/labstack/echo"
	"strconv"
	"strings"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(e *echo.Echo, usecase Usecase) Handler {
	h := Handler{usecase: usecase}

	e.POST("/api/user/:nickname/create", h.handleUserCreate)
	e.GET("/api/user/:nickname/profile", h.handleGetUserProfile)
	e.POST("/api/user/:nickname/profile", h.handleUserUpdate)

	e.POST("/api/forum/create", h.handleForumCreate)
	e.POST("/api/forum/:slug/create", h.handleThreadCreate)
	e.GET("/api/forum/:slug/details", h.handleGetForumDetails)
	e.GET("/api/forum/:slug/threads", h.handleGetForumThreads)
	e.GET("/api/forum/:slug/users", h.handleGetForumUsers)

	e.POST("/api/thread/:slug_or_id/create", h.handlePostCreate)
	e.POST("/api/thread/:slug_or_id/vote", h.handleVoteForThread)
	e.GET("/api/thread/:slug_or_id/details", h.handleGetThreadDetails)
	e.POST("/api/thread/:slug_or_id/details", h.handleThreadUpdate)
	e.GET("/api/thread/:slug_or_id/posts", h.handleGetThreadPosts)

	e.GET("/api/post/:id/details", h.handleGetPostDetails)
	e.POST("/api/post/:id/details", h.handlePostUpdate)

	return h
}

func (h *Handler) handleUserCreate(c echo.Context) error {
	u := forumModel.UserInput{}
	if err := c.Bind(&u); err != nil {
		return delivery.BadRequest(c, err)
	}
	users, err := h.usecase.createUser(c.Param("nickname"), u.Email, u.Fullname, u.About)
	if errors.Is(err, consts.ErrConflict) {
		return delivery.Conflict(c, users)
	}
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Created(c, users[0])
}

func (h *Handler) handleGetUserProfile(c echo.Context) error {
	u, err := h.usecase.getUserByNickname(c.Param("nickname"))
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, u)
}

func (h *Handler) handleUserUpdate(c echo.Context) error {
	u := forumModel.UserInput{}
	if err := c.Bind(&u); err != nil {
		return delivery.BadRequest(c, err)
	}
	user, err := h.usecase.updateUser(c.Param("nickname"), u.Email, u.Fullname, u.About)
	if errors.Is(err, consts.ErrConflict) {
		return delivery.ConflictWithMessage(c, err)
	}
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, user)
}

func (h Handler) handleForumCreate(c echo.Context) error {
	forumToCreate := forumModel.ForumCreate{}
	if err := c.Bind(&forumToCreate); err != nil {
		return delivery.BadRequest(c, err)
	}
	forum, err := h.usecase.createForum(forumToCreate.Title, forumToCreate.Slug, forumToCreate.User)
	if errors.Is(err, consts.ErrConflict) {
		return delivery.Conflict(c, forum)
	}
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Created(c, forum)
}

func (h *Handler) handleThreadCreate(c echo.Context) error {
	thread := forumModel.ThreadCreate{}
	if err := c.Bind(&thread); err != nil {
		return delivery.BadRequest(c, err)
	}
	forum := c.Param("slug")
	result, err := h.usecase.createThread(forum, thread)
	if errors.Is(err, consts.ErrConflict) {
		return delivery.Conflict(c, result)
	}
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Created(c, result)
}

func (h *Handler) handleGetForumDetails(c echo.Context) error {
	slug := c.Param("slug")
	forum, err := h.usecase.getForum(slug)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, forum)
}

func (h *Handler) handleGetForumThreads(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	desc, _ := strconv.ParseBool(c.QueryParam("desc"))
	threads, err := h.usecase.getForumThreads(c.Param("slug"), c.QueryParam("since"), limit, desc)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, threads)
}

func (h *Handler) handleGetForumUsers(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	desc, _ := strconv.ParseBool(c.QueryParam("desc"))
	users, err := h.usecase.getForumUsers(c.Param("slug"), c.QueryParam("since"), limit, desc)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, users)
}

func (h *Handler) handlePostCreate(c echo.Context) error {
	var posts []*forumModel.PostCreate
	if err := c.Bind(&posts); err != nil {
		return delivery.BadRequest(c, err)
	}
	result, err := h.usecase.createPosts(c.Param("slug_or_id"), posts)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Created(c, result)
}

func (h *Handler) handleVoteForThread(c echo.Context) error {
	var vote forumModel.Vote
	if err := c.Bind(&vote); err != nil {
		return delivery.BadRequest(c, err)
	}
	thread, err := h.usecase.voteForThread(c.Param("slug_or_id"), vote)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, thread)
}

func (h *Handler) handleGetThreadDetails(c echo.Context) error {
	thread, err := h.usecase.getThread(c.Param("slug_or_id"))
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, thread)
}

func (h *Handler) handleThreadUpdate(c echo.Context) error {
	t := forumModel.ThreadUpdate{}
	if err := c.Bind(&t); err != nil {
		return delivery.BadRequest(c, err)
	}
	thread, err := h.usecase.updateThread(c.Param("slug_or_id"), t.Message, t.Title)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, thread)
}

func (h *Handler) handleGetThreadPosts(c echo.Context) error {
	sp := c.QueryParam("since")
	var since *int = nil
	if sp != "" {
		n, _ := strconv.Atoi(sp)
		since = &n
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	desc, _ := strconv.ParseBool(c.QueryParam("desc"))
	posts, err := h.usecase.getThreadPosts(
		c.Param("slug_or_id"),
		limit,
		since,
		c.QueryParam("sort"),
		desc,
	)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, posts)
}

func (h *Handler) handleGetPostDetails(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	related := strings.Split(c.QueryParam("related"), ",")
	details, err := h.usecase.getPostDetails(id, related)
	if err != nil {
		return delivery.Error(c, err)
	}
	result := map[string]interface{}{
		"post": details.Post,
	}
	for _, r := range related {
		switch r {
		case "user":
			result["author"] = details.Author
		case "forum":
			result["forum"] = details.Forum
		case "thread":
			result["thread"] = details.Thread
		}
	}
	return delivery.Ok(c, result)
}

func (h *Handler) handlePostUpdate(c echo.Context) error {
	p := forumModel.PostUpdate{}
	if err := c.Bind(&p); err != nil {
		return delivery.BadRequest(c, err)
	}
	id, _ := strconv.Atoi(c.Param("id"))
	post, err := h.usecase.updatePost(id, p.Message)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, post)
}
