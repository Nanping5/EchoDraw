package main

import (
	"log"
	"time"

	"github.com/echodraw/server/internal/api"
	"github.com/echodraw/server/internal/config"
	"github.com/echodraw/server/internal/engine"
	"github.com/echodraw/server/internal/llm"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// 结构化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	gin.SetMode(gin.ReleaseMode)

	// 依赖装配
	eng := engine.New(cfg.CanvasW, cfg.CanvasH)
	var llmClient *llm.Client
	if cfg.LLMAPIKey != "" {
		llmClient = llm.New(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
		logger.Info("LLM enabled",
			zap.String("model", cfg.LLMModel),
			zap.String("base_url", cfg.LLMBaseURL),
		)
	} else {
		logger.Warn("LLM disabled (set LLM_API_KEY to enable)")
	}

	// 路由
	srv := api.New(eng, llmClient)
	r := srv.Routes()

	logger.Info("EchoDraw server starting",
		zap.String("port", cfg.Port),
		zap.Int("canvas_w", cfg.CanvasW),
		zap.Int("canvas_h", cfg.CanvasH),
	)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
	_ = time.Second
}
