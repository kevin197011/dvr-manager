package service

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"dvr-vod-system/internal/config"
	"dvr-vod-system/pkg/httpclient"
)

// ProxyService 代理服务接口
type ProxyService interface {
	ProxyStream(ctx context.Context, recordID, realURL string, w http.ResponseWriter, rangeHeader string) error
}

type proxyService struct {
	clientMu   sync.Mutex
	httpClient *http.Client
	clientTLS  bool
	clientTO   time.Duration
}

// NewProxyService 创建代理服务
func NewProxyService(_ *config.Config) ProxyService {
	return &proxyService{}
}

func (s *proxyService) streamClient(cfg *config.Config) *http.Client {
	timeout := 5 * time.Minute
	if cfg != nil && cfg.DVR.Timeout > 0 {
		timeout = cfg.DVR.Timeout * 30
		if timeout < 60*time.Second {
			timeout = 60 * time.Second
		}
	}
	skipTLS := cfg != nil && cfg.DVR.SkipTLSVerify
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	if s.httpClient == nil || s.clientTLS != skipTLS || s.clientTO != timeout {
		s.httpClient = httpclient.New(timeout, skipTLS, true)
		s.clientTLS = skipTLS
		s.clientTO = timeout
	}
	return s.httpClient
}

// ProxyStream 代理视频流
func (s *proxyService) ProxyStream(ctx context.Context, recordID, realURL string, w http.ResponseWriter, rangeHeader string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", realURL, nil)
	if err != nil {
		return err
	}
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	client := s.streamClient(config.GetConfig())
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] 代理请求失败 - 编号: %s, Error: %v", recordID, err)
		return err
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	written, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("[WARN] 流传输中断 - 编号: %s, 已传输: %d bytes, Error: %v", recordID, written, err)
		return err
	}
	log.Printf("[SUCCESS] 流传输完成 - 编号: %s, 传输: %d bytes", recordID, written)
	return nil
}
