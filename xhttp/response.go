package xhttp

import (
	"net/http"

	"github.com/richxan/xcommon/xerror"

	"github.com/gin-gonic/gin"
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code"`               // 错误码
	Message string      `json:"message"`            // 错误信息
	Data    interface{} `json:"data,omitempty"`     // 数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	resp := &Response{
		Code:    xerror.Success.Code,
		Message: xerror.Success.Message,
		Data:    data,
	}
	if traceID := c.GetString("trace_id"); traceID != "" {
		resp.TraceID = traceID
	}
	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	var resp *Response
	if e, ok := err.(*xerror.Error); ok {
		resp = &Response{
			Code:    e.Code,
			Message: e.Message,
		}
	} else {
		resp = &Response{
			Code:    xerror.SystemError.Code,
			Message: err.Error(),
		}
	}
	if traceID := c.GetString("trace_id"); traceID != "" {
		resp.TraceID = traceID
	}

	// 根据错误码设置 HTTP 状态码
	httpStatus := getHTTPStatus(resp.Code)
	c.JSON(httpStatus, resp)
}

// getHTTPStatus 根据错误码获取 HTTP 状态码
func getHTTPStatus(code int) int {
	switch code {
	case xerror.Success.Code:
		return http.StatusOK
	case xerror.ParamError.Code:
		return http.StatusBadRequest
	case xerror.Unauthorized.Code:
		return http.StatusUnauthorized
	case xerror.Forbidden.Code:
		return http.StatusForbidden
	case xerror.NotFound.Code:
		return http.StatusNotFound
	case xerror.MethodNotAllow.Code:
		return http.StatusMethodNotAllowed
	case xerror.TooManyRequests.Code:
		return http.StatusTooManyRequests
	default:
		if code >= 500 {
			return http.StatusInternalServerError
		}
		return http.StatusBadRequest
	}
}

// List 列表响应
type List struct {
	Total int64       `json:"total"` // 总数
	Items interface{} `json:"items"` // 数据列表
}

// Page 分页响应
type Page struct {
	List
	Page     int `json:"page"`      // 当前页码
	PageSize int `json:"page_size"` // 每页数量
}

// NewList 创建列表响应
func NewList(total int64, items interface{}) *List {
	return &List{
		Total: total,
		Items: items,
	}
}

// NewPage 创建分页响应
func NewPage(page, pageSize int, total int64, items interface{}) *Page {
	return &Page{
		List: List{
			Total: total,
			Items: items,
		},
		Page:     page,
		PageSize: pageSize,
	}
}
