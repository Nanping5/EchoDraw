// Package engine 规则引擎: 处理 80% 的结构化语音指令 (PRD Class 1-3)。
//
// 设计原则:
//   - 优先规则命中, 未命中返回 Hit=false 给调用方决定是否走 LLM
//   - 4 类 Action 判定 (delta/redraw/modify/clear) 对应 PRD 附录 A.7 伪代码
//   - 纯函数为主, 少 closure 少 map[string]func, 易测试
package engine

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/echodraw/server/internal/model"
	"github.com/google/uuid"
)

// 默认尺寸常量
const (
	sizeSmall = 40
	sizeMid   = 80
	sizeBig   = 140
)

// Engine 规则引擎
type Engine struct {
	canvasW, canvasH int
}

func New(canvasW, canvasH int) *Engine {
	return &Engine{canvasW: canvasW, canvasH: canvasH}
}

// Result 理解结果
type Result struct {
	Intents []model.Intent
	Hit     bool   // 规则是否命中, 未命中交给 LLM
	Reply   string // 系统回复文本
}

// Understand 入口: 把自然语言解析为 Intent 列表。
// ctx 是当前画布上的图元, 用于指代消解。
func (e *Engine) Understand(text string, ctx []model.Shape) Result {
	t := strings.TrimSpace(text)
	if t == "" {
		return Result{Hit: true, Reply: ""}
	}

	// 1) 系统级命令
	if cmd, ok := matchSystem(t); ok {
		return Result{
			Hit:     true,
			Intents: []model.Intent{{Cmd: cmd, Action: model.ActionClear, Raw: t, Reply: replyFor(cmd)}},
			Reply:   replyFor(cmd),
		}
	}

	// 2) 删除
	if intent, ok := e.matchDelete(t, ctx); ok {
		return Result{Hit: true, Intents: []model.Intent{intent}, Reply: "已删除"}
	}

	// 3) 选中
	if intent, ok := e.matchSelect(t, ctx); ok {
		return Result{Hit: true, Intents: []model.Intent{intent}, Reply: "已选中"}
	}

	// 4) 变换/重画/修改
	if intent, ok := e.matchTransform(t, ctx); ok {
		return Result{Hit: true, Intents: []model.Intent{intent}, Reply: "已修改"}
	}

	// 5) 创建
	if intent, ok := e.matchCreate(t); ok {
		return Result{Hit: true, Intents: []model.Intent{intent}, Reply: "已创建"}
	}

	// 6) 未命中
	return Result{Hit: false, Intents: []model.Intent{{Cmd: model.CmdUnknown, Raw: t}}}
}

// ----------------- 系统级 -----------------

func matchSystem(t string) (model.CommandType, bool) {
	var bestHit model.CommandType
	bestLen := 0
	for kw, cmd := range model.SystemCommands {
		if strings.Contains(t, kw) && len(kw) > bestLen {
			bestHit = cmd
			bestLen = len(kw)
		}
	}
	return bestHit, bestHit != ""
}

func replyFor(cmd model.CommandType) string {
	switch cmd {
	case model.CmdUndo:
		return "已撤销"
	case model.CmdRedo:
		return "已重做"
	case model.CmdClear:
		return "已清空画布"
	case model.CmdExport:
		return "正在保存"
	}
	return "好的"
}

// ----------------- 创建 -----------------

