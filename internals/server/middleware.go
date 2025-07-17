package server

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func TracingMiddleware(logger *zap.Logger, serviceName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

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

func CORSMiddleware() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"*"},
	})
}

func RecoverMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Error("panic recovered", zap.Error(err), zap.ByteString("stack", stack))
			return nil
		},
	})
}
