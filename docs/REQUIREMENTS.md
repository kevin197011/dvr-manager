# DVR Manager 产品需求文档（PRD）

| 属性 | 值 |
|------|-----|
| 文档版本 | 1.1.0 |
| 更新日期 | 2026-07-07 |
| 项目代号 | dvr-manager |
| 状态 | 已实现（含 v1.1 Dashboard） |

> 本文档基于当前代码库梳理，描述**系统实际应具备的行为**，作为后续功能迭代、缺陷修复、运维交接的需求基线。新增需求请在本文件追加变更记录。

---

## 1. 背景与目标

### 1.1 背景

企业内部部署多台 DVR 录像服务器，录像文件以 `{录像编号}.mp4` 形式存储。业务人员需要按编号查询并播放录像，但：

- DVR 服务器地址分散，不宜直接暴露给终端用户；
- 多台 DVR 可能存有相同编号规则的文件，需要自动发现可用源；
- 需要基本的访问控制与操作审计。

### 1.2 产品目标

1. 提供统一的 Web 入口，按录像编号查询、在线播放、下载录像；
2. 通过后端代理隐藏真实 DVR 地址；
3. 支持多 DVR 服务器并发探测，先命中先返回；
4. 提供用户认证、角色权限、管理后台；
5. 支持 OIDC 单点登录（可选）；
6. 记录关键操作审计日志；
7. **（v1.1）** 为管理员提供使用统计 Dashboard，掌握查询与播放调用量及日趋势。

### 1.3 非目标（当前版本不做）

- 录像文件本地存储/转码/切片（HLS/DASH）；
- 细粒度 RBAC（仅 `admin` / `user` 两角色）；
- JWT 服务端吊销（登出为客户端删除 Token）；
- 多租户隔离；
- SAML / LDAP 等非 OIDC 协议（仅 OIDC）；
- Dashboard v1.1 不做独立埋点库 / Prometheus / 小时级实时大屏（见 §3.10）。

---

## 2. 用户角色与权限

| 角色 | 标识 | 权限范围 |
|------|------|----------|
| 管理员 | `admin` | 录像查询、播放、下载；**使用统计 Dashboard**；系统配置；DVR 服务器管理；用户管理；SSO 配置；审计日志查看 |
| 普通用户 | `user` | 录像查询、播放、下载 |
| 未登录用户 | — | 可访问登录页；**录像 API 为可选认证**（见 §6.4 安全说明） |

### 2.1 用户来源

| 来源 | `source` 字段 | 说明 |
|------|---------------|------|
| 本地账号 | `local` | 用户名密码登录，bcrypt 存储 |
| SSO | `oidc:{provider_id}` | OIDC 登录，首次自动创建，默认角色 `user` |

### 2.2 用户管理约束

- 密码最少 6 位；
- SSO 用户禁止本地密码登录；
- 管理员不能将自己降级为非 admin；
- 管理员不能删除自己；
- 系统至少保留一个 admin 账号；
- 首次启动且用户表为空时，按环境变量种子账号（见 §8.1）。

---

## 3. 功能需求

### 3.1 录像查询（FR-PLAY）

**入口**：首页 `/`（需登录）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-PLAY-01 | 单个录像查询 | 输入一个编号，调用 `POST /api/play`，返回 `proxy_url` |
| FR-PLAY-02 | 批量录像查询 | 多行输入（每行一个编号），调用 `POST /api/play` 带 `record_ids` |
| FR-PLAY-03 | 编号规范化 | 后端自动拼接 `.mp4` 后缀查询 |
| FR-PLAY-04 | 查询结果展示 | 表格显示编号、状态（已找到/未找到）、操作按钮 |
| FR-PLAY-05 | 未找到处理 | 不弹全局错误，在结果行展示「未找到」及 Tooltip 详情 |
| FR-PLAY-06 | GET 查询兼容 | `GET /api/play?record_id=xxx` 同等支持 |

**DVR 探测逻辑**：

1. 从数据库 `dvr_servers` 表读取服务器列表（空则回退配置 JSON）；
2. **并发**向所有服务器发起 HEAD 请求，URL 规则：`{server_url}/{record_id}.mp4`（server_url 无尾斜杠时自动补 `/`）；
3. 单服务器支持重试，次数由 `dvr.retry` 配置（默认 3），指数退避（500ms × 重试次数）；
4. HTTP 200 或 302 视为存在；404 不重试；先成功者优先返回；
5. 单服务器超时由 `dvr.timeout` 控制（默认 10s）；
6. 支持跳过 TLS 证书验证（`dvr.skip_tls_verify`，默认 true）。

