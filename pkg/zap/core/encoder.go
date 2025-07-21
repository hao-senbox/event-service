package core

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

func getEncoder(cfg *viper.Viper) (zapcore.Encoder, error) {
	encoderType := cfg.GetString("encoding")
	if encoderType == "" {
		encoderType = "json" // Default to JSON if not specified
	}

	var encoderConfig zapcore.EncoderConfig
	encoderConfig.TimeKey = "[TIME]"
	encoderConfig.LevelKey = "[LEVEL]"
	encoderConfig.NameKey = "[SERVICE]"
	encoderConfig.MessageKey = "[MESSAGE]"
	encoderConfig.StacktraceKey = "[STACKTRACE]"
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	switch encoderType {
	case "json":
		encoderConfig.FunctionKey = "[CALLER]"
		encoderConfig.EncodeName = zapcore.FullNameEncoder
		return zapcore.NewJSONEncoder(encoderConfig), nil
	case "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		encoderConfig.ConsoleSeparator = " | "
		return zapcore.NewConsoleEncoder(encoderConfig), nil
	default:
		return nil, fmt.Errorf("unsupported encoder type: %s", encoderType)
	}
}
