# 安全说明

## 敏感信息保护

### 配置文件安全

**重要提示**：`config.yml` 包含敏感的 DVR 服务器地址和配置信息。

#### ✅ 正确做法

1. **不要提交 config.yml**
   - `config.yml` 已添加到 `.gitignore`
   - 仅提交 `config.example.yml` 作为模板

2. **使用环境变量（可选）**
   ```bash
   export DVR_SERVER_1="http://your-server1.com:8080/record"
   export DVR_SERVER_2="https://your-server2.com:8080/path"
   ```

3. **生产环境部署**
   - 使用 Docker secrets 或 Kubernetes secrets
   - 限制配置文件的访问权限：`chmod 600 config.yml`

#### ❌ 错误做法

- ❌ 将 `config.yml` 提交到 Git
- ❌ 在公开文档中暴露真实的 DVR 地址
- ❌ 在日志中输出完整的 DVR URL（已优化）
- ❌ 在前端暴露真实的 DVR 地址（已使用代理）

## 安全特性

### 1. 视频流代理

系统使用代理模式隐藏真实的 DVR 服务器地址：

- ✅ 前端只能看到代理 URL：`/stream/xxx.mp4`
- ✅ 真实 DVR 地址仅在后端使用
- ✅ 日志中不输出完整的 DVR URL

### 2. HTTPS 证书验证

系统支持配置 HTTPS 证书验证：

```yaml
dvr:
  skip_tls_verify: true  # 跳过证书验证（开发/内网环境）
```

**使用场景**：
- ✅ 开发环境：使用自签名证书
- ✅ 内网环境：DVR 服务器使用自签名证书
- ✅ 测试环境：快速部署，无需配置证书

**安全建议**：
- ⚠️ 生产环境建议使用有效的 SSL 证书
- ⚠️ 公网部署时设置 `skip_tls_verify: false`
- ⚠️ 使用 Let's Encrypt 等免费证书服务

### 2. 访问控制

建议在生产环境中添加：

- 用户认证（JWT、OAuth 等）
- IP 白名单
- 请求频率限制
- HTTPS 加密传输

### 3. 日志安全

系统日志已优化：

- ✅ 只记录录像编号，不记录完整 URL
- ✅ 记录客户端 IP，便于审计
- ✅ 敏感信息已脱敏

## 生产环境建议

### 1. 使用 HTTPS

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://dvr-vod:8080;
    }
}
```

### 2. 添加认证

在 Nginx 中添加基本认证：

```nginx
location / {
    auth_basic "DVR VOD System";
    auth_basic_user_file /etc/nginx/.htpasswd;
    proxy_pass http://dvr-vod:8080;
}
```

### 3. 限制访问

```nginx
# IP 白名单
location / {
    allow 192.168.1.0/24;
    deny all;
    proxy_pass http://dvr-vod:8080;
}

# 请求频率限制
limit_req_zone $binary_remote_addr zone=dvr_limit:10m rate=10r/s;

location /api/ {
    limit_req zone=dvr_limit burst=20;
    proxy_pass http://dvr-vod:8080;
}
```

### 4. 配置文件权限

```bash
# 限制配置文件权限
chmod 600 config.yml
chown root:root config.yml

# Docker 部署时使用 secrets
docker secret create dvr_config config.yml
```

## 漏洞报告

如果发现安全问题，请通过以下方式报告：

1. **不要**在公开的 issue 中讨论安全问题
2. 发送邮件到：security@your-domain.com
3. 提供详细的漏洞描述和复现步骤

## 更新日志

- 2024-11-10: 添加视频流代理，隐藏真实 DVR 地址
- 2024-11-10: 优化日志输出，移除敏感信息
- 2024-11-10: 添加 config.yml 到 .gitignore

## 参考资源

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [Nginx Security](https://nginx.org/en/docs/http/ngx_http_ssl_module.html)
