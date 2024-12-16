package xmiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDKey 请求ID的键名
	RequestIDKey = "request_id"
	// RequestIDHeader 请求ID的请求头
	RequestIDHeader = "X-Request-ID"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从请求头获取
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			// 如果请求头中没有，则生成新的
			requestID = uuid.New().String()
		}

		// 设置到上下文
		c.Set(RequestIDKey, requestID)
		// 设置响应头
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
} 