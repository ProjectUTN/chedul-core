package logger

import (
	"chedul-core/util"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger() error {
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

	var err error
	Logger, err = config.Build()

	return err
}

func GetLogger() *zap.Logger {
	return Logger
}

func GetLoggerFromCtx(c echo.Context) *zap.Logger {
	if logger, ok := c.Get("Logger").(*zap.Logger); ok {
		return logger
	}

	return GetLogger()
}
