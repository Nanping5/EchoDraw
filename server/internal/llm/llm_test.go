package llm

import (
	"strings"
	"testing"

	"github.com/echodraw/server/internal/model"
)

func TestParseIntentsArray(t *testing.T) {
	raw := `[{"cmd":"create","action":"delta","shape":{"id":"x","type":"circle","x":100,"y":100,"radius":50}}]`
	intents, err := parseIntents(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(intents) != 1 {
		t.Fatalf("want 1 intent, got %d", len(intents))
	}
	if intents[0].Cmd != model.CmdCreate {
		t.Errorf("want create, got %s", intents[0].Cmd)
	}
	if intents[0].Shape == nil || intents[0].Shape.Type != model.ShapeCircle {
		t.Errorf("shape wrong: %+v", intents[0].Shape)
	}
}

func TestParseIntentsWrapped(t *testing.T) {
	raw := `{"intents":[{"cmd":"ask_back","question":"什么颜色？"}]}`
	intents, err := parseIntents(raw)
	if err != nil {
		t.Fatal(err)
	}
	if intents[0].Cmd != model.CmdAskBack || intents[0].Question != "什么颜色？" {
		t.Errorf("got %+v", intents[0])
	}
}

func TestParseIntentsInvalid(t *testing.T) {
	_, err := parseIntents("not json")
	if err == nil {
		t.Errorf("want error for invalid JSON")
	}
}

func TestParseIntentsScene(t *testing.T) {
	raw := `[{"cmd":"scene","action":"delta","scenes":[
		{"id":"1","type":"rect","x":600,"y":400,"width":1200,"height":800,"style":{"fill":"#000"}},
		{"id":"2","type":"circle","x":900,"y":200,"radius":80,"style":{"fill":"#fff"}}
	]}]`
	intents, err := parseIntents(raw)
	if err != nil {
		t.Fatal(err)
	}
	if intents[0].Cmd != model.CmdScene {
		t.Errorf("want scene, got %s", intents[0].Cmd)
	}
	if len(intents[0].Scenes) != 2 {
		t.Errorf("want 2 scenes, got %d", len(intents[0].Scenes))
	}
}

func TestSystemPromptHasAllConstraints(t *testing.T) {
	// 防止 prompt 退化
	must := []string{"circle", "rect", "ellipse", "line", "triangle", "star", "text", "arrow", "create", "scene", "ask_back", "夜空", "海面", "花园", "咖啡馆", "房子", "#0d1b2a", "#fdd835", "50"}
	for _, k := range must {
		if !strings.Contains(systemPrompt, k) {
			t.Errorf("systemPrompt 缺关键内容: %s", k)
		}
	}
}

func TestClientNilSafe(t *testing.T) {
	var c *Client
	r, err := c.Understand(nil, "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Intents[0].Cmd != model.CmdAskBack {
		t.Errorf("nil client should ask back, got %s", r.Intents[0].Cmd)
	}
}
