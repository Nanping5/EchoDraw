package engine

import (
	"testing"

	"github.com/echodraw/server/internal/model"
)

// 测试辅助: 构造一个 2 元素的 ctx
func makeCtx() []model.Shape {
	return []model.Shape{
		{ID: "a", Type: model.ShapeCircle, X: 100, Y: 100, Radius: 50, Style: model.Style{Fill: "#e53935"}},
		{ID: "b", Type: model.ShapeRect, X: 500, Y: 400, Width: 80, Height: 80, Style: model.Style{Fill: "#1e88e5"}},
	}
}

func newEng() *Engine { return New(1200, 800) }

// ----------------- 系统级 -----------------

func TestSystemCommands(t *testing.T) {
	e := newEng()
	cases := []struct {
		text string
		cmd  model.CommandType
	}{
		{"撤销", model.CmdUndo},
		{"撤回", model.CmdUndo},
		{"重做", model.CmdRedo},
		{"清空", model.CmdClear},
		{"清空画布", model.CmdClear},
		{"保存", model.CmdExport},
		{"导出", model.CmdExport},
	}
	for _, c := range cases {
		r := e.Understand(c.text, nil)
		if !r.Hit {
			t.Errorf("expected hit for %q, got miss", c.text)
			continue
		}
		if len(r.Intents) != 1 || r.Intents[0].Cmd != c.cmd {
			t.Errorf("%q: expected cmd=%s, got %v", c.text, c.cmd, r.Intents)
		}
	}
}

// ----------------- 创建 -----------------

func TestCreateSimple(t *testing.T) {
	e := newEng()
	r := e.Understand("画一个圆", nil)
	if !r.Hit || r.Intents[0].Cmd != model.CmdCreate {
		t.Fatalf("expected create, got %v", r)
	}
	s := r.Intents[0].Shape
	if s.Type != model.ShapeCircle || s.Radius != 60 {
		t.Errorf("expected circle r=60, got type=%s r=%v", s.Type, s.Radius)
	}
}

func TestCreateColoredBig(t *testing.T) {
	e := newEng()
	r := e.Understand("画一个红色的大圆", nil)
	s := r.Intents[0].Shape
	if s.Style.Fill != "#e53935" {
		t.Errorf("fill: want #e53935, got %s", s.Style.Fill)
	}
	if s.Radius != 140 {
		t.Errorf("radius: want 140 (big), got %v", s.Radius)
	}
}

func TestCreateSmall(t *testing.T) {
	e := newEng()
	r := e.Understand("画一个小的圆", nil)
	if r.Intents[0].Shape.Radius != 40 {
		t.Errorf("radius: want 40 (small), got %v", r.Intents[0].Shape.Radius)
	}
}

func TestCreateWithRadius(t *testing.T) {
	e := newEng()
	r := e.Understand("画一个半径 100 的圆", nil)
	if r.Intents[0].Shape.Radius != 100 {
		t.Errorf("radius: want 100, got %v", r.Intents[0].Shape.Radius)
	}
}

func TestCreateWithPosition(t *testing.T) {
	e := newEng()
	r := e.Understand("画一个圆在左边", nil)
	s := r.Intents[0].Shape
	if s.X != 240 || s.Y != 400 { // 0.2 * 1200, 0.5 * 800
		t.Errorf("position: want (240,400), got (%v,%v)", s.X, s.Y)
	}
}

func TestCreateWithCoord(t *testing.T) {
	e := newEng()
	r := e.Understand("画一个圆在 (300, 400)", nil)
	s := r.Intents[0].Shape
	if s.X != 300 || s.Y != 400 {
		t.Errorf("coord: want (300,400), got (%v,%v)", s.X, s.Y)
	}
}

func TestCreateDeltaTrigger(t *testing.T) {
	e := newEng()
	r := e.Understand("再画一个圆", nil)
	if r.Intents[0].Action != model.ActionDelta {
		t.Errorf("action: want delta, got %s", r.Intents[0].Action)
	}
}

func TestCreateText(t *testing.T) {
	e := newEng()
	r := e.Understand("添加文字 \"你好\"", nil)
	if !r.Hit {
		t.Fatalf("expected hit for text")
	}
	if r.Intents[0].Shape.Type != model.ShapeText {
		t.Errorf("want text, got %s", r.Intents[0].Shape.Type)
	}
	if r.Intents[0].Shape.Text != "你好" {
		t.Errorf("text: want 你好, got %s", r.Intents[0].Shape.Text)
	}
}

// ----------------- 变换 / 修改 -----------------

func TestUpdateColor(t *testing.T) {
	e := newEng()
	r := e.Understand("把它改成蓝色", makeCtx())
	if !r.Hit {
		t.Fatal("expected hit")
	}
	it := r.Intents[0]
	if it.Cmd != model.CmdUpdate || it.Action != model.ActionModify {
		t.Errorf("want update/modify, got %s/%s", it.Cmd, it.Action)
	}
	if it.Patch == nil || it.Patch.Fill != "#1e88e5" {
		t.Errorf("patch: want fill=#1e88e5, got %+v", it.Patch)
	}
}

func TestUpdateMoveTo(t *testing.T) {
	e := newEng()
	r := e.Understand("把它移到中间", makeCtx())
	it := r.Intents[0]
	if it.MoveTo == nil || it.MoveTo.X != 600 || it.MoveTo.Y != 400 {
		t.Errorf("moveTo: want (600,400), got %+v", it.MoveTo)
	}
}