### 3.2 视频播放（FR-STREAM）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-STREAM-01 | 代理播放 | 通过 `/stream/{record_id}.mp4` 播放，不暴露真实 DVR URL |
| FR-STREAM-02 | Range 支持 | 转发客户端 `Range` 头，支持拖动进度条 |
| FR-STREAM-03 | 直接访问流 | 无需先调 `/api/play`，缓存未命中时自动查 DVR |
| FR-STREAM-04 | 内联播放 | 首页表格展开行内嵌 `VideoPlayer`，同时仅一个展开 |
| FR-STREAM-05 | 流式传输 | 后端 `io.Copy` 流式转发，不整文件缓冲 |

### 3.3 视频下载（FR-DOWNLOAD）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-DOWNLOAD-01 | 浏览器下载 | 通过 `<a download>` 指向 `proxy_url` 触发下载，避免大文件 blob 占用内存 |
| FR-DOWNLOAD-02 | 进度提示 | 下载中/完成/失败 message 提示 |

### 3.4 录像 URL 缓存（FR-CACHE）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-CACHE-01 | 查询后缓存 | `/api/play` 成功后将 `record_id → real_url` 写入 `recording_cache` |
| FR-CACHE-02 | 代理命中缓存 | `/stream` 优先读缓存，避免重复 HEAD 探测 |
| FR-CACHE-03 | TTL | 默认 30 天，环境变量 `RECORD_CACHE_TTL_DAYS` 可配置 |
| FR-CACHE-04 | 过期清理 | 启动时清理 + 每日 00:00 定时清理过期条目 |

### 3.5 认证（FR-AUTH）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-AUTH-01 | 本地登录 | `POST /api/auth/login`，返回 JWT + 用户信息 |
| FR-AUTH-02 | Token 有效期 | JWT 24 小时（HS256） |
| FR-AUTH-03 | 当前用户 | `GET /api/auth/me` 验证 Token |
| FR-AUTH-04 | 登出 | `POST /api/auth/logout`，客户端清除 Token |
| FR-AUTH-05 | 修改密码 | `POST /api/auth/change-password`，需校验旧密码 |
| FR-AUTH-06 | 401 处理 | 前端拦截器清除 storage 并跳转 `/login` |
| FR-AUTH-07 | 路由守卫 | 业务页需 `ProtectedRoute`；管理页需 `AdminRoute` |

### 3.6 SSO / OIDC（FR-SSO）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-SSO-01 | 提供商列表（公开） | `GET /api/auth/sso/providers` 返回已启用且初始化成功的提供商 |
| FR-SSO-02 | 发起登录 | `GET /api/auth/sso/oidc/:id/login` 跳转 IdP，设置 state Cookie |
| FR-SSO-03 | 回调处理 | `GET /api/auth/sso/oidc/:id/callback` 校验 state、换 token、提取用户名 |
| FR-SSO-04 | 自动建号 | 用户名不存在则创建 `role=user` 的 SSO 用户 |
| FR-SSO-05 | 前端回调 | 重定向 `/sso-callback#token=...`（URL fragment），`SsoCallback` 页写入 auth store |
| FR-SSO-06 | 管理 CRUD | 管理员可增删改查、启用/停用 SSO 提供商 |
| FR-SSO-07 | OIDC 必填字段 | `issuer`, `client_id`, `client_secret`, `redirect_url` |
| FR-SSO-08 | 用户名 Claim | 默认 `preferred_username`，可配置；回退 `email` / `sub` |

**OIDC 配置字段**（`config_json`）：

```json
{
  "issuer": "https://idp.example.com",
  "client_id": "...",
  "client_secret": "...",
  "redirect_url": "https://app.example.com/api/auth/sso/oidc/1/callback",
  "scopes": ["openid", "profile", "email"],
  "username_claim": "preferred_username",
  "skip_tls_verify": false
}
```

### 3.7 管理后台 — 系统配置（FR-ADMIN-CFG）

**入口**：`/admin`（admin 角色）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-ADMIN-CFG-01 | 查看配置 | 加载 DVR 服务器列表 + 完整系统配置 |
| FR-ADMIN-CFG-02 | 服务器配置 | 端口、请求总超时 |
| FR-ADMIN-CFG-03 | DVR 配置 | 单服务器超时、重试次数、跳过 TLS 验证 |
| FR-ADMIN-CFG-04 | CORS 配置 | 开关、origins、methods、headers |
| FR-ADMIN-CFG-05 | DVR 服务器 CRUD | 列表增删改，保存时同步 DB |
| FR-ADMIN-CFG-06 | 保存配置 | `POST /api/admin/config` |
| FR-ADMIN-CFG-07 | 重载配置 | `POST /api/admin/reload` 从 DB 刷新内存 |
| FR-ADMIN-CFG-08 | 非空校验 | DVR 服务器列表不能为空 |

**默认配置值**：

