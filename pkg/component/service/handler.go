package service

import (
	"github.com/kzon/technopark-sem2-db/pkg/delivery"
	"github.com/labstack/echo"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(e *echo.Echo, usecase Usecase) Handler {
	h := Handler{usecase: usecase}
	e.GET("/api/service/status", h.handleStatus)
	e.POST("/api/service/clear", h.handleClear)
	return h
}

func (h *Handler) handleStatus(c echo.Context) error {
	status, err := h.usecase.getStatus()
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, status)
}

func (h *Handler) handleClear(c echo.Context) error {
	err := h.usecase.clear()
	if err != nil {
		return delivery.Error(c, err)
	}
	return delivery.Ok(c, nil)
}
