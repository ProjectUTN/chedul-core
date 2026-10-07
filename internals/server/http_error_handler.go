package server

import (
	"chedul-core/internals/handlers"
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HttpErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var apiErr handlers.ApiError
	var httpErr *echo.HTTPError

	switch {
	case errors.As(err, &apiErr):
		c.JSON(apiErr.StatusCode, apiErr)
	case errors.As(err, &httpErr):
		c.JSON(httpErr.Code, map[string]any{
			"statusCode": httpErr.Code,
			"msg":        httpErr.Message,
		})
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, map[string]any{
			"statusCode": http.StatusNotFound,
			"msg":        "no encontrado",
		})
	default:
		c.JSON(http.StatusInternalServerError, map[string]any{
			"statusCode": http.StatusInternalServerError,
			"msg":        "internal server error",
		})
	}
}
