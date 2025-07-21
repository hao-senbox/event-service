package core

import (
	"go.uber.org/zap"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

type StreamCore struct{}

const TypeCoreStream = "stream"

func (c *StreamCore) Create(cfg *viper.Viper) (_ zapcore.Core, err error) {
	var encoder zapcore.Encoder
	if encoder, err = getEncoder(cfg); err != nil {
		return nil, err
	}

	var level zapcore.Level
	if level, err = getLevel(cfg); err != nil {
		return nil, err
	}

	return zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), zap.NewAtomicLevelAt(level)), nil
}
