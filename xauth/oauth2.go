package xauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var (
	ErrInvalidProvider = errors.New("invalid OAuth provider")
	ErrUserInfoFailed  = errors.New("failed to get user info")
)

// 自定义 OAuth2 端点
var (
	WechatEndpoint = oauth2.Endpoint{
		AuthURL:  "https://open.weixin.qq.com/connect/qrconnect",
		TokenURL: "https://api.weixin.qq.com/sns/oauth2/access_token",
	}

	QQEndpoint = oauth2.Endpoint{
		AuthURL:  "https://graph.qq.com/oauth2.0/authorize",
		TokenURL: "https://graph.qq.com/oauth2.0/token",
	}

	WeiboEndpoint = oauth2.Endpoint{
		AuthURL:  "https://api.weibo.com/oauth2/authorize",
		TokenURL: "https://api.weibo.com/oauth2/access_token",
	}
)

// GenerateState 生成OAuth状态值
func GenerateState() string {
	return uuid.New().String()
}

// OAuthUserInfo 统一的用户信息结构
type OAuthUserInfo struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	Name      string                 `json:"name"`
	AvatarURL string                 `json:"avatar_url"`
	Provider  string                 `json:"provider"`
	OpenID    string                 `json:"open_id"`
	UnionID   string                 `json:"union_id,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// OAuthProvider OAuth2提供商配置
type OAuthProvider struct {
	Config     *oauth2.Config
	GetUserInfo func(ctx context.Context, client *http.Client) (*OAuthUserInfo, error)
}

// OAuthConfig OAuth2配置
type OAuthConfig struct {
	Providers map[string]*OAuthProvider
}

// NewOAuthConfig 创建新的OAuth配置
func NewOAuthConfig(baseURL string, configs map[string]map[string]string) *OAuthConfig {
	providers := make(map[string]*OAuthProvider)

	// GitHub配置
	if cfg, ok := configs["github"]; ok {
		providers["github"] = &OAuthProvider{
			Config: &oauth2.Config{
				ClientID:     cfg["client_id"],
				ClientSecret: cfg["client_secret"],
				RedirectURL:  baseURL + "/auth/github/callback",
				Scopes:      []string{"user:email"},
				Endpoint:    github.Endpoint,
			},
			GetUserInfo: getGitHubUserInfo,
		}
	}

	// Google配置
	if cfg, ok := configs["google"]; ok {
		providers["google"] = &OAuthProvider{
			Config: &oauth2.Config{
				ClientID:     cfg["client_id"],
				ClientSecret: cfg["client_secret"],
				RedirectURL:  baseURL + "/auth/google/callback",
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
					"https://www.googleapis.com/auth/userinfo.profile",
				},
				Endpoint: google.Endpoint,
			},
			GetUserInfo: getGoogleUserInfo,
		}
	}

	// 微信配置
	if cfg, ok := configs["wechat"]; ok {
		providers["wechat"] = &OAuthProvider{
			Config: &oauth2.Config{
				ClientID:     cfg["client_id"],
				ClientSecret: cfg["client_secret"],
				RedirectURL:  baseURL + "/auth/wechat/callback",
				Scopes:      []string{cfg["scope"]},
				Endpoint:    WechatEndpoint,
			},
			GetUserInfo: getWechatUserInfo,
		}
	}

	// QQ配置
	if cfg, ok := configs["qq"]; ok {
		providers["qq"] = &OAuthProvider{
			Config: &oauth2.Config{
				ClientID:     cfg["client_id"],
				ClientSecret: cfg["client_secret"],
				RedirectURL:  baseURL + "/auth/qq/callback",
				Scopes:      []string{cfg["scope"]},
				Endpoint:    QQEndpoint,
			},
			GetUserInfo: getQQUserInfo,
		}
	}

	// 微博配置
	if cfg, ok := configs["weibo"]; ok {
		providers["weibo"] = &OAuthProvider{
			Config: &oauth2.Config{
				ClientID:     cfg["client_id"],
				ClientSecret: cfg["client_secret"],
				RedirectURL:  baseURL + "/auth/weibo/callback",
				Scopes:      []string{cfg["scope"]},
				Endpoint:    WeiboEndpoint,
			},
			GetUserInfo: getWeiboUserInfo,
		}
	}

	return &OAuthConfig{
		Providers: providers,
	}
}

// GetProvider 获取指定的OAuth提供商
func (c *OAuthConfig) GetProvider(name string) (*OAuthProvider, error) {
	provider, ok := c.Providers[name]
	if !ok {
		return nil, ErrInvalidProvider
	}
	return provider, nil
}

// getGitHubUserInfo 获取GitHub用户信息
func getGitHubUserInfo(ctx context.Context, client *http.Client) (*OAuthUserInfo, error) {
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Login     string `json:"login"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        fmt.Sprintf("%d", data.ID),
		Email:     data.Email,
		Name:      data.Name,
		AvatarURL: data.AvatarURL,
		Provider:  "github",
		OpenID:    fmt.Sprintf("%d", data.ID),
		Extra: map[string]interface{}{
			"login": data.Login,
		},
	}, nil
}

