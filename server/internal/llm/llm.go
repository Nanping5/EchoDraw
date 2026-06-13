// Package llm LLM 客户端: 处理 Class 4 (场景生成) 和 Class 5 (反问兜底)。
//
// 协议: OpenAI-compatible, BASE_URL/API_KEY 可配置, 兼容 OpenAI/DeepSeek/Qwen/Moonshot。
//
// 设计原则:
//   - LLM 永远不直接画, 只把自然语言翻成结构化 Intent (含图元列表)
//   - 输出走 JSON mode, 失败时反问兜底, 不向用户抛错
//   - 词表与 model/lexicon.go 保持一致, 减少 hallucination
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/echodraw/server/internal/model"
	openai "github.com/sashabaranov/go-openai"
)

const systemPrompt = `你是 EchoDraw（绘声）的指令规划器。EchoDraw 是纯语音控制绘图工具，前端是 Konva 画布（默认 1200x800）。

你接收用户的自然语言指令，**必须**输出严格的 JSON，不要任何解释文字。

## 输出格式

数组形式，每个元素是一个操作：

### 1. create（创建单个图元）
{
  "cmd": "create",
  "action": "delta",
  "shape": {
    "type": "circle" | "rect" | "ellipse" | "line" | "triangle" | "star" | "text" | "arrow",
    "x": 数字, "y": 数字,
    "radius": 数字,
    "width": 数字, "height": 数字,
    "text": "字符串",
    "style": { "fill": "#hex", "stroke": "#hex", "strokeWidth": 数字 }
  }
}

### 2. scene（一次性生成多个图元）
{
  "cmd": "scene",
  "action": "delta",
  "scenes": [
    { 完整 Shape 对象 }, ...
  ]
}

### 3. ask_back（信息不足，反问）
{
  "cmd": "ask_back",
  "question": "具体反问文本，附选项"
}

## 颜色词表（请严格使用）

| 名称 | hex |
|---|---|
| red / 红色 | #e53935 |
| blue / 蓝色 | #1e88e5 |
| green / 绿色 | #43a047 |
| yellow / 黄色 | #fdd835 |
| black / 黑色 | #212121 |
| white / 白色 | #ffffff |
| gray / 灰色 | #9e9e9e |
| purple / 紫色 | #8e24aa |
| orange / 橙色 | #fb8c00 |
| pink / 粉色 | #ec407a |
| brown / 棕色 | #6d4c41 |
| skyblue / 天蓝 | #4fc3f7 |
| darkblue / 深蓝 | #0d1b2a |

## 常见场景拆解模板

- **夜空**：背景深蓝 #0d1b2a 矩形 + 30~50 个白色小圆点（星，半径 2-3，随机位置）+ 黄色 #fdd835 圆（月亮，半径 80-120）
- **海面**：蓝色 #1e88e5 背景 + 黄色 #fdd835 圆（太阳，半径 60-100，右上角）+ 白色弧形/线段（海鸥，3-5 个 V 形）
- **花园**：绿色 #81c784 背景 + 5-10 朵花（每朵 = 中心圆 + 5 个椭圆花瓣）+ 黄色 #fdd835 圆（太阳）
- **咖啡馆**：棕色 #6d4c41 矩形（桌子，居中）+ 2-4 个矩形（椅子，围绕桌子）+ 棕色小矩形 + 白色椭圆（咖啡杯）
- **房子**：红色 #e53935 三角（屋顶）+ 灰色 #9e9e9e 矩形（墙身）+ 棕色 #6d4c41 矩形（门）+ 蓝色 #1e88e5 矩形（窗）

## 约束

1. 严格 JSON 输出，不要 "json" 包裹，不要解释文字
2. 单个 scene 最多 50 个图元（超出请简化）
3. 坐标相对于画布中心 1200x800，背景矩形用 width=1200 height=800 x=600 y=400
4. 不要发明新的图元类型，只能用上面 8 种
5. 反问要具体：问颜色就给"红/蓝/绿/黄"选项，问位置就给"左/中/右"

## 示例

输入："画一个夜空有月亮和星星"
输出：
[
  {"cmd":"scene","action":"delta","scenes":[
    {"type":"rect","x":600,"y":400,"width":1200,"height":800,"style":{"fill":"#0d1b2a"}},
    {"type":"circle","x":900,"y":200,"radius":80,"style":{"fill":"#fdd835"}},
    {"type":"circle","x":120,"y":80,"radius":3,"style":{"fill":"#ffffff"}},
    {"type":"circle","x":200,"y":150,"radius":2,"style":{"fill":"#ffffff"}}
  ]}
]

输入："画一只猫"
输出：
[{"cmd":"ask_back","question":"您想画什么颜色的猫？放在什么位置？大概多大？"}]
`

// Client OpenAI-compatible LLM 客户端
type Client struct {
	c     *openai.Client
	model string
}

func New(baseURL, apiKey, modelName string) *Client {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return &Client{c: openai.NewClientWithConfig(cfg), model: modelName}
}

// Understand 调用 LLM 解析自然语言为结构化 Intent。
// shapesCtx 用于指代消解时让 LLM 知道当前画布有什么。
func (l *Client) Understand(ctx context.Context, text string, shapesCtx []model.Shape) (model.UnderstandResponse, error) {
	if l == nil || l.c == nil {
		return model.UnderstandResponse{
			Intents: []model.Intent{{Cmd: model.CmdAskBack, Question: "LLM 未配置", Raw: text}},
		}, nil
	}

	ctxJSON, _ := json.Marshal(shapesCtx)
	userMsg := fmt.Sprintf("当前画布上有 %d 个图元: %s\n\n用户指令: %s", len(shapesCtx), string(ctxJSON), text)

	resp, err := l.c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: l.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMsg},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
		Temperature:    0.3,
	})
	if err != nil {
		return model.UnderstandResponse{}, err
	}
	if len(resp.Choices) == 0 {
		return model.UnderstandResponse{}, fmt.Errorf("empty LLM response")
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	intents, err := parseIntents(raw)
	if err != nil {
		// 解析失败, 反问兜底
		return model.UnderstandResponse{
			Intents: []model.Intent{{Cmd: model.CmdAskBack, Question: "我没理解，请换个说法", Raw: text}},
		}, nil
	}

	// 给 scene 里的图元补 ID, 给 shape 补 ID
	for i := range intents {
		intents[i].Raw = text
		if intents[i].Cmd == model.CmdScene {
			for j := range intents[i].Scenes {
				if intents[i].Scenes[j].ID == "" {
					intents[i].Scenes[j].ID = fmt.Sprintf("s_%d_%d", timeNow(), j)
				}
			}
		}
		if intents[i].Shape != nil && intents[i].Shape.ID == "" {
			intents[i].Shape.ID = fmt.Sprintf("s_%d", timeNow())
		}
	}

	return model.UnderstandResponse{Intents: intents, Reply: "已生成"}, nil
}

// parseIntents 解析 LLM 返回的 JSON。支持两种格式:
//   - 直接数组: [...]
//   - 对象包装: {"intents": [...]}
func parseIntents(raw string) ([]model.Intent, error) {
	var arr []model.Intent
	if err := json.Unmarshal([]byte(raw), &arr); err == nil {
		return arr, nil
	}
	var wrapped struct {
		Intents []model.Intent `json:"intents"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapped); err == nil {
		return wrapped.Intents, nil
	}
	return nil, fmt.Errorf("invalid JSON shape: %s", raw)
}

// 单独抽出来避免循环依赖
var timeNow = func() int64 { return 0 } // 测试时可以替换
