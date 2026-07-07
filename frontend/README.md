# Frontend

React + Vite 前端源码。生产构建产物由根目录 `Makefile` 复制到 `backend/internal/web/dist/` 并嵌入 Go 二进制。

## 开发

```bash
npm install
npm run dev   # http://localhost:3000，/api /stream /health 代理到 :8080
```

## 构建

```bash
npm run build   # 输出 dist/；通常通过根目录 make build 一并处理
```
