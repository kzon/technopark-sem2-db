package api

import (
	"encoding/json"
	"errors"
	"github.com/buaazp/fasthttprouter"
	apiModel "github.com/kzon/technopark-sem2-db/pkg/api/model"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/deliv"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

type Handler struct {
	usecase *Usecase
	router  *fasthttprouter.Router
}

func NewHandler(usecase Usecase) Handler {
	h := Handler{
		usecase: &usecase,
		router:  fasthttprouter.New(),
	}

	h.router.POST("/api/user/:nickname/create", h.handleUserCreate)
	h.router.GET("/api/user/:nickname/profile", h.handleGetUserProfile)
	h.router.POST("/api/user/:nickname/profile", h.handleUserUpdate)

	h.router.POST("/api/forum/:slug/create", h.handleForumOrThreadCreate)
	h.router.GET("/api/forum/:slug/details", h.handleGetForumDetails)
	h.router.GET("/api/forum/:slug/threads", h.handleGetForumThreads)
	h.router.GET("/api/forum/:slug/users", h.handleGetForumUsers)

	h.router.POST("/api/thread/:slug_or_id/create", h.handlePostCreate)
	h.router.POST("/api/thread/:slug_or_id/vote", h.handleVoteForThread)
	h.router.GET("/api/thread/:slug_or_id/details", h.handleGetThreadDetails)
	h.router.POST("/api/thread/:slug_or_id/details", h.handleThreadUpdate)
	h.router.GET("/api/thread/:slug_or_id/posts", h.handleGetThreadPosts)

	h.router.GET("/api/post/:id/details", h.handleGetPostDetails)
	h.router.POST("/api/post/:id/details", h.handlePostUpdate)

	h.router.GET("/api/service/status", h.handleStatus)
	h.router.POST("/api/service/clear", h.handleClear)

	return h
}

func (h *Handler) GetHandleFunc() fasthttp.RequestHandler {
	return h.router.Handler
}

func (h *Handler) handleUserCreate(c *fasthttp.RequestCtx) {
	u := apiModel.UserInput{}
	if err := json.Unmarshal(c.PostBody(), &u); err != nil {
		deliv.BadRequest(c, err)
		return
	}
	users, err := h.usecase.createUser(deliv.PathParam(c, "nickname"), u.Email, u.Fullname, u.About)
	if errors.Is(err, consts.ErrConflict) {
		deliv.Conflict(c, users)
		return
	}
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Created(c, users[0])
}

func (h *Handler) handleGetUserProfile(c *fasthttp.RequestCtx) {
	u, err := h.usecase.getUserByNickname(deliv.PathParam(c, "nickname"))
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, u)
}

func (h *Handler) handleUserUpdate(c *fasthttp.RequestCtx) {
	u := apiModel.UserInput{}
	if err := json.Unmarshal(c.PostBody(), &u); err != nil {
		deliv.BadRequest(c, err)
		return
	}
	nick := deliv.PathParam(c, "nickname")
	user, err := h.usecase.updateUser(nick, u.Email, u.Fullname, u.About)
	if errors.Is(err, consts.ErrConflict) {
		deliv.ConflictWithMessage(c, err)
		return
	}
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, user)
}

func (h *Handler) handleForumOrThreadCreate(c *fasthttp.RequestCtx) {
	// Hack fasthttprouter path params matching
	if string(c.Path()) == "/api/forum/create" {
		h.handleForumCreate(c)
	} else {
		h.handleThreadCreate(c)
	}
}

func (h *Handler) handleForumCreate(c *fasthttp.RequestCtx) {
	forumToCreate := apiModel.ForumCreate{}
	if err := json.Unmarshal(c.PostBody(), &forumToCreate); err != nil {
		deliv.BadRequest(c, err)
		return
	}
	forum, err := h.usecase.createForum(forumToCreate.Title, forumToCreate.Slug, forumToCreate.User)
	if errors.Is(err, consts.ErrConflict) {
		deliv.Conflict(c, forum)
		return
	}
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Created(c, forum)
}

func (h *Handler) handleThreadCreate(c *fasthttp.RequestCtx) {
	thread := apiModel.ThreadCreate{}
	if err := json.Unmarshal(c.PostBody(), &thread); err != nil {
		deliv.BadRequest(c, err)
		return
	}
	forum := deliv.PathParam(c, "slug")
	result, err := h.usecase.createThread(forum, thread)
	if errors.Is(err, consts.ErrConflict) {
		deliv.Conflict(c, result)
		return
	}
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Created(c, result)
}

