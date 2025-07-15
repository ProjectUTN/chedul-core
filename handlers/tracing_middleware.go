package handlers

import (
	"chedul-core/util"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TracingMiddleware(serviceName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			logger := initLogger()

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

func initLogger() *zap.Logger {
	var config zap.Config
	if util.IsEnvProd() {
		config = zap.NewProductionConfig()
		config.DisableStacktrace = true
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger, _ := config.Build()
	return logger
}
