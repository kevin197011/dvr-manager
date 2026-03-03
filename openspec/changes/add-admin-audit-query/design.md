# Design: 后台审计查询

## Context

- 现有技术栈：后端 Go/Gin + SQLite（`pkg/db`），前端 React + Ant Design；管理端路由已受 `AuthMiddleware` + `AdminMiddleware` 保护。
- 需满足：仅管理员可见与可访问、审计数据保留 3 个月。

## Goals / Non-Goals

- Goals: 可审计关键操作（登录、录像查询、配置变更）；管理员可查询与筛选审计日志；3 个月自动过期。
- Non-Goals: 不实现审计日志导出、不实现实时告警、不改变现有非管理员功能。

## Decisions

- **存储**：在现有 SQLite 中新增 `audit_log` 表（id, created_at, action, username, role, client_ip, resource, detail, status），便于与现有 db 包一致，无需新依赖。
- **写入方式**：在现有 handler 或 middleware 中在关键操作后同步写入审计记录（保持简单；若后续量大可改为异步）。
- **保留策略**：3 个月保留。实现方式二选一：定时任务（如启动时或 cron）删除 `created_at < 3 个月前` 的记录；或查询时过滤并仅返回 3 个月内数据（不删库，仅隐藏）。推荐定时删除以控制库体积。
- **API**：`GET /api/admin/audit`，查询参数：`from`, `to`（时间范围）, `action`（可选）, `username`（可选）, `page`, `page_size`；响应分页列表。沿用现有 admin 组与 AdminMiddleware，不新增权限模型。
- **前端**：在侧栏为 admin 增加「审计查询」菜单项，路由如 `/admin/audit`，页面内表格 + 时间范围与动作类型筛选，调用上述 API。

## Risks / Trade-offs

- 同步写审计可能略增延迟：可接受；若单次写入 < 5ms 不单独优化。
- 定时删除依赖进程存活或外部 cron：文档中说明部署时需配置或内置轻量调度。

## Migration Plan

- 新表通过现有 `createTables` 或迁移逻辑创建；无旧数据迁移。
- 若验证失败可关闭审计写入或删除新路由与菜单，回滚无数据依赖。

## Open Questions

- 无。
