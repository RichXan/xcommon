package xoauth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

const (
	// AccessTokenExpiry 访问令牌过期时间
	AccessTokenExpiry = 2 * time.Hour
	// RefreshTokenExpiry 刷新令牌过期时间
	RefreshTokenExpiry = 7 * 24 * time.Hour
	// MaxLoginAttempts 最大登录尝试次数
	MaxLoginAttempts = 5
	// PEM类型常量
	PrivateKeyPEMType = "PRIVATE KEY"
	PublicKeyPEMType  = "PUBLIC KEY"
)

// TokenPair 访问令牌和刷新令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims 自定义的 JWT Claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TokenID  string `json:"token_id"`
	jwt.RegisteredClaims

	// 密钥对 - 不导出且不序列化
	privateKey ed25519.PrivateKey `json:"-"`
	publicKey  ed25519.PublicKey  `json:"-"`
}

// NewClaims 创建新的Claims实例
func NewClaims() *Claims {
	return &Claims{}
}

func NewClaimsWithKeyPair(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *Claims {
	return &Claims{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// GenerateKeyPair 生成新的Ed25519密钥对
func (c *Claims) GenerateKeyPair() error {
	// 生成新的密钥对
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// 直接保存密钥
	c.privateKey = privateKey
	c.publicKey = publicKey

	return nil
}

// SaveKeyPair 将密钥对保存到文件
func (c *Claims) SaveKeyPair(dir string) error {
	// 确保目录存在
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// 创建PEM块
	privatePEM := &pem.Block{
		Type:  PrivateKeyPEMType,
		Bytes: c.privateKey,
	}

	publicPEM := &pem.Block{
		Type:  PublicKeyPEMType,
		Bytes: c.publicKey,
	}

	// 保存私钥
	privateKeyFile := filepath.Join(dir, "private.pem")
	privateKeyHandle, err := os.OpenFile(privateKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer privateKeyHandle.Close()

	if err := pem.Encode(privateKeyHandle, privatePEM); err != nil {
		return err
	}

	// 保存公钥
	publicKeyFile := filepath.Join(dir, "public.pem")
	publicKeyHandle, err := os.OpenFile(publicKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer publicKeyHandle.Close()

	return pem.Encode(publicKeyHandle, publicPEM)
}

// LoadKeyPair 从文件加载密钥对
func (c *Claims) LoadKeyPair(dir string) error {
	// 读取私钥
	privateKeyFile := filepath.Join(dir, "private.pem")
	privateKeyBytes, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return err
	}

	privatePEM, _ := pem.Decode(privateKeyBytes)
	if privatePEM == nil {
		return errors.New("failed to decode private key PEM")
	}

	// 读取公钥
	publicKeyFile := filepath.Join(dir, "public.pem")
	publicKeyBytes, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return err
	}

	publicPEM, _ := pem.Decode(publicKeyBytes)
	if publicPEM == nil {
		return errors.New("failed to decode public key PEM")
	}

	// 验证密钥长度
	if len(privatePEM.Bytes) != ed25519.PrivateKeySize {
		return errors.New("invalid private key size")
	}
	if len(publicPEM.Bytes) != ed25519.PublicKeySize {
		return errors.New("invalid public key size")
	}

	// 直接保存密钥
	c.privateKey = ed25519.PrivateKey(privatePEM.Bytes)
	c.publicKey = ed25519.PublicKey(publicPEM.Bytes)

	return nil
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (c *Claims) GenerateTokenPair(userID, username string) (*TokenPair, error) {
	tokenID := uuid.New().String()

	// 生成访问令牌
	accessToken, err := c.generateAccessToken(userID, username, tokenID)
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := c.generateRefreshToken(userID, username, tokenID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateAccessToken 生成访问令牌
func (c *Claims) generateAccessToken(userID, username, tokenID string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xan",
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(c.privateKey)
}

// generateRefreshToken 生成刷新令牌
func (c *Claims) generateRefreshToken(userID, username, tokenID string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xan",
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(c.privateKey)
}

// ParseToken 解析令牌
func (c *Claims) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.publicKey, nil
	})

	if err != nil {
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
func (c *Claims) RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	// 解析刷新令牌
	claims, err := c.ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 生成新的令牌对
	return c.GenerateTokenPair(claims.UserID, claims.Username)
}

// ValidateToken 验证令牌
func (c *Claims) ValidateToken(tokenString string) bool {
	_, err := c.ParseToken(tokenString)
	return err == nil
}

// ParseAccessToken 解析访问令牌
func (c *Claims) ParseAccessToken(tokenString string) (*Claims, error) {
	return c.ParseToken(tokenString)
}

// ParseRefreshToken 解析刷新令牌
func (c *Claims) ParseRefreshToken(tokenString string) (*Claims, error) {
	return c.ParseToken(tokenString)
}

// ValidateAccessToken 验证访问令牌
func (c *Claims) ValidateAccessToken(tokenString string) bool {
	return c.ValidateToken(tokenString)
}
