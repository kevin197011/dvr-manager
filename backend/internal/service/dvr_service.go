package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"dvr-manager/internal/config"
	"dvr-manager/internal/repository"
	"dvr-manager/pkg/httpclient"
)

// DVRService DVR 服务接口
type DVRService interface {
	FindRecording(ctx context.Context, recordID string) (string, error)
}

type dvrService struct {
	repo       repository.DVRRepository
	clientMu   sync.Mutex
	httpClient *http.Client
	clientTLS  bool
	clientTO   time.Duration
}

// NewDVRService 创建 DVR 服务
func NewDVRService(_ *config.Config, repo repository.DVRRepository) DVRService {
	return &dvrService{repo: repo}
}

func (s *dvrService) probeClient(cfg *config.Config) *http.Client {
	timeout := cfg.DVR.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	if s.httpClient == nil || s.clientTLS != cfg.DVR.SkipTLSVerify || s.clientTO != timeout {
		s.httpClient = httpclient.New(timeout, cfg.DVR.SkipTLSVerify, false)
		s.clientTLS = cfg.DVR.SkipTLSVerify
		s.clientTO = timeout
	}
	return s.httpClient
}

// DVRQueryResult 查询结果
type DVRQueryResult struct {
	URL       string
	ServerIdx int
	Error     error
}

// FindRecording 并发查询多个 DVR 服务器
func (s *dvrService) FindRecording(ctx context.Context, recordID string) (string, error) {
	cfg := config.GetConfig()
	if cfg == nil {
		return "", fmt.Errorf("config not loaded")
	}

	filename := recordID + ".mp4"
	client := s.probeClient(cfg)

	maxRetries := cfg.DVR.Retry
	if maxRetries == 0 {
		maxRetries = 3
	}

	dvrServers := s.listServers(cfg)
	if len(dvrServers) == 0 {
		return "", fmt.Errorf("no dvr servers configured")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultChan := make(chan DVRQueryResult, len(dvrServers))

	for i, serverURL := range dvrServers {
		go func(serverIdx int, server string) {
			url := server
			if url[len(url)-1] != '/' {
				url += "/"
			}
			url += filename
			foundURL, err := s.queryServer(ctx, client, url, serverIdx, len(dvrServers), maxRetries)
			select {
			case <-ctx.Done():
				return
			case resultChan <- DVRQueryResult{URL: foundURL, ServerIdx: serverIdx, Error: err}:
			}
		}(i, serverURL)
	}

	var lastErr error
	for i := 0; i < len(dvrServers); i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case result := <-resultChan:
			if result.Error == nil && result.URL != "" {
				cancel()
				log.Printf("[SUCCESS] 录像找到 - 编号: %s, URL: %s", recordID, result.URL)
				return result.URL, nil
			}
			lastErr = result.Error
		}
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("recording not found: %s", recordID)
}

func (s *dvrService) listServers(cfg *config.Config) []string {
	var raw []string
	if s.repo != nil {
		if servers, err := s.repo.GetAll(); err == nil && len(servers) > 0 {
			raw = servers
		}
	}
	if len(raw) == 0 {
		raw = cfg.DVRServers
	}
	out := make([]string, 0, len(raw))
	for _, u := range raw {
		u = strings.TrimSpace(u)
		if u != "" {
			out = append(out, u)
		}
	}
	return out
}

func (s *dvrService) queryServer(ctx context.Context, client *http.Client, url string, serverIdx, totalServers, maxRetries int) (string, error) {
	var lastErr error

	for retry := 0; retry < maxRetries; retry++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if retry > 0 {
			time.Sleep(time.Duration(retry) * 500 * time.Millisecond)
		}

		req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
			return url, nil
		}
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("not found")
		}
		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return "", lastErr
}
