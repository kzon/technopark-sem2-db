package forum

import (
	"errors"
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
	return handler
}

func (h Handler) handleForumCreate(c echo.Context) error {
	forumToCreate := forumCreate{}
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
	thread := threadCreate{}
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
	var posts []postCreate
	if err := c.Bind(&posts); err != nil {
		return delivery.BadRequest(c, err)
	}
	thread := c.Param("slug_or_id")
	result, err := h.usecase.createPosts(thread, posts)
	if errors.Is(err, consts.ErrConflict) {
		return delivery.Conflict(c, result)
	}
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Created(c, result)
}
