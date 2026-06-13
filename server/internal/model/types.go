// Package model 定义 EchoDraw 前后端共享的领域类型。
//
// 这些类型是 voice → intent → render 流水线的契约:
// 1. 前端从语音识别拿到 text
// 2. 后端 /api/understand 把 text 解析为 []Intent
// 3. 前端把每个 Intent 翻译为画布操作
package model

// ShapeType 图元类型 (前端 Konva 节点类型, 保持一致)
type ShapeType string

const (
	ShapeCircle   ShapeType = "circle"
	ShapeRect     ShapeType = "rect"
	ShapeLine     ShapeType = "line"
	ShapeEllipse  ShapeType = "ellipse"
	ShapeTriangle ShapeType = "triangle"
	ShapeStar     ShapeType = "star"
	ShapeText     ShapeType = "text"
	ShapeArrow    ShapeType = "arrow"
)

// Style 图元样式
type Style struct {
	Fill       string  `json:"fill,omitempty"`
	Stroke     string  `json:"stroke,omitempty"`
	StrokeW    int     `json:"strokeWidth,omitempty"`
	FontSize   int     `json:"fontSize,omitempty"`
	Opacity    float64 `json:"opacity,omitempty"`
	Rotation   float64 `json:"rotation,omitempty"`
}

// Shape 画布上的图元 (前后端共享)
// X/Y 是中心坐标 (前端 Konva 用法); 线段/箭头用 Points 存相对偏移。
type Shape struct {
	ID     string    `json:"id"`
	Type   ShapeType `json:"type"`
	X      float64   `json:"x"`
	Y      float64   `json:"y"`
	Width  float64   `json:"width,omitempty"`
	Height float64   `json:"height,omitempty"`
	Radius float64   `json:"radius,omitempty"`
	Points []float64 `json:"points,omitempty"` // line/arrow 相对中心点偏移
	Text   string    `json:"text,omitempty"`
	Style  Style     `json:"style"`
}

// CommandType 操作类型
type CommandType string

const (
	// Create 创建单个图元
	CmdCreate CommandType = "create"
	// Update 就地修改属性 (颜色/位置/大小/旋转)
	CmdUpdate CommandType = "update"
	// Delete 删除
	CmdDelete CommandType = "delete"
	// Select 选中 (用于前端高亮)
	CmdSelect CommandType = "select"
	// Undo 撤销
	CmdUndo CommandType = "undo"
	// Redo 重做
	CmdRedo CommandType = "redo"
	// Clear 清空画布
	CmdClear CommandType = "clear"
	// Export 保存/导出
	CmdExport CommandType = "export"
	// AskBack 信息不足, 反问用户
	CmdAskBack CommandType = "ask_back"
	// Scene AI 生成的复合场景 (含多个图元)
	CmdScene CommandType = "scene"
)

// Action 增量/重画/修改/清空 的分类 (PRD 附录 A.1)
type Action string

const (
	ActionDelta   Action = "delta"   // 增量 (追加图元)
	ActionRedraw  Action = "redraw"  // 重画 (删旧 + 建新)
	ActionModify  Action = "modify"  // 就地修改
	ActionClear   Action = "clear"   // 清空
)

// Selection 目标选择 (指代消解结果)
type Selection struct {
	Ref    string   `json:"ref,omitempty"`    // "last" | "first" | "all"
	Filter string   `json:"filter,omitempty"` // "red" | "circle" | "red+circle"
	IDs    []string `json:"ids,omitempty"`    // 解析出的具体 ID
}

// Intent 一次意图理解的结果
type Intent struct {
	Cmd      CommandType `json:"cmd"`
	Action   Action      `json:"action,omitempty"`
	Shape    *Shape      `json:"shape,omitempty"`    // create 时用
	Scenes   []Shape     `json:"scenes,omitempty"`   // scene 时用
	Target   *Selection  `json:"target,omitempty"`   // update/delete/select 的目标
	Patch    *Style      `json:"patch,omitempty"`    // update 的属性补丁
	MoveTo   *struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"moveTo,omitempty"`                    // update 的位置移动
	Scale    float64     `json:"scale,omitempty"`   // update 的缩放倍数
	Rotation float64     `json:"rotation,omitempty"` // update 的旋转角度
	Question string      `json:"question,omitempty"` // ask_back 的反问文本
	Reply    string      `json:"reply,omitempty"`    // 系统的语音回复
	Raw      string      `json:"raw"`               // 原始文本
}

// UnderstandRequest 理解请求
type UnderstandRequest struct {
	Text    string  `json:"text"`
	Context []Shape `json:"context,omitempty"` // 当前画布上的图元, 用于指代消解
}

// UnderstandResponse 理解响应
type UnderstandResponse struct {
	Intents []Intent `json:"intents"`
	Reply   string   `json:"reply,omitempty"`
}
