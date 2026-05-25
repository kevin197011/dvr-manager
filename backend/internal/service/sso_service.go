package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"dvr-vod-system/internal/repository"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
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

// SAMLConfig SAML 提供商配置
type SAMLConfig struct {
	IdPMetadataURL string `json:"idp_metadata_url"` // 优先使用 URL
	IdPMetadataXML string `json:"idp_metadata_xml"` // 或者粘贴 XML
	SPEntityID     string `json:"sp_entity_id"`     // 通常等于 base_url 或自定义 URI
	BaseURL        string `json:"base_url"`         // SP 对外可访问地址，例如 https://dvr.example.com
	UsernameAttr   string `json:"username_attr"`    // 用户名属性，默认 NameID
	SPCertPEM      string `json:"sp_cert_pem"`      // SP 证书（PEM）
	SPKeyPEM       string `json:"sp_key_pem"`       // SP 私钥（PEM）
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

// samlRuntime 已初始化的 SAML 运行时
type samlRuntime struct {
	id   int64
	name string
	cfg  SAMLConfig
	sp   *samlsp.Middleware
}

// SSOService SSO 服务
type SSOService interface {
	ListProviders() ([]SSOProviderInfo, error)
	Reload() error

	// OIDC
	BuildOIDCAuthURL(id int64, state string) (string, error)
	ExchangeOIDC(ctx context.Context, id int64, code string) (username string, err error)

	// SAML
	GetSAMLMiddleware(id int64) (*samlsp.Middleware, error)
	GetSAMLConfig(id int64) (SAMLConfig, error)
}

type ssoService struct {
	repo repository.SSORepository

	mu      sync.RWMutex
	oidcSet map[int64]*oidcRuntime
	samlSet map[int64]*samlRuntime
	all     []repository.SSOProvider
}

// NewSSOService 创建 SSO 服务
func NewSSOService(repo repository.SSORepository) SSOService {
	s := &ssoService{repo: repo}
	if err := s.Reload(); err != nil {
		// 首次加载失败仅记录到 stderr，不阻塞启动
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
	newSAML := make(map[int64]*samlRuntime)
	for _, p := range all {
		if !p.Enabled {
			continue
		}
		switch p.Type {
		case repository.SSOTypeOIDC:
			rt, err := buildOIDCRuntime(p)
			if err != nil {
				fmt.Printf("[SSO] OIDC provider %d (%s) init failed: %v\n", p.ID, p.Name, err)
				continue
			}
			newOIDC[p.ID] = rt
		case repository.SSOTypeSAML:
			rt, err := buildSAMLRuntime(p)
			if err != nil {
				fmt.Printf("[SSO] SAML provider %d (%s) init failed: %v\n", p.ID, p.Name, err)
				continue
			}
			newSAML[p.ID] = rt
		}
	}

	s.mu.Lock()
	s.oidcSet = newOIDC
	s.samlSet = newSAML
	s.all = all
	s.mu.Unlock()
	return nil
}

// ListProviders 列出已启用的提供商（仅展示信息）
func (s *ssoService) ListProviders() ([]SSOProviderInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	infos := make([]SSOProviderInfo, 0)
	for _, p := range s.all {
		if !p.Enabled {
			continue
		}
		// 只输出确实成功加载的
		if p.Type == repository.SSOTypeOIDC {
			if _, ok := s.oidcSet[p.ID]; !ok {
				continue
			}
		}
		if p.Type == repository.SSOTypeSAML {
			if _, ok := s.samlSet[p.ID]; !ok {
				continue
			}
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

// ---------------- SAML ----------------

func buildSAMLRuntime(p repository.SSOProvider) (*samlRuntime, error) {
	var cfg SAMLConfig
	if err := json.Unmarshal([]byte(p.ConfigJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid saml config: %w", err)
	}
	if cfg.BaseURL == "" {
		return nil, errors.New("base_url 不能为空")
	}
	if cfg.SPCertPEM == "" || cfg.SPKeyPEM == "" {
		return nil, errors.New("sp_cert_pem / sp_key_pem 不能为空")
	}

	keyBlock, _ := pem.Decode([]byte(cfg.SPKeyPEM))
	if keyBlock == nil {
		return nil, errors.New("解析 sp_key_pem 失败")
	}
	var rsaKey *rsa.PrivateKey
	if k, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes); err == nil {
		rsaKey = k
	} else if pk, err2 := x509.ParsePKCS8PrivateKey(keyBlock.Bytes); err2 == nil {
		if rk, ok := pk.(*rsa.PrivateKey); ok {
			rsaKey = rk
		} else {
			return nil, errors.New("sp 私钥不是 RSA")
		}
	} else {
		return nil, fmt.Errorf("解析私钥失败: %v / %v", err, err2)
	}

	certBlock, _ := pem.Decode([]byte(cfg.SPCertPEM))
	if certBlock == nil {
		return nil, errors.New("解析 sp_cert_pem 失败")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("base_url 不合法: %w", err)
	}

	// IdP 元数据：URL 优先，否则使用 XML
	var idpMeta *saml.EntityDescriptor
	if strings.TrimSpace(cfg.IdPMetadataURL) != "" {
		idpURL, err := url.Parse(cfg.IdPMetadataURL)
		if err != nil {
			return nil, fmt.Errorf("idp_metadata_url 不合法: %w", err)
		}
		md, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpURL)
		if err != nil {
			return nil, fmt.Errorf("拉取 IdP metadata 失败: %w", err)
		}
		idpMeta = md
	} else if strings.TrimSpace(cfg.IdPMetadataXML) != "" {
		md, err := samlsp.ParseMetadata([]byte(cfg.IdPMetadataXML))
		if err != nil {
			return nil, fmt.Errorf("解析 IdP metadata 失败: %w", err)
		}
		idpMeta = md
	} else {
		return nil, errors.New("必须提供 idp_metadata_url 或 idp_metadata_xml")
	}

	sp, err := samlsp.New(samlsp.Options{
		EntityID:    cfg.SPEntityID,
		URL:         *baseURL,
		Key:         rsaKey,
		Certificate: cert,
		IDPMetadata: idpMeta,
	})
	if err != nil {
		return nil, fmt.Errorf("init saml sp: %w", err)
	}
	// 每个提供商使用独立的 ACS / metadata 路径：/api/auth/sso/saml/:id/{acs|metadata|login}
	prefix := fmt.Sprintf("/api/auth/sso/saml/%d", p.ID)
	sp.ServiceProvider.AcsURL = *resolveURL(baseURL, prefix+"/acs")
	sp.ServiceProvider.MetadataURL = *resolveURL(baseURL, prefix+"/metadata")
	sp.ServiceProvider.SloURL = *resolveURL(baseURL, prefix+"/slo")

	return &samlRuntime{id: p.ID, name: p.Name, cfg: cfg, sp: sp}, nil
}

func resolveURL(base *url.URL, path string) *url.URL {
	ref, _ := url.Parse(path)
	u := base.ResolveReference(ref)
	return u
}

// GetSAMLMiddleware 获取指定 ID 的 SAML SP 中间件
func (s *ssoService) GetSAMLMiddleware(id int64) (*samlsp.Middleware, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rt := s.samlSet[id]
	if rt == nil {
		return nil, errors.New("SAML 提供商不存在或未启用")
	}
	return rt.sp, nil
}

// GetSAMLConfig 返回 SAML 配置（供 handler 提取用户名属性）
func (s *ssoService) GetSAMLConfig(id int64) (SAMLConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rt := s.samlSet[id]
	if rt == nil {
		return SAMLConfig{}, errors.New("SAML 提供商不存在或未启用")
	}
	return rt.cfg, nil
}

// GenerateState 生成随机 state（OIDC CSRF）
func GenerateState() (string, error) {
	b := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