| 配置项 | 默认值 |
|--------|--------|
| `server.port` | 8080 |
| `server.timeout` | 30s |
| `dvr.timeout` | 10s |
| `dvr.retry` | 3 |
| `dvr.skip_tls_verify` | true |
| `cors.enabled` | true |
| `cors.allow_origins` | `*` |
| `cors.allow_methods` | `POST, GET, OPTIONS` |
| `cors.allow_headers` | `Content-Type` |

> **注意**：修改 `server.port` 后需重启进程才能生效（当前实现）。

### 3.8 管理后台 — 用户管理（FR-ADMIN-USER）

**入口**：`/admin/users`

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-ADMIN-USER-01 | 用户列表 | `GET /api/admin/users` |
| FR-ADMIN-USER-02 | 创建用户 | `POST /api/admin/users`，指定 username/password/role |
| FR-ADMIN-USER-03 | 修改角色 | `PUT /api/admin/users/:id/role` |
| FR-ADMIN-USER-04 | 重置密码 | `POST /api/admin/users/:id/reset-password` |
| FR-ADMIN-USER-05 | 删除用户 | `DELETE /api/admin/users/:id`，受 §2.2 约束 |

### 3.9 管理后台 — 审计日志（FR-ADMIN-AUDIT）

**入口**：`/admin/audit`

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-ADMIN-AUDIT-01 | 分页查询 | `GET /api/admin/audit`，默认 page=1, page_size=20，最大 100 |
| FR-ADMIN-AUDIT-02 | 筛选 | 支持 `from`/`to`（RFC3339）、`action`、`username` |
| FR-ADMIN-AUDIT-03 | 保留策略 | 默认保留 **3 个月**，超出部分硬删除（非软删） |
| FR-ADMIN-AUDIT-04 | 手动清理 | `POST /api/admin/audit/cleanup` 立即删除保留期外记录 |
| FR-ADMIN-AUDIT-05 | **自动清理（必须）** | 见下方「审计日志生命周期」 |

**审计动作类型（action）**：

| action | 触发场景 |
|--------|----------|
| `login_success` | 本地/SSO 登录成功 |
| `login_fail` | 登录失败 |
| `logout` | （预留，当前登出未写审计） |
| `password_change` | 用户修改密码 |
| `play` | 单个录像查询（`/api/play` 单条） |
| `play_batch` | 批量录像查询 |
| `stream` | 流代理访问（`/stream`，v1.1 起独立 action；历史数据可能仍为 `play`+`流代理:` 前缀） |
| `config_save` | 保存配置或 DVR 列表 |
| `config_reload` | 重载配置 |
| `user_create` | 创建用户 |
| `user_update_role` | 修改角色 |
| `user_reset_password` | 重置密码 |
| `user_delete` | 删除用户 |
| `sso_create` / `sso_update` / `sso_toggle` / `sso_delete` | SSO 提供商管理 |

**审计记录字段**：`id`, `created_at`, `action`, `username`, `role`, `client_ip`, `resource`, `detail`, `status`（`success` / `fail`）

**审计日志生命周期**（防止 `audit_log` 无限增长影响 SQLite 查询性能）：

| 时机 | 行为 |
|------|------|
| 进程启动 | 立即删除 `created_at` 早于保留截止日的记录 |
| 每日 00:00（进程本地时区） | 后台 goroutine 自动执行相同清理，并写 `[Audit] daily cleanup` 日志 |
| 管理端查询 | `List` 仅返回保留期内的记录 |
| 手动触发 | `POST /api/admin/audit/cleanup`（与自动清理规则一致） |

保留月数默认 **3**，环境变量 `AUDIT_RETENTION_MONTHS` 可在启动时配置（修改后需重启进程）。

验收：启动日志含 `startup + daily 00:00 cleanup enabled`；存在超期数据时启动或次日凌晨后 `audit_log` 行数下降。

验收：启动日志含 `startup + daily 00:00 cleanup enabled`；存在超期数据时启动或次日凌晨后 `audit_log` 行数下降。

### 3.10 管理后台 — 使用统计 Dashboard（FR-ADMIN-DASH）

> **版本**：v1.1 已实现  
> **设计原则**：复用现有 `audit_log`，不新增统计表；与审计保留期（默认 3 个月）一致；仅 **admin** 可访问。

**入口**：`/admin/dashboard`（建议作为管理后台菜单第一项）

#### 3.10.1 目标

让管理员快速回答：

1. 今天 / 本周 / 本月系统被**查询**了多少次？**播放（流代理）**了多少次？
2. 每日调用量趋势如何（时间序列）？
3. 查询成功率、活跃用户数大致水平？

#### 3.10.2 指标口径（基于 audit_log）