func (e *Engine) matchCreate(t string) (model.Intent, bool) {
	triggered := containsAny(t, model.CreateTriggers) || containsAny(t, model.DeltaTriggers)
	if !triggered {
		return model.Intent{}, false
	}

	if isTextIntent(t) {
		return e.createText(t), true
	}

	shape, baseSize, ok := pickShape(t)
	if !ok {
		return model.Intent{}, false
	}

	size := baseSize
	if hasAny(t, []string{"大", "巨大", "超大", "很大"}) {
		size = sizeBig
	}
	if hasAny(t, []string{"小", "很小", "迷你", "tiny", "小的"}) {
		size = sizeSmall
	}
	if n, ok := pickNumber(t, []string{"半径", "直径", "size"}); ok {
		size = n
	}
	if n, ok := pickScale(t); ok {
		size = baseSize * n
	}

	fill := "#90a4ae"
	if c, ok := pickColor(t); ok {
		fill = c
	}

	x, y := e.defaultPos()
	if pos, ok := pickPosition(t); ok {
		x = float64(e.canvasW) * pos.X
		y = float64(e.canvasH) * pos.Y
	}
	if ax, ay, ok := pickCoord(t); ok {
		x, y = ax, ay
	}

	shape.X = x
	shape.Y = y
	shape.ID = uuid.NewString()
	shape.Style = model.Style{Fill: fill, Stroke: "#37474f", StrokeW: 2}

	switch shape.Type {
	case model.ShapeCircle:
		shape.Radius = size
	case model.ShapeRect, model.ShapeEllipse, model.ShapeTriangle, model.ShapeStar:
		shape.Width = size
		shape.Height = size
	case model.ShapeLine, model.ShapeArrow:
		shape.X = x - size/2
		shape.Points = []float64{0, 0, size, 0}
	}

	return model.Intent{
		Cmd:    model.CmdCreate,
		Action: model.ActionDelta,
		Shape:  &shape,
		Raw:    t,
		Reply:  "已创建",
	}, true
}

func (e *Engine) createText(t string) model.Intent {
	x, y := e.defaultPos()
	if pos, ok := pickPosition(t); ok {
		x = float64(e.canvasW) * pos.X
		y = float64(e.canvasH) * pos.Y
	}
	fill := "#212121"
	if c, ok := pickColor(t); ok {
		fill = c
	}
	text := "文本"
	if m := textQuoteRe.FindStringSubmatch(t); m != nil {
		text = m[1]
	}
	return model.Intent{
		Cmd:    model.CmdCreate,
		Action: model.ActionDelta,
		Shape: &model.Shape{
			ID:    uuid.NewString(),
			Type:  model.ShapeText,
			X:     x, Y: y,
			Width: 200, Height: 40,
			Text:  text,
			Style: model.Style{Fill: fill, FontSize: 24},
		},
		Raw:   t,
		Reply: "已创建",
	}
}

var textQuoteRe = regexp.MustCompile(`[""''「『"'](.+?)[""''」』"']`)

func isTextIntent(t string) bool {
	return containsAny(t, []string{"文字", "文本", "字"})
}

func pickShape(t string) (model.Shape, float64, bool) {
	var best model.Shape
	var bestSize float64
	bestLen := 0
	for kw, info := range model.ShapeWords {
		if strings.Contains(t, kw) && len(kw) > bestLen {
			best = model.Shape{Type: info.Type}
			bestSize = info.Size
			bestLen = len(kw)
		}
	}
	return best, bestSize, bestLen > 0
}

func pickColor(t string) (string, bool) {
	if m := hexRe.FindString(t); m != "" {
		return m, true
	}
	low := strings.ToLower(t)
	for k, v := range model.ColorHex {
		if strings.Contains(t, k) || strings.Contains(low, k) {
			return v, true
		}
	}
	return "", false
}

var hexRe = regexp.MustCompile(`#[0-9a-fA-F]{6}`)

func pickPosition(t string) (struct{ X, Y float64 }, bool) {
	for k, v := range model.Position {
		if strings.Contains(t, k) {
			return v, true
		}
	}
	return struct{ X, Y float64 }{}, false
}

var coordRe = regexp.MustCompile(`\(\s*(\d+)\s*[,，]\s*(\d+)\s*\)`)

