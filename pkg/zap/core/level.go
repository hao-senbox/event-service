package core

import (
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// For mapping config logger to email_service logger levels
var loggerLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func getLevel(cfg *viper.Viper) (zapcore.Level, error) {
	levelCfg := cfg.GetString("level") // Assuming "level" key in your config
	level, exist := loggerLevelMap[levelCfg]
	if !exist {
		level = zapcore.DebugLevel
	}

	return level, nil
}
