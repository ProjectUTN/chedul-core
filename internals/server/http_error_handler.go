package server

import (
	"chedul-core/internals/handlers"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HttpErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if apiErr, ok := err.(handlers.ApiError); ok {
		c.JSON(apiErr.StatusCode, apiErr)
	} else if httpErr, ok := err.(*echo.HTTPError); ok {
		errResp := map[string]any{
			"statusCode": httpErr.Code,
			"msg":        httpErr.Message,
		}
		c.JSON(httpErr.Code, errResp)
	} else {
		errResp := map[string]any{
			"statusCode": http.StatusInternalServerError,
			"msg":        "internal server error",
		}
		c.JSON(http.StatusInternalServerError, errResp)
	}

	c.Logger().Error("HTTP API Error ", "err ", err.Error(), "path ", c.Request().URL.Path)

}
