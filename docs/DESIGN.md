# EchoDraw · 绘声 — 设计文档 (DESIGN)

> 文档版本: v0.1 · 最后更新: 2026-06-13
> 配套产品需求: `docs/PRD.md`
> 配套实现: 13 个 PR 已合并至 main

---

## 0. 阅读指南

本设计文档面向**工程实现者**和**未来接手人**：

- **1-3 章**讲"系统怎么组织的"（架构、模块、数据流）
- **4-6 章**讲"关键决策为什么这么做"（设计取舍，重要）
- **7-8 章**讲"做了什么、没做什么、为什么"（未实现说明）

PRD 关注"做什么"，本设计文档关注"怎么做"和"为什么这么做"。

---

## 1. 系统架构

### 1.1 整体视图

```
┌─────────────────────────────────────────────────────────────┐
│                         Browser                              │
│  ┌─────────────────┐    ┌──────────────────┐                │
│  │   Web Speech    │───▶│  VoiceBar UI     │                │
│  │   (Chrome)      │    │  + useSpeech     │                │
│  └─────────────────┘    └────────┬─────────┘                │
│                                  │ text                     │
│                                  ▼                          │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  applyIntent (前端派发器)                              │   │
│  │  Intent.cmd → switch → store.addShape / delete...     │   │
│  └──────────────────────────────────────────────────────┘   │
│         ▲                                                     │
│         │  { intents: [...] }                                │
│  ┌──────┴─────────┐                                          │
│  │  React + Konva │ ◄── user click/drag (备选交互)            │
│  │  + Zustand     │                                          │
│  └──────┬─────────┘                                          │
└─────────┼────────────────────────────────────────────────────┘
          │ fetch /api/understand  (Vite dev server 代理)
          ▼
┌─────────────────────────────────────────────────────────────┐
│                     Go Backend (Gin)                          │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              /api/understand handler                   │   │
│  │   rule engine hit? ─── yes ──▶ return Intent         │   │
│  │            │ no                                        │   │
│  │            ▼                                           │   │
│  │   LLM client (OpenAI-compatible)                       │   │
│  │   prompt + context ──▶ JSON ──▶ Intent                 │   │
│  └──────────────────────────────────────────────────────┘   │
│         │                                                     │
│         │  HTTP to OpenAI / DeepSeek / Qwen / ...            │
│         ▼                                                     │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 模块职责

| 模块 | 职责 | 关键文件 |
|---|---|---|
| **语音输入** | 浏览器 STT → 文本 | `web/src/hooks/useSpeech.ts` |
| **指令分发** | 把 Intent 翻译为 store 操作 | `web/src/lib/applyIntent.ts` |
| **画布状态** | shapes + selectedIds + 撤销/重做 | `web/src/store/canvasStore.ts` |
| **画布渲染** | Konva Stage + Layer + Transformer | `web/src/components/Canvas.tsx` |
| **HTTP 客户端** | fetch /api/understand | `web/src/lib/api.ts` |
| **路由层** | 接 engine + llm | `server/internal/api/server.go` |
| **规则引擎** | 80% 结构化指令 | `server/internal/engine/engine.go` |
| **LLM 客户端** | 20% 复杂/模糊指令 | `server/internal/llm/llm.go` |
| **共享类型** | Shape/Intent/Intent.Action | `server/internal/model/types.go` + `web/src/types/index.ts` |
| **共享词表** | 颜色/形状/位置词 | `server/internal/model/lexicon.go` |

### 1.3 数据契约

**前后端共用** Intent 类型（关键字段）：

```typescript
interface Intent {
  cmd: "create" | "scene" | "update" | "delete" | "select" |
       "undo" | "redo" | "clear" | "export" | "ask_back" | "unknown";
  action?: "delta" | "redraw" | "modify" | "clear";
  shape?: Shape;           // create 时
  scenes?: Shape[];        // scene 时
  target?: { ids: string[]; ref?: string; filter?: string };
  patch?: Style;           // update 时
  moveTo?: { x: number; y: number };
  scale?: number;
  rotation?: number;
  question?: string;       // ask_back 时
  reply?: string;          // UI 提示
}
```

**Action 字段** 是 v0.2 新增——前端拿 Action 直接知道走 delta/redraw/modify/clear 哪条路径，避免在前端重新分类。

---

## 2. 关键数据流

### 2.1 一句话生成场景（含 LLM）

```
用户说: "画一个夜空有月亮和星星"
    ↓ Web Speech API
