package xmiddleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/RichXan/xcommon/xlog"

	"github.com/gin-gonic/gin"
)

// bodyLogWriter 是一个自定义的响应写入器，用于捕获响应body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// formatJSON 格式化JSON字符串
func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Compact(&prettyJSON, data); err != nil {
		return string(data)
	}
	return prettyJSON.String()
}

// Logger 日志中间件
func Logger(logger *xlog.Logger, debug bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 读取请求body
		var requestBody []byte
		if debug {
			if c.Request.Body != nil {
				requestBody, _ = io.ReadAll(c.Request.Body)
				// 重新设置请求body，因为读取后需要重置
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
		}

		// 设置自定义ResponseWriter来捕获响应body
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 结束时间
		latency := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// 使用结构化日志记录请求信息
		logEvent := logger.Info().
			Int("status", c.Writer.Status()).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("ip", c.ClientIP()).
			Dur("latency", latency).
			Str("user_agent", c.Request.UserAgent()).
			Str("request_id", c.GetString("request_id"))

		// 添加请求body（如果存在）
		if len(requestBody) > 0 {
			logEvent.Str("request_body", formatJSON(requestBody))
		}

		// 添加响应body（如果存在）
		if blw.body.Len() > 0 {
			logEvent.Str("response_body", formatJSON(blw.body.Bytes()))
		}

		logEvent.Msg("HTTP Request")
	}
}
