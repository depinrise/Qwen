#!/bin/bash

# Script untuk menjalankan Telegram AI Bot

echo "🤖 Starting Telegram AI Bot with Streaming Support..."
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ File .env tidak ditemukan!"
    echo "📋 Silakan salin env.example ke .env dan isi dengan credentials Anda:"
    echo "   cp env.example .env"
    echo "   nano .env"
    exit 1
fi

# Check if bot binary exists
if [ ! -f bot ]; then
    echo "🔨 Building bot..."
    go build -o bot cmd/main.go
    if [ $? -ne 0 ]; then
        echo "❌ Build failed!"
        exit 1
    fi
fi

echo "🚀 Starting bot..."
echo "📡 WebSocket interface akan tersedia di: http://localhost:${HTTP_PORT:-8080}"
echo "🔄 Real-time streaming dengan thinking process aktif"
echo ""
echo "📱 Features:"
echo "   - Streaming AI responses di Telegram"
echo "   - WebSocket testing interface"
echo "   - Real-time thinking/reasoning display"
echo "   - Model: qwen-mt-turbo"
echo ""
echo "⏹️  Press Ctrl+C to stop"
echo "🐛 For debug mode, use: ./debug.sh"
echo ""

# Run with error handling
./bot

# If bot crashes, show helpful message
if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Bot crashed or exited with error!"
    echo "🔧 Try these steps:"
    echo "   1. Check your .env file credentials"
    echo "   2. Run ./debug.sh for detailed logs"
    echo "   3. Test WebSocket at http://localhost:8080"
    echo ""
fi
