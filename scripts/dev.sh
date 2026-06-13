#!/usr/bin/env bash
# 启动 server + web 开发环境
# 用法: ./scripts/dev.sh
# 终止: Ctrl+C (会同时关掉两个进程)

set -e
cd "$(dirname "$0")/.."

cleanup() {
  echo ""
  echo "Stopping..."
  kill $SERVER_PID $WEB_PID 2>/dev/null || true
  wait 2>/dev/null
}
trap cleanup EXIT INT TERM

echo "▶ Starting server (Go) on :8080..."
(cd server && go run ./cmd/server) &
SERVER_PID=$!

echo "▶ Starting web (Vite) on :5173..."
(cd web && npm run dev) &
WEB_PID=$!

echo ""
echo "==========================================="
echo "  EchoDraw · 绘声 已启动"
echo "  Web:    http://localhost:5173"
echo "  Server: http://localhost:8080/health"
echo "  Press Ctrl+C to stop"
echo "==========================================="
wait
