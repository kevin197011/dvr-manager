# 前端 Dockerfile
# 构建阶段
FROM node:20-alpine AS builder

WORKDIR /app

# 复制 package 文件
COPY package.json vite.config.js index.html ./
COPY src/ ./src/
COPY public/ ./public/

# 安装依赖
RUN npm install

# 构建前端
RUN npm run build

# 运行阶段 - 使用 Nginx 提供静态文件服务
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 创建必要的目录
RUN mkdir -p /usr/share/nginx/html

# 暴露端口
EXPOSE 80

# 启动 Nginx
CMD ["nginx", "-g", "daemon off;"]
