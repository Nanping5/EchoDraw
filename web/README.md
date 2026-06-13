# EchoDraw Web (React + Vite)

EchoDraw 的前端，纯语音控制绘图工具的画布与交互。

## 快速开始

```bash
# 1. 安装依赖
npm install

# 2. 启动 dev server (默认 5173)
npm run dev

# 3. 打开浏览器
open http://localhost:5173
```

Vite dev server 已配置 `/api` 代理到 `http://localhost:8080`，需要先启动后端。

## 可用脚本

| 命令 | 说明 |
|---|---|
| `npm run dev` | 启动 Vite dev server |
| `npm run build` | 生产构建到 `dist/` |
| `npm run preview` | 预览生产构建 |
| `npm run test` | 跑 vitest |

## 目录结构

```
web/
├── src/
│   ├── App.tsx              # 根组件
│   ├── main.tsx             # 入口
│   └── index.css            # Tailwind 入口
├── index.html
├── vite.config.ts
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json
└── package.json
```

## 当前状态

✅ 脚手架 (PR #8) - Vite + React + TS + Konva + Tailwind
🔜 Zustand store + 撤销/重做 (PR #9)
🔜 Konva 画布组件 (PR #10)
🔜 Web Speech API Hook (PR #11)
🔜 指令分发层 (PR #12)
🔜 联调 + 启动脚本 (PR #13)
