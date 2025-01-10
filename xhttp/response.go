package xhttp

import (
	"encoding/json"
	"net/http"

	"github.com/RichXan/xcommon/xerror"

	"github.com/gin-gonic/gin"
)

// APIResponse 标准响应结构
type APIResponse struct {
	Code    int         `json:"code"`               // 业务编码
	Status  bool        `json:"status"`             // 请求是否成功
	Message string      `json:"message,omitempty"`  // 错误描述
	Current int         `json:"current,omitempty"`  // 当前页码
	Size    int         `json:"size,omitempty"`     // 当前页数量
	PerPage int         `json:"per_page,omitempty"` // 每页数量
	Total   int64       `json:"total,omitempty"`    // 总数量
	Data    interface{} `json:"data,omitempty"`     // 数据
	Order   string      `json:"order,omitempty"`    // 排序字段
	TraceID string      `json:"trace_id,omitempty"` // 追踪ID
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	resp := &APIResponse{
		Data:    data,
		Status:  true,
		Code:    xerror.Success.Code,
		Message: xerror.Success.Message,
	}
	if traceID := c.GetString("trace_id"); traceID != "" {
		resp.TraceID = traceID
	}
	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	var resp *APIResponse
	if e, ok := err.(*xerror.Error); ok {
		resp = &APIResponse{
			Code:    e.Code,
			Message: e.Message,
		}
	} else {
		resp = &APIResponse{
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

func NewResponse(err *xerror.Error) *APIResponse {
	status := false
	if err.Code == 0 {
		status = true
	}
	return &APIResponse{
		Status:  status,
		Code:    err.Code,
		Message: err.Message,
	}
}

func NewResponseMessage(err *xerror.Error, message string) *APIResponse {
	status := false
	if err.Code == 0 {
		status = true
	}
	return &APIResponse{
		Status:  status,
		Code:    err.Code,
		Message: err.Message + " : " + message,
	}
}

func NewResponseData(err *xerror.Error, data interface{}) *APIResponse {
	status := false
	if err.Code == 0 {
		status = true
	}
	return &APIResponse{
		Status:  status,
		Code:    err.Code,
		Message: err.Message,
		Data:    data,
	}
}

func (res *APIResponse) WithData(data interface{}) *APIResponse {
	res.Data = data
	return res
}

func (res *APIResponse) WithTraceID(traceID string) *APIResponse {
	res.TraceID = traceID
	return res
}

func (res *APIResponse) WithMessage(msg string) *APIResponse {
	res.Message = msg
	return res
}

func (res *APIResponse) WithTotal(total int64) *APIResponse {
	res.Total = total
	return res
}

func (res *APIResponse) WithSize(size int) *APIResponse {
	res.Size = size
	return res
}

func (res *APIResponse) WithCurrent(current int) *APIResponse {
	res.Current = current
	return res
}

func (res *APIResponse) WithPerPage(perPage int) *APIResponse {
	res.PerPage = perPage
	return res
}

// ToString 返回 JSON 格式的错误详情
func (res *APIResponse) ToString() string {
	err := &APIResponse{
		Code:    res.Code,
		Message: res.Message,
		Data:    res.Data,
		TraceID: res.TraceID,
	}

	raw, _ := json.Marshal(err)
	return string(raw)
}
