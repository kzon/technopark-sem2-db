package delivery

import (
	"errors"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/labstack/echo"
	"net/http"
)

func Ok(c echo.Context, body interface{}) error {
	return c.JSON(http.StatusOK, body)
}

func Created(c echo.Context, body interface{}) error {
	return c.JSON(http.StatusCreated, body)
}

func BadRequest(c echo.Context, err error) error {
	return responseWithMessage(c, http.StatusBadRequest, err)
}

func Error(c echo.Context, err error) error {
	if errors.Is(err, consts.ErrNotFound) {
		return notFound(c, err)
	}
	if errors.Is(err, consts.ErrConflict) {
		return ConflictWithMessage(c, err)
	}
	if err != nil {
		return internalError(c, err)
	}
	return Ok(c, "")
}

func notFound(c echo.Context, err error) error {
	return responseWithMessage(c, http.StatusNotFound, err)
}

func internalError(c echo.Context, err error) error {
	return responseWithMessage(c, http.StatusInternalServerError, err)
}

func Conflict(c echo.Context, body interface{}) error {
	return c.JSON(http.StatusConflict, body)
}

func ConflictWithMessage(c echo.Context, err error) error {
	return responseWithMessage(c, http.StatusConflict, err)
}

func responseWithMessage(c echo.Context, status int, err error) error {
	return c.JSON(status, map[string]string{
		"message": err.Error(),
	})
}
