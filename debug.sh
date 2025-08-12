#!/bin/bash

# Debug script untuk Telegram AI Bot

echo "🐛 Debug Mode: Telegram AI Bot"
echo "=============================="

# Kill existing bot process if any
echo "🔄 Stopping existing bot processes..."
pkill -f "./bot" 2>/dev/null
sleep 1

# Build with debug info
echo "🔨 Building bot with debug info..."
go build -o bot cmd/main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "🚀 Starting bot in debug mode..."
echo "📱 Bot logs will show detailed error information"
echo "🔍 WebSocket interface: http://localhost:8080"
echo "⏹️  Press Ctrl+C to stop"
echo ""

# Run with verbose output
./bot 2>&1 | tee bot.log
