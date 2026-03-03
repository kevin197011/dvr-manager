# Tasks: add-admin-audit-query

## 1. Backend – 存储与写入

- [x] 1.1 在 `pkg/db` 中新增 `audit_log` 表及索引（created_at, action），并在 InitDB/createTables 中创建。
- [x] 1.2 新增 `internal/repository/audit_repository.go`：写入单条审计记录、按条件分页查询、删除早于指定时间的记录。
- [x] 1.3 在登录成功/失败、录像查询（play）、管理端配置变更处调用 audit repository 写入审计记录（包含 action, username, client_ip, resource, detail, status 等）。

## 2. Backend – 审计查询 API 与清理

- [x] 2.1 新增 `internal/handler/audit_handler.go`：`GET /api/admin/audit`，解析 from/to/action/username/page/page_size，调用 repository 分页查询，仅返回 3 个月内数据（或由 repository 强制过滤）。
- [x] 2.2 在 `router.go` 的 admin 组下注册 `GET /api/admin/audit`，使用现有 AuthMiddleware + AdminMiddleware。
- [x] 2.3 实现 3 个月保留：在启动或定时逻辑中调用 repository 删除 `created_at < 3 个月前` 的记录；或在文档中说明由外部 cron 调用清理接口（若提供 `POST /api/admin/audit/cleanup`）。

## 3. Frontend – 审计查询页与权限

- [x] 3.1 在 `adminService` 中新增 `getAuditLogs(params)` 调用 `GET /api/admin/audit`。
- [x] 3.2 新增路由 `/admin/audit` 及页面组件（如 `pages/Audit.jsx`），使用 AdminRoute 包裹，仅管理员可访问。
- [x] 3.3 Layout 侧栏中在「系统管理」旁（或下）增加「审计查询」菜单项，仅当 `user?.role === 'admin'` 时显示。
- [x] 3.4 审计页：表格展示时间、动作、用户、IP、资源、详情、状态；支持时间范围与动作类型筛选及分页。

## 4. Validation

- [x] 4.1 以管理员身份登录，确认侧栏有「审计查询」且可打开页面；非管理员无该菜单且直接访问 `/admin/audit` 被拒绝。
- [x] 4.2 执行登录、录像查询、配置变更，确认审计列表中出现对应记录且时间、用户、IP 正确。
- [x] 4.3 确认仅能查询到 3 个月内数据；可选：验证清理逻辑会删除超过 3 个月的旧记录。
