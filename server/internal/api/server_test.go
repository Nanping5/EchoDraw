package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/echodraw/server/internal/engine"
	"github.com/echodraw/server/internal/model"
)

func newTestServer() *Server {
	return New(engine.New(1200, 800), nil) // 无 LLM
}

func postJSON(t *testing.T, s *Server, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := s.Routes()
	req := httptest.NewRequest("POST", path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealth(t *testing.T) {
	s := newTestServer()
	r := s.Routes()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["ok"] != true {
		t.Errorf("want ok=true")
	}
	if body["llm_enabled"] != false {
		t.Errorf("want llm_enabled=false")
	}
}

func TestUnderstandHitRule(t *testing.T) {
	s := newTestServer()
	w := postJSON(t, s, "/api/understand", `{"text":"画一个红色的大圆","context":[]}`)
	if w.Code != 200 {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.UnderstandResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Intents) == 0 || resp.Intents[0].Cmd != model.CmdCreate {
		t.Errorf("want create intent, got %+v", resp)
	}
}

func TestUnderstandMissAskBack(t *testing.T) {
	s := newTestServer()
	w := postJSON(t, s, "/api/understand", `{"text":"画一只猫","context":[]}`)
	if w.Code != 200 {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp model.UnderstandResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Intents[0].Cmd != model.CmdAskBack {
		t.Errorf("want ask_back when no LLM, got %s", resp.Intents[0].Cmd)
	}
	if resp.Intents[0].Question == "" {
		t.Errorf("question should not be empty")
	}
}

func TestUnderstandWithContext(t *testing.T) {
	s := newTestServer()
	ctxJSON := `[{"id":"a","type":"circle","x":100,"y":100,"radius":50,"style":{"fill":"#e53935"}}]`
	w := postJSON(t, s, "/api/understand", `{"text":"把它改成蓝色","context":`+ctxJSON+`}`)
	if w.Code != 200 {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.UnderstandResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Intents[0].Cmd != model.CmdUpdate {
		t.Errorf("want update, got %s", resp.Intents[0].Cmd)
	}
}

func TestUnderstandBadJSON(t *testing.T) {
	s := newTestServer()
	w := postJSON(t, s, "/api/understand", `not json`)
	if w.Code != 400 {
		t.Errorf("want 400, got %d", w.Code)
	}
}
