package handlers

import (
	"chedul-core/logger"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func TracingMiddleware(serviceName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			logger := logger.GetLogger()

			requestLogger := logger.With(
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.String("remote_addr", c.RealIP()),
			)

			c.Set("Logger", requestLogger)
			requestLogger.Info("Request started")

			err := next(c)
			duration := time.Since(start)

			status := c.Response().Status
			if err != nil {
				if httpErr, ok := err.(*echo.HTTPError); ok {
					status = httpErr.Code
				}
			}

			logFields := []zap.Field{
				zap.Int("status", status),
				zap.Duration("duration", duration),
				zap.Int64("response_size", c.Response().Size),
			}

			if err != nil {
				logFields = append(logFields, zap.Error(err))
				requestLogger.Error("Request failed", logFields...)
			} else {
				requestLogger.Info("Request completed", logFields...)
			}

			return err

		}
	}
}
