# DVR Manager

DVR 录像点播管理系统：按编号在多台 DVR 服务器上查找录像，通过代理 URL 在线播放与下载，不暴露真实 DVR 地址。前后端分离，带用户认证与管理后台。

## 功能

- 录像单个/批量查询，并发探测多台 DVR
- 视频流代理播放（支持 Range 拖动）与下载
- JWT 登录、角色权限（admin / user）
- OIDC 单点登录（可选）
- 管理后台：DVR 服务器、系统配置、用户、SSO、审计日志
- 审计日志默认保留 3 个月，启动与每日自动清理

## 技术栈

Go · Gin · SQLite · React 18 · Ant Design 5 · Vite · Docker

## 快速开始

### Docker（推荐）

```bash
./deploy.sh
# 或
docker compose up -d --build
```

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3000 |
| 后端 API | http://localhost:8080 |

默认账号（首次启动、用户表为空时创建）：

| 角色 | 用户名 | 密码 |
|------|--------|------|
| 管理员 | `admin` | `admin123` |
| 普通用户 | `user` | `user123` |

```bash
docker compose logs -f    # 查看日志
docker compose down       # 停止
```

数据持久化在 `./data`（SQLite 数据库）。

### 本地开发

```bash
# 后端
cd backend && go mod download && go run ./cmd/server

# 前端（另开终端）
cd frontend && npm install && npm run dev
```

前端 http://localhost:3000 ，API 自动代理到后端 :8080。

## 生产环境

部署前请至少修改以下环境变量（见 `docker-compose.yml`）：

| 变量 | 说明 |
|------|------|
| `JWT_SECRET` | JWT 签名密钥 |
| `ADMIN_USERNAME` / `ADMIN_PASSWORD` | 初始管理员 |
| `USER_USERNAME` / `USER_PASSWORD` | 初始普通用户（可选） |

其他常用变量：`DATA_DIR`、`RECORD_CACHE_TTL_DAYS`（默认 30）、`AUDIT_RETENTION_MONTHS`（默认 3）、`REQUIRE_AUTH_FOR_PLAY`（默认 false，设为 true 时播放需登录）。

使用 Nginx 反向代理时参考 [NGINX.md](NGINX.md)。

## 文档

| 文档 | 说明 |
|------|------|
| [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) | 完整需求、API、数据模型、架构与安全说明 |
| [NGINX.md](NGINX.md) | Nginx 反向代理配置 |

## 许可证

MIT
