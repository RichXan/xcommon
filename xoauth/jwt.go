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
)

const (
	// AccessTokenExpiry 访问令牌过期时间
	AccessTokenExpiry = 2 * time.Hour
	// RefreshTokenExpiry 刷新令牌过期时间
	RefreshTokenExpiry = 2 * 24 * time.Hour
	// MaxLoginAttempts 最大登录尝试次数
	MaxLoginAttempts = 5
	// PEM类型常量
	PrivateKeyPEMType = "PRIVATE KEY"
	PublicKeyPEMType  = "PUBLIC KEY"

	PrivateKeyFileName = "private.pem"
	PublicKeyFileName  = "public.pem"
)

// TokenPair 访问令牌和刷新令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// 保存的信息
type Info struct {
	Issuer   string `json:"issuer"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Data     any    `json:"data"`
}

type Config struct {
	TokenExpiryHours        int `yaml:"token_expiry_hours" json:"token_expiry_hours"`
	RefreshTokenExpiryHours int `yaml:"refresh_token_expiry_hours" json:"refresh_token_expiry_hours"`

	PrivateKey string `yaml:"private_key" json:"private_key"`
	PublicKey  string `yaml:"public_key" json:"public_key"`
	// 密钥对 - 不导出且不序列化
	privateKey ed25519.PrivateKey `json:"-"`
	publicKey  ed25519.PublicKey  `json:"-"`
}

// Claims 自定义的 JWT Claims
type Claims struct {
	Info
	Config
	jwt.RegisteredClaims
}

type Claim interface {
	GetPublicKey() ed25519.PublicKey
	GetPrivateKey() ed25519.PrivateKey
	GenerateKeyPair(dir string) error
	GenerateTokenPair(info Info) (*TokenPair, error)
	RefreshTokenPair(refreshToken string) (*TokenPair, error)
	ParseAccessToken(tokenString string) (*Claims, error)
}

// NewClaims 创建新的Claims实例
func NewClaims(config *Config) Claim {
	config = checkConfig(config)
	return &Claims{
		Config: *config,
	}
}

func NewClaimsWithKeyPairFromPEM(config *Config) (Claim, error) {
	config = checkConfig(config)

	privateKey, publicKey, err := decodePEMBytes([]byte(config.PrivateKey), []byte(config.PublicKey))
	if err != nil {
		return nil, err
	}

	config.privateKey = privateKey
	config.publicKey = publicKey

	return &Claims{
		Config: *config,
	}, nil
}

func (c *Claims) GetPrivateKey() ed25519.PrivateKey {
	return c.privateKey
}

func (c *Claims) GetPublicKey() ed25519.PublicKey {
	return c.publicKey
}

// GenerateKeyPair 生成新的Ed25519密钥对
func (c *Claims) GenerateKeyPair(dir string) error {
	// 生成新的密钥对
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// 直接保存密钥
	c.privateKey = privateKey
	c.publicKey = publicKey

	// 保存密钥对到文件
	return c.saveKeyPair(dir)
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (c *Claims) GenerateTokenPair(info Info) (*TokenPair, error) {
	tokenID := uuid.New().String()

	// 生成访问令牌
	accessToken, err := c.generateAccessToken(info, tokenID)
	if err != nil {
		return nil, err
	}

	// 生成刷新令牌
	refreshToken, err := c.generateRefreshToken(info, tokenID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshTokenPair 使用刷新令牌生成新的令牌对
func (c *Claims) RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	// 解析刷新令牌
	claims, err := c.parseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 生成新的token ID
	tokenID := uuid.New().String()

	// 生成新的访问令牌
	accessToken, err := c.generateAccessToken(claims.Info, tokenID)
	if err != nil {
		return nil, err
	}

	// 返回令牌对，保持原有的刷新令牌
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ParseAccessToken 解析访问令牌
func (c *Claims) ParseAccessToken(tokenString string) (*Claims, error) {
	return c.parseToken(tokenString)
}

// 添加新的工具方法用于生成 Claims
func (c *Claims) newTokenClaims(info Info, tokenID string, expiry time.Duration) Claims {
	return Claims{
		Info: info,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Issuer:    "xan",
			Subject:   info.UserID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
}

// checkConfig 检查并返回配置
func checkConfig(config *Config) *Config {
	if config == nil {
		return &Config{
			TokenExpiryHours:        int(AccessTokenExpiry.Hours()),
			RefreshTokenExpiryHours: int(RefreshTokenExpiry.Hours()),
		}
	}

	if config.TokenExpiryHours == 0 {
		config.TokenExpiryHours = int(AccessTokenExpiry.Hours())
	}
	if config.RefreshTokenExpiryHours == 0 {
		config.RefreshTokenExpiryHours = int(RefreshTokenExpiry.Hours())
	}
	return config
}

func decodePEMBytes(privateKeyPEMBytes, publicKeyPEMBytes []byte) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	privateKey, _ := pem.Decode(privateKeyPEMBytes)
	if privateKey == nil {
		return nil, nil, fmt.Errorf("failed to decode private key PEM")
	}

	publicKey, _ := pem.Decode(publicKeyPEMBytes)
	if publicKey == nil {
		return nil, nil, fmt.Errorf("failed to decode public key PEM")
	}

	if err := validateKeySize(privateKey.Bytes, publicKey.Bytes); err != nil {
		return nil, nil, err
	}
	return ed25519.PrivateKey(privateKey.Bytes), ed25519.PublicKey(publicKey.Bytes), nil
}

// SaveKeyPair 将密钥对保存到文件
func (c *Claims) saveKeyPair(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	privatePEM := &pem.Block{
		Type:  PrivateKeyPEMType,
		Bytes: c.privateKey,
	}

	publicPEM := &pem.Block{
		Type:  PublicKeyPEMType,
		Bytes: c.publicKey,
	}

	// 保存私钥
	if err := savePEMToFile(getKeyFilePath(dir, PrivateKeyFileName), privatePEM, 0600); err != nil {
		return err
	}

	// 保存公钥
	return savePEMToFile(getKeyFilePath(dir, PublicKeyFileName), publicPEM, 0644)
}

func savePEMToFile(filePath string, pemBlock *pem.Block, perm os.FileMode) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	return pem.Encode(file, pemBlock)
}

// generateAccessToken 生成访问令牌
func (c *Claims) generateAccessToken(info Info, tokenID string) (string, error) {
	claims := c.newTokenClaims(info, tokenID, time.Duration(c.Config.TokenExpiryHours)*time.Hour)
	return c.generateToken(claims)
}

func (c *Claims) generateRefreshToken(info Info, tokenID string) (string, error) {
	claims := c.newTokenClaims(info, tokenID, time.Duration(c.Config.RefreshTokenExpiryHours)*time.Hour)
	return c.generateToken(claims)
}

func (c *Claims) generateToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(c.privateKey)
}

func validateKeySize(privateKey, publicKey []byte) error {
	if len(privateKey) != ed25519.PrivateKeySize {
		return errors.New("invalid private key size")
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return errors.New("invalid public key size")
	}
	return nil
}

func getKeyFilePath(dir string, fileName string) string {
	return filepath.Join(dir, fileName)
}

// parseToken 解析令牌
func (c *Claims) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return c.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
