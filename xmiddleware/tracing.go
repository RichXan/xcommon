package xmiddleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// TracingMiddleware 链路追踪中间件
func TracingMiddleware(tracer opentracing.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头中提取span上下文
		spanCtx, err := tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)

		var span opentracing.Span
		if err != nil {
			// 如果没有上下文，创建新的根span
			span = tracer.StartSpan(fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()))
		} else {
			// 如果有上下文，创建子span
			span = tracer.StartSpan(
				fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()),
				ext.RPCServerOption(spanCtx),
			)
		}
		defer span.Finish()

		// 设置标签
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.String())
		ext.Component.Set(span, "gin")

		// 将span注入到请求上下文
		c.Set("span", span)

		// 处理请求
		c.Next()

		// 设置响应状态码
		ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))
		if c.Writer.Status() >= 400 {
			ext.Error.Set(span, true)
		}
	}
} 