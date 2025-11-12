# Nginx 反向代理部署指南

本文档介绍如何使用 Nginx 作为 DVR 点播系统的反向代理。

## 为什么使用 Nginx？

- ✅ **性能优化**：静态文件缓存、连接复用
- ✅ **负载均衡**：支持多实例部署
- ✅ **HTTPS 支持**：SSL/TLS 加密传输
- ✅ **访问控制**：IP 白名单、认证
- ✅ **日志管理**：统一的访问日志
- ✅ **安全防护**：防止直接访问后端服务

## 部署方式

### 方式一：Docker Compose（推荐）

**1. 使用 Nginx 配置启动**

```bash
# 使用 Nginx 版本的 docker-compose
docker-compose -f docker-compose.nginx.yml up -d
```

**2. 访问服务**

- HTTP: http://localhost
- HTTPS: https://localhost (需配置 SSL 证书)

**3. 查看日志**

```bash
# DVR VOD 日志
docker-compose -f docker-compose.nginx.yml logs -f dvr-vod

# Nginx 日志
docker-compose -f docker-compose.nginx.yml logs -f nginx

# 或查看日志文件
tail -f logs/nginx/access.log
tail -f logs/nginx/error.log
```

### 方式二：系统 Nginx

**1. 安装 Nginx**

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install nginx

# CentOS/RHEL
sudo yum install nginx

# macOS
brew install nginx
```

**2. 复制配置文件**

```bash
# 复制配置文件
sudo cp nginx.conf /etc/nginx/sites-available/dvr-vod

# 创建软链接
sudo ln -s /etc/nginx/sites-available/dvr-vod /etc/nginx/sites-enabled/

# 或直接复制到 conf.d
sudo cp nginx.conf /etc/nginx/conf.d/dvr-vod.conf
```

**3. 修改配置**

编辑配置文件，修改以下内容：

```nginx
# 修改域名
server_name your-domain.com;

# 修改后端地址（如果不是本地）
upstream dvr_vod_backend {
    server 127.0.0.1:8080;
}
```

**4. 测试配置**

```bash
sudo nginx -t
```

**5. 重启 Nginx**

```bash
# Ubuntu/Debian
sudo systemctl restart nginx

# CentOS/RHEL
sudo systemctl restart nginx

# macOS
brew services restart nginx
```

## HTTPS 配置

### 使用 Let's Encrypt（免费）

**1. 安装 Certbot**

```bash
# Ubuntu/Debian
sudo apt-get install certbot python3-certbot-nginx

# CentOS/RHEL
sudo yum install certbot python3-certbot-nginx

# macOS
brew install certbot
```

**2. 获取证书**

```bash
sudo certbot --nginx -d your-domain.com
```

**3. 自动续期**

```bash
# 测试续期
sudo certbot renew --dry-run

# 添加定时任务
sudo crontab -e
# 添加以下行（每天凌晨 2 点检查续期）
0 2 * * * certbot renew --quiet
```

### 使用自签名证书（测试）

```bash
# 创建 SSL 目录
mkdir -p ssl

# 生成自签名证书
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/key.pem \
  -out ssl/cert.pem \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=DVR/CN=dvr.example.com"

# 修改 nginx.conf 中的证书路径
ssl_certificate /path/to/ssl/cert.pem;
ssl_certificate_key /path/to/ssl/key.pem;
```

## 高级配置

### 1. 负载均衡

编辑 `nginx.conf`，添加多个后端服务器：

```nginx
upstream dvr_vod_backend {
    # 轮询（默认）
    server 127.0.0.1:8080;
    server 127.0.0.1:8081;
    server 127.0.0.1:8082;
    
    # 或使用 IP Hash（同一客户端总是访问同一服务器）
    # ip_hash;
    
    # 或使用最少连接
    # least_conn;
    
    keepalive 32;
}
```

### 2. IP 白名单

```nginx
location / {
    # 只允许特定 IP 访问
    allow 192.168.1.0/24;
    allow 10.0.0.0/8;
    deny all;
    
    proxy_pass http://dvr_vod_backend;
}
```

### 3. 基本认证

```bash
# 创建密码文件
sudo htpasswd -c /etc/nginx/.htpasswd admin

