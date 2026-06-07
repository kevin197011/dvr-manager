# Graph Report - dvr-manager  (2026-06-07)

## Corpus Check
- 68 files · ~32,000 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 866 nodes · 961 edges · 71 communities (62 shown, 9 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 44 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `5c78edf4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 56|Community 56]]
- [[_COMMUNITY_Community 57|Community 57]]
- [[_COMMUNITY_Community 58|Community 58]]
- [[_COMMUNITY_Community 59|Community 59]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Community 61|Community 61]]
- [[_COMMUNITY_Community 62|Community 62]]
- [[_COMMUNITY_Community 63|Community 63]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Community 66|Community 66]]
- [[_COMMUNITY_Community 67|Community 67]]

## God Nodes (most connected - your core abstractions)
1. `importMap` - 79 edges
2. `importMap` - 79 edges
3. `exportsByPath` - 54 edges
4. `NewRouter()` - 29 edges
5. `byLanguage` - 16 edges
6. `byLanguage` - 16 edges
7. `OpenSpec Instructions` - 15 edges
8. `DVR 点播系统` - 14 edges
9. `useAuthStore` - 13 edges
10. `AuthService` - 12 edges

## Surprising Connections (you probably didn't know these)
- `NewRouter()` --calls--> `NewPlayHandler()`  [INFERRED]
  backend/internal/router/router.go → backend/internal/handler/play_handler.go
- `NewRouter()` --calls--> `NewAuthHandler()`  [INFERRED]
  backend/internal/router/router.go → backend/internal/handler/auth_handler.go
- `NewRouter()` --calls--> `NewHealthHandler()`  [INFERRED]
  backend/internal/router/router.go → backend/internal/handler/health_handler.go
- `NewRouter()` --calls--> `NewAdminHandler()`  [INFERRED]
  backend/internal/router/router.go → backend/internal/handler/admin_handler.go
- `NewRouter()` --calls--> `NewProxyHandler()`  [INFERRED]
  backend/internal/router/router.go → backend/internal/handler/proxy_handler.go

## Communities (71 total, 9 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.03
Nodes (79): importMap, AGENTS.md, backend/cmd/server/main.go, backend/Dockerfile, backend/go.mod, backend/go.sum, backend/internal/config/config.go, backend/internal/handler/admin_handler.go (+71 more)

### Community 1 - "Community 1"
Cohesion: 0.03
Nodes (79): importMap, AGENTS.md, backend/cmd/server/main.go, backend/Dockerfile, backend/go.mod, backend/go.sum, backend/internal/config/config.go, backend/internal/handler/admin_handler.go (+71 more)

### Community 2 - "Community 2"
Cohesion: 0.04
Nodes (54): exportsByPath, backend/cmd/server/main.go, backend/internal/config/config.go, backend/internal/handler/admin_handler.go, backend/internal/handler/audit_handler.go, backend/internal/handler/auth_handler.go, backend/internal/handler/config_handler.go, backend/internal/handler/health_handler.go (+46 more)

### Community 3 - "Community 3"
Cohesion: 0.08
Nodes (20): AdminRoute(), Layout(), ProtectedRoute(), designSystem, getAntdTheme(), getCSSVariables(), ACTION_OPTIONS, Login() (+12 more)

### Community 4 - "Community 4"
Cohesion: 0.06
Nodes (34): code, config, docs, infra, markup, script, conf, css (+26 more)

### Community 5 - "Community 5"
Cohesion: 0.06
Nodes (30): code, config, docs, infra, markup, script, conf, css (+22 more)

### Community 6 - "Community 6"
Cohesion: 0.08
Nodes (25): dependencies, @ant-design/icons, antd, axios, react, react-dom, react-router-dom, zustand (+17 more)

### Community 7 - "Community 7"
Cohesion: 0.12
Nodes (12): NewSSOHandler(), oidcStateCookieName(), parseID(), SSOHandler, OIDCConfig, oidcRuntime, buildOIDCRuntime(), GenerateState() (+4 more)

### Community 8 - "Community 8"
Cohesion: 0.10
Nodes (20): API 接口, code:http (POST /api/auth/login), code:json ({), code:http (GET /api/auth/me), code:http (POST /api/auth/logout), code:http (POST /api/play), code:http (GET /api/play?record_id=GT03225A120DV), code:http (POST /api/play) (+12 more)

### Community 9 - "Community 9"
Cohesion: 0.15
Nodes (15): BM25, detect_domain(), _load_csv(), Lowercase, split, remove punctuation, filter short words, Build BM25 index from documents, Score all documents against query, Load CSV and return list of dicts, Core search function using BM25 (+7 more)

### Community 10 - "Community 10"
Cohesion: 0.11
Nodes (18): ADDED Requirements, Requirement: 审计数据保留 3 个月, Requirement: 审计日志记录, Requirement: 审计查询仅管理员可见与可访问, Requirement: 审计查询接口与筛选, Scenario: 分页查询审计列表, Scenario: 审计 API 仅管理员可调用, Scenario: 录像查询写入审计 (+10 more)

### Community 11 - "Community 11"
Cohesion: 0.20
Nodes (5): hashPassword(), NewAuthService(), toUser(), AuthService, User

### Community 12 - "Community 12"
Cohesion: 0.15
Nodes (8): NewAuthHandler(), AuthHandler, ChangePasswordRequest, Claims, LoginRequest, LoginResponse, UserInfo, VerifyResponse

### Community 13 - "Community 13"
Cohesion: 0.17
Nodes (7): Config, SetConfig(), CORSConfig, DVRConfig, ServerConfig, NewConfigService(), ConfigService

### Community 14 - "Community 14"
Cohesion: 0.15
Nodes (13): 1. 负载均衡, 2. IP 白名单, 3. 基本认证, 4. 请求频率限制, 5. 视频流优化, 6. 缓存配置, code:nginx (upstream dvr_vod_backend {), code:nginx (location / {) (+5 more)

### Community 15 - "Community 15"
Cohesion: 0.18
Nodes (6): NewAdminHandler(), AdminHandler, GetDVRServersResponse, UpdateConfigRequest, UpdateDVRServersRequest, UpdateDVRServersResponse

### Community 16 - "Community 16"
Cohesion: 0.23
Nodes (4): NewSSORepository(), scanSSO(), SSOProvider, SSORepository

### Community 17 - "Community 17"
Cohesion: 0.23
Nodes (5): CreateUserRequest, ResetPasswordRequest, UpdateRoleRequest, NewUserHandler(), UserHandler

### Community 18 - "Community 18"
Cohesion: 0.17
Nodes (11): Architecture Patterns, Code Style, Domain Context, External Dependencies, Git Workflow, Important Constraints, Project Context, Project Conventions (+3 more)

### Community 19 - "Community 19"
Cohesion: 0.29
Nodes (4): NewSSOAdminHandler(), validateProviderConfig(), SSOAdminHandler, SSOProviderRequest

### Community 20 - "Community 20"
Cohesion: 0.18
Nodes (5): GetDB(), NewDVRRepository(), DVRRepository, User, NewUserRepository()

### Community 22 - "Community 22"
Cohesion: 0.24
Nodes (6): AdminMiddleware(), AuthMiddleware(), OptionalAuthMiddleware(), CORSMiddleware(), LoggerMiddleware(), NewRouter()

### Community 23 - "Community 23"
Cohesion: 0.20
Nodes (10): code:bash (# 使用 Nginx 版本的 docker-compose), code:bash (# DVR VOD 日志), code:bash (# Ubuntu/Debian), code:bash (# 复制配置文件), code:nginx (# 修改域名), code:bash (sudo nginx -t), code:bash (# Ubuntu/Debian), 方式一：Docker Compose（推荐） (+2 more)

### Community 24 - "Community 24"
Cohesion: 0.20
Nodes (10): code:bash (./deploy.sh), code:bash (# 1. 登录 GitHub Container Registry), code:bash (# 参考 DEPLOY.md 中的 Nginx 配置), code:bash (go build -o dvr-vod-system ./cmd/server), code:bash (./dvr-vod-system), code:ini ([Unit]), code:bash (sudo systemctl enable dvr-vod), Docker 部署（推荐） (+2 more)

### Community 25 - "Community 25"
Cohesion: 0.20
Nodes (9): code:block25 (dvr-manager/), DVR 点播系统, 功能特点, 工作原理, 并发处理, 技术栈, 生成 Favicon, 许可证 (+1 more)

### Community 26 - "Community 26"
Cohesion: 0.20
Nodes (9): code:bash (# 安装依赖), code:bash (# 构建生产版本), code:bash (# 预览构建结果), code:block4 (frontend/), DVR Manager Frontend, 开发, 构建, 项目结构 (+1 more)

### Community 27 - "Community 27"
Cohesion: 0.20
Nodes (9): Before Any Task, code:bash (# 1) Explore current state), code:block2 (openspec/), Directory Structure, Happy Path Script, OpenSpec Instructions, Search Guidance, TL;DR Quick Checklist (+1 more)

### Community 28 - "Community 28"
Cohesion: 0.25
Nodes (6): BatchPlayResponse, NewPlayHandler(), PlayHandler, PlayRequest, PlayResponse, RecordingResult

### Community 29 - "Community 29"
Cohesion: 0.39
Nodes (7): Close(), createTables(), InitDB(), main(), recordingCacheTTLDays(), runAuditDailyCleanup(), runRecordingCacheDailyCleanup()

### Community 30 - "Community 30"
Cohesion: 0.25
Nodes (7): Context, Decisions, Design: 后台审计查询, Goals / Non-Goals, Migration Plan, Open Questions, Risks / Trade-offs

### Community 31 - "Community 31"
Cohesion: 0.25
Nodes (7): API 接口, code:block1 (backend/), code:bash (# 安装依赖), DVR Manager Backend, 数据目录, 本地开发, 项目结构

### Community 32 - "Community 32"
Cohesion: 0.25
Nodes (8): code:bash (./deploy.sh), code:bash (# 1. 启动服务（前后端分离）), code:bash (docker-compose logs -f), code:bash (docker-compose down), code:bash (# 使用 Nginx 版本部署), code:bash (# 1. 登录 GitHub Container Registry（首次使用需要）), code:yaml (services:), 方式一：Docker 部署（推荐）

### Community 33 - "Community 33"
Cohesion: 0.25
Nodes (8): code:bash (export JWT_SECRET="your-secret-key-here"), CORS 配置, DVR 服务器配置, DVR 查询配置, 服务器配置, 环境变量, 配置说明, 配置项说明

### Community 34 - "Community 34"
Cohesion: 0.25
Nodes (8): code:block3 (New request?), code:markdown (# Change: [Brief description of change]), code:markdown (## ADDED Requirements), code:markdown (## 1. Implementation), code:markdown (## Context), Creating Change Proposals, Decision Tree, Proposal Structure

### Community 35 - "Community 35"
Cohesion: 0.25
Nodes (8): code:markdown (## RENAMED Requirements), code:markdown (#### Scenario: User login success), code:markdown (- **Scenario: User login**  ❌), Critical: Scenario Formatting, Delta Operations, Requirement Wording, Spec File Format, When to use ADDED vs MODIFIED

### Community 36 - "Community 36"
Cohesion: 0.29
Nodes (3): NewAuditRepository(), AuditEntry, AuditRepository

### Community 37 - "Community 37"
Cohesion: 0.29
Nodes (7): 502 Bad Gateway, 504 Gateway Timeout, code:bash (# 检查后端服务是否运行), code:nginx (# 增加超时时间), code:bash (# 检查 Nginx 用户), 故障排查, 权限问题

### Community 38 - "Community 38"
Cohesion: 0.29
Nodes (7): code:bash (# 测试续期), code:bash (# 创建 SSL 目录), code:bash (# Ubuntu/Debian), code:bash (sudo certbot --nginx -d your-domain.com), HTTPS 配置, 使用 Let's Encrypt（免费）, 使用自签名证书（测试）

### Community 39 - "Community 39"
Cohesion: 0.29
Nodes (7): code:bash (# 检查配置), code:bash (# 实时查看访问日志), code:nginx (# 在 nginx.conf 的 http 块中添加), 性能优化, 日志分析, 查看 Nginx 状态, 监控和维护

### Community 40 - "Community 40"
Cohesion: 0.33
Nodes (3): NewDVRService(), DVRQueryResult, DVRService

### Community 41 - "Community 41"
Cohesion: 0.33
Nodes (5): 1. Backend – 存储与写入, 2. Backend – 审计查询 API 与清理, 3. Frontend – 审计查询页与权限, 4. Validation, Tasks: add-admin-audit-query

### Community 43 - "Community 43"
Cohesion: 0.33
Nodes (3): Cache, NewSQLiteCache(), sqliteCache

### Community 44 - "Community 44"
Cohesion: 0.33
Nodes (3): NewAuditHandler(), AuditHandler, ListQuery

### Community 46 - "Community 46"
Cohesion: 0.33
Nodes (6): code:bash (sudo apt-get update && sudo apt-get upgrade nginx), code:nginx (http {), code:nginx (if ($request_method !~ ^(GET|POST|HEAD)$ ) {), code:nginx (limit_conn_zone $binary_remote_addr zone=addr:10m;), code:bash (sudo ufw allow 80/tcp), 安全建议

### Community 47 - "Community 47"
Cohesion: 0.33
Nodes (6): code:bash (curl -X POST http://localhost:8080/api/play \), code:bash (curl -X POST http://localhost:8080/api/play \), code:bash (curl http://localhost:8080/health), code:bash (curl http://localhost:8080/api/config), code:json ({), 测试

### Community 48 - "Community 48"
Cohesion: 0.33
Nodes (5): scriptCompleted, stats, filesScanned, filesWithImports, totalEdges

### Community 49 - "Community 49"
Cohesion: 0.33
Nodes (6): Best Practices, Capability Naming, Change ID Naming, Clear References, Complexity Triggers, Simplicity First

### Community 50 - "Community 50"
Cohesion: 0.33
Nodes (5): algorithm, batches, schemaVersion, totalBatches, totalFiles

### Community 51 - "Community 51"
Cohesion: 0.40
Nodes (4): Change: 增加后台审计查询功能, Impact, What Changes, Why

### Community 52 - "Community 52"
Cohesion: 0.40
Nodes (3): NewConfigHandler(), ConfigHandler, ConfigResponse

### Community 53 - "Community 53"
Cohesion: 0.40
Nodes (5): CLI Essentials, code:bash (openspec list              # What's in progress?), File Purposes, Quick Reference, Stage Indicators

### Community 57 - "Community 57"
Cohesion: 0.50
Nodes (3): Nginx 反向代理部署指南, 为什么使用 Nginx？, 参考资源

### Community 58 - "Community 58"
Cohesion: 0.50
Nodes (4): CI/CD, code:bash (# 1. 创建 GitHub Personal Access Token (PAT)), 使用预构建镜像, 工作流说明

### Community 59 - "Community 59"
Cohesion: 0.50
Nodes (4): code:bash (# 1. 进入后端目录), code:bash (# 1. 进入前端目录), 快速开始, 方式二：本地运行

### Community 60 - "Community 60"
Cohesion: 0.50
Nodes (4): CLI Commands, code:bash (# Essential commands), Command Flags, Quick Start

### Community 61 - "Community 61"
Cohesion: 0.50
Nodes (4): code:bash (# Always use strict mode for comprehensive checks), Common Errors, Troubleshooting, Validation Tips

### Community 62 - "Community 62"
Cohesion: 0.50
Nodes (4): Change Conflicts, Error Recovery, Missing Context, Validation Failures

### Community 63 - "Community 63"
Cohesion: 0.50
Nodes (4): code:block13 (openspec/changes/add-2fa-notify/), code:markdown (## ADDED Requirements), code:markdown (## ADDED Requirements), Multi-Capability Example

### Community 64 - "Community 64"
Cohesion: 0.50
Nodes (4): Stage 1: Creating Changes, Stage 2: Implementing Changes, Stage 3: Archiving Changes, Three-Stage Workflow

## Knowledge Gaps
- **499 isolated node(s):** `name`, `version`, `type`, `dev`, `build` (+494 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewRouter()` connect `Community 22` to `Community 7`, `Community 11`, `Community 12`, `Community 13`, `Community 15`, `Community 16`, `Community 17`, `Community 19`, `Community 20`, `Community 28`, `Community 29`, `Community 36`, `Community 40`, `Community 42`, `Community 43`, `Community 44`, `Community 45`, `Community 52`, `Community 54`, `Community 55`, `Community 56`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `importMap` connect `Community 1` to `Community 4`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **Why does `main()` connect `Community 29` to `Community 36`, `Community 42`, `Community 13`, `Community 45`, `Community 22`?**
  _High betweenness centrality (0.010) - this node is a cross-community bridge._
- **Are the 28 inferred relationships involving `NewRouter()` (e.g. with `main()` and `LoggerMiddleware()`) actually correct?**
  _`NewRouter()` has 28 INFERRED edges - model-reasoned connections that need verification._
- **What connects `name`, `version`, `type` to the rest of the system?**
  _509 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.02531645569620253 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.02531645569620253 - nodes in this community are weakly interconnected._