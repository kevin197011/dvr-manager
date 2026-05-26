package service

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"dvr-vod-system/internal/repository"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCConfig OIDC 提供商配置
type OIDCConfig struct {
	Issuer        string   `json:"issuer"`
	ClientID      string   `json:"client_id"`
	ClientSecret  string   `json:"client_secret"`
	RedirectURL   string   `json:"redirect_url"`
	Scopes        []string `json:"scopes"`
	UsernameClaim string   `json:"username_claim"` // 默认 preferred_username，缺失时回退到 email
	SkipTLSVerify bool     `json:"skip_tls_verify"`
}

// SSOProviderInfo 用于前端展示
type SSOProviderInfo struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// oidcRuntime 已初始化的 OIDC 运行时实体
type oidcRuntime struct {
	id          int64
	name        string
	cfg         OIDCConfig
	provider    *oidc.Provider
	verifier    *oidc.IDTokenVerifier
	oauthConfig *oauth2.Config
}

// SSOService SSO 服务（仅 OIDC）
type SSOService interface {
	ListProviders() ([]SSOProviderInfo, error)
	Reload() error

	BuildOIDCAuthURL(id int64, state string) (string, error)
	ExchangeOIDC(ctx context.Context, id int64, code string) (username string, err error)
}

type ssoService struct {
	repo repository.SSORepository

	mu      sync.RWMutex
	oidcSet map[int64]*oidcRuntime
	all     []repository.SSOProvider
}

// NewSSOService 创建 SSO 服务
func NewSSOService(repo repository.SSORepository) SSOService {
	s := &ssoService{repo: repo}
	if err := s.Reload(); err != nil {
		fmt.Printf("[SSO] reload providers failed: %v\n", err)
	}
	return s
}

// Reload 重新加载所有已启用提供商
func (s *ssoService) Reload() error {
	all, err := s.repo.List()
	if err != nil {
		return err
	}

	newOIDC := make(map[int64]*oidcRuntime)
	for _, p := range all {
		if !p.Enabled {
			continue
		}
		if p.Type != repository.SSOTypeOIDC {
			continue
		}
		rt, err := buildOIDCRuntime(p)
		if err != nil {
			fmt.Printf("[SSO] OIDC provider %d (%s) init failed: %v\n", p.ID, p.Name, err)
			continue
		}
		newOIDC[p.ID] = rt
	}

	s.mu.Lock()
	s.oidcSet = newOIDC
	s.all = all
	s.mu.Unlock()
	return nil
}

// ListProviders 列出已成功加载、且启用的提供商
func (s *ssoService) ListProviders() ([]SSOProviderInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	infos := make([]SSOProviderInfo, 0)
	for _, p := range s.all {
		if !p.Enabled || p.Type != repository.SSOTypeOIDC {
			continue
		}
		if _, ok := s.oidcSet[p.ID]; !ok {
			continue
		}
		infos = append(infos, SSOProviderInfo{ID: p.ID, Type: p.Type, Name: p.Name})
	}
	return infos, nil
}

// ---------------- OIDC ----------------

func buildOIDCRuntime(p repository.SSOProvider) (*oidcRuntime, error) {
	var cfg OIDCConfig
	if err := json.Unmarshal([]byte(p.ConfigJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid oidc config: %w", err)
	}
	if cfg.Issuer == "" || cfg.ClientID == "" || cfg.RedirectURL == "" {
		return nil, errors.New("issuer / client_id / redirect_url 不能为空")
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}
	if cfg.UsernameClaim == "" {
		cfg.UsernameClaim = "preferred_username"
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	if cfg.SkipTLSVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	ctx := oidc.ClientContext(context.Background(), httpClient)

	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("discover issuer: %w", err)
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
	}
	return &oidcRuntime{
		id:          p.ID,
		name:        p.Name,
		cfg:         cfg,
		provider:    provider,
		verifier:    verifier,
		oauthConfig: oauthCfg,
	}, nil
}

// BuildOIDCAuthURL 生成 OIDC 授权 URL
func (s *ssoService) BuildOIDCAuthURL(id int64, state string) (string, error) {
	s.mu.RLock()
	rt := s.oidcSet[id]
	s.mu.RUnlock()
	if rt == nil {
		return "", errors.New("OIDC 提供商不存在或未启用")
	}
	return rt.oauthConfig.AuthCodeURL(state), nil
}

// ExchangeOIDC 用 authorization code 换取 ID token，并提取用户名
func (s *ssoService) ExchangeOIDC(ctx context.Context, id int64, code string) (string, error) {
	s.mu.RLock()
	rt := s.oidcSet[id]
	s.mu.RUnlock()
	if rt == nil {
		return "", errors.New("OIDC 提供商不存在或未启用")
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	if rt.cfg.SkipTLSVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	ctx = oidc.ClientContext(ctx, httpClient)

	tok, err := rt.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("exchange code: %w", err)
	}
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", errors.New("响应中缺少 id_token")
	}
	idToken, err := rt.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", fmt.Errorf("verify id_token: %w", err)
	}
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("decode claims: %w", err)
	}
	username := pickStringClaim(claims, rt.cfg.UsernameClaim, "preferred_username", "email", "sub")
	if username == "" {
		return "", errors.New("无法从 id_token 提取用户名")
	}
	return username, nil
}

func pickStringClaim(claims map[string]interface{}, names ...string) string {
	for _, n := range names {
		if v, ok := claims[n].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// GenerateState 生成随机 state（OIDC CSRF）
func GenerateState() (string, error) {
	b := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
