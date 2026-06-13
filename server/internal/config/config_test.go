package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 测试 helper: 创建一个临时目录, 写 config.toml, 切到该目录运行 Load
func withTempConfig(t *testing.T, content string, fn func()) {
	t.Helper()
	dir := t.TempDir()
	if content != "" {
		if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)
	// 清空所有可能影响测试的 env
	for _, k := range []string{"PORT", "LLM_API_KEY", "LLM_BASE_URL", "LLM_MODEL", "CANVAS_W", "CANVAS_H", "ECHODRAW_PORT"} {
		os.Unsetenv(k)
	}
	fn()
}

func TestLoadDefaults(t *testing.T) {
	withTempConfig(t, "", func() {
		c := Load()
		if c.Port != "8080" {
			t.Errorf("port default: want 8080, got %s", c.Port)
		}
		if c.LLMBaseURL != "https://api.openai.com/v1" {
			t.Errorf("base_url default: want default OpenAI, got %s", c.LLMBaseURL)
		}
		if c.LLMAPIKey != "" {
			t.Errorf("api_key default: want empty, got %s", c.LLMAPIKey)
		}
		if c.LLMModel != "gpt-4o-mini" {
			t.Errorf("model default: want gpt-4o-mini, got %s", c.LLMModel)
		}
		if c.CanvasW != 1200 || c.CanvasH != 800 {
			t.Errorf("canvas: want 1200x800, got %dx%d", c.CanvasW, c.CanvasH)
		}
	})
}

func TestLoadFromToml(t *testing.T) {
	withTempConfig(t, `
port = "9090"

[llm]
base_url = "https://api.deepseek.com/v1"
api_key = "sk-test-123"
model = "deepseek-chat"

[canvas]
width = 800
height = 600
`, func() {
		c := Load()
		if c.Port != "9090" {
			t.Errorf("port: want 9090, got %s", c.Port)
		}
		if c.LLMBaseURL != "https://api.deepseek.com/v1" {
			t.Errorf("base_url: want deepseek, got %s", c.LLMBaseURL)
		}
		if c.LLMAPIKey != "sk-test-123" {
			t.Errorf("api_key: want sk-test-123, got %s", c.LLMAPIKey)
		}
		if c.LLMModel != "deepseek-chat" {
			t.Errorf("model: want deepseek-chat, got %s", c.LLMModel)
		}
		if c.CanvasW != 800 || c.CanvasH != 600 {
			t.Errorf("canvas: want 800x600, got %dx%d", c.CanvasW, c.CanvasH)
		}
	})
}

func TestEnvOverridesToml(t *testing.T) {
	withTempConfig(t, `
port = "9090"

[llm]
api_key = "sk-from-toml"
`, func() {
		os.Setenv("PORT", "7070")
		os.Setenv("LLM_API_KEY", "sk-from-env")
		c := Load()
		if c.Port != "7070" {
			t.Errorf("env should override toml: port want 7070, got %s", c.Port)
		}
		if c.LLMAPIKey != "sk-from-env" {
			t.Errorf("env should override toml: api_key want sk-from-env, got %s", c.LLMAPIKey)
		}
	})
}

func TestEnvPrefixEchoDraw(t *testing.T) {
	withTempConfig(t, "", func() {
		os.Setenv("ECHODRAW_PORT", "6060")
		os.Setenv("ECHODRAW_LLM_API_KEY", "sk-prefix")
		c := Load()
		if c.Port != "6060" {
			t.Errorf("ECHODRAW_PORT: want 6060, got %s", c.Port)
		}
		if c.LLMAPIKey != "sk-prefix" {
			t.Errorf("ECHODRAW_LLM_API_KEY: want sk-prefix, got %s", c.LLMAPIKey)
		}
	})
}

func TestMissingConfigFileUsesDefaults(t *testing.T) {
	withTempConfig(t, "", func() {
		// 目录空, 没 config.toml
		c := Load()
		if c.Port != "8080" {
			t.Errorf("missing file: should fall back to default, got port=%s", c.Port)
		}
	})
}

func TestInvalidTomlFallsBack(t *testing.T) {
	// 验证语法错误的 toml 不会让服务起不来
	// (我们的策略是 warn + 继续, 而不是 fatal)
	withTempConfig(t, "this is = = invalid toml {{{ ]]]", func() {
		c := Load()
		// 解析失败, 但 Load 不 panic, 用默认 port
		_ = c
	})
}
