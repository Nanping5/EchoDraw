# EchoDraw · 绘声

> 纯语音控制的 AI 绘图工具。说一句话就能生成场景，并能继续用语音修改它。

![Version](https://img.shields.io/badge/version-0.1.0-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## ✨ 核心特性

- 🎤 **纯语音输入**：浏览器原生 Web Speech API，中文优先
- 🤖 **AI 场景生成**："画一个夜空有月亮和星星" → 自动拆解为 30+ 个图元
- 🖊️ **结构化图元**：圆/矩形/线/椭圆/三角/星/箭头/文本 8 种
- 🎨 **增量 + 重画 + 修改**：4 种 Action 模式自由切换
- ↩️ **撤销/重做**：50 步历史栈
- 💾 **PNG 导出**：一键保存画布

## 🚀 快速开始

### 1. 一键启动（推荐）

```bash
git clone https://github.com/Nanping5/EchoDraw.git
cd EchoDraw
./scripts/dev.sh
```

会自动启动：
- **后端**：http://localhost:8080
- **前端**：http://localhost:5173

浏览器打开 http://localhost:5173 即可。

### 2. 分别启动

```bash
# 终端 1: 启动后端
cd server
go run ./cmd/server

# 终端 2: 启动前端
cd web
npm install  # 首次需要
npm run dev
```

### 3. 启用 LLM（可选，用于复杂指令）

```bash
# server 默认无 LLM, 只能走规则引擎 (覆盖 ~80% 指令)
# 启用 LLM 后, 复杂场景如"画一个夜空"才能生成多图元

export LLM_API_KEY=sk-...
export LLM_BASE_URL=https://api.openai.com/v1  # 或 DeepSeek/Qwen/Moonshot
export LLM_MODEL=gpt-4o-mini

./scripts/dev.sh
```

## 🎙️ 试试这些指令

打开 http://localhost:5173，点击麦克风按钮，允许权限，然后说：

| 指令 | 效果 |
|---|---|
| 画一个红色的大圆 | 创建大红圆 |
| 画一个蓝色矩形在左边 | 左侧创建蓝矩形 |
| 画一个夜空有月亮和星星 | LLM 生成 30+ 个图元 |
| 再画一个圆 | 追加一个圆（增量） |
| 把它改成黄色 | 改色 |
| 把它放大两倍 | 缩放 |
| 把它移到中间 | 移动 |
| 旋转 45 度 | 旋转 |
| 撤销 / 重做 | 历史栈 |
| 清空画布 | 全部清除 |

## 🏗️ 架构

```
┌─────────────────┐     WebSocket      ┌──────────────────┐
│  React Frontend │ ◄────────────────► │   Go Backend     │
│  - Web Speech   │                    │   - Gin          │
│  - React-Konva  │   HTTP             │   - 规则引擎     │
│  - Zustand      │ ◄────────────────► │   - LLM 客户端   │
└─────────────────┘                    └──────────────────┘
                                                │
                                                ▼
                                       ┌──────────────────┐
                                       │  LLM Provider    │
                                       │  (OpenAI compat) │
                                       └──────────────────┘
```

**关键设计**：
- **规则引擎优先**：80% 常见指令走规则引擎，毫秒级响应
- **LLM 兜底**：复杂场景 + 模糊指令走 LLM
- **前端为权威**：画布状态在 React/Zustand
- **结构化输出**：LLM 输出结构化 JSON 图元列表

详细架构见 `docs/PRD.md`。

## 📁 目录结构

```
EchoDraw/
├── docs/
│   ├── PRD.md             # 产品需求文档
│   └── DESIGN.md          # 设计文档
├── server/                # Go + Gin 后端
│   ├── cmd/server/        # 入口
│   ├── internal/
│   │   ├── api/           # HTTP 路由
│   │   ├── config/        # 环境变量
│   │   ├── engine/        # 规则引擎
│   │   ├── llm/           # LLM 客户端
│   │   └── model/         # 共享类型
│   ├── go.mod
│   └── Makefile
├── web/                   # React + Vite 前端
│   ├── src/
│   │   ├── components/    # Canvas / VoiceBar / ShapeNode
│   │   ├── hooks/         # useSpeech
│   │   ├── lib/           # api / applyIntent
│   │   ├── store/         # Zustand canvasStore
│   │   └── types/         # 共享类型
│   ├── package.json
│   └── vite.config.ts
├── scripts/
│   └── dev.sh             # 一键启动
└── README.md
```

## 🧪 测试

```bash
# 后端 (38 个测试)
cd server && go test ./...

# 前端 (15 个测试)
cd web && npm test
```

## 🌐 浏览器支持

- ✅ Chrome / Edge 桌面版（推荐）
- ❌ Firefox / Safari（Web Speech API 识别支持有限）
- ❌ 移动端（暂未适配）

## 🔮 后续 Roadmap

- v1.1：自由画笔 + 图层管理
- v1.2：AI 真实生图（"导入素材"）
- v1.3：协作模式
- v2.0：移动端 / PWA

## 📄 文档

- [PRD](docs/PRD.md) - 产品需求文档
- [DESIGN](docs/DESIGN.md) - 设计文档（含未实现说明）

## 📝 License

MIT
