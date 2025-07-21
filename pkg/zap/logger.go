package zap

import (
	"os"
	"event-service/config"
	"event-service/pkg/constants"
	"time"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger methods interface
type Logger interface {
	GetLogger() *zap.Logger
	Sync() error
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	WarnMsg(msg string, err error)
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Err(msg string, err error)
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Printf(template string, args ...interface{})
	WithName(name string)
	HttpMiddlewareAccessLogger(method string, uri string, status int, size int64, time time.Duration)
	GrpcMiddlewareAccessLogger(method string, time time.Duration, metaData map[string][]string, err error)
	GrpcClientInterceptorLogger(method string, req interface{}, reply interface{}, time time.Duration, metaData map[string][]string, err error)
	KafkaProcessMessage(topic string, partition int, message string, workerID int, offset int64, time time.Time)
	KafkaLogCommittedMessage(topic string, partition int, offset int64)
	ProjectionEvent(projectionName string, groupName string, event *esdb.ResolvedEvent, workerID int)
}

// Application logger
type appLogger struct {
	level       string
	devMode     bool
	sugarLogger *zap.SugaredLogger
	logger      *zap.Logger
}

func New(cfg *config.Config) (*appLogger, error) {
	// Create a new Viper instance
	v := viper.New()
	v.Set("zap.development", cfg.Zap.Development)
	v.Set("zap.caller", cfg.Zap.Caller)
	v.Set("zap.stacktrace", cfg.Zap.Stacktrace)
	v.Set("zap.cores.console.type", cfg.Zap.Cores.Console.Type)
	v.Set("zap.cores.console.level", cfg.Zap.Cores.Console.Level)
	v.Set("zap.cores.console.encoding", cfg.Zap.Cores.Console.Encoding)

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "[TIME]",
		LevelKey:       "[LEVEL]",
		NameKey:        "[SERVICE]",
		MessageKey:     "[MESSAGE]",
		StacktraceKey:  "[STACKTRACE]",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	// Create encoder based on configuration
	var encoder zapcore.Encoder
	if cfg.Zap.Cores.Console.Encoding == "json" {
		encoderConfig.FunctionKey = "[CALLER]"
		encoderConfig.EncodeName = zapcore.FullNameEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		encoderConfig.ConsoleSeparator = " | "
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create core
	level := zapcore.InfoLevel
	if cfg.Zap.Cores.Console.Level == "debug" {
		level = zapcore.DebugLevel
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(zapcore.AddSync(zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)))),
		zap.NewAtomicLevelAt(level),
	)

	// Create logger options
	var opts []zap.Option
	if cfg.Zap.Development {
		opts = append(opts, zap.Development())
	}
	if cfg.Zap.Caller {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	if cfg.Zap.Stacktrace != "" {
		var stackLevel = zap.NewAtomicLevel()
		if err := stackLevel.UnmarshalText([]byte(cfg.Zap.Stacktrace)); err != nil {
			return nil, err
		}
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	// Create logger
	logger := zap.New(core, opts...)
	return &appLogger{
		level:       cfg.Zap.Cores.Console.Level,
		devMode:     cfg.Zap.Development,
		logger:      logger,
		sugarLogger: logger.Sugar(),
	}, nil
}

// GetLogger methods
func (l *appLogger) GetLogger() *zap.Logger {
	return l.logger
}

// WithName add logger microservice name
func (l *appLogger) WithName(name string) {
	l.logger = l.logger.Named(name)
	l.sugarLogger = l.sugarLogger.Named(name)
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *appLogger) Debug(args ...interface{}) {
	l.sugarLogger.Debug(args...)
}

// Debugf uses fmt.Sprintf to log a templated message
func (l *appLogger) Debugf(template string, args ...interface{}) {
	l.sugarLogger.Debugf(template, args...)
}

// Info uses fmt.Sprint to construct and log a message
func (l *appLogger) Info(args ...interface{}) {
	l.sugarLogger.Info(args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *appLogger) Infof(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

// Printf uses fmt.Sprintf to log a templated message
func (l *appLogger) Printf(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *appLogger) Warn(args ...interface{}) {
	l.sugarLogger.Warn(args...)
}

// WarnMsg log error message with warn level.
func (l *appLogger) WarnMsg(msg string, err error) {
	l.logger.Warn(msg, zap.String("error", err.Error()))
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *appLogger) Warnf(template string, args ...interface{}) {
	l.sugarLogger.Warnf(template, args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *appLogger) Error(args ...interface{}) {
	l.sugarLogger.Error(args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *appLogger) Errorf(template string, args ...interface{}) {
	l.sugarLogger.Errorf(template, args...)
}

// Err uses error to log a message.
func (l *appLogger) Err(msg string, err error) {
	l.logger.Error(msg, zap.Error(err))
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the logger then panics. (See DPanicLevel for details.)
func (l *appLogger) DPanic(args ...interface{}) {
	l.sugarLogger.DPanic(args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the logger then panics. (See DPanicLevel for details.)
func (l *appLogger) DPanicf(template string, args ...interface{}) {
	l.sugarLogger.DPanicf(template, args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (l *appLogger) Panic(args ...interface{}) {
	l.sugarLogger.Panic(args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics
func (l *appLogger) Panicf(template string, args ...interface{}) {
	l.sugarLogger.Panicf(template, args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (l *appLogger) Fatal(args ...interface{}) {
	l.sugarLogger.Fatal(args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (l *appLogger) Fatalf(template string, args ...interface{}) {
	l.sugarLogger.Fatalf(template, args...)
}

// Sync flushes any buffered log entries
func (l *appLogger) Sync() error {
	go l.logger.Sync() // nolint: errcheck
	return l.sugarLogger.Sync()
}

func (l *appLogger) HttpMiddlewareAccessLogger(method, uri string, status int, size int64, time time.Duration) {
	l.logger.Info(
		constants.HTTP,
		zap.String(constants.METHOD, method),
		zap.String(constants.URI, uri),
		zap.Int(constants.STATUS, status),
		zap.Int64(constants.SIZE, size),
		zap.Duration(constants.TIME, time),
	)
}

func (l *appLogger) GrpcMiddlewareAccessLogger(method string, time time.Duration, metaData map[string][]string, err error) {
	if err != nil {
		l.logger.Info(
			constants.GRPC,
			zap.String(constants.METHOD, method),
			zap.Duration(constants.TIME, time),
			zap.Any(constants.METADATA, metaData),
			zap.String(constants.ERROR, err.Error()),
		)
		return
	}
	l.logger.Info(constants.GRPC, zap.String(constants.METHOD, method), zap.Duration(constants.TIME, time), zap.Any(constants.METADATA, metaData))
}

func (l *appLogger) GrpcClientInterceptorLogger(method string, req, reply interface{}, time time.Duration, metaData map[string][]string, err error) {
	if err != nil {
		l.logger.Info(
			constants.GRPC,
			zap.String(constants.METHOD, method),
			zap.Any(constants.REQUEST, req),
			zap.Any(constants.REPLY, reply),
			zap.Duration(constants.TIME, time),
			zap.Any(constants.METADATA, metaData),
			zap.String(constants.ERROR, err.Error()),
		)
		return
	}
	l.logger.Info(
		constants.GRPC,
		zap.String(constants.METHOD, method),
		zap.Any(constants.REQUEST, req),
		zap.Any(constants.REPLY, reply),
		zap.Duration(constants.TIME, time),
		zap.Any(constants.METADATA, metaData),
	)
}

func (l *appLogger) KafkaProcessMessage(topic string, partition int, message string, workerID int, offset int64, time time.Time) {
	l.logger.Debug(
		"Processing Kafka message",
		zap.String(constants.Topic, topic),
		zap.Int(constants.Partition, partition),
		zap.String(constants.Message, message),
		zap.Int(constants.WorkerID, workerID),
		zap.Int64(constants.Offset, offset),
		zap.Time(constants.Time, time),
	)
}

func (l *appLogger) KafkaLogCommittedMessage(topic string, partition int, offset int64) {
	l.logger.Info(
		"Committed Kafka message",
		zap.String(constants.Topic, topic),
		zap.Int(constants.Partition, partition),
		zap.Int64(constants.Offset, offset),
	)
}

func (l *appLogger) ProjectionEvent(projectionName string, groupName string, event *esdb.ResolvedEvent, workerID int) {
	l.logger.Debug(
		projectionName,
		zap.String(constants.GroupName, groupName),
		zap.String(constants.StreamID, event.OriginalEvent().StreamID),
		zap.String(constants.EventID, event.OriginalEvent().EventID.String()),
		zap.String(constants.EventType, event.OriginalEvent().EventType),
		zap.Uint64(constants.EventNumber, event.OriginalEvent().EventNumber),
		zap.Time(constants.CreatedDate, event.OriginalEvent().CreatedDate),
		zap.String(constants.UserMetadata, string(event.OriginalEvent().UserMetadata)),
		zap.Int(constants.WorkerID, workerID),
	)
}
