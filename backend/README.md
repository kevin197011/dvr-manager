# DVR Manager Backend

后端服务，基于 Go + Gin Framework，采用 Nunu 项目结构。

## 项目结构

```
backend/
├── cmd/
│   └── server/          # 应用入口
├── internal/
│   ├── config/          # 配置管理
│   ├── handler/         # HTTP 处理器
│   ├── middleware/      # 中间件
│   ├── repository/      # 数据访问层
│   ├── router/          # 路由
│   └── service/         # 业务逻辑
├── pkg/
│   ├── cache/           # 缓存包
│   └── db/              # 数据库包
├── go.mod               # Go 模块
└── go.sum               # Go 依赖
```

## 本地开发

```bash
# 安装依赖
go mod download

# 运行服务器
go run ./cmd/server

# 编译
go build -o dvr-vod-system ./cmd/server
```

## 数据目录

- **本地开发**: `../data` (相对于 backend 目录)
- **Docker 环境**: `/app/data` (通过环境变量 `DATA_DIR` 配置)

## API 接口

- `GET /api/config` - 获取配置信息
- `POST /api/play` - 播放录像
- `GET /api/play` - 播放录像（GET）
- `GET /api/admin/config` - 获取完整配置
- `GET /api/admin/dvr-servers` - 获取 DVR 服务器列表
- `POST /api/admin/dvr-servers` - 更新 DVR 服务器列表
- `POST /api/admin/reload` - 重新加载配置
- `GET /stream/:filename` - 视频流代理
- `GET /health` - 健康检查
