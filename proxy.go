package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// StreamProxyHandler 处理视频流代理请求
func StreamProxyHandler(c *gin.Context) {
	// 从路径中提取录像编号
	filename := c.Param("filename")
	recordID := strings.TrimSuffix(filename, ".mp4")

	log.Printf("[INFO] 流代理请求 - IP: %s, 编号: %s", c.ClientIP(), recordID)

	// 从缓存获取真实 URL
	realURL, exists := GetCachedURL(recordID)
	if !exists {
		log.Printf("[WARN] 流代理失败 - 编号: %s, 原因: URL 未缓存", recordID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "recording not found or expired",
		})
		return
	}

	// 创建代理请求
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", realURL, nil)
	if err != nil {
		log.Printf("[ERROR] 创建代理请求失败 - 编号: %s, Error: %v", recordID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create proxy request",
		})
		return
	}

	// 转发客户端的 Range 请求头（支持视频拖动）
	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	// 获取配置
	config := GetConfig()
	skipTLSVerify := false
	if config != nil {
		skipTLSVerify = config.DVR.SkipTLSVerify
	}
	
	// 创建自定义 Transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify, // 根据配置决定是否跳过证书验证
		},
	}
	
	// 发送请求到真实 DVR 服务器
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] 代理请求失败 - 编号: %s, Error: %v", recordID, err)
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "failed to fetch video from DVR server",
		})
		return
	}
	defer resp.Body.Close()

	log.Printf("[INFO] 流代理成功 - 编号: %s, 状态码: %d, 大小: %s",
		recordID, resp.StatusCode, resp.Header.Get("Content-Length"))

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 设置状态码
	c.Status(resp.StatusCode)

	// 流式传输视频数据
	written, err := io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("[WARN] 流传输中断 - 编号: %s, 已传输: %d bytes, Error: %v",
			recordID, written, err)
		return
	}

	log.Printf("[SUCCESS] 流传输完成 - 编号: %s, 传输: %d bytes", recordID, written)
}