| 指标键 | 含义 | SQL / 规则口径 |
|--------|------|----------------|
| `query_single` | 单次录像查询（`POST/GET /api/play` 单条） | `action = 'play'` 且 `detail` 为 `录像已找到` 或 `录像未找到` |
| `query_batch` | 批量查询 API 调用次数 | `action = 'play_batch'`（每条审计 = 1 次批量请求） |
| `query_batch_records` | 批量查询涉及的录像条数（可选展示） | 解析 `play_batch` 的 `detail`：`批量查询 N 条，找到 M 条` 中的 **N** 累加 |
| `stream` | 流代理访问（播放 / 下载走 `/stream`） | `action = 'stream'`，或历史 `action = 'play'` 且 `detail` 以 `流代理:` 开头 |
| `query_success` / `query_fail` | 单次查询成功 / 失败 | 同上 `query_single`，按 `status` 区分 |
| `stream_success` / `stream_fail` | 流代理成功 / 失败 | 同上 `stream`，按 `status` 区分 |
| `active_users` | 活跃用户数 | 时间范围内 `username` 非空的去重计数（统计查询 + 流代理相关 action） |
| `login_success` | 登录次数（辅助） | `action = 'login_success'`（可选卡片，默认折叠或次要展示） |

**说明**：

- 流代理自 v1.1 起写入 `action=stream`；统计 SQL 同时兼容历史 `play`+`流代理:` 记录。
- 批量查询在 API 层记 **1 条** `play_batch`，不在每条录像上重复记 `play`。
- 统计时间按进程时区（Docker 默认 `Asia/Shanghai`）取 `DATE(created_at)` 聚合。
- 查询范围不得超过 `AUDIT_RETENTION_MONTHS` 保留窗口；超出部分无数据属预期。

#### 3.10.3 功能需求

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-ADMIN-DASH-01 | 汇总卡片 | 展示今日 / 近 7 日 / 近 30 日（或自定义区间）的查询次数、流访问次数、查询成功率 |
| FR-ADMIN-DASH-02 | 日时间序列图 | 折线或柱状图：X 轴为日期，Y 轴为次数；至少支持切换「查询」「流访问」两条序列 |
| FR-ADMIN-DASH-03 | 时间范围筛选 | 日期选择器：`from` / `to`，默认最近 30 天；最长不超过审计保留期 |
| FR-ADMIN-DASH-04 | 数据接口 | `GET /api/admin/dashboard/stats`，仅 admin |
| FR-ADMIN-DASH-05 | 空数据态 | 无审计数据时展示 0 与友好提示，不报错 |
| FR-ADMIN-DASH-06 | 与审计一致 | 统计数据仅来自保留期内 `audit_log`，与 §3.9 清理策略一致 |
| FR-ADMIN-DASH-07 | 导航入口 | 管理侧边栏增加「使用统计」，路由 `/admin/dashboard` |
| FR-ADMIN-DASH-08 | 懒加载 | 前端页面与其他管理页一致 `React.lazy`，图表库按需加载 |

#### 3.10.4 API 设计

**`GET /api/admin/dashboard/stats`**

| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `from` | date `YYYY-MM-DD` | 今天 - 29 天 | 起始日（含） |
| `to` | date `YYYY-MM-DD` | 今天 | 结束日（含） |
| `granularity` | `day` | `day` | v1.1 仅支持按日；预留 `week` |

**响应示例**：

```json
{
  "success": true,
  "range": { "from": "2026-06-08", "to": "2026-07-07", "granularity": "day" },
  "summary": {
    "query_single": 120,
    "query_batch": 8,
    "query_batch_records": 240,
    "stream": 95,
    "query_success_rate": 0.92,
    "stream_success_rate": 0.98,
    "active_users": 12,
    "login_success": 45
  },
  "series": [
    {
      "date": "2026-07-01",
      "query_single": 10,
      "query_batch": 1,
      "stream": 8,
      "query_success": 9,
      "query_fail": 1,
      "stream_success": 8,
      "stream_fail": 0
    }
  ]
}
```

**校验**：

- `from` > `to` → 400；
- 区间超过保留期 → 自动截断到 `RetentionCutoff()` 并可在响应中加 `truncated: true`（可选）；
- 单次聚合区间建议上限 90 天（与默认保留 3 个月匹配）。

#### 3.10.5 页面布局（线框）

```
┌─────────────────────────────────────────────────────────────┐
│ 使用统计          [从 2026-06-08] [至 2026-07-07] [查询]     │
├──────────┬──────────┬──────────┬──────────┬──────────────────┤
│ 单次查询  │ 批量查询  │ 流访问    │ 查询成功率│ 活跃用户(区间)   │
│   120    │    8     │   95     │  92%     │      12          │
├─────────────────────────────────────────────────────────────┤
│ 每日调用趋势（折线图）                                        │
│  — 单次查询  — 批量查询  — 流访问                             │
│     ╱╲    ╱╲                                               │
│    ╱  ╲__╱  ╲___                                           │
├─────────────────────────────────────────────────────────────┤
│ 提示：数据来自审计日志，保留 3 个月；明细见「审计查询」        │
└─────────────────────────────────────────────────────────────┘
```

