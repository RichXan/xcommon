package xmiddleware

import (
	"regexp"

	"github.com/gin-gonic/gin"
)

type copyWriter struct {
	gin.ResponseWriter
}

func (cw copyWriter) Write(b []byte) (int, error) {
	// 匹配零值时间格式 (0001-01-01 00:00:00 或 0001-01-01T00:00:00Z 等)
	zeroTimePattern := regexp.MustCompile(`["|']?\d{4}-01-01[T ]00:00:00\.?\d*Z?["|']?`)
	s := string(b)

	// 将零值时间替换为空字符串
	s = zeroTimePattern.ReplaceAllString(s, `""`)

	// 匹配 ISO8601/RFC3339 格式的时间字符串
	timePatterns := []*regexp.Regexp{
		// 匹配带毫秒的格式
		regexp.MustCompile(`(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})\.?\d*([+-]\d{2}:?\d{2})?`),
		// 匹配不带毫秒的格式
		regexp.MustCompile(`(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})([+-]\d{2}:?\d{2})?`),
		// 匹配Z结尾的UTC时间格式
		regexp.MustCompile(`(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})\.?\d*Z`),
	}

	// 依次应用所有正则表达式格式化非零时间
	for _, pattern := range timePatterns {
		s = pattern.ReplaceAllString(s, "$1 $2")
	}

	return cw.ResponseWriter.WriteString(s)
}

func TimeFormat(ctx *gin.Context) {
	cw := &copyWriter{ResponseWriter: ctx.Writer}
	ctx.Writer = cw
	ctx.Next()
}
