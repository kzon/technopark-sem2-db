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

func (h *Handler) handleUserCreate(c echo.Context) error {
	u := userInput{}
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
	u := userInput{}
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