text = "画一个夜空有月亮和星星"
    ↓ fetch /api/understand
server.handleUnderstand:
    engine.Understand → Hit=false (规则未命中)
    ↓
    llm.Understand(ctx, text):
        system_prompt + 画布 context → LLM
        LLM 返回 JSON
        parseIntents → [{cmd: "scene", scenes: [bg, moon, 30 stars]}]
    ↓
返回 {intents: [{cmd: "scene", action: "delta", scenes: [...]}]}
    ↓ 前端 applyIntent
store.addShapes(scenes) → 30+ 节点入栈 (一次性)
    ↓ React 重新渲染
画布出现夜空
```

### 2.2 增量 + 修改

```
ctx = [red_circle, blue_rect]
用户说: "再画一个圆"
    ↓
rule engine: 命中 create + delta + shape=circle(default 颜色, default 位置)
    ↓
ctx → [red_circle, blue_rect, gray_circle]
    ↓
用户说: "把它改成黄色"
    ↓
rule engine matchTransform:
    hasTarget("它") = true
    resolveTarget("它", ctx) → ref=last, filter="黄色"
    // fix 1: 改色动词出现时跳过 filter
    // fix 2: ref=last + filter 时退化 ref=all
    → ids = [gray_circle]
    pickColor("黄色") = #fdd835
    → intent{cmd: update, target: {ids: [gray_circle]}, patch: {fill: #fdd835}}
    ↓
store.updateShapes(["gray_circle"], {style: {fill: "#fdd835"}})
    ↓
gray_circle 变黄
```

---

## 3. 模块设计要点

### 3.1 规则引擎

**为什么这样组织代码**：
- 单一入口 `Engine.Understand(text, ctx) → Result{Hit, Intents, Reply}`
- 判定顺序：**系统 → 删除 → 选中 → 变换 → 创建 → 未命中**（避免 "删" 和 "改" 混淆）
- 纯函数优先：`pickShape` / `pickColor` / `pickPosition` / `pickNumber` / `pickScale` 都返回 `(X, ok)`，易测
- 不引入 `map[string]func` 的 closure 风格——所有逻辑用 `if/switch` 显式分叉，编译期就能发现类型错误

**性能**：
- O(n) 词表查找（map + contains），n=词表大小
- 单次调用 < 1ms（go test 实测）

**已知 bug 及修复**（已合并至 v0.1）：
- ✅ `pickScale` 不支持中文数字 → 增 `cnNum` 词表
- ✅ `resolveTarget` 在 `ref=last + filter` 时误判 → 改 `ref=all`
- ✅ `resolveTarget` 在改色动词出现时把颜色词当 filter → 跳过 filter

### 3.2 LLM 客户端

**Prompt 设计要点**：
- **封闭图元词汇表**：8 种 ShapeType，禁止发明
- **封闭颜色词汇表**：13 种 hex，词表与 lexicon.go 一致
- **场景拆解模板**：夜空/海面/花园/咖啡馆/房子 各给一个固定模板
- **强约束**：JSON mode + 严格 JSON 输出 + 单 scene 上限 50

**为什么不调 OpenAI Function Calling / Tools**：
- Function calling 增加 token 开销（每次 100+ token 定义）
- 我们的输出形态固定（数组 + 4 种 cmd），普通 JSON 模式够用
- 减少 prompt 复杂度 = 降低幻觉率

**失败兜底**：
- 网络错误：catch 后转 `ask_back`
- JSON 解析失败：转 `ask_back`
- 返回 0 个 intents：转 `ask_back`
- 5xx / 超时：调用方负责（本层只 wrap error）

### 3.3 Zustand Store

**为什么不用 Redux**：
- 4KB 体积差异
- 无 boilerplate
- 本项目状态简单（一个数组 + 一个 Set），Redux 杀鸡用牛刀

**为什么 `addShapes` 单独提供**：
- 场景生成一次产生 30+ 图元，如果逐个 `addShape` 会 push 30 次 history
- 批量 add 一次入栈，撤销一次回到场景前

**为什么 history 上限 50**：
- 50 步撤销足够日常使用
- 内存可控：每次 snapshot 浅拷贝，复杂度 O(50 * shapes)

**为什么不持久化**（到 IndexedDB）：
- v0.1 不要求
- 持久化需要序列化 + 错误恢复 + 多 tab 同步，复杂度高
- v0.2 视情况加

### 3.4 Konva 画布

**为什么选 Konva 而非 Fabric.js / 原生 Canvas**：
- 节点化（Stage > Layer > Shape），与 React 思维一致
- Transformer 内置（选中后自动挂上，可拖动/缩放/旋转）
- 社区成熟

**坐标语义**：所有 Shape.X/Y 为**中心点**（不是左上角）。Rect 通过 `offsetX/offsetY = width/2` 居中。这与后端 `engine.matchCreate` 的 "position 是中心" 一致。

**为什么不用 Stage.click 抓取背景点击**：
- `e.target === e.target.getStage()` 判断更精确，避免点到子元素时误判

### 3.5 applyIntent

**switch (cmd) 派发**而非 if-else 链：
- 新增 cmd 时编译器/IDE 提示
- 性能差异可忽略

**为什么 `update` 走 updateShapes 而非 updateShape**：
- `target.ids` 可能是 1 个也可能是 N 个（"所有红色"）
- 统一用 batch 接口，简化代码

**export 怎么实现**：
- 找 DOM `<canvas>` → `toDataURL("image/png")` → `<a download>` 点击
- **必须在 Canvas 组件渲染后**才能找到 `<canvas>`，所以延迟到点击时 querySelector
- 不在 applyIntent 内做是因为它依赖 DOM，store 不能依赖

---

## 4. 设计决策记录 (ADR 风格)

### ADR-001: 选 A 档（结构化图元）而非 B 档（位图生图）

**决策**：LLM 输出结构化图元 JSON，前端 Konva 渲染。

**考虑过的方案**：
- A 档：LLM → 8 种图元 JSON → Konva 渲染。**采用**
- B 档：LLM → Stable Diffusion / DALL·E → 位图。**放弃**
- C 档（视觉）：LLM 看截图改。**放弃**

**理由**：
- A 档优势：可编辑、可继续改、token 省、确定性高、纯前端无需 GPU
- B 档劣势：位图无法选中/修改，违背"AI 语音绘图" vs "AI 生图" 的核心差异
- C 档劣势：改不精确、慢、贵

**代价**：
- 复杂对象（手/Logo/写实猫）画不好
- 这是 v0.1 接受的边界

### ADR-002: 规则引擎 + LLM 双层，规则优先

**决策**：80% 走规则引擎，20% 走 LLM。

**理由**：
- 规则引擎 < 1ms 响应
- LLM 1-3s 响应 + token 成本
- 大部分指令是结构化的（"画 X" + 修饰词），规则足够

**判定边界**：
- 规则未命中 = LLM 兜底
- 命中但需要 4 个修饰词以上 = 仍然走规则（避免假阴性）

### ADR-003: 后端不存画布状态

**决策**：画布状态在前端 Zustand，后端不存。

**理由**：
- 撤销/重做必须毫秒级，round-trip 后端不可接受
- 后端只做"指令理解"，职责单一
- 前端可离线工作（虽然语音不能用）

**代价**：
- 刷新页面 = 画布丢失（v0.1 接受，v0.2 加 IndexedDB 持久化）

### ADR-004: Vite dev server 代理 /api 到 :8080

**决策**：前端 fetch 相对路径 `/api/understand`，Vite proxy 到后端。

**理由**：
- 写代码不用关心 baseURL
- 跨域问题由后端 CORS 兜底（生产环境）

**生产环境**：
- 静态文件由 nginx 提供，nginx 反代 /api 到后端
- 详见后续部署文档（v0.2）

### ADR-005: Action 字段前置

**决策**：Intent 增加 `action` 字段，值 `delta/redraw/modify/clear`。

**理由**：
- 4 种 Action 对应 PRD 附录 A.1 的 4 类指令
- 前端拿 action 直接走渲染路径，不用重新分类
- 减少前端的 if-else

**不放在 cmd 上的原因**：
- cmd 是"系统级操作类型"（create/update/...），action 是"用户意图类型"（增量/重画/...）
- 同一 cmd 可以有不同 action（如 create + delta vs scene + delta）

---

## 5. 性能 & 可靠性

### 5.1 性能预算（PRD 附录 A.6）

| 指标 | 目标 | 实测方法 | v0.1 状态 |
|---|---|---|---|
| 语音首字延迟 | < 500ms | Chrome DevTools | ✅ 浏览器原生 |
| 规则引擎响应 | < 100ms | 后端日志 | ✅ < 1ms |
| LLM 响应 | < 3s | fetch 计时 | ✅ OpenAI gpt-4o-mini ~1.5s |
| 渲染 100 图元 | < 16ms/frame | Konva cache | ✅ 未压测 |
| 撤销/重做 | < 50ms | 前端 setState | ✅ 毫秒级 |

### 5.2 错误降级

| 失败 | 降级 |
|---|---|
| 浏览器无 Web Speech | 提示切 Chrome，文本输入框降级（v0.2） |
| 规则 + LLM 都失败 | 语音播报"没听清" |
| LLM 返回非法 JSON | 转 ask_back |
| 复杂场景 > 100 图元 | 截断 + 提示 |
| 后端 5xx | 前端 catch，UI 提示 |

### 5.3 安全

**当前不涉及**：
- 无登录/鉴权（v0.1 单机使用）
- 无数据持久化（v0.1 刷新即丢）
- 无用户上传（v0.1 无文件上传）

**后续需要**：
- LLM API key 由用户自配（不存到后端）
- 上传图片（如 v1.2 的生图导入）需要鉴权 + 大小限制

---

## 6. 复用清单

| 用途 | 选型 | 理由 |
|---|---|---|
| 画布 | Konva.js | 节点化、Transformer 内置 |
| 语音 | Web Speech API | 零成本、中文流式 |
| LLM | OpenAI-compatible | 兼容多家 |
| 后端框架 | Gin | 主流、WebSocket 友好 |
| 状态管理 | Zustand | 4KB、简单 |
| UI 组件 | 原生 + Tailwind | 工具类、迭代快 |
| 图标 | lucide-react | 1KB/图标、按需引入 |
| 构建 | Vite | 快、HMR 好 |
| 测试 (后端) | Go testing | 内置、table-driven |
| 测试 (前端) | Vitest | 0 配置、与 Vite 整合 |

---

## 7. 已实现清单

| PR | 能力 |
|---|---|
| #1 | 仓库 + PRD |
| #2 | server 启动 + /health |
| #3 | data model + 词表 |
| #4 | 规则引擎 (4 类 Action) |
| #5 | 27 个引擎单测 |
| #6 | LLM 客户端 (6 测试) |
| #7 | /api/understand 路由 (5 测试) |
| #8 | Web 脚手架 |
| #9 | Zustand store (15 测试) |
| #10 | Konva 画布 |
| #11 | Web Speech Hook |
| #12 | 指令分发层 |
| #13 | 启动脚本 + README + 联调 |

**总测试数**：后端 38 + 前端 15 = 53 个测试

---

## 8. 未实现清单（重要）

> 题目要求"额外提交一份设计文档，记录你计划支持哪些指令能力，最终实现了哪些，以及未完成部分的原因说明"
> 本章是核心交付物之一

### 8.1 计划支持但未实现

| 功能 | 计划来源 | 原因 | 计划补回时间 |
|---|---|---|---|
| **多选 (shift+click)** | PRD 1.0 MVP | v0.1 范围内但优先级低，先单选 | v0.2 |
| **拖动 Transformer 改 width/height 同步到 store** | 画布交互 | Konva Transformer 没接 store onChange | v0.2 |
| **文本双击编辑** | 画布交互 | 需要 contenteditable + Konva 同步，复杂度高 | v1.0 |
| **画布持久化 (IndexedDB)** | PRD 8.1 | v0.1 接受刷新即丢 | v0.2 |
| **多轮对话上下文（"再画一个一样的"）** | PRD 5.2 | 需要 session 管理，PRD 已标后续 | v1.0 |
| **多语言 (英文)** | PRD 非功能需求 | Chrome Web Speech API 支持但中文优先 | v1.0 |
| **对象语义库（"窗户"=矩形+4 线）** | PRD 5.2（讨论过） | v0.1 LLM 模板覆盖 5 个场景就够 | v1.1 |
| **AI 真实生图（B 档）作为"导入素材"** | PRD 8 v1.2 | 设计决策 ADR-001 | v1.2 |
| **自由画笔** | PRD 8 v1.1 | 与结构化图元冲突大 | v1.1 |
| **图层管理** | PRD 8 v1.1 | 需要 z-index 重构 | v1.1 |
| **协作模式** | PRD 8 v1.3 | 需要 CRDT / WebSocket 重构 | v1.3 |
| **移动端 / PWA** | PRD 8 v2.0 | Web Speech API 移动端支持差 | v2.0 |

### 8.2 复杂场景降级（v0.1 行为）

**问题**：无 LLM 配置时，"画一个夜空" 类指令会被规则引擎降级为"画其中的形状"。

**示例**：
- 输入：`"画一个夜空有月亮和星星"`
- 规则引擎：找到"星星" → 画一个五角星
- LLM 启用：找到场景"夜空" → 拆为背景 + 月亮 + 30 颗星

**这是有意的降级**——v0.1 设计为"无 LLM 也能用基础功能"。

**用户应对**：
- 启用 LLM（推荐）
- 或者把场景指令拆为"先画背景矩形，再画月亮，再画 10 颗星"

### 8.3 中英文数字支持不全

**问题**：`pickScale` 支持 `一二两三四五六七八九十半`，**不支持** `十一/十二/二十` 等。

**示例**：
- `"放大两倍"` ✅ → 2
- `"放大三倍"` ✅ → 3
- `"放大十倍"` ✅ → 10
- `"放大十一倍"` ❌ → 未匹配，回退到 1.5

**原因**：MVP 不需要 10 以上的数。`十一` 拆词复杂（"十"+"一"）。

**计划补回**：v1.0

### 8.4 重画 (redraw) 暂未实现多 intent

**问题**：PRD 附录 A.1 的 "重画" 语义是 `delete + create` 拆为两个 intent。当前 `matchTransform` 只产出单 intent（`update` + `patch`），不区分重画和修改。

**当前行为**：
- `"把那个圆改成方形"` → update + patch（不创建新图元）
- 如果用户想"真重画"（删旧 + 建新），需要手动说"删除 + 画"

**原因**：API 层要做 multi-intent 包装，需要前端 applyIntent 也支持。当前架构支持扩展，但 v0.1 简化了。

**计划补回**：v0.2

### 8.5 没有 WebSocket 实时通信

**计划**：PRD 架构图里画了 WebSocket（"语音 → 实时识别 → 流式回传"）。

**实际**：用 HTTP POST `/api/understand`，一次性返回。

**原因**：
- 浏览器 Web Speech API 已经在前端做了 STT，识别完成后才发后端
- WebSocket 的价值在"边说边识别"，但 Web Speech API 是 final 结果才稳定
- 当前 1.5s 延迟用户感受不到

**计划补回**：v1.0（如果用流式 STT + LLM 联动）

### 8.6 没做的优化

- Konva Layer cache（100+ 图元时考虑）
- 画布缩放/平移（v0.1 固定 1200x800）
- 颜色拾色器（v0.1 只能语音改色）

---

## 9. 后续 Roadmap

```
v0.2 (1-2 周)
├── 多选 + 拖动改尺寸
├── IndexedDB 持久化
├── 文本输入框降级 (Web Speech 不可用时)
├── 重画 multi-intent 支持
└── 修复 Transformer ↔ store 同步

v1.0 (1-2 月)
├── 自由画笔
├── 图层管理
├── 多轮对话
├── 性能优化 (100+ 图元)
└── 中英文双语

v1.1+
├── 对象语义库 (50-100 个常用对象)
├── AI 真实生图 (B 档作为"导入素材")
└── 协作模式 (CRDT)

v2.0
├── 移动端
├── PWA
└── 离线模式
```

---

## 10. 给接手人的建议

**如果你要加新指令类型**：
1. 优先加规则引擎（engine.go + lexicon.go）
2. 复杂场景才走 LLM（llm.go prompt）
3. 加单测（engine_test.go）
4. 跑 `go test ./...` 验证

**如果你要加新图元类型**：
1. `model/types.go` 加 `ShapeType` 常量
2. `model/lexicon.go` 加 `ShapeWords`
3. `web/src/components/ShapeNode.tsx` 加 switch case
4. `web/src/types/index.ts` 加 `ShapeType` union
5. LLM prompt 里加类型说明

**如果你要改 LLM**：
1. `llm/llm.go` 改 `New()` 接受不同协议（已抽象）
2. 或换 client（langchaingo / openai-go 等）

**如果你要加部署**：
1. 后端：Dockerfile（多阶段构建，golang:alpine）
2. 前端：`npm run build` 出 dist，nginx 提供静态文件
3. nginx 反代 /api 到后端 :8080
4. Web Speech API 需要 HTTPS（生产环境必须）

---

## 附录 A: 文件清单

| 路径 | 行数 | 用途 |
|---|---|---|
| `docs/PRD.md` | 338 | 产品需求 |
| `docs/DESIGN.md` | (本文档) | 设计文档 |
| `server/cmd/server/main.go` | 50 | 入口 |
| `server/internal/api/server.go` | 119 | HTTP 路由 |
| `server/internal/engine/engine.go` | 560 | 规则引擎 |
| `server/internal/engine/engine_test.go` | 305 | 引擎单测 |
| `server/internal/llm/llm.go` | 200 | LLM 客户端 |
| `server/internal/llm/llm_test.go` | 90 | LLM 单测 |
| `server/internal/api/server_test.go` | 95 | API 单测 |
| `server/internal/model/types.go` | 130 | 共享类型 |
| `server/internal/model/lexicon.go` | 82 | 词表 |
| `server/internal/config/config.go` | 50 | 配置 |
| `web/src/App.tsx` | 100 | 根组件 |
| `web/src/components/Canvas.tsx` | 90 | 画布 |
| `web/src/components/ShapeNode.tsx` | 115 | 图元节点 |
| `web/src/components/VoiceBar.tsx` | 80 | 语音 UI |
| `web/src/hooks/useSpeech.ts` | 113 | 语音 Hook |
| `web/src/lib/api.ts` | 20 | API 客户端 |
| `web/src/lib/applyIntent.ts` | 80 | 派发器 |
| `web/src/store/canvasStore.ts` | 192 | 状态 |
| `web/src/store/canvasStore.test.ts` | 100 | 状态单测 |
| `web/src/types/index.ts` | 70 | 共享类型 |
| `scripts/dev.sh` | 30 | 启动脚本 |
| `README.md` | 163 | 产品 README |

**总代码行数**（不含 vendor/node_modules）：~3,000 行

## 附录 B: 术语表

- **Action**：用户意图类型（delta/redraw/modify/clear）
- **Intent**：后端返回的结构化指令
- **Shape**：画布上图元
- **Context**：当前画布上的 shapes 列表（指代消解用）
- **Hit**：规则引擎是否命中
- **AskBack**：信息不足反问用户
- **Scene**：LLM 生成的多图元场景