func TestUpdateScale(t *testing.T) {
	e := newEng()
	r := e.Understand("把它放大两倍", makeCtx())
	if r.Intents[0].Scale != 2 {
		t.Errorf("scale: want 2, got %v", r.Intents[0].Scale)
	}
}

func TestUpdateScaleModifier(t *testing.T) {
	e := newEng()
	r := e.Understand("把它放大", makeCtx())
	if r.Intents[0].Scale != 1.5 {
		t.Errorf("scale: want 1.5, got %v", r.Intents[0].Scale)
	}
}

func TestUpdateRotate(t *testing.T) {
	e := newEng()
	r := e.Understand("把它旋转 45 度", makeCtx())
	if r.Intents[0].Rotation != 45 {
		t.Errorf("rotation: want 45, got %v", r.Intents[0].Rotation)
	}
}

// ----------------- 删除 -----------------

func TestDelete(t *testing.T) {
	e := newEng()
	r := e.Understand("删除它", makeCtx())
	if !r.Hit || r.Intents[0].Cmd != model.CmdDelete {
		t.Errorf("want delete, got %v", r)
	}
	if len(r.Intents[0].Target.IDs) == 0 {
		t.Errorf("target.ids empty")
	}
}

func TestDeleteFilterByColor(t *testing.T) {
	e := newEng()
	r := e.Understand("把那个红色的删除", makeCtx())
	if r.Intents[0].Target.IDs[0] != "a" {
		t.Errorf("want target=a (red circle), got %v", r.Intents[0].Target.IDs)
	}
}

// ----------------- 选中 -----------------

func TestSelect(t *testing.T) {
	e := newEng()
	r := e.Understand("选中最新的", makeCtx())
	if r.Intents[0].Cmd != model.CmdSelect {
		t.Errorf("want select, got %s", r.Intents[0].Cmd)
	}
	if r.Intents[0].Target.IDs[0] != "b" {
		t.Errorf("want target=b (newest), got %v", r.Intents[0].Target.IDs)
	}
}

// ----------------- 未命中 -----------------

func TestMiss(t *testing.T) {
	e := newEng()
	r := e.Understand("画一只猫", nil)
	if r.Hit {
		t.Errorf("want miss for '画一只猫', got hit")
	}
	if r.Intents[0].Cmd != model.CmdUnknown {
		t.Errorf("want unknown, got %s", r.Intents[0].Cmd)
	}
}

func TestEmptyString(t *testing.T) {
	e := newEng()
	r := e.Understand("   ", nil)
	if !r.Hit {
		t.Errorf("empty string should hit (no-op), got miss")
	}
}

// ----------------- 指代消解 -----------------

func TestResolveLast(t *testing.T) {
	sel := resolveTarget("刚才那个", makeCtx())
	if sel == nil || sel.IDs[0] != "b" {
		t.Errorf("want last=b, got %+v", sel)
	}
}

func TestResolveFirst(t *testing.T) {
	sel := resolveTarget("首个", makeCtx())
	if sel == nil || sel.IDs[0] != "a" {
		t.Errorf("want first=a, got %+v", sel)
	}
}

func TestResolveAllByColor(t *testing.T) {
	ctx := []model.Shape{
		{ID: "a", Type: model.ShapeCircle, Style: model.Style{Fill: "#e53935"}},
		{ID: "b", Type: model.ShapeCircle, Style: model.Style{Fill: "#e53935"}},
		{ID: "c", Type: model.ShapeCircle, Style: model.Style{Fill: "#1e88e5"}},
	}
	sel := resolveTarget("所有红色的", ctx)
	if sel == nil || len(sel.IDs) != 2 {
		t.Errorf("want 2 red ids, got %+v", sel)
	}
}

func TestResolveByShape(t *testing.T) {
	sel := resolveTarget("那个矩形", makeCtx())
	if sel == nil || sel.IDs[0] != "b" {
		t.Errorf("want rect=b, got %+v", sel)
	}
}

// ----------------- 词表单测 -----------------

func TestPickColor(t *testing.T) {
	c, ok := pickColor("红色的")
	if !ok || c != "#e53935" {
		t.Errorf("红色 → #e53935, got %s ok=%v", c, ok)
	}
	c, ok = pickColor("#ff0000")
	if !ok || c != "#ff0000" {
		t.Errorf("hex → #ff0000, got %s ok=%v", c, ok)
	}
	_, ok = pickColor("随便")
	if ok {
		t.Errorf("'随便' should not match any color")
	}
}

func TestPickPosition(t *testing.T) {
	p, ok := pickPosition("中间")
	if !ok || p.X != 0.5 || p.Y != 0.5 {
		t.Errorf("中间 → {0.5, 0.5}, got %+v ok=%v", p, ok)
	}
}

func TestPickCoord(t *testing.T) {
	x, y, ok := pickCoord("在 (300, 400) 这里")
	if !ok || x != 300 || y != 400 {
		t.Errorf("coord → (300,400), got (%v,%v) ok=%v", x, y, ok)
	}
}

func TestPickScale(t *testing.T) {
	n, ok := pickScale("放大 3 倍")
	if !ok || n != 3 {
		t.Errorf("3 倍 → 3, got %v ok=%v", n, ok)
	}
}
