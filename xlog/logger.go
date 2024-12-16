package xlog

import (
	"github.com/rs/zerolog"
)

// 打印调用函数时，跳过栈帧的数量
const ZeroLogEventCallerSkipFrameCount = 2
const DefaultLoggerName = "x_logger"

type Logger struct {
	zeroLoger *zerolog.Logger
	context   map[string]interface{}
	Config    LoggerConfig
}

func (l *Logger) doLogEvent(zeroLogEventFunc func() *zerolog.Event) *zerolog.Event {
	// ZeroEventCallerSkipFrameCount 打印上一个调用函数的文件和行号
	var e *zerolog.Event
	if len(l.context) == 0 {
		e = zeroLogEventFunc().Caller(ZeroLogEventCallerSkipFrameCount)
	} else {
		e = zeroLogEventFunc().Caller(ZeroLogEventCallerSkipFrameCount).Fields(l.context)
	}
	return e
}

func (l *Logger) Debug() *zerolog.Event {
	return l.doLogEvent(l.zeroLoger.Debug)
}

func (l *Logger) Info() *zerolog.Event {
	return l.doLogEvent(l.zeroLoger.Info)
}

func (l *Logger) Error() *zerolog.Event {
	return l.doLogEvent(l.zeroLoger.Error)
}

func (l *Logger) Panic() *zerolog.Event {
	return l.doLogEvent(l.zeroLoger.Panic)
}

// 可以把 request id ，uin 等放到 context 里面，统一打印
func (l *Logger) ContextLogger(ctx map[string]interface{}) *Logger {
	al := Logger{zeroLoger: l.zeroLoger, context: ctx}
	return &al
}

func NewLogger(cfg LoggerConfig) *Logger {
	if cfg.LoggerName == "" {
		cfg.LoggerName = DefaultLoggerName
	}
	zl := newZeroLogger(cfg)
	return &Logger{zeroLoger: &zl, Config: cfg}
}