// getGoogleUserInfo 获取Google用户信息
func getGoogleUserInfo(ctx context.Context, client *http.Client) (*OAuthUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        data.ID,
		Email:     data.Email,
		Name:      data.Name,
		AvatarURL: data.Picture,
		Provider:  "google",
		OpenID:    data.ID,
		Extra: map[string]interface{}{
			"verified_email": data.VerifiedEmail,
		},
	}, nil
}

// getWechatUserInfo 获取微信用户信息
func getWechatUserInfo(ctx context.Context, client *http.Client) (*OAuthUserInfo, error) {
	// 首先获取用户OpenID和访问令牌
	token, err := getWechatToken(client)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	resp, err := client.Get(fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		token.AccessToken,
		token.OpenID,
	))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		OpenID     string   `json:"openid"`
		Nickname   string   `json:"nickname"`
		HeadImgURL string   `json:"headimgurl"`
		UnionID    string   `json:"unionid"`
		Privilege  []string `json:"privilege"`
		Sex        int      `json:"sex"`
		Country    string   `json:"country"`
		Province   string   `json:"province"`
		City       string   `json:"city"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        data.UnionID,
		
		Email:     data.OpenID + "@wechat.com",
		Name:      data.Nickname,
		AvatarURL: data.HeadImgURL,
		Provider:  "wechat",
		OpenID:    data.OpenID,
		UnionID:   data.UnionID,
		Extra: map[string]interface{}{
			"sex":       data.Sex,
			"country":   data.Country,
			"province":  data.Province,
			"city":      data.City,
			"privilege": data.Privilege,
		},
	}, nil
}

// getQQUserInfo 获取QQ用户信息
func getQQUserInfo(ctx context.Context, client *http.Client) (*OAuthUserInfo, error) {
	// 首先获取用户OpenID
	openID, err := getQQOpenID(client)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	resp, err := client.Get(fmt.Sprintf(
		"https://graph.qq.com/user/get_user_info?openid=%s",
		openID,
	))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Nickname     string `json:"nickname"`
		Figureurl    string `json:"figureurl"`
		FigureurlQQ1 string `json:"figureurl_qq_1"`
		FigureurlQQ2 string `json:"figureurl_qq_2"`
		Gender       string `json:"gender"`
		Level        string `json:"level"`
		VipInfo      int    `json:"vip_info"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	avatarURL := data.FigureurlQQ2
	if avatarURL == "" {
		avatarURL = data.FigureurlQQ1
	}
	if avatarURL == "" {
		avatarURL = data.Figureurl
	}

	return &OAuthUserInfo{
		ID:        openID,
		Email:     openID + "@qq.com",
		Name:      data.Nickname,
		AvatarURL: avatarURL,
		Provider:  "qq",
		OpenID:    openID,
		Extra: map[string]interface{}{
			"gender":   data.Gender,
			"level":    data.Level,
			"vip_info": data.VipInfo,
		},
	}, nil
}

// getWeiboUserInfo 获取微博用户信息
func getWeiboUserInfo(ctx context.Context, client *http.Client) (*OAuthUserInfo, error) {
	// 获取用户UID
	uid, err := getWeiboUID(client)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	resp, err := client.Get(fmt.Sprintf(
		"https://api.weibo.com/2/users/show.json?uid=%s",
		uid,
	))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID              int64  `json:"id"`
		ScreenName      string `json:"screen_name"`
		Name            string `json:"name"`
		ProfileImageURL string `json:"profile_image_url"`
		Email           string `json:"email"`
		Gender          string `json:"gender"`
		Location        string `json:"location"`
		Description     string `json:"description"`
		FollowersCount  int    `json:"followers_count"`
		FriendsCount    int    `json:"friends_count"`
		StatusesCount   int    `json:"statuses_count"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        fmt.Sprintf("%d", data.ID),
		Email:     data.Email,
		Name:      data.ScreenName,
		AvatarURL: data.ProfileImageURL,
		Provider:  "weibo",
		OpenID:    fmt.Sprintf("%d", data.ID),
		Extra: map[string]interface{}{
			"name":            data.Name,
			"gender":          data.Gender,
			"location":        data.Location,
			"description":     data.Description,
			"followers_count": data.FollowersCount,
			"friends_count":   data.FriendsCount,
			"statuses_count": data.StatusesCount,
		},
	}, nil
}

// 辅助函数：获取微信令牌信息
type wechatToken struct {
	AccessToken string `json:"access_token"`
	OpenID      string `json:"openid"`
}

func getWechatToken(client *http.Client) (*wechatToken, error) {
	resp, err := client.Get("https://api.weixin.qq.com/sns/oauth2/access_token")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token wechatToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// 辅助函数：获取QQ OpenID
func getQQOpenID(client *http.Client) (string, error) {
	resp, err := client.Get("https://graph.qq.com/oauth2.0/me")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// QQ返回的是JSONP格式，需要处理
	data := string(body)
	start := strings.Index(data, "{")
	end := strings.LastIndex(data, "}")
	if start == -1 || end == -1 {
		return "", errors.New("invalid response format")
	}

	var result struct {
		OpenID string `json:"openid"`
	}

	if err := json.Unmarshal([]byte(data[start:end+1]), &result); err != nil {
		return "", err
	}

	return result.OpenID, nil
}

// 辅助函数：获取微博UID
func getWeiboUID(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.weibo.com/2/account/get_uid.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		UID int64 `json:"uid"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", result.UID), nil
} 