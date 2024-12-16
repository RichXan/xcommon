package xconfig

import (
	"time"
)

// SocialConfig 社交账号配置
type SocialConfig struct {
	// 同步配置
	Sync SyncConfig `yaml:"sync"`
	// OAuth配置
	OAuth OAuthConfig `yaml:"oauth"`
	// 安全配置
	Security SecurityConfig `yaml:"security"`
}

// SyncConfig 同步配置
type SyncConfig struct {
	// 自动同步间隔
	Interval time.Duration `yaml:"interval"`
	// 最大并发同步数
	MaxConcurrent int `yaml:"max_concurrent"`
	// 同步超时时间
	Timeout time.Duration `yaml:"timeout"`
	// 同步重试次数
	MaxRetries int `yaml:"max_retries"`
	// 同步重试间隔
	RetryInterval time.Duration `yaml:"retry_interval"`
	// 同步历史保留天数
	HistoryRetention int `yaml:"history_retention"`
	// 同步项配置
	Items SyncItems `yaml:"items"`
}

// SyncItems 同步项配置
type SyncItems struct {
	// 是否同步头像
	Avatar bool `yaml:"avatar"`
	// 是否同步昵称
	Nickname bool `yaml:"nickname"`
	// 是否同步个人信息
	Profile bool `yaml:"profile"`
	// 是否同步状态
	Status bool `yaml:"status"`
}

// OAuthConfig OAuth配置
type OAuthConfig struct {
	// 提供商配置
	Providers map[string]ProviderConfig `yaml:"providers"`
	// 回调URL基础路径
	CallbackBaseURL string `yaml:"callback_base_url"`
	// 状态令牌过期时间
	StateExpiry time.Duration `yaml:"state_expiry"`
	// 是否允许自动创建用户
	AutoCreateUser bool `yaml:"auto_create_user"`
	// 默认用户角色
	DefaultRole string `yaml:"default_role"`
}

// ProviderConfig OAuth提供商配置
type ProviderConfig struct {
	// 客户端ID
	ClientID string `yaml:"client_id"`
	// 客户端密钥
	ClientSecret string `yaml:"client_secret"`
	// 授权范围
	Scopes []string `yaml:"scopes"`
	// 是否启用
	Enabled bool `yaml:"enabled"`
	// 自定义配置
	Extra map[string]interface{} `yaml:"extra"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 最大绑定账号数
	MaxBindings int `yaml:"max_bindings"`
	// 是否允许解绑最后一个账号
	AllowUnbindLast bool `yaml:"allow_unbind_last"`
	// 是否允许合并账号
	AllowMerge bool `yaml:"allow_merge"`
	// 合并确认超时时间
	MergeConfirmTimeout time.Duration `yaml:"merge_confirm_timeout"`
	// 是否启用IP限制
	EnableIPLimit bool `yaml:"enable_ip_limit"`
	// IP限制配置
	IPLimit IPLimitConfig `yaml:"ip_limit"`
}

// IPLimitConfig IP限制配置
type IPLimitConfig struct {
	// 时间窗口
	Window time.Duration `yaml:"window"`
	// 最大请求次数
	MaxRequests int `yaml:"max_requests"`
	// 封禁时间
	BanDuration time.Duration `yaml:"ban_duration"`
	// 白名单
	Whitelist []string `yaml:"whitelist"`
}

// DefaultConfig 默认配置
func DefaultConfig() *SocialConfig {
	return &SocialConfig{
		Sync: SyncConfig{
			Interval:         6 * time.Hour,
			MaxConcurrent:    5,
			Timeout:          30 * time.Second,
			MaxRetries:       3,
			RetryInterval:    time.Minute,
			HistoryRetention: 30,
			Items: SyncItems{
				Avatar:   true,
				Nickname: true,
				Profile:  true,
				Status:   true,
			},
		},
		OAuth: OAuthConfig{
			Providers:      make(map[string]ProviderConfig),
			StateExpiry:    15 * time.Minute,
			AutoCreateUser: true,
			DefaultRole:    "user",
		},
		Security: SecurityConfig{
			MaxBindings:         5,
			AllowUnbindLast:     false,
			AllowMerge:          true,
			MergeConfirmTimeout: 24 * time.Hour,
			EnableIPLimit:       true,
			IPLimit: IPLimitConfig{
				Window:      time.Hour,
				MaxRequests: 100,
				BanDuration: 24 * time.Hour,
				Whitelist:   []string{},
			},
		},
	}
}

// Validate 验证配置
func (c *SocialConfig) Validate() error {
	// 验证同步配置
	if c.Sync.Interval < time.Minute {
		c.Sync.Interval = time.Minute
	}
	if c.Sync.MaxConcurrent < 1 {
		c.Sync.MaxConcurrent = 1
	}
	if c.Sync.Timeout < time.Second {
		c.Sync.Timeout = time.Second
	}
	if c.Sync.MaxRetries < 0 {
		c.Sync.MaxRetries = 0
	}
	if c.Sync.RetryInterval < time.Second {
		c.Sync.RetryInterval = time.Second
	}
	if c.Sync.HistoryRetention < 1 {
		c.Sync.HistoryRetention = 1
	}

	// 验证OAuth配置
	if c.OAuth.StateExpiry < time.Minute {
		c.OAuth.StateExpiry = time.Minute
	}
	if c.OAuth.DefaultRole == "" {
		c.OAuth.DefaultRole = "user"
	}

	// 验证安全配置
	if c.Security.MaxBindings < 1 {
		c.Security.MaxBindings = 1
	}
	if c.Security.MergeConfirmTimeout < time.Minute {
		c.Security.MergeConfirmTimeout = time.Minute
	}
	if c.Security.EnableIPLimit {
		if c.Security.IPLimit.Window < time.Second {
			c.Security.IPLimit.Window = time.Second
		}
		if c.Security.IPLimit.MaxRequests < 1 {
			c.Security.IPLimit.MaxRequests = 1
		}
		if c.Security.IPLimit.BanDuration < time.Minute {
			c.Security.IPLimit.BanDuration = time.Minute
		}
	}

	return nil
}
