package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"
)

// DVRQueryResult 查询结果
type DVRQueryResult struct {
	URL       string
	ServerIdx int
	Error     error
}

// FindRecording 并发查询多个 DVR 服务器查找录像
func FindRecording(ctx context.Context, recordID string) (string, error) {
	config := GetConfig()
	if config == nil {
		log.Printf("[ERROR] 配置未加载")
		return "", fmt.Errorf("config not loaded")
	}

	filename := recordID + ".mp4"
	log.Printf("[DEBUG] 开始查询录像 - 编号: %s, 文件名: %s", recordID, filename)

	// 创建带超时的 HTTP 客户端，使用配置的超时时间
	timeout := config.DVR.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second // 默认 10 秒
	}
	
	// 创建自定义 Transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.DVR.SkipTLSVerify, // 根据配置决定是否跳过证书验证
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 不跟随重定向
		},
	}
	
	log.Printf("[DEBUG] HTTP 客户端配置 - 跳过 HTTPS 证书验证: %v", config.DVR.SkipTLSVerify)
	
	// 获取重试次数配置
	maxRetries := config.DVR.Retry
	if maxRetries == 0 {
		maxRetries = 3 // 默认重试 3 次
	}
	
	log.Printf("[DEBUG] DVR 查询配置 - 超时: %v, 重试次数: %d, 并发查询: %d 个服务器", timeout, maxRetries, len(config.DVRServers))

	// 使用通道接收查询结果
	resultChan := make(chan DVRQueryResult, len(config.DVRServers))
	
	// 并发查询所有 DVR 服务器
	for i, serverURL := range config.DVRServers {
		go func(serverIdx int, server string) {
			url := server
			if url[len(url)-1] != '/' {
				url += "/"
			}
			url += filename
			
			log.Printf("[DEBUG] 并发查询服务器 [%d/%d] - URL: %s", serverIdx+1, len(config.DVRServers), url)
			
			// 查询该服务器
			foundURL, err := queryServer(ctx, client, url, serverIdx, len(config.DVRServers), maxRetries)
			
			resultChan <- DVRQueryResult{
				URL:       foundURL,
				ServerIdx: serverIdx,
				Error:     err,
			}
		}(i, serverURL)
	}
	
	// 等待第一个成功的结果或所有失败
	var lastErr error
	for i := 0; i < len(config.DVRServers); i++ {
		select {
		case <-ctx.Done():
			log.Printf("[WARN] 查询被取消 - 编号: %s", recordID)
			return "", ctx.Err()
		case result := <-resultChan:
			if result.Error == nil && result.URL != "" {
				log.Printf("[SUCCESS] 录像找到 - 编号: %s, URL: %s, 服务器: [%d/%d]",
					recordID, result.URL, result.ServerIdx+1, len(config.DVRServers))
				return result.URL, nil
			}
			lastErr = result.Error
		}
	}

	log.Printf("[WARN] 所有服务器均未找到录像 - 编号: %s", recordID)
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("recording not found: %s", recordID)
}

// queryServer 查询单个服务器（带重试）
func queryServer(ctx context.Context, client *http.Client, url string, serverIdx, totalServers, maxRetries int) (string, error) {
	var lastErr error
	
	for retry := 0; retry < maxRetries; retry++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if retry > 0 {
			log.Printf("[INFO] 重试 [%d/%d] - 服务器: [%d/%d], URL: %s",
				retry, maxRetries-1, serverIdx+1, totalServers, url)
			// 重试前等待一小段时间（指数退避）
			time.Sleep(time.Duration(retry) * 500 * time.Millisecond)
		}

		// 发送 HEAD 请求检查文件是否存在
		req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
		if err != nil {
			log.Printf("[ERROR] 创建请求失败 - URL: %s, Error: %v", url, err)
			lastErr = err
			continue
		}

		startTime := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(startTime)

		if err != nil {
			lastErr = err
			if retry < maxRetries-1 {
				log.Printf("[WARN] 请求失败 [%d/%d] (将重试) - URL: %s, 耗时: %v, Error: %v",
					serverIdx+1, totalServers, url, duration, err)
			} else {
				log.Printf("[WARN] 请求失败 [%d/%d] (已达最大重试) - URL: %s, 耗时: %v, Error: %v",
					serverIdx+1, totalServers, url, duration, err)
			}
			continue
		}
		resp.Body.Close()

		log.Printf("[DEBUG] 服务器响应 [%d/%d] - URL: %s, 状态码: %d, 耗时: %v",
			serverIdx+1, totalServers, url, resp.StatusCode, duration)

		// 如果返回 200 或 302，认为文件存在
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
			if retry > 0 {
				log.Printf("[SUCCESS] 录像找到 (重试 %d 次后成功) - URL: %s, 服务器: [%d/%d]",
					retry, url, serverIdx+1, totalServers)
			}
			return url, nil
		}

		// 如果是 404，不需要重试，直接返回
		if resp.StatusCode == http.StatusNotFound {
			log.Printf("[DEBUG] 文件不存在 (404) - 服务器: [%d/%d]", serverIdx+1, totalServers)
			return "", fmt.Errorf("not found")
		}

		// 其他状态码，记录并重试
		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	return "", lastErr
}
