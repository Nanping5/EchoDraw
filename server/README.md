# EchoDraw Server (Go + Gin)

EchoDraw 的后端服务，负责：
- 接收前端传来的语音识别文本
- 走规则引擎（80% 常见指令）或 LLM（20% 复杂/模糊）解析为结构化意图
- 返回统一的 `UnderstandResponse` 给前端渲染

## 快速开始

```bash
# 1. 安装依赖
go mod tidy

# 2. 启动 (无 LLM, 仅规则引擎)
make run

# 3. 启动 (启用 LLM)
export LLM_API_KEY=sk-...
export LLM_BASE_URL=https://api.openai.com/v1  # 或其他 OpenAI-compatible 服务
export LLM_MODEL=gpt-4o-mini
make run
```

服务默认监听 `:8080`。

## 环境变量

| 变量 | 默认 | 说明 |
|---|---|---|
| `PORT` | `8080` | HTTP 监听端口 |
| `LLM_API_KEY` | (空) | 留空则禁用 LLM, 仅走规则引擎 |
| `LLM_BASE_URL` | `https://api.openai.com/v1` | OpenAI-compatible 服务地址 |
| `LLM_MODEL` | `gpt-4o-mini` | 调用的模型名 |
| `CANVAS_W` | `1200` | 画布宽（用于位置计算） |
| `CANVAS_H` | `800` | 画布高 |

## 目录结构

```
server/
├── cmd/server/         # 入口 main
├── internal/
│   ├── api/            # HTTP 路由 (后续 PR)
│   ├── config/         # 环境变量加载
│   ├── engine/         # 规则引擎 (后续 PR)
│   ├── llm/            # LLM 客户端 (后续 PR)
│   └── model/          # 数据模型 (后续 PR)
├── go.mod
└── Makefile
```

## 当前状态

✅ 启动 + `/health` 检查 (PR #2)
🔜 规则引擎 (PR #3)
🔜 LLM 客户端 (PR #4)
