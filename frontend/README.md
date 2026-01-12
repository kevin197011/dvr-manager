# DVR Manager Frontend

前端应用，使用 Vite 构建。

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

开发服务器运行在 `http://localhost:3000`，API 请求会自动代理到后端 `http://localhost:8080`。

## 构建

```bash
# 构建生产版本
npm run build
```

构建产物输出到 `dist/` 目录。

## 预览

```bash
# 预览构建结果
npm run preview
```

## 项目结构

```
frontend/
├── src/           # 源代码
│   ├── index.html      # 主页面
│   └── admin.html      # 管理后台
├── public/        # 静态资源
│   └── favicon.*  # 图标文件
├── dist/          # 构建产物（不提交到 Git）
├── package.json   # 依赖配置
├── vite.config.js # Vite 配置
└── Dockerfile     # Docker 构建文件
```
