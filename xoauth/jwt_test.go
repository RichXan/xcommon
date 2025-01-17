package xoauth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID   = "test_user_123"
	testUsername = "test_user"
	testKeyDir   = "./test_keys"
)

func TestSaveKeyPair(t *testing.T) {
	claims := NewClaims(nil)
	err := claims.GenerateKeyPair(testKeyDir)
	require.NoError(t, err)

	// 验证文件内容
	privateKeyPath := filepath.Join(testKeyDir, "private.pem")
	publicKeyPath := filepath.Join(testKeyDir, "public.pem")

	privateKeyData, err := os.ReadFile(privateKeyPath)
	require.NoError(t, err)
	assert.NotEmpty(t, privateKeyData)

	publicKeyData, err := os.ReadFile(publicKeyPath)
	require.NoError(t, err)
	assert.NotEmpty(t, publicKeyData)

	assert.NotNil(t, claims.GetPrivateKey())
	assert.NotNil(t, claims.GetPublicKey())
}

func TestNewClaimsWithKeyPairFromPEM(t *testing.T) {
	claims := NewClaims(nil)
	err := claims.GenerateKeyPair(testKeyDir)
	require.NoError(t, err)

	privateKeyBytes, err := os.ReadFile(testKeyDir + "/private.pem")
	require.NoError(t, err)
	assert.NotEmpty(t, privateKeyBytes)

	publicKeyBytes, err := os.ReadFile(testKeyDir + "/public.pem")
	require.NoError(t, err)
	assert.NotEmpty(t, publicKeyBytes)

	claims, err = NewClaimsWithKeyPairFromPEM(&Config{
		PrivateKey: string(privateKeyBytes),
		PublicKey:  string(publicKeyBytes),
	})
	require.NoError(t, err)
	assert.NotNil(t, claims)
}

func TestMain(m *testing.M) {
	code := m.Run()
	// 注释掉清理代码
	os.Remove(testKeyDir + "/private.pem")
	os.Remove(testKeyDir + "/public.pem")
	os.Exit(code)
}

func TestNewClaims(t *testing.T) {
	claims := NewClaims(nil)
	assert.NotNil(t, claims)
}

func TestKeyPairOperations(t *testing.T) {
	// 测试生成密钥对
	claims := NewClaims(nil)
	err := claims.GenerateKeyPair(testKeyDir)
	require.NoError(t, err)
	assert.NotNil(t, claims.GetPrivateKey())
	assert.NotNil(t, claims.GetPublicKey())

	// 创建测试目录
	err = os.MkdirAll(testKeyDir, 0700)
	require.NoError(t, err)

	// 验证文件是否存在
	privateKeyPath := filepath.Join(testKeyDir, "private.pem")
	publicKeyPath := filepath.Join(testKeyDir, "public.pem")
	assert.FileExists(t, privateKeyPath)
	assert.FileExists(t, publicKeyPath)

	// 测试加载密钥对
	newClaims := NewClaims(nil)
	err = newClaims.GenerateKeyPair(testKeyDir)
	require.NoError(t, err)
	assert.NotNil(t, newClaims.GetPrivateKey())
	assert.NotNil(t, newClaims.GetPublicKey())
}

func TestGenerateAndValidateTokenPair(t *testing.T) {
	claims := NewClaims(nil)
	err := claims.GenerateKeyPair(testKeyDir)
	require.NoError(t, err)

	// 生成token对
	tokenPair, err := claims.GenerateTokenPair(Info{
		UserID:   testUserID,
		Username: testUsername,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)

	// 解析访问令牌
	parsedClaims, err := claims.ParseAccessToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, testUserID, parsedClaims.UserID)
}

func TestRefreshTokenPair(t *testing.T) {
	claims := NewClaims(nil)
	err := claims.GenerateKeyPair(testKeyDir)
	require.NoError(t, err)

	// 生成初始token对
	originalPair, err := claims.GenerateTokenPair(Info{
		UserID:   testUserID,
		Username: testUsername,
	})
	require.NoError(t, err)

	// 使用刷新令牌生成新的token对
	newPair, err := claims.RefreshTokenPair(originalPair.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newPair.AccessToken)
	assert.NotEmpty(t, newPair.RefreshToken)
	assert.NotEqual(t, originalPair.AccessToken, newPair.AccessToken)
}

func TestTokenExpirationScenario(t *testing.T) {
	claims := NewClaims(nil)
	err := claims.GenerateKeyPair(testKeyDir)
	require.NoError(t, err)

	// 生成一个短期token
	shortLivedClaims := Claims{
		Info: Info{
			UserID:   testUserID,
			Username: testUsername,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "xan",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, shortLivedClaims)
	tokenString, err := token.SignedString(claims.GetPrivateKey())
	require.NoError(t, err)

	// 等待令牌过期
	time.Sleep(2 * time.Second)

	// 验证过期的访问令牌
	_, err = claims.ParseAccessToken(tokenString)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestInvalidToken(t *testing.T) {
	claims := NewClaims(nil)
	err := claims.GenerateKeyPair(testKeyDir)
	require.NoError(t, err)

	// 测试无效的token
	_, err = claims.ParseAccessToken("invalid.token.string")
	assert.Error(t, err)

	// 测试空token
	_, err = claims.ParseAccessToken("")
	assert.Error(t, err)
}
