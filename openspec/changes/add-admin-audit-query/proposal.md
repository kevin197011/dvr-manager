# Change: 增加后台审计查询功能

## Why

需要为管理员提供可追溯的操作与访问记录，便于安全排查与合规；同时控制存储成本，仅保留 3 个月历史数据。

## What Changes

- 新增审计日志能力：记录登录（成功/失败）、录像查询请求、管理端配置变更等可审计事件。
- 新增管理员专属的「审计查询」入口与页面，仅管理员可见与可访问。
- 提供审计日志的查询 API（分页、按时间范围与动作类型筛选），受现有 Admin 鉴权保护。
- 历史数据保留 3 个月：超过 3 个月的审计记录自动删除或归档后不可查。

## Impact

- Affected specs: 新增 `admin-audit` capability。
- Affected code:
  - Backend: 新增 `audit` 表与 repository、审计写入点（auth/play/admin handlers 或 middleware）、admin 路由下新增 `GET /api/admin/audit`（及可选的清理/定时任务）。
  - Frontend: 新增审计查询页与路由、Layout 中仅 admin 可见的菜单项、调用审计 API 的服务与组件。