func (h *Handler) handleGetForumDetails(c *fasthttp.RequestCtx) {
	slug := deliv.PathParam(c, "slug")
	forum, err := h.usecase.getForum(slug)
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, forum)
}

func (h *Handler) handleGetForumThreads(c *fasthttp.RequestCtx) {
	limit, _ := strconv.Atoi(deliv.QueryParam(c, "limit"))
	desc, _ := strconv.ParseBool(deliv.QueryParam(c, "desc"))
	threads, err := h.usecase.getForumThreads(deliv.PathParam(c, "slug"), deliv.QueryParam(c, "since"), limit, desc)
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, threads)
}

func (h *Handler) handleGetForumUsers(c *fasthttp.RequestCtx) {
	limit, _ := strconv.Atoi(deliv.QueryParam(c, "limit"))
	desc, _ := strconv.ParseBool(deliv.QueryParam(c, "desc"))
	users, err := h.usecase.getForumUsers(deliv.PathParam(c, "slug"), deliv.QueryParam(c, "since"), limit, desc)
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, users)
}

func (h *Handler) handlePostCreate(c *fasthttp.RequestCtx) {
	var posts []*apiModel.PostCreate
	if err := json.Unmarshal(c.PostBody(), &posts); err != nil {
		deliv.BadRequest(c, err)
		return
	}
	result, err := h.usecase.createPosts(deliv.PathParam(c, "slug_or_id"), posts)
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Created(c, result)
}

func (h *Handler) handleVoteForThread(c *fasthttp.RequestCtx) {
	var vote apiModel.Vote
	if err := json.Unmarshal(c.PostBody(), &vote); err != nil {
		deliv.BadRequest(c, err)
		return
	}
	thread, err := h.usecase.voteForThread(deliv.PathParam(c, "slug_or_id"), vote)
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, thread)
}

func (h *Handler) handleGetThreadDetails(c *fasthttp.RequestCtx) {
	thread, err := h.usecase.getThread(deliv.PathParam(c, "slug_or_id"))
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, thread)
}

func (h *Handler) handleThreadUpdate(c *fasthttp.RequestCtx) {
	t := apiModel.ThreadUpdate{}
	if err := json.Unmarshal(c.PostBody(), &t); err != nil {
		deliv.BadRequest(c, err)
		return
	}
	thread, err := h.usecase.updateThread(deliv.PathParam(c, "slug_or_id"), t.Message, t.Title)
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, thread)
}

func (h *Handler) handleGetThreadPosts(c *fasthttp.RequestCtx) {
	sp := deliv.QueryParam(c, "since")
	var since *int = nil
	if sp != "" {
		n, _ := strconv.Atoi(sp)
		since = &n
	}
	limit, _ := strconv.Atoi(deliv.QueryParam(c, "limit"))
	desc, _ := strconv.ParseBool(deliv.QueryParam(c, "desc"))
	posts, err := h.usecase.getThreadPosts(
		deliv.PathParam(c, "slug_or_id"),
		limit,
		since,
		deliv.QueryParam(c, "sort"),
		desc,
	)
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, posts)
}

func (h *Handler) handleGetPostDetails(c *fasthttp.RequestCtx) {
	id, _ := strconv.Atoi(deliv.PathParam(c, "id"))
	related := strings.Split(deliv.QueryParam(c, "related"), ",")
	details, err := h.usecase.getPostDetails(id, related)
	if err != nil {
		deliv.Error(c, err)
		return
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
	deliv.Ok(c, result)
}

func (h *Handler) handlePostUpdate(c *fasthttp.RequestCtx) {
	t := apiModel.PostUpdate{}
	if err := json.Unmarshal(c.PostBody(), &t); err != nil {
		deliv.BadRequest(c, err)
		return
	}
	id, _ := strconv.Atoi(deliv.PathParam(c, "id"))
	thread, err := h.usecase.updatePost(id, t.Message)
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, thread)
}

func (h *Handler) handleStatus(c *fasthttp.RequestCtx) {
	status, err := h.usecase.getStatus()
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, status)
}

func (h *Handler) handleClear(c *fasthttp.RequestCtx) {
	err := h.usecase.clear()
	if err != nil {
		deliv.Error(c, err)
		return
	}
	deliv.Ok(c, nil)
	return
}