# 在 nginx.conf 中添加
location / {
    auth_basic "DVR VOD System";
    auth_basic_user_file /etc/nginx/.htpasswd;
    proxy_pass http://dvr_vod_backend;
}
```

### 4. 请求频率限制

```nginx
# 在 http 块中定义限制区域
http {
    limit_req_zone $binary_remote_addr zone=dvr_limit:10m rate=10r/s;
    
    server {
        location /api/ {
            limit_req zone=dvr_limit burst=20 nodelay;
            proxy_pass http://dvr_vod_backend;
        }
    }
}
```

### 5. 视频流优化

```nginx
location /stream/ {
    # 禁用缓冲，实现真正的流式传输
    proxy_buffering off;
    proxy_cache off;
    
    # 支持断点续传
    proxy_set_header Range $http_range;
    proxy_set_header If-Range $http_if_range;
    
    # 限速（可选）
    limit_rate 5m;  # 限制每个连接 5MB/s
    
    # 超时设置
    proxy_read_timeout 300s;
    
    proxy_pass http://dvr_vod_backend;
}
```

### 6. 缓存配置

```nginx
# 在 http 块中定义缓存路径
http {
    proxy_cache_path /var/cache/nginx/dvr 
                     levels=1:2 
                     keys_zone=dvr_cache:10m 
                     max_size=1g 
                     inactive=60m;
    
    server {
        location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
            proxy_cache dvr_cache;
            proxy_cache_valid 200 1h;
            proxy_cache_use_stale error timeout http_500 http_502 http_503 http_504;
            
            proxy_pass http://dvr_vod_backend;
        }
    }
}
```

## 监控和维护

### 查看 Nginx 状态

```bash
# 检查配置
sudo nginx -t

# 查看进程
ps aux | grep nginx

# 查看端口
sudo netstat -tlnp | grep nginx
```

### 日志分析

```bash
# 实时查看访问日志
tail -f /var/log/nginx/dvr-vod-access.log

# 统计访问量
cat /var/log/nginx/dvr-vod-access.log | wc -l

# 统计 IP 访问次数
awk '{print $1}' /var/log/nginx/dvr-vod-access.log | sort | uniq -c | sort -rn | head -10

# 统计状态码
awk '{print $9}' /var/log/nginx/dvr-vod-access.log | sort | uniq -c | sort -rn
```

### 性能优化

```nginx
# 在 nginx.conf 的 http 块中添加
http {
    # 工作进程数（通常设置为 CPU 核心数）
    worker_processes auto;
    
    # 每个进程的最大连接数
    events {
        worker_connections 1024;
        use epoll;  # Linux 下使用 epoll
    }
    
    # 开启 gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript 
               application/json application/javascript application/xml+rss;
    
    # 文件缓存
    open_file_cache max=1000 inactive=20s;
    open_file_cache_valid 30s;
    open_file_cache_min_uses 2;
}
```

## 故障排查

### 502 Bad Gateway

```bash
# 检查后端服务是否运行
curl http://localhost:8080/health

# 检查防火墙
sudo ufw status

# 检查 SELinux（CentOS/RHEL）
sudo setenforce 0  # 临时禁用
```

### 504 Gateway Timeout

```nginx
# 增加超时时间
location / {
    proxy_connect_timeout 120s;
    proxy_send_timeout 120s;
    proxy_read_timeout 120s;
    proxy_pass http://dvr_vod_backend;
}
```

### 权限问题

```bash
# 检查 Nginx 用户
ps aux | grep nginx

# 修改文件权限
sudo chown -R nginx:nginx /var/log/nginx
sudo chmod -R 755 /var/log/nginx
```

## 安全建议

1. **定期更新 Nginx**
   ```bash
   sudo apt-get update && sudo apt-get upgrade nginx
   ```

2. **隐藏 Nginx 版本**
   ```nginx
   http {
       server_tokens off;
   }
   ```

3. **限制请求方法**
   ```nginx
   if ($request_method !~ ^(GET|POST|HEAD)$ ) {
       return 405;
   }
   ```

4. **防止 DDoS**
   ```nginx
   limit_conn_zone $binary_remote_addr zone=addr:10m;
   limit_conn addr 10;
   ```

5. **配置防火墙**
   ```bash
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

## 参考资源

- [Nginx 官方文档](https://nginx.org/en/docs/)
- [Let's Encrypt](https://letsencrypt.org/)
- [Nginx 性能优化](https://www.nginx.com/blog/tuning-nginx/)
