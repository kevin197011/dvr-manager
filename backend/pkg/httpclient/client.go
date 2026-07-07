package httpclient

import (
	"crypto/tls"
	"net/http"
	"time"
)

// New 创建 HTTP 客户端；followRedirect=false 时不跟随重定向（DVR HEAD 探测用）。
func New(timeout time.Duration, skipTLSVerify, followRedirect bool) *http.Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	if !followRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client
}
