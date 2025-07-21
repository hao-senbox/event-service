// core/file.go

package core

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"runtime"
	"strings"
)

type FileCore struct{}

const TypeCoreFile = "file"

func (c *FileCore) Create(cfg *viper.Viper) (zapcore.Core, error) {
	filePath := cfg.GetString("file.path")
	if filePath == "" {
		return nil, fmt.Errorf("file path not specified in configuration")
	}

	maxBackups := cfg.GetInt("file.max_backups")
	maxAge := cfg.GetInt("file.max_age")
	maxSize := cfg.GetInt("file.max_size")

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := dir + filePath

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0777)
		if err != nil {
			panic(err)
		}
	}
	if strings.Contains(runtime.GOOS, "window") {
		path = path + "\\"
	} else {
		path = path + "/"
	}

	// Create a new log file
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   path + "log.log",
		MaxSize:    maxSize, // megabytes
		MaxAge:     maxAge,  // days
		MaxBackups: maxBackups,
		LocalTime:  true,
		Compress:   true,
	})

	encoder, err := getEncoder(cfg)
	if err != nil {
		return nil, err
	}

	level, err := getLevel(cfg)
	if err != nil {
		return nil, err
	}

	return zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stderr), zapcore.AddSync(writer)), zap.NewAtomicLevelAt(level)), nil
}