- 图表：推荐 `@ant-design/plots`（与 Ant Design 5 一致）；不引入重量级 BI 组件。
- 汇总卡片与图表共用同一 `from`/`to` 请求，避免重复拉取。

#### 3.10.6 实现要点（供开发）

| 层 | 建议 |
|----|------|
| Repository | `AuditRepository.Stats(from, to)`：`GROUP BY date(created_at)` + 条件聚合；复用 `idx_audit_log_created_at` |
| Handler | `DashboardHandler.GetStats`，注册 `admin.GET("/dashboard/stats", ...)` |
| 前端 | `pages/Dashboard.jsx`；`services/dashboardService.js` |
| 测试 | 插入固定 `audit_log` 夹具，断言聚合结果与口径表一致 |

#### 3.10.7 后续可选（非 v1.1）

- 按用户 Top N、按录像编号 Top N；
- 独立 `stream` action 后简化 SQL；
- 导出 CSV；
- 小时级粒度（数据量大时再评估）。

### 3.11 管理后台 — SSO 配置（FR-ADMIN-SSO）

**入口**：`/admin/sso`

管理 OIDC 提供商的完整 CRUD + 启用/停用切换。

### 3.12 公开接口（FR-PUBLIC）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-PUBLIC-01 | 健康检查 | `GET/HEAD /health` 返回 200 |
| FR-PUBLIC-02 | 公开配置摘要 | `GET /api/config` 返回端口、DVR 数量、重试信息、版本号（无敏感信息） |

### 3.13 前端通用（FR-UI）

| 编号 | 需求 | 验收标准 |
|------|------|----------|
| FR-UI-01 | 主题切换 | 支持明/暗主题（Zustand `themeStore`） |
| FR-UI-02 | 侧边导航 | 普通用户见「录像查询」；管理员额外见管理菜单（含使用统计 Dashboard） |
| FR-UI-03 | 用户菜单 | 修改密码、登出 |
| FR-UI-04 | SPA 路由 | Go `embed` 静态资源；未匹配 API 路径时回退 `index.html` |

---

## 4. 非功能需求

### 4.1 性能

| 编号 | 需求 | 指标 |
|------|------|------|
| NFR-PERF-01 | 并发查询隔离 | 每个 HTTP 请求独立 context，超时互不影响 |
| NFR-PERF-02 | DVR 并发探测 | 多服务器 goroutine 并发 HEAD |
| NFR-PERF-03 | 连接复用 | HTTP Transport `MaxIdleConns=100` |
| NFR-PERF-04 | 视频代理 | Go `io.Copy` 流式转发，不整文件缓冲 |
| NFR-PERF-05 | 审计日志容量 | 默认仅保留 3 个月；启动 + 每日自动硬删除，避免 `audit_log` 堆积拖慢查询 |
| NFR-PERF-06 | Dashboard 聚合 | 单次 `stats` 请求在 3 个月审计数据量下 P95 < 500ms；使用 `created_at` 索引按日 `GROUP BY` |

### 4.2 可用性

| 编号 | 需求 |
|------|------|
| NFR-AVAIL-01 | Docker 健康检查：`wget /health`，30s 间隔 |
| NFR-AVAIL-02 | 单容器部署，无前后端启动顺序依赖 |
| NFR-AVAIL-03 | 容器 `restart: unless-stopped` |

### 4.3 可维护性

| 编号 | 需求 |
|------|------|
| NFR-MAINT-01 | 分层架构：Handler → Service → Repository |
| NFR-MAINT-02 | 配置存 SQLite，管理后台可热更新（端口除外） |
| NFR-MAINT-03 | 结构化日志（标准 log，含 IP/编号/耗时） |
| NFR-MAINT-04 | GitHub Actions 自动构建推送 GHCR 单镜像（前端已嵌入） |

### 4.4 兼容性

| 编号 | 需求 |
|------|------|
| NFR-COMPAT-01 | 浏览器：现代浏览器（Chrome/Firefox/Safari/Edge） |
| NFR-COMPAT-02 | DVR 协议：HTTP/HTTPS HEAD + GET |
| NFR-COMPAT-03 | 视频格式：MP4（依赖 DVR 源站 Content-Type） |

---

## 5. 系统架构

### 5.1 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.25, Gin, JWT, go-oidc, modernc.org/sqlite |
| 前端 | React 18, Ant Design 5, Vite 5, Zustand, Axios |
| 部署 | Docker Compose 单服务；前端 `dist` 经 `go:embed` 打入二进制 |
| CI | GitHub Actions → GHCR（`ghcr.io/<repo>`） |

