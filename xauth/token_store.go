package xauth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// 令牌黑名单前缀
	tokenBlacklistPrefix = "token:blacklist:"
	// 登录尝试次数前缀
	loginAttemptsPrefix = "login:attempts:"
	// 登录锁定前缀
	loginLockPrefix = "login:lock:"

	// 默认配置
	maxLoginAttempts   = 5         // 最大登录尝试次数
	loginLockDuration  = time.Hour // 登录锁定时长
	loginAttemptExpiry = time.Hour // 登录尝试记录过期时间
)

// TokenStore 令牌存储接口
type TokenStore interface {
	// 令牌黑名单相关
	RevokeToken(ctx context.Context, tokenID string, expiration time.Duration) error
	IsTokenRevoked(ctx context.Context, tokenID string) bool

	// 登录频率限制相关
	IncrLoginAttempts(ctx context.Context, identifier string) (int64, error)
	IsLoginLocked(ctx context.Context, identifier string) bool
	LockLogin(ctx context.Context, identifier string) error
	ResetLoginAttempts(ctx context.Context, identifier string) error
}

// RedisTokenStore Redis实现的令牌存储
type RedisTokenStore struct {
	client *redis.Client
}

// NewRedisTokenStore 创建新的Redis令牌存储
func NewRedisTokenStore(client *redis.Client) TokenStore {
	return &RedisTokenStore{
		client: client,
	}
}

// RevokeToken 撤销令牌
func (s *RedisTokenStore) RevokeToken(ctx context.Context, tokenID string, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s", tokenBlacklistPrefix, tokenID)
	return s.client.Set(ctx, key, "revoked", expiration).Err()
}

// IsTokenRevoked 检查令牌是否已被撤销
func (s *RedisTokenStore) IsTokenRevoked(ctx context.Context, tokenID string) bool {
	key := fmt.Sprintf("%s%s", tokenBlacklistPrefix, tokenID)
	exists, _ := s.client.Exists(ctx, key).Result()
	return exists > 0
}

// IncrLoginAttempts 增加登录尝试次数
func (s *RedisTokenStore) IncrLoginAttempts(ctx context.Context, identifier string) (int64, error) {
	key := fmt.Sprintf("%s%s", loginAttemptsPrefix, identifier)
	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, loginAttemptExpiry)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// IsLoginLocked 检查登录是否被锁定
func (s *RedisTokenStore) IsLoginLocked(ctx context.Context, identifier string) bool {
	key := fmt.Sprintf("%s%s", loginLockPrefix, identifier)
	exists, _ := s.client.Exists(ctx, key).Result()
	return exists > 0
}

// LockLogin 锁定登录
func (s *RedisTokenStore) LockLogin(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("%s%s", loginLockPrefix, identifier)
	return s.client.Set(ctx, key, "locked", loginLockDuration).Err()
}

// ResetLoginAttempts 重置登录尝试次数
func (s *RedisTokenStore) ResetLoginAttempts(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("%s%s", loginAttemptsPrefix, identifier)
	return s.client.Del(ctx, key).Err()
}
