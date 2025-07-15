package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func HttpErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if apiErr, ok := err.(ApiError); ok {
		c.JSON(apiErr.StatusCode, apiErr)
	} else {
		errResp := map[string]any{
			"statusCode": http.StatusInternalServerError,
			"msg":        "internal server error",
		}
		c.JSON(http.StatusInternalServerError, errResp)
	}

	c.Logger().Error("HTTP API Error ", "err ", err.Error(), " path:", c.Request().URL.Path)

}
