package core

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

type Core interface {
	Create(cfg *viper.Viper) (zapcore.Core, error)
}

func Create(cfg *viper.Viper, path string) (zapcore.Core, error) {
	cfgCore := cfg.Sub(path)
	if cfgCore == nil {
		return nil, fmt.Errorf("core config at path '%s' not found", path)
	}

	cfgCore.SetDefault("type", TypeCoreStream)
	cfgCore.SetDefault("level", "debug")

	var core Core
	coreType := cfgCore.GetString("type")
	switch coreType {
	case TypeCoreStream:
		core = &StreamCore{}
	case TypeCoreFile:
		core = &FileCore{}
	default:
		return nil, fmt.Errorf("unsupported core type: %s", coreType)
	}

	return core.Create(cfgCore)
}
