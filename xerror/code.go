package xerror

const (
	CodeSuccess         = 0
	CodeSystemError     = 1
	CodeParamError      = 2
	CodeUnauthorized    = 3
	CodeForbidden       = 4
	CodeNotFound        = 5
	CodeMethodNotAllow  = 6
	CodeTimeout         = 7
	CodeTooManyRequests = 8
	CodeServerBusy      = 9
	CodeRequestRejected = 10
)

// 预定义错误
var (
	// 系统级错误码 (1-99)
	Success         = NewError(CodeSuccess, "success")                   // 成功
	SystemError     = NewError(CodeSystemError, "system error")          // 系统错误
	ParamError      = NewError(CodeParamError, "parameter error")        // 参数错误
	Unauthorized    = NewError(CodeUnauthorized, "unauthorized")         // 未授权
	Forbidden       = NewError(CodeForbidden, "forbidden")               // 禁止访问
	NotFound        = NewError(CodeNotFound, "not found")                // 资源不存在
	MethodNotAllow  = NewError(CodeMethodNotAllow, "method not allowed") // 方法不允许
	Timeout         = NewError(CodeTimeout, "timeout")                   // 超时
	TooManyRequests = NewError(CodeTooManyRequests, "too many requests") // 请求过多
	ServerBusy      = NewError(CodeServerBusy, "server is busy")         // 服务器繁忙
	RequestRejected = NewError(CodeRequestRejected, "request rejected")  // 请求被拒绝

	// 用户相关错误码 (100-199)
	UserNotFound        = NewError(100, "user not found")          // 用户不存在
	PasswordError       = NewError(101, "password error")          // 密码错误
	UserExists          = NewError(102, "user already exists")     // 用户已存在
	TokenExpired        = NewError(103, "token expired")           // Token过期
	TokenInvalid        = NewError(104, "token invalid")           // Token无效
	RefreshTokenExpired = NewError(105, "refresh token expired")   // 刷新Token过期
	RefreshTokenInvalid = NewError(106, "refresh token invalid")   // 刷新Token无效
	UserDisabled        = NewError(107, "user disabled")           // 用户被禁用
	EmailExists         = NewError(108, "email already exists")    // 邮箱已存在
	EmailInvalid        = NewError(109, "invalid email format")    // 邮箱格式无效
	UsernameInvalid     = NewError(110, "invalid username format") // 用户名格式无效
	PasswordInvalid     = NewError(111, "invalid password format") // 密码格式无效

	// 文章相关错误码 (200-299)
	PostNotFound     = NewError(200, "post not found")            // 文章不存在
	PostForbidden    = NewError(201, "forbidden to operate post") // 无权操作文章
	PostTitleEmpty   = NewError(202, "post title is empty")       // 文章标题为空
	PostContentEmpty = NewError(203, "post content is empty")     // 文章内容为空
	PostDeleted      = NewError(204, "post has been deleted")     // 文章已删除
	PostDraft        = NewError(205, "post is draft")             // 文章为草稿状态

	// 评论相关错误码 (300-399)
	CommentNotFound  = NewError(300, "comment not found")            // 评论不存在
	CommentForbidden = NewError(301, "forbidden to operate comment") // 无权操作评论
	CommentEmpty     = NewError(302, "comment content is empty")     // 评论内容为空
	CommentDeleted   = NewError(303, "comment has been deleted")     // 评论已删除

	// 点赞相关错误码 (400-499)
	LikeExists    = NewError(400, "already liked")             // 已点赞
	LikeNotFound  = NewError(401, "like not found")            // 点赞不存在
	LikeForbidden = NewError(402, "forbidden to operate like") // 无权操作点赞

	// OAuth相关错误码 (500-599)
	OAuthFailed     = NewError(500, "oauth authentication failed") // OAuth认证失败
	OAuthCanceled   = NewError(501, "oauth canceled")              // OAuth取消
	OAuthTimeout    = NewError(502, "oauth timeout")               // OAuth超时
	OAuthStateError = NewError(503, "oauth state error")           // OAuth状态错误
	OAuthBound      = NewError(504, "oauth account already bound") // OAuth账号已绑定

)
