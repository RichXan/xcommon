package xlog

import "github.com/rs/zerolog"

var sugarLogger *zerolog.Logger

func Debug() *zerolog.Event {
	return sugarLogger.Debug()
}

func Info() *zerolog.Event {
	return sugarLogger.Info()
}

func Error() *zerolog.Event {
	return sugarLogger.Error()
}

func Warn() *zerolog.Event {
	return sugarLogger.Warn()
}

func Panic() *zerolog.Event {
	return sugarLogger.Panic()
}

func Fatal() *zerolog.Event {
	return sugarLogger.Fatal()
}
