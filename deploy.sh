#!/bin/bash

# DVR 点播系统 - 快速部署脚本

set -e

echo "=========================================="
echo "DVR 点播系统 - 快速部署"
echo "=========================================="
echo ""

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装，请先安装 Docker"
    echo "安装指南: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose 未安装，请先安装 Docker Compose"
    echo "安装指南: https://docs.docker.com/compose/install/"
    exit 1
fi

echo "✅ Docker 环境检查通过"
echo ""

# 检查配置文件
if [ ! -f "config.yml" ]; then
    echo "⚠️  未找到 config.yml，从示例文件创建..."
    if [ -f "config.example.yml" ]; then
        cp config.example.yml config.yml
        echo "✅ 已创建 config.yml，请编辑配置文件设置 DVR 服务器地址"
        echo ""
        read -p "是否现在编辑配置文件? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            ${EDITOR:-vi} config.yml
        fi
    else
        echo "❌ 未找到 config.example.yml"
        exit 1
    fi
fi

echo ""
echo "开始构建和启动服务..."
echo ""

# 构建并启动
docker-compose up -d --build

echo ""
echo "=========================================="
echo "✅ 部署完成！"
echo "=========================================="
echo ""
echo "服务地址: http://localhost:8080"
echo ""
echo "常用命令:"
echo "  查看日志: docker-compose logs -f"
echo "  停止服务: docker-compose down"
echo "  重启服务: docker-compose restart"
echo "  查看状态: docker-compose ps"
echo ""
echo "详细文档请查看: DEPLOY.md"
echo ""
