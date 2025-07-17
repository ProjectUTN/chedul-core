package server

import (
	"chedul-core/internals/handlers"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
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

	slog.Error("HTTP API Error ",
		zap.String("err", err.Error()),
		zap.String("path", c.Request().URL.Path),
	)
}
