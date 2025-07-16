package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type AppContext struct {
	echo.Context
	conn   *bun.DB
	logger *zap.Logger
}

func AppContextMiddleware(db *bun.DB, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &AppContext{
				Context: c,
				conn:    db,
				logger:  logger,
			}
			return next(cc)
		}
	}
}