func pickCoord(t string) (float64, float64, bool) {
	m := coordRe.FindStringSubmatch(t)
	if m == nil {
		return 0, 0, false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	if a <= 0 || b <= 0 {
		return 0, 0, false
	}
	return float64(a), float64(b), true
}

func pickNumber(t string, prefixes []string) (float64, bool) {
	for _, p := range prefixes {
		re := regexp.MustCompile(p + `\s*(\d+(?:\.\d+)?)`)
		if m := re.FindStringSubmatch(t); m != nil {
			if n, err := strconv.ParseFloat(m[1], 64); err == nil && n > 0 {
				return n, true
			}
		}
	}
	re := regexp.MustCompile(`(\d+)\s*(?:像素|px|个像素)`)
	if m := re.FindStringSubmatch(t); m != nil {
		if n, err := strconv.ParseFloat(m[1], 64); err == nil && n > 0 {
			return n, true
		}
	}
	return 0, false
}

// 中文数字 → 阿拉伯数字
var cnNum = map[string]float64{
	"一": 1, "二": 2, "两": 2, "三": 3, "四": 4, "五": 5,
	"六": 6, "七": 7, "八": 8, "九": 9, "十": 10, "半": 0.5,
}

func pickScale(t string) (float64, bool) {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?|[一二两三四五六七八九十半])\s*倍`)
	if m := re.FindStringSubmatch(t); m != nil {
		if n, err := strconv.ParseFloat(m[1], 64); err == nil && n > 0 {
			return n, true
		}
		if n, ok := cnNum[m[1]]; ok {
			return n, true
		}
	}
	return 0, false
}

// ----------------- 变换 / 重画 / 修改 -----------------

func (e *Engine) matchTransform(t string, ctx []model.Shape) (model.Intent, bool) {
	if !hasTarget(t) {
		return model.Intent{}, false
	}
	target := resolveTarget(t, ctx)
	if target == nil {
		return model.Intent{}, false
	}

	// 1) 改色
	if hasAny(t, []string{"改色", "变色", "改颜色"}) {
		if c, ok := pickColor(t); ok {
			return model.Intent{
				Cmd:    model.CmdUpdate,
				Action: model.ActionModify,
				Target: target,
				Patch:  &model.Style{Fill: c},
				Raw:    t,
				Reply:  "已修改",
			}, true
		}
	}
	// 隐式改色: "把它改成红色" / "把那个圆换成蓝色"
	if hasAny(t, []string{"变成", "改为", "换成", "改成"}) {
		if c, ok := pickColor(t); ok {
			return model.Intent{
				Cmd:    model.CmdUpdate,
				Action: model.ActionModify,
				Target: target,
				Patch:  &model.Style{Fill: c},
				Raw:    t,
				Reply:  "已修改",
			}, true
		}
	}

	// 2) 移动
	if hasAny(t, []string{"移到", "移动到", "放到", "挪到"}) {
		if pos, ok := pickPosition(t); ok {
			return model.Intent{
				Cmd:    model.CmdUpdate,
				Action: model.ActionModify,
				Target: target,
				MoveTo: &struct {
					X float64 `json:"x"`
					Y float64 `json:"y"`
				}{X: float64(e.canvasW) * pos.X, Y: float64(e.canvasH) * pos.Y},
				Raw:   t,
				Reply: "已移动",
			}, true
		}
	}

	// 3) 倍数缩放
	if n, ok := pickScale(t); ok {
		return model.Intent{
			Cmd:    model.CmdUpdate,
			Action: model.ActionModify,
			Target: target,
			Scale:  n,
			Raw:    t,
			Reply:  "已缩放",
		}, true
	}
	// 4) 修饰词缩放
	if hasAny(t, []string{"放大", "变大", "大一点"}) {
		return model.Intent{
			Cmd:    model.CmdUpdate,
			Action: model.ActionModify,
			Target: target,
			Scale:  1.5,
			Raw:    t,
			Reply:  "已放大",
		}, true
	}
	if hasAny(t, []string{"缩小", "变小", "小一点"}) {
		return model.Intent{
			Cmd:    model.CmdUpdate,
			Action: model.ActionModify,
			Target: target,
			Scale:  0.7,
			Raw:    t,
			Reply:  "已缩小",
		}, true
	}

	// 5) 旋转
	if m := rotateRe.FindStringSubmatch(t); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return model.Intent{
				Cmd:      model.CmdUpdate,
				Action:   model.ActionModify,
				Target:   target,
				Rotation: float64(n),
				Raw:      t,
				Reply:    "已旋转",
			}, true
		}
	}

	return model.Intent{}, false
}

var rotateRe = regexp.MustCompile(`旋转\s*(\d+)\s*度`)

// ----------------- 删除 -----------------

func (e *Engine) matchDelete(t string, ctx []model.Shape) (model.Intent, bool) {
	if !hasAny(t, []string{"删除", "去掉", "移除", "擦掉"}) {
		return model.Intent{}, false
	}
	target := resolveTarget(t, ctx)
	if target == nil {
		return model.Intent{}, false
	}
	return model.Intent{
		Cmd:    model.CmdDelete,
		Target: target,
		Raw:    t,
		Reply:  "已删除",
	}, true
}

// ----------------- 选中 -----------------

func (e *Engine) matchSelect(t string, ctx []model.Shape) (model.Intent, bool) {
	if !hasAny(t, []string{"选中", "选择", "选"}) {
		return model.Intent{}, false
	}
	target := resolveTarget(t, ctx)
	if target == nil {
		return model.Intent{}, false
	}
	return model.Intent{
		Cmd:    model.CmdSelect,
		Target: target,
		Raw:    t,
		Reply:  "已选中",
	}, true
}

// ----------------- 指代消解 -----------------

func hasTarget(t string) bool {
	for k := range model.RefWords {
		if strings.Contains(t, k) {
			return true
		}
	}
	return false
}

func resolveTarget(t string, ctx []model.Shape) *model.Selection {
	if len(ctx) == 0 {
		return nil
	}
	sel := &model.Selection{}

	for k, v := range model.RefWords {
		if strings.Contains(t, k) {
			sel.Ref = v
			break
		}
	}

	// 仅当文本里没有"改色动词"时, 才把颜色词当作 filter (避免"改成蓝色"误把蓝色当 filter)
	if !hasAny(t, []string{"改成", "变成", "改为", "换成", "改色", "变色", "改颜色"}) {
		for k := range model.ColorHex {
			if strings.Contains(t, k) {
				sel.Filter = k
				break
			}
		}
	}
	for k := range model.ShapeWords {
		if strings.Contains(t, k) {
			if sel.Filter != "" {
				sel.Filter += "+"
			}
			sel.Filter += k
			break
		}
	}

	// filter 存在时, ref 退化为"all" 让 filter 单独决定 (否则"那个红色的"会被 last 误判)
	if sel.Ref == "last" && sel.Filter != "" {
		sel.Ref = "all"
	}
	sel.IDs = filterIDs(ctx, sel.Ref, sel.Filter)

	// 无 filter 时退到 ctx 最后一项
	if len(sel.IDs) == 0 && sel.Filter == "" && (sel.Ref == "last" || sel.Ref == "") && len(ctx) > 0 {
		sel.IDs = []string{ctx[len(ctx)-1].ID}
	}
	if len(sel.IDs) == 0 {
		return nil
	}
	return sel
}

func filterIDs(ctx []model.Shape, ref, filter string) []string {
	out := []string{}
	for i, s := range ctx {
		match := true
		switch ref {
		case "last":
			if i != len(ctx)-1 {
				match = false
			}
		case "first":
			if i != 0 {
				match = false
			}
		}
		if match && filter != "" {
			if !matchFilter(s, filter) {
				match = false
			}
		}
		if match {
			out = append(out, s.ID)
		}
	}
	return out
}

func matchFilter(s model.Shape, filter string) bool {
	for _, p := range strings.Split(filter, "+") {
		if p == "" {
			continue
		}
		if hex, ok := model.ColorHex[p]; ok && strings.EqualFold(s.Style.Fill, hex) {
			return true
		}
		if info, ok := model.ShapeWords[p]; ok && s.Type == info.Type {
			return true
		}
	}
	return false
}

// ----------------- 工具 -----------------

func (e *Engine) defaultPos() (float64, float64) {
	return float64(e.canvasW) / 2, float64(e.canvasH) / 2
}

func containsAny(s string, keys []string) bool {
	return hasAny(s, keys)
}

func hasAny(s string, keys []string) bool {
	for _, k := range keys {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}
