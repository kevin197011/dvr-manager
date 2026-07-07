# Backend

Go + Gin API 服务。生产环境同时通过 `internal/web` 嵌入并提供前端静态资源。

## 开发

```bash
go mod download
go run ./cmd/server   # :8080，仅 API（无完整前端，见 internal/web/dist 占位页）
```

## 生产构建

在仓库根目录执行：

```bash
make build          # 构建前端 → embed → ./dvr-vod-system
```

或 `docker compose up -d --build`（根目录 `Dockerfile` 多阶段构建）。

## 数据目录

- 本地：`../data`
- Docker：`/app/data`（`DATA_DIR`）

完整 API 与架构见 [docs/REQUIREMENTS.md](../docs/REQUIREMENTS.md)。
