package xerror

const (
	CodeSuccess         = 0
	CodeSystemError     = 10000
	CodeUnauthorized    = 10001
	CodeForbidden       = 10002
	CodeMethodNotAllow  = 10003
	CodeTimeout         = 10004
	CodeTooManyRequests = 10005
	CodeServerBusy      = 10006
	CodeRequestRejected = 10007

	// 服务内部错误码 11000-11999
	CodeParamError         = 11000
	CodeCreateError        = 11001
	CodeDeleteError        = 11002
	CodeUpdateError        = 11003
	CodeGetError           = 11004
	CodeJsonMarshalError   = 11005
	CodeJsonUnmarshalError = 11006
)

// 预定义错误
var (
	// 系统级错误码
	Success         = New(CodeSuccess, "success")                   // 成功
	SystemError     = New(CodeSystemError, "system error")          // 系统错误
	Unauthorized    = New(CodeUnauthorized, "unauthorized")         // 未授权
	Forbidden       = New(CodeForbidden, "forbidden")               // 禁止访问
	MethodNotAllow  = New(CodeMethodNotAllow, "method not allowed") // 方法不允许
	Timeout         = New(CodeTimeout, "timeout")                   // 超时
	TooManyRequests = New(CodeTooManyRequests, "too many requests") // 请求过多
	ServerBusy      = New(CodeServerBusy, "server is busy")         // 服务器繁忙
	RequestRejected = New(CodeRequestRejected, "request rejected")  // 请求被拒绝

	// 服务内部错误码 11000
	ParamError         = New(CodeParamError, "parameter error")              // 参数错误
	CreateError        = New(CodeCreateError, "create resource error")       // 创建错误
	DeleteError        = New(CodeDeleteError, "delete resource error")       // 删除错误
	UpdateError        = New(CodeUpdateError, "update resource error")       // 更新错误
	GetError           = New(CodeGetError, "resource not found")             // 获取错误
	JsonMarshalError   = New(CodeJsonMarshalError, "json marshal error")     // JSON 序列化错误
	JsonUnmarshalError = New(CodeJsonUnmarshalError, "json unmarshal error") // JSON 反序列化错误

	// 用户相关错误码 (100-199)
	// UserNotFound        = New(100, "user not found")          // 用户不存在
	// PasswordError       = New(101, "password error")          // 密码错误
	// UserExists          = New(102, "user already exists")     // 用户已存在
	// TokenExpired        = New(103, "token expired")           // Token过期
	// TokenInvalid        = New(104, "token invalid")           // Token无效
	// RefreshTokenExpired = New(105, "refresh token expired")   // 刷新Token过期
	// RefreshTokenInvalid = New(106, "refresh token invalid")   // 刷新Token无效
	// UserDisabled        = New(107, "user disabled")           // 用户被禁用
	// EmailExists         = New(108, "email already exists")    // 邮箱已存在
	// EmailInvalid        = New(109, "invalid email format")    // 邮箱格式无效
	// UsernameInvalid     = New(110, "invalid username format") // 用户名格式无效
	// PasswordInvalid     = New(111, "invalid password format") // 密码格式无效

	// // OAuth相关错误码 (500-599)
	// OAuthFailed     = New(500, "oauth authentication failed") // OAuth认证失败
	// OAuthCanceled   = New(501, "oauth canceled")              // OAuth取消
	// OAuthTimeout    = New(502, "oauth timeout")               // OAuth超时
	// OAuthStateError = New(503, "oauth state error")           // OAuth状态错误
	// OAuthBound      = New(504, "oauth account already bound") // OAuth账号已绑定

)
