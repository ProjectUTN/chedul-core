package logger

import (
	"chedul-core/pkg/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(logLevel zapcore.Level) (*zap.Logger, error) {
	var encoderCfg zapcore.EncoderConfig
	if config.IsProd() {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}

	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(logLevel),
		Development:       true,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stderr",
		},
		ErrorOutputPaths: []string{
			"stderr",
		},
		InitialFields: map[string]any{
			"pid": os.Getpid(),
		},
	}

	return config.Build()
}
