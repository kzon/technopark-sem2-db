package forum

import (
	"errors"
	forumModel "github.com/kzon/technopark-sem2-db/pkg/component/forum/model"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/delivery"
	"github.com/labstack/echo"
	"strconv"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(e *echo.Echo, usecase Usecase) Handler {
	handler := Handler{usecase: usecase}
	e.POST("/api/forum/create", handler.handleForumCreate)
	e.POST("/api/forum/:slug/create", handler.handleThreadCreate)
	e.GET("/api/forum/:slug/details", handler.handleGetForumDetails)
	e.GET("/api/forum/:slug/threads", handler.handleGetForumThreads)
	e.POST("/api/thread/:slug_or_id/create", handler.handlePostCreate)
	e.POST("/api/thread/:slug_or_id/vote", handler.handleVoteForThread)
	e.GET("/api/thread/:slug_or_id/details", handler.handleGetThreadDetails)
	e.POST("/api/thread/:slug_or_id/details", handler.handleThreadUpdate)
	e.GET("/api/thread/:slug_or_id/posts", handler.handleGetThreadPosts)
	return handler
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

func (h *Handler) handlePostCreate(c echo.Context) error {
	var posts []forumModel.PostCreate
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
