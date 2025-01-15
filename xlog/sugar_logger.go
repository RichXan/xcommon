package xlog

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var sugaredlogger *zap.SugaredLogger

// 设置日志级别、输出格式和日志文件的路径
func SetLogs(logLevel string, logFormat, fileName string, skip int) {
	var core zapcore.Core

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        TIME_KEY,
		LevelKey:       LEVLE_KEY,
		NameKey:        NAME_KEY,
		CallerKey:      CALLER_KEY,
		MessageKey:     MESSAGE_KEY,
		StacktraceKey:  STACKTRACE_KEY,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // 大写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短路径编码器(相对路径+行号)
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志输出格式
	var encoder zapcore.Encoder
	switch logFormat {
	case LOGFORMAT_JSON:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	if len(fileName) > 0 {
		fmt.Println("init logger to file")
		// 添加日志切割归档功能
		hook := lumberjack.Logger{
			Filename:   fileName,    // 日志文件路径
			MaxSize:    MAX_SIZE,    // 每个日志文件保存的最大尺寸 单位：M
			MaxBackups: MAX_BACKUPS, // 日志文件最多保存多少个备份
			MaxAge:     MAX_AGE,     // 文件最多保存多少天
			Compress:   true,        // 是否压缩
		}

		core = zapcore.NewCore(
			encoder, // 编码器配置
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stderr), zapcore.AddSync(&hook)), // 打印到控制台和文件
			zap.NewAtomicLevelAt(levelMap[logLevel]),                                        // 日志级别
		)
	} else {
		fmt.Println("init logger to stdout")
		syncWriter := zapcore.AddSync(os.Stdout)
		core = zapcore.NewCore(encoder, syncWriter, zap.NewAtomicLevelAt(levelMap[logLevel]))
	}

	// 开启文件及行号
	caller := zap.AddCaller()
	// 开启开发模式，堆栈跟踪
	// development := zap.Development()
	// 构造日志
	logger := zap.New(core, caller, zap.AddCallerSkip(skip))

	sugaredlogger = logger.Sugar()

	// 将自定义的logger替换为全局的logger
	zap.ReplaceGlobals(logger)
}

func Debug(args ...interface{}) {
	sugaredlogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	sugaredlogger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	sugaredlogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	sugaredlogger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	sugaredlogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	sugaredlogger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	sugaredlogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	sugaredlogger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	sugaredlogger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	sugaredlogger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	sugaredlogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	sugaredlogger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	sugaredlogger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	sugaredlogger.Fatalf(template, args...)
}
