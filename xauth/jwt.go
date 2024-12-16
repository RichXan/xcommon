package xauth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	accessSecretKey  = []byte("xblog-access-secret-key")
	refreshSecretKey = []byte("xblog-refresh-secret-key")
)

const (
	// AccessTokenExpiry 访问令牌过期时间
	AccessTokenExpiry = 15 * time.Minute
	// RefreshTokenExpiry 刷新令牌过期时间
	RefreshTokenExpiry = 7 * 24 * time.Hour
	// MaxLoginAttempts 最大登录尝试次数
	MaxLoginAttempts = 5
)

// TokenPair 访问令牌和刷新令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims 自定义的 JWT Claims
type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	TokenID  string `json:"token_id"`
	jwt.RegisteredClaims
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func GenerateTokenPair(userID uint64, username string) (*TokenPair, error) {
	tokenID := uuid.New().String()

	// 生成访问令牌
	accessToken, err := generateAccessToken(userID, username, tokenID)
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := generateRefreshToken(userID, username, tokenID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateAccessToken 生成访问令牌
func generateAccessToken(userID uint64, username, tokenID string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xblog",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessSecretKey)
}

// generateRefreshToken 生成刷新令牌
func generateRefreshToken(userID uint64, username, tokenID string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xblog",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecretKey)
}

// ParseAccessToken 解析访问令牌
func ParseAccessToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, accessSecretKey)
}

// ParseRefreshToken 解析刷新令牌
func ParseRefreshToken(tokenString string) (*Claims, error) {
	return parseToken(tokenString, refreshSecretKey)
}

// parseToken 解析令牌
func parseToken(tokenString string, secretKey []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		// 检查是否是过期错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RefreshTokenPair 使用刷新令牌生成新的令牌对
func RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	// 解析刷新令牌
	claims, err := ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 生成新的令牌对
	return GenerateTokenPair(claims.UserID, claims.Username)
}

// ValidateAccessToken 验证访问令牌
func ValidateAccessToken(tokenString string) bool {
	_, err := ParseAccessToken(tokenString)
	return err == nil
}
