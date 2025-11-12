# DVR 点播系统

一个基于 Golang 的 DVR 录像点播系统后端，支持多个 DVR 服务器串行查询，多用户并发访问互不干扰。

## 功能特点

- ✅ 串行查询多个 DVR 服务器
- ✅ 自动组合录像文件名（编号 + .mp4）
- ✅ 支持多用户并发访问，互不干扰
- ✅ 请求超时控制
- ✅ RESTful API 接口
- ✅ 跨域支持（CORS）
- ✅ 简洁的前端界面
- ✅ **视频流代理，隐藏真实 DVR 地址**
- ✅ 支持视频拖动（Range 请求）
- ✅ 系统运行状态监控
- ✅ 视频下载功能

## 技术栈

- **后端**: Golang + Gin Framework
- **前端**: HTML + CSS + JavaScript

## 快速开始

### 方式一：Docker 部署（推荐）

**一键部署：**
```bash
./deploy.sh
```

**或手动部署：**
```bash
# 1. 配置 DVR 服务器（编辑 config.yml）
cp config.example.yml config.yml
vi config.yml

# 2. 启动服务
docker-compose up -d

# 3. 访问服务
# 浏览器打开 http://localhost:8080
```

**查看日志：**
```bash
docker-compose logs -f
```

**停止服务：**
```bash
docker-compose down
```

**使用 Nginx 反向代理**：
```bash
# 使用 Nginx 版本部署
docker-compose -f docker-compose.nginx.yml up -d

# 访问服务
# HTTP: http://localhost
# HTTPS: https://localhost (需配置 SSL)
```

详细 Nginx 配置请查看 [NGINX.md](NGINX.md)

### 方式二：本地运行

**1. 安装依赖**
```bash
go mod download
```

**2. 配置 DVR 服务器**

编辑 `config.yml` 文件：
```yaml
server:
  port: 8080
  timeout: 30s

dvr:
  timeout: 10s
  retry: 3

dvr_servers:
  - http://your-dvr-server1.com:8080/record
  - https://your-dvr-server2.com:8080/path
  - https://your-dvr-server3.com:8080/videos

cors:
  enabled: true
  allow_origins: "*"
```

**3. 运行服务器**
```bash
go run .
```

**4. 访问服务**

浏览器打开：http://localhost:8080

## API 接口

### 播放录像

**单个查询**

```http
POST /api/play
Content-Type: application/json

{
  "record_id": "GT03225A120DV"
}
```

或使用 GET 请求：

```http
GET /api/play?record_id=GT03225A120DV
```

**批量查询**

```http
POST /api/play
Content-Type: application/json

{
  "record_ids": ["GT03225A120DV", "GT03225A120DW", "GT03225A120DX"]
}
```

**响应**

单个查询成功：
```json
{
  "success": true,
  "proxy_url": "/stream/GT03225A120DV.mp4",
  "message": "recording found"
}
```

批量查询成功：
```json
{
  "success": true,
  "results": [
    {
      "record_id": "GT03225A120DV",
      "found": true,
      "proxy_url": "/stream/GT03225A120DV.mp4"
    },
    {
      "record_id": "GT03225A120DW",
      "found": false
    }
  ],
  "message": "batch query completed"
}
```

失败：
```json
{
  "success": false,
  "message": "recording not found"
}
```

### 健康检查

```http
GET /health
```

## 工作原理

1. 前端输入录像编号（如 `GT03225A120DV`）
2. 后端自动组合成 `GT03225A120DV.mp4`
3. **并发查询**所有配置的 DVR 服务器（而非串行）
4. 使用 HEAD 请求检查文件是否存在
5. 支持自动重试机制（默认 3 次）
6. 找到第一个可用录像后立即返回
7. 生成代理 URL（如 `/stream/GT03225A120DV.mp4`）
8. 前端通过代理 URL 播放视频
9. 后端代理转发视频流，隐藏真实 DVR 地址

## 并发处理

- 每个请求使用独立的 context，互不干扰
- 支持多用户同时查询不同的录像
- 请求超时自动取消，不影响其他请求

## 配置说明

所有配置都在 `config.yml` 文件中：

### 服务器配置
- `server.port`: 服务器监听端口（默认 8080）
- `server.timeout`: 请求总超时时间

### DVR 查询配置
- `dvr.timeout`: 单个 DVR 服务器查询超时时间（建议 10-15 秒）
- `dvr.retry`: 查询失败时的重试次数（建议 2-3 次）
  - 支持自动重试机制
  - 使用指数退避策略（每次重试间隔递增）
  - 404 错误不会重试，直接尝试下一个服务器
- `dvr.skip_tls_verify`: 跳过 HTTPS 证书验证（默认 true）
  - 适用于自签名证书或内网环境
  - 生产环境建议使用有效证书并设置为 false

### DVR 服务器配置
- `dvr_servers`: DVR 服务器列表，按顺序串行查询
  - 直接填写服务器 URL 地址
  - 系统会自动在 URL 后添加 `/录像编号.mp4`
  - 如果 URL 不以 `/` 结尾，会自动添加
  - 支持失败自动重试

### CORS 配置
- `cors.enabled`: 是否启用跨域支持
- `cors.allow_origins`: 允许的来源（`*` 表示所有）
- `cors.allow_methods`: 允许的 HTTP 方法
- `cors.allow_headers`: 允许的请求头

## 测试

**单个查询**：
```bash
curl -X POST http://localhost:8080/api/play \
  -H "Content-Type: application/json" \
  -d '{"record_id":"GT03225A120DV"}'
```

**批量查询**：
```bash
curl -X POST http://localhost:8080/api/play \
  -H "Content-Type: application/json" \
  -d '{"record_ids":["GT03225A120DV","GT03225A120DW","GT03225A120DX"]}'
```

**健康检查**：
```bash
curl http://localhost:8080/health
```

**获取配置信息**：
```bash
curl http://localhost:8080/api/config
```

响应示例：
```json
{
  "server_port": 8080,
  "dvr_count": 3,
  "retry_enabled": true,
  "retry_count": 3,
  "version": "1.0.0"
}
```

## 生产部署

### Docker 部署（推荐）

详细部署文档请查看 [DEPLOY.md](DEPLOY.md)

**快速部署：**
```bash
./deploy.sh
```

**使用 Nginx 反向代理：**
```bash
# 参考 DEPLOY.md 中的 Nginx 配置
```

### 传统部署

**编译：**
```bash
go build -o dvr-vod-system
```

**运行：**
```bash
./dvr-vod-system
```

**使用 systemd（Linux）：**

创建 `/etc/systemd/system/dvr-vod.service`：
```ini
[Unit]
Description=DVR VOD System
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/dvr-vod-system
ExecStart=/path/to/dvr-vod-system
Restart=always

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl enable dvr-vod
sudo systemctl start dvr-vod
```

## 生成 Favicon

在浏览器中打开 `generate_ico.html`，点击"下载 favicon.ico"按钮，将下载的文件放到项目根目录即可。

## 许可证

MIT
