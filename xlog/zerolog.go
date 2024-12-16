package xlog

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/richxan/xcommon/xutil"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func newZeroLogger(cfg LoggerConfig) zerolog.Logger {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		panic(fmt.Errorf("fatal error, init logger error: Parse log level error %s ", err.Error()))
	}
	zerolog.SetGlobalLevel(level)

	var writers []io.Writer
	writers = append(writers, os.Stdout)
	if cfg.SaveLoggerAsFile {
		writers = append(writers, newRollingFile(cfg.Directory, cfg.ProjectName, cfg.LoggerName, cfg.MaxSize, cfg.MaxBackups))
	}
	mw := io.MultiWriter(writers...)
	l := zerolog.New(mw).With().Timestamp().Logger()
	return l
}

// 创建文件
func newRollingFile(dir, projectName, loggerName string, maxSize, maxBackups int) io.Writer {
	if dir == "" || projectName == "" || loggerName == "" {
		panic(fmt.Errorf("fatal error, init logger error: log director or project name is nil "))
	}

	loggerNameLen := len(loggerName)
	if loggerName[loggerNameLen-4:loggerNameLen-1] != ".log" {
		loggerName = loggerName + ".log"
	}

	// make sure the log file permission is 644
	filename := path.Join(dir, projectName, loggerName)
	if err := xutil.SetFileModeWithCreating(filename, fs.FileMode(0644)); err != nil {
		panic(fmt.Errorf("fatal error, init logger error: Set log file mode error %s ", err))
	}

	return &lumberjack.Logger{
		Filename:   filename,   //日志文件
		MaxBackups: maxBackups, //保留旧文件的最大数量
		MaxSize:    maxSize,    //单文件最大容量(单位MB)
		Compress:   false,
	}
}
