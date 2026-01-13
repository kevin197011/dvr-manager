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
- ✅ **用户认证和权限管理（JWT）**
- ✅ **React + Ant Design 现代化前端界面**
- ✅ **管理后台（配置管理、DVR 服务器管理）**

## 技术栈

- **后端**: Golang + Gin Framework (采用 Nunu 项目结构)
- **前端**: React 18 + Ant Design 5 + Vite
- **认证**: JWT (JSON Web Token)
- **架构**: 前后端分离 + 分层架构 (Handler -> Service -> Repository)
- **数据库**: SQLite (嵌入式关系型数据库)

## 快速开始

### 方式一：Docker 部署（推荐）

**一键部署：**
```bash
./deploy.sh
```

**或手动部署：**
```bash
# 1. 启动服务（前后端分离）
docker-compose up -d

# 2. 访问服务
# 前端: http://localhost (端口 80)
# 后端 API: http://localhost:8080
# 默认账号:
#   - 管理员: admin / admin123
#   - 普通用户: user / user123
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

**使用 GitHub Packages 镜像：**

项目已配置 GitHub Actions 自动构建并推送 Docker 镜像到 GitHub Packages。你可以直接使用预构建的镜像：

```bash
# 1. 登录 GitHub Container Registry（首次使用需要）
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# 2. 拉取镜像
docker pull ghcr.io/OWNER/dvr-manager/backend:latest
docker pull ghcr.io/OWNER/dvr-manager/frontend:latest

# 3. 更新 docker-compose.yml 使用预构建镜像
# 将 build 部分替换为 image 配置
```

或者直接修改 `docker-compose.yml`：

```yaml
services:
  backend:
    image: ghcr.io/OWNER/dvr-manager/backend:latest
    # 移除 build 配置
  frontend:
    image: ghcr.io/OWNER/dvr-manager/frontend:latest
    # 移除 build 配置
```

**注意**：将 `OWNER` 替换为你的 GitHub 用户名或组织名。

### 方式二：本地运行

**后端：**
```bash
# 1. 进入后端目录
cd backend

# 2. 安装依赖
go mod download

# 3. 运行后端服务器
go run ./cmd/server

# 后端运行在 http://localhost:8080
```

**前端：**
```bash
# 1. 进入前端目录
cd frontend

# 2. 安装依赖
npm install

# 3. 启动开发服务器
npm run dev

# 前端运行在 http://localhost:3000
# API 请求自动代理到后端
```

**访问服务：**
- 前端开发服务器: http://localhost:3000
- 默认登录账号:
  - 管理员: admin / admin123（可访问所有功能）
  - 普通用户: user / user123（只能访问录像查询）
- 后端 API: http://localhost:8080

## API 接口

### 认证接口

**登录**
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

响应：
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "username": "admin",
    "role": "admin"
  }
}
```

**获取当前用户信息**
```http
GET /api/auth/me
Authorization: Bearer <token>
```

**登出**
```http
POST /api/auth/logout
Authorization: Bearer <token>
```

### 播放录像（可选认证）

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

### 管理后台 API（需要管理员权限）

**获取系统配置**
```http
GET /api/admin/config
Authorization: Bearer <token>
```

**获取 DVR 服务器列表**
```http
GET /api/admin/dvr-servers
Authorization: Bearer <token>
```

**更新 DVR 服务器列表**
```http
POST /api/admin/dvr-servers
Authorization: Bearer <token>
Content-Type: application/json

{
  "servers": [
    "http://dvr1.example.com:8080/record",
    "http://dvr2.example.com:8080/record"
  ]
}
```

