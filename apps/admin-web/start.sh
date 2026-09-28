#!/bin/bash

echo "🚀 聊天记录管理系统启动脚本"
echo "=================================="

# 检查 Node.js 是否安装
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装，请先安装 Node.js (版本 >= 14.0)"
    exit 1
fi

# 检查 npm 是否安装
if ! command -v npm &> /dev/null; then
    echo "❌ npm 未安装，请先安装 npm"
    exit 1
fi

echo "✅ Node.js 版本: $(node --version)"
echo "✅ npm 版本: $(npm --version)"
echo ""

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
    echo "📦 正在安装项目依赖..."
    npm install
    
    if [ $? -eq 0 ]; then
        echo "✅ 依赖安装成功"
    else
        echo "❌ 依赖安装失败"
        exit 1
    fi
else
    echo "✅ 依赖已安装"
fi

echo ""
echo "🎯 系统功能说明："
echo "1. 仪表盘 - 系统统计和快速操作"
echo "2. 聊天记录 - 强大的聊天记录查询功能"
echo "3. 联系人管理 - 联系人信息管理"
echo "4. 群聊管理 - 群聊信息和成员管理"
echo "5. 会话列表 - 最近会话快速访问"
echo "6. 多媒体管理 - 图片、视频、语音、文件管理"
echo ""

echo "🔧 API 配置说明："
echo "默认 API 地址: http://127.0.0.1:5030"
echo "如需修改，请编辑 src/api/index.js 文件"
echo ""

echo "🌐 正在启动开发服务器..."
echo "访问地址: http://localhost:8080"
echo ""

# 启动开发服务器
npm run serve 