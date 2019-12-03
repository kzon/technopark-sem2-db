package user

import (
	"errors"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/kzon/technopark-sem2-db/pkg/delivery"
	"github.com/labstack/echo"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(e *echo.Echo, usecase Usecase) Handler {
	handler := Handler{usecase: usecase}
	e.POST("/api/user/:nickname/create", handler.handleUserCreate)
	e.GET("/api/user/:nickname/profile", handler.handleGetUserProfile)
	e.POST("/api/user/:nickname/profile", handler.handleUserUpdate)
	return handler
}

const nicknameRequestParam = "nickname"

func (h *Handler) handleUserCreate(c echo.Context) error {
	userToInput := userToInput{}
	if err := c.Bind(&userToInput); err != nil {
		return delivery.BadRequest(c, err)
	}
	nickname := c.Param(nicknameRequestParam)
	users, err := h.usecase.createUser(nickname, userToInput.Email, userToInput.Fullname, userToInput.About)
	if errors.Is(err, consts.ErrConflict) {
		return delivery.Conflict(c, users)
	}
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Created(c, users[0])
}

func (h *Handler) handleGetUserProfile(c echo.Context) error {
	nickname := c.Param(nicknameRequestParam)
	user, err := h.usecase.getUserByNickname(nickname)
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, user)
}

func (h *Handler) handleUserUpdate(c echo.Context) error {
	userToInput := userToInput{}
	if err := c.Bind(&userToInput); err != nil {
		return delivery.BadRequest(c, err)
	}
	nickname := c.Param(nicknameRequestParam)
	user, err := h.usecase.updateUser(nickname, userToInput.Email, userToInput.Fullname, userToInput.About)
	if errors.Is(err, consts.ErrConflict) {
		return delivery.ConflictWithMessage(c, err)
	}
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, user)
}
