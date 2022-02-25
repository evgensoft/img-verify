package logger

import (
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logMaxAgeDays = 30
)

var (
	logger *zap.SugaredLogger
	once   sync.Once
)

// GetLogger is getter to create new logger instance.
func GetLogger() *zap.SugaredLogger {
	once.Do(func() {
		logger = newLogger()
	})

	return logger
}

func newLogger() *zap.SugaredLogger {
	var coreArr []zapcore.Core

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	highPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev < zap.ErrorLevel && lev >= zap.DebugLevel
	})

	infoFileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./log/info.log",
		MaxSize:    1,
		MaxBackups: 10,
		MaxAge:     logMaxAgeDays,
		Compress:   false,
	})
	infoFileCore := zapcore.NewCore(encoder,
		zapcore.NewMultiWriteSyncer(infoFileWriteSyncer, zapcore.AddSync(os.Stdout)), lowPriority)

	errorFileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "./log/error.log",
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     logMaxAgeDays,
		Compress:   false,
	})
	errorFileCore := zapcore.NewCore(encoder,
		zapcore.NewMultiWriteSyncer(errorFileWriteSyncer, zapcore.AddSync(os.Stdout)), highPriority)

	coreArr = append(coreArr, infoFileCore)
	coreArr = append(coreArr, errorFileCore)
	zaplog := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller()).Sugar()

	return zaplog
}
