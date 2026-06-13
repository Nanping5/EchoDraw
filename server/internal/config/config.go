// Package config 运行时配置加载。
//
// 配置优先级 (高 → 低):
//   1. 环境变量 (PORT, LLM_API_KEY, ECHODRAW_*)  -- 容器/生产
//   2. config.local.toml                          -- 本地覆盖 (git ignore)
//   3. config.toml                                -- dev 默认 (git 跟踪)
//   4. 代码内置默认值                             -- 兜底
//
// 配置文件查找路径 (按顺序, 找到第一个存在):
//   - ./config.toml         (从 server/ 目录运行时)
//   - ./config.local.toml   (本地覆盖, 不进 git)
//
// 字段含义不变, 调用方 (main.go) 代码零改动。
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 运行时配置
type Config struct {
	Port       string
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
	CanvasW    int
	CanvasH    int
}

// Load 从 config.toml + config.local.toml + 环境变量加载配置
func Load() *Config {
	v := viper.New()
	setDefaults(v)

	readFile(v, "config")      // 读 config.toml
	readFile(v, "config.local") // 读 config.local.toml 覆盖

	// 环境变量 (最高优先级, 覆盖一切)
	// 兼容历史: 旧名 PORT/LLM_* (无前缀) + 新名 ECHODRAW_*
	_ = v.BindEnv("port", "PORT", "ECHODRAW_PORT")
	_ = v.BindEnv("llm.api_key", "LLM_API_KEY", "ECHODRAW_LLM_API_KEY")
	_ = v.BindEnv("llm.base_url", "LLM_BASE_URL", "ECHODRAW_LLM_BASE_URL")
	_ = v.BindEnv("llm.model", "LLM_MODEL", "ECHODRAW_LLM_MODEL")
	_ = v.BindEnv("canvas.width", "CANVAS_W", "ECHODRAW_CANVAS_W")
	_ = v.BindEnv("canvas.height", "CANVAS_H", "ECHODRAW_CANVAS_HEIGHT")
	v.AutomaticEnv()

	return &Config{
		Port:       v.GetString("port"),
		LLMBaseURL: v.GetString("llm.base_url"),
		LLMAPIKey:  v.GetString("llm.api_key"),
		LLMModel:   v.GetString("llm.model"),
		CanvasW:    v.GetInt("canvas.width"),
		CanvasH:    v.GetInt("canvas.height"),
	}
}

// readFile 读单个 toml 配置文件, 文件不存在静默跳过, 解析错误 warn。
// 每次调用都用独立的 viper 实例, 避免 state 污染。
func readFile(target *viper.Viper, name string) {
	tmp := viper.New()
	tmp.SetConfigName(name)
	tmp.SetConfigType("toml")
	tmp.AddConfigPath(".")
	if err := tmp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("warn: read %s.toml: %v\n", name, err)
		}
		return
	}
	// merge 到 target (target 的已有值优先级更高 → 后面 merge 的会覆盖)
	if err := target.MergeConfigMap(tmp.AllSettings()); err != nil {
		fmt.Printf("warn: merge %s.toml: %v\n", name, err)
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("port", "8080")
	v.SetDefault("llm.base_url", "https://api.openai.com/v1")
	v.SetDefault("llm.api_key", "")
	v.SetDefault("llm.model", "gpt-4o-mini")
	v.SetDefault("canvas.width", 1200)
	v.SetDefault("canvas.height", 800)
}
