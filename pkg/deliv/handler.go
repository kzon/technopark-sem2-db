package deliv

import (
	"encoding/json"
	"errors"
	"github.com/kzon/technopark-sem2-db/pkg/consts"
	"github.com/valyala/fasthttp"
	"net/http"
)

func PathParam(c *fasthttp.RequestCtx, param string) string {
	return c.UserValue(param).(string)
}

func QueryParam(c *fasthttp.RequestCtx, param string) string {
	return string(c.FormValue(param))
}

func Ok(c *fasthttp.RequestCtx, body interface{}) {
	sendJSON(c, http.StatusOK, body)
}

func Created(c *fasthttp.RequestCtx, body interface{}) {
	sendJSON(c, http.StatusCreated, body)
}

func BadRequest(c *fasthttp.RequestCtx, err error) {
	sendMessage(c, http.StatusBadRequest, err)
}

func Error(c *fasthttp.RequestCtx, err error) {
	if errors.Is(err, consts.ErrNotFound) {
		notFound(c, err)
		return
	}
	if errors.Is(err, consts.ErrConflict) {
		ConflictWithMessage(c, err)
		return
	}
	if err != nil {
		internalError(c, err)
		return
	}
	Ok(c, "")
}

func notFound(c *fasthttp.RequestCtx, err error) {
	sendMessage(c, http.StatusNotFound, err)
}

func internalError(c *fasthttp.RequestCtx, err error) {
	sendMessage(c, http.StatusInternalServerError, err)
}

func Conflict(c *fasthttp.RequestCtx, body interface{}) {
	sendJSON(c, http.StatusConflict, body)
}

func ConflictWithMessage(c *fasthttp.RequestCtx, err error) {
	sendMessage(c, http.StatusConflict, err)
}

func sendJSON(c *fasthttp.RequestCtx, status int, body interface{}) {
	c.SetStatusCode(status)
	c.SetContentType("application/json")
	response, _ := json.Marshal(body)
	c.Write(response)
}

func sendMessage(c *fasthttp.RequestCtx, status int, err error) {
	sendJSON(c, status, map[string]string{
		"message": err.Error(),
	})
}