**重新加载配置**
```http
POST /api/admin/reload
Authorization: Bearer <token>
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

## 项目结构

```
dvr-manager/
├── backend/             # 后端项目
│   ├── cmd/
│   │   └── server/      # 应用入口
│   ├── internal/
│   │   ├── config/      # 配置管理
│   │   ├── handler/     # HTTP 处理器
│   │   ├── middleware/  # 中间件
│   │   ├── repository/  # 数据访问层
│   │   ├── router/      # 路由
│   │   └── service/     # 业务逻辑
│   ├── pkg/
│   │   ├── cache/       # 缓存包
│   │   └── db/          # 数据库包
│   ├── go.mod           # Go 模块
│   └── go.sum           # Go 依赖
├── frontend/            # 前端项目
│   ├── src/             # 源代码
│   │   ├── main.jsx     # 应用入口
│   │   ├── App.jsx      # 根组件
│   │   ├── pages/       # 页面组件
│   │   │   ├── Login.jsx    # 登录页
│   │   │   ├── Home.jsx     # 录像查询页
│   │   │   └── Admin.jsx    # 管理后台页
│   │   ├── components/  # 组件
│   │   ├── services/    # API 服务
│   │   └── store/       # 状态管理（Zustand）
│   ├── public/          # 静态资源
│   ├── index.html       # HTML 入口
│   ├── package.json     # 依赖配置
│   ├── vite.config.js   # Vite 配置
│   └── Dockerfile       # 前端 Dockerfile
├── data/                # 数据目录（数据库文件）
├── Dockerfile.backend   # 后端 Dockerfile
└── docker-compose.yml   # Docker Compose 配置
```

## 配置说明

所有配置都存储在 SQLite 数据库中，可通过管理后台进行配置。首次启动时，系统会使用默认配置。

### 环境变量

后端支持以下环境变量：

- `DATA_DIR`: 数据目录路径（默认：`/app/data` 或 `../data`）
- `JWT_SECRET`: JWT 密钥（默认：使用固定密钥，生产环境请务必修改）
- `ADMIN_USERNAME`: 管理员用户名（默认：`admin`）
- `ADMIN_PASSWORD`: 管理员密码（默认：`admin123`）
- `USER_USERNAME`: 普通用户名（默认：`user`）
- `USER_PASSWORD`: 普通用户密码（默认：`user123`）

**生产环境建议：**
```bash
export JWT_SECRET="your-secret-key-here"
export ADMIN_USERNAME="your-admin-username"
export ADMIN_PASSWORD="your-strong-password"
```

### 配置项说明

所有配置可通过管理后台（`/admin`）进行修改：

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

## CI/CD

项目已配置 GitHub Actions 自动构建和推送 Docker 镜像到 GitHub Packages。

### 工作流说明

- **触发条件**：
  - 推送到 `main` 或 `develop` 分支
  - 创建版本标签（`v*`）
  - 手动触发（workflow_dispatch）
  - Pull Request（仅构建，不推送）

- **构建内容**：
  - 后端 Docker 镜像（`ghcr.io/OWNER/dvr-manager/backend`）
  - 前端 Docker 镜像（`ghcr.io/OWNER/dvr-manager/frontend`）

- **镜像标签**：
  - `latest`：主分支的最新构建
  - `分支名`：对应分支的构建
  - `v1.0.0`：版本标签构建
  - `分支名-SHA`：包含提交 SHA 的构建

### 使用预构建镜像

```bash
# 1. 创建 GitHub Personal Access Token (PAT)
# 需要 packages:read 权限

# 2. 登录 GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# 3. 拉取镜像
docker pull ghcr.io/OWNER/dvr-manager/backend:latest
docker pull ghcr.io/OWNER/dvr-manager/frontend:latest
```

## 生产部署

### Docker 部署（推荐）

详细部署文档请查看 [DEPLOY.md](DEPLOY.md)

**快速部署：**
```bash
./deploy.sh
```

**使用预构建镜像部署：**
```bash
# 1. 登录 GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# 2. 使用预构建镜像（需要将 docker-compose.packages.yml 中的 OWNER 替换为你的 GitHub 用户名）
docker-compose -f docker-compose.packages.yml up -d
```

**使用 Nginx 反向代理：**
```bash
# 参考 DEPLOY.md 中的 Nginx 配置
```

### 传统部署

**编译：**
```bash
go build -o dvr-vod-system ./cmd/server
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