### 5.2 部署拓扑

生产环境为**单进程、单容器**：Go 服务同时提供 API 与 SPA 静态资源。

```
用户浏览器
    │
    ▼
┌─────────────────────────────────────┐     ┌─────────────────┐
│  app:8080 (Go)                      │────▶│  DVR 服务器 ×N  │
│  /api/* /stream/* /health  → API    │     │  (HEAD/GET)     │
│  /*                        → SPA    │     └─────────────────┘
└────────────────┬────────────────────┘
                 │
                 ▼
        ┌─────────────────┐
        │  SQLite         │
        │  ./data/        │
        └─────────────────┘
```

**本地开发**仍前后端分离：Vite dev server `:3000` 代理 `/api`、`/stream`、`/health` 到 `:8080`。

**构建链路**（生产镜像 / 本地单二进制）：

```
frontend/npm run build → dist/
        ↓
backend/internal/web/dist/  (//go:embed)
        ↓
go build → 单一可执行文件
```

### 5.3 目录结构

```
dvr-manager/
├── backend/
│   ├── cmd/server/main.go       # 入口、定时任务
│   ├── internal/
│   │   ├── config/              # 配置结构体
│   │   ├── handler/             # HTTP 处理器
│   │   ├── middleware/          # 认证、CORS、日志
│   │   ├── repository/          # 数据访问
│   │   ├── router/              # 路由注册
│   │   ├── service/             # 业务逻辑
│   │   └── web/                 # 嵌入的前端 dist（构建时生成）
│   └── pkg/
│       ├── cache/               # 录像 URL 缓存
│       └── db/                  # SQLite 初始化
├── frontend/                    # 前端源码（开发 / 构建）
│   └── src/
│       ├── pages/               # 页面
│       ├── components/          # 组件
│       ├── services/            # API 封装
│       └── store/               # 状态
├── data/                        # 运行时数据（gitignore）
├── docs/                        # 文档
├── Dockerfile                   # 多阶段：Node 构建前端 + Go embed
├── Makefile                     # make build 本地单二进制
├── docker-compose.yml           # 单服务 app:8080
└── deploy.sh
```

### 5.4 请求链路（录像播放）

```
1. 用户输入编号 → POST /api/play
2. DVRService.FindRecording() 并发 HEAD 探测
3. 命中 → cache.Set(record_id, real_url)
4. 返回 proxy_url: /stream/{id}.mp4
5. 用户播放 → GET /stream/{id}.mp4
6. cache.Get() 或重新 FindRecording()
7. ProxyService.ProxyStream() 转发 Range/Body
```

---

## 6. 数据模型

### 6.1 ER 关系

```
config (KV)
dvr_servers (URL 列表)
users (账号)
sso_providers (OIDC 配置)
audit_log (操作日志)
recording_cache (录像 URL 缓存)
```

### 6.2 表结构

#### config

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| key | TEXT UNIQUE | 固定 `main` |
| value | TEXT | JSON 序列化的 Config 结构 |
| updated_at | DATETIME | |

#### dvr_servers

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| server | TEXT UNIQUE | DVR 基础 URL，如 `http://dvr1:8080/record` |
| created_at | DATETIME | |

#### users

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| username | TEXT UNIQUE | |
| password_hash | TEXT | bcrypt；SSO 用户为占位哈希 |
| role | TEXT | `admin` / `user` |
| source | TEXT | `local` / `oidc:{id}` |
| created_at / updated_at | DATETIME | |

#### sso_providers

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| type | TEXT | 仅 `oidc` |
| name | TEXT | 显示名称 |
| enabled | INTEGER | 0/1 |
| config_json | TEXT | OIDC 配置 JSON |
| created_at / updated_at | DATETIME | |

#### audit_log

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| created_at | DATETIME | |
| action | TEXT | 见 §3.9 |
| username / role | TEXT | 可空 |
| client_ip | TEXT | |
| resource | TEXT | 如录像编号、配置项名 |
| detail | TEXT | 人类可读描述 |
| status | TEXT | `success` / `fail` |

#### recording_cache

| 字段 | 类型 | 说明 |
|------|------|------|
| record_id | TEXT PK | 录像编号（无 .mp4） |
| real_url | TEXT | DVR 真实 URL |
| created_at | DATETIME | |
| expires_at | DATETIME | TTL 到期时间 |

---

## 7. API 规格摘要

### 7.1 认证头

```
Authorization: Bearer <jwt_token>
```

