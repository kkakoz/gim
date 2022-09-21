package logger

import (
	"github.com/kkakoz/gim/pkg/conf"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
)

var once sync.Once

var logger *zap.Logger
var sugarLogger *zap.SugaredLogger

func L() *zap.Logger {
	once.Do(func() {
		initLog(conf.Conf())
	})
	return logger
}

func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

func WithFields(field ...zap.Field) *zap.Logger {
	return L().With(field...)
}

func Sugar() *zap.SugaredLogger {
	once.Do(func() {
		initLog(conf.Conf())
	})
	return sugarLogger
}

func initLog(viper *viper.Viper) {
	viper.SetDefault("log.path", "temp.log")
	viper.SetDefault("log.maxSize", 10)
	viper.SetDefault("log.maxBackups", 5)
	viper.SetDefault("log.maxAge", 30)
	viper.SetDefault("log.stdout", true)

	encoder := getEncoder()

	var core zapcore.Core
	if viper.GetBool("log.stdout") {
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core = zapcore.NewCore(consoleEncoder, os.Stdout, zapcore.DebugLevel)
	} else {
		fw := zapcore.AddSync(&lumberjack.Logger{
			Filename:   viper.GetString("log.path"),
			MaxSize:    viper.GetInt("log.maxSize"),    // 日志文件最大大小(MB)
			MaxBackups: viper.GetInt("log.maxBackups"), // 保留旧文件最大数量
			MaxAge:     viper.GetInt("log.maxAge"),     // 保留旧文件最长天数
		})
		core = zapcore.NewCore(encoder, fw, zapcore.InfoLevel)
	}
	logger = zap.New(core)
	sugarLogger = logger.Sugar()

	zap.ReplaceGlobals(logger)
}

func getEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(config)
}
