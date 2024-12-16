package xlog

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// 设置测试配置
	cfg := LoggerConfig{
		Level:            "debug",
		Directory:        "../../logs",
		ProjectName:      "test",
		LoggerName:       "test.log",
		MaxSize:          10,
		MaxBackups:       3,
		SaveLoggerAsFile: false,
	}

	// 创建日志实例
	logger := NewLogger(cfg)
	assert.NotNil(t, logger)

	// 测试基本日志级别
	t.Run("Test log levels", func(t *testing.T) {
		logger.Debug().Msg("This is a debug message")
		logger.Info().Msg("This is an info message")
		logger.Error().Msg("This is an error message")
	})

	// 测试带上下文的日志
	t.Run("Test context logger", func(t *testing.T) {
		ctx := map[string]interface{}{
			"request_id": "123456",
			"user_id":    "user_789",
		}

		contextLogger := logger.ContextLogger(ctx)
		assert.NotNil(t, contextLogger)

		contextLogger.Info().Str("action", "test").Msg("This is a context log message")
	})

	// 测试控制台输出
	t.Run("Test console writer", func(t *testing.T) {
		consoleLogger := NewLogger(LoggerConfig{
			Level:            "debug",
			SaveLoggerAsFile: false,
		})
		consoleLogger.Info().Msg("This is a console message")
	})

	// 清理测试文件
	t.Cleanup(func() {
		os.RemoveAll("../../logs/test")
	})
}

// 测试日志配置验证
func TestLoggerConfig(t *testing.T) {
	t.Run("Test invalid log level", func(t *testing.T) {
		cfg := LoggerConfig{
			Level:       "invalid_level",
			Directory:   "../../logs",
			ProjectName: "test",
			LoggerName:  "test.log",
		}

		assert.Panics(t, func() {
			NewLogger(cfg)
		})
	})

	t.Run("Test default logger name", func(t *testing.T) {
		cfg := LoggerConfig{
			Level:       "debug",
			Directory:   "../../logs",
			ProjectName: "test",
		}

		logger := NewLogger(cfg)
		assert.NotNil(t, logger)
		assert.Equal(t, DefaultLoggerName, cfg.LoggerName)
	})
}
