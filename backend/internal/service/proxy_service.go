package service

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net/http"

	"dvr-vod-system/internal/config"
)

// ProxyService 代理服务接口
type ProxyService interface {
	ProxyStream(ctx context.Context, recordID, realURL string, w http.ResponseWriter, rangeHeader string) error
}

// proxyService 代理服务实现
type proxyService struct {
	config *config.Config
}

// NewProxyService 创建新的代理服务
func NewProxyService(cfg *config.Config) ProxyService {
	return &proxyService{
		config: cfg,
	}
}

// ProxyStream 代理视频流
func (s *proxyService) ProxyStream(ctx context.Context, recordID, realURL string, w http.ResponseWriter, rangeHeader string) error {
	log.Printf("[INFO] 流代理请求 - 编号: %s", recordID)

	// 创建代理请求
	req, err := http.NewRequestWithContext(ctx, "GET", realURL, nil)
	if err != nil {
		log.Printf("[ERROR] 创建代理请求失败 - 编号: %s, Error: %v", recordID, err)
		return err
	}

	// 转发客户端的 Range 请求头（支持视频拖动）
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	// 动态获取最新配置
	cfg := config.GetConfig()
	skipTLSVerify := false
	if cfg != nil {
		skipTLSVerify = cfg.DVR.SkipTLSVerify
	}

	// 创建自定义 Transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
		},
	}

	// 发送请求到真实 DVR 服务器
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] 代理请求失败 - 编号: %s, Error: %v", recordID, err)
		return err
	}
	defer resp.Body.Close()

	log.Printf("[INFO] 流代理成功 - 编号: %s, 状态码: %d, 大小: %s",
		recordID, resp.StatusCode, resp.Header.Get("Content-Length"))

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 流式传输视频数据
	written, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("[WARN] 流传输中断 - 编号: %s, 已传输: %d bytes, Error: %v",
			recordID, written, err)
		return err
	}

	log.Printf("[SUCCESS] 流传输完成 - 编号: %s, 传输: %d bytes", recordID, written)
	return nil
}