### 7.2 接口清单

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/auth/login` | 无 | 登录 |
| GET | `/api/auth/me` | 可选 | 当前用户 |
| POST | `/api/auth/logout` | 无 | 登出 |
| POST | `/api/auth/change-password` | 必须 | 改密 |
| GET | `/api/auth/sso/providers` | 无 | SSO 列表 |
| GET | `/api/auth/sso/oidc/:id/login` | 无 | 跳转 IdP |
| GET | `/api/auth/sso/oidc/:id/callback` | 无 | OIDC 回调 |
| POST/GET | `/api/play` | 可选 | 录像查询 |
| GET | `/api/config` | 可选 | 公开配置 |
| GET | `/stream/:filename` | 可选 | 视频代理 |
| GET/HEAD | `/health` | 无 | 健康检查 |
| GET | `/api/admin/config` | admin | 完整配置 |
| POST | `/api/admin/config` | admin | 更新配置 |
| GET | `/api/admin/dvr-servers` | admin | DVR 列表 |
| POST | `/api/admin/dvr-servers` | admin | 更新 DVR 列表 |
| POST | `/api/admin/reload` | admin | 重载配置 |
| GET | `/api/admin/audit` | admin | 审计日志 |
| GET | `/api/admin/dashboard/stats` | admin | 使用统计（v1.1） |
| POST | `/api/admin/audit/cleanup` | admin | 清理审计 |
| GET/POST/PUT/DELETE | `/api/admin/users/...` | admin | 用户管理 |
| GET/POST/PUT/DELETE | `/api/admin/sso/providers/...` | admin | SSO 管理 |

### 7.3 关键响应示例

**登录成功**：
```json
{
  "success": true,
  "token": "eyJ...",
  "user": { "username": "admin", "role": "admin" }
}
```

**单个查询成功**：
```json
{
  "success": true,
  "proxy_url": "/stream/GT03225A120DV.mp4",
  "message": "recording found"
}
```

**批量查询**：
```json
{
  "success": true,
  "results": [
    { "record_id": "GT03225A120DV", "found": true, "proxy_url": "/stream/GT03225A120DV.mp4" },
    { "record_id": "GT03225A120DW", "found": false }
  ],
  "message": "batch query completed"
}
```

**Dashboard 统计（v1.1）**：
```json
{
  "success": true,
  "range": { "from": "2026-06-08", "to": "2026-07-07", "granularity": "day" },
  "summary": {
    "query_single": 120,
    "query_batch": 8,
    "stream": 95,
    "query_success_rate": 0.92,
    "active_users": 12
  },
  "series": [
    { "date": "2026-07-07", "query_single": 10, "query_batch": 1, "stream": 8 }
  ]
}
```

---

## 8. 部署与运维

### 8.1 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DATA_DIR` | `../data` 或 `/app/data` | SQLite 目录 |
| `JWT_SECRET` | 固定字符串 | **生产必改** |
| `ADMIN_USERNAME` | `admin` | 种子管理员 |
| `ADMIN_PASSWORD` | `admin123` | 种子管理员密码 |
| `USER_USERNAME` | `user` | 种子普通用户 |
| `USER_PASSWORD` | `user123` | 种子普通用户密码 |
| `RECORD_CACHE_TTL_DAYS` | `30` | 录像缓存天数 |
| `AUDIT_RETENTION_MONTHS` | `3` | 审计日志保留月数；启动时 + 每日 00:00 自动清理超期记录 |
| `REQUIRE_AUTH_FOR_PLAY` | `false` | 设为 `true` 时 `/api/play` 与 `/stream` 强制登录 |
| `TZ` | — | 时区（Docker 默认 Asia/Shanghai）；影响每日清理触发时刻 |
| `VITE_API_BASE_URL` | `/api` | 前端 API 基址（构建时） |

### 8.2 部署命令

```bash
# 推荐（Docker 单容器）
./deploy.sh

# 或
docker compose up -d --build

# 本地开发（前后端分离，热更新）
cd backend && go run ./cmd/server          # :8080，仅 API
cd frontend && npm install && npm run dev  # :3000，代理到 :8080

# 本地单二进制（与生产一致）
make build && ./dvr-manager             # :8080，API + SPA
```

### 8.3 端口

| 场景 | 端口 |
|------|------|
| 生产 / Docker（`app` 服务） | 8080（由外层代理转发） |
| 本地开发（Vite） | 3000（API 代理到 8080） |

### 8.4 静态资源嵌入说明

| 项 | 说明 |
|----|------|
| 嵌入路径 | `backend/internal/web/dist/` |
| 实现 | `internal/web/static.go` 使用 `//go:embed dist/*` |
| 路由 | API/stream/health 优先注册；`NoRoute` 提供静态文件与 SPA fallback |
| 占位页 | 未执行 `make build` 时 `go run` 仅显示占位 `index.html` |

### 8.5 备份

