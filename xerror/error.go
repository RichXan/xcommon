package xerror

import "fmt"

// Error 自定义错误
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var codes = map[int]string{}

// Error 实现 error 接口
func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// New 创建新的错误
func New(code int, message string) *Error {
	// 预定义错误码
	// if _, ok := codes[code]; ok {
	// 	panic(fmt.Sprintf("code %d already exists", code))
	// }
	// codes[code] = message
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装错误
func Wrap(err error, code int, message string) *Error {
	if err == nil {
		return New(code, message)
	}
	return New(code, fmt.Sprintf("%s: %v", message, err))
}