定期备份 `data/dvr-manager.db`（含用户、配置、审计、缓存）。

### 8.6 日志

- 容器日志：`docker compose logs -f`
- 日志轮转：json-file driver，max-size 10m × 3 files

---

## 9. 安全需求与风险

### 9.1 安全要求

| 编号 | 要求 |
|------|------|
| SEC-01 | 生产环境必须修改 `JWT_SECRET` 及默认账号密码 |
| SEC-02 | 密码 bcrypt 存储，不明文 |
| SEC-03 | OIDC state Cookie 防 CSRF，HttpOnly |
| SEC-04 | 管理接口强制 admin 角色 |
| SEC-05 | SSO client_secret 仅存数据库，前端展示需脱敏 |

### 9.2 已知风险 / 待改进

| 风险 | 说明 | 建议 |
|------|------|------|
| 录像 API 可选认证 | `/api/play`、`/stream` 默认未强制登录 | 生产可设 `REQUIRE_AUTH_FOR_PLAY=true` |
| JWT 无吊销 | 登出不清服务端 Token | 敏感环境可加快过期或加黑名单 |
| 默认 TLS 跳过验证 | `skip_tls_verify=true` | 内网可接受；公网 DVR 应关闭 |
| CORS 默认 `*` | 允许任意来源 | 生产限制 `allow_origins` |

---

## 10. 测试验收清单

### 10.1 冒烟测试

```bash
# 健康检查
curl http://localhost:8080/health

# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 单个查询
curl -X POST http://localhost:8080/api/play \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"record_id":"GT03225A120DV"}'

# 公开配置
curl http://localhost:8080/api/config
```

### 10.2 功能回归要点

- [ ] 单条/批量录像查询
- [ ] 播放（含拖动）
- [ ] 下载
- [ ] 本地登录 / 登出 / 改密
- [ ] SSO 登录全流程（如已配置）
- [ ] 管理后台保存配置并重载
- [ ] DVR 服务器增删
- [ ] 用户 CRUD 及权限约束
- [ ] 审计日志写入与查询
- [ ] Dashboard 汇总与按日时间序列与审计口径一致（v1.1）
- [ ] Docker 重启后数据持久化

---

## 11. 术语表

| 术语 | 定义 |
|------|------|
| 录像编号 | 业务标识，如 `GT03225A120DV`，查询时自动加 `.mp4` |
| DVR 服务器 | 存储录像的 HTTP 源站，配置为基础 URL |
| 代理 URL | `/stream/{编号}.mp4`，对外暴露的播放地址 |
| 真实 URL | DVR 上的完整文件地址，仅后端知晓 |
| 可选认证 | 有 Token 则解析用户信息写审计；无 Token 仍放行 |
| Dashboard | 管理后台使用统计页，基于 `audit_log` 聚合展示调用量与日趋势 |

---

## 12. 变更记录

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|----------|
| 1.1.0 | 2026-07-07 | — | 规划管理后台 Dashboard：基于 audit_log 的使用统计与日时间序列（§3.10） |
| 1.0.3 | 2026-07-07 | — | 前端嵌入 Go 二进制：单容器部署；清理废弃的前端/后端 Docker 与 Nginx 配置 |
| 1.0.2 | 2026-07-07 | — | 代码质量优化：配置热更新、JWT 抽离、批量并发、SSO fragment、可选强制播放鉴权 |
| 1.0.1 | 2026-07-07 | — | 明确审计日志 3 个月保留 + 启动/每日自动清理；`AUDIT_RETENTION_MONTHS` |
| 1.0.0 | 2026-07-07 | — | 基于代码库初始梳理，建立 As-Is 需求基线 |

---

## 附录 A：前端路由表

| 路径 | 组件 | 权限 |
|------|------|------|
| `/login` | Login | 公开 |
| `/sso-callback` | SsoCallback | 公开 |
| `/` | Home | 登录 |
| `/admin` | Admin | admin |
| `/admin/dashboard` | Dashboard | admin |
| `/admin/users` | Users | admin |
| `/admin/audit` | Audit | admin |
| `/admin/sso` | SsoConfig | admin |

## 附录 B：配置热更新 vs 重启

| 配置项 | 热更新 | 说明 |
|--------|--------|------|
| DVR 服务器列表 | ✅ | `config.SetConfig` 即时生效 |
| dvr.timeout / retry / skip_tls_verify | ✅ | 每次查询读全局配置 |
| cors.* | ✅ | 中间件读配置 |
| server.port | ❌ | 需重启进程 |
| JWT_SECRET | ❌ | 需重启（环境变量） |
| RECORD_CACHE_TTL_DAYS | ❌ | 仅启动时读取 |
| AUDIT_RETENTION_MONTHS | ❌ | 仅启动时读取；每日清理使用启动时配置 |
