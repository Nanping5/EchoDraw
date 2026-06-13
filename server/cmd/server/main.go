package main

import (
	"log"
	"net/http"
	"time"

	"github.com/echodraw/server/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// 初始化结构化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// /health 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ok":      true,
			"name":    "EchoDraw",
			"version": "0.1.0",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	logger.Info("EchoDraw server starting",
		zap.String("port", cfg.Port),
		zap.String("llm_model", cfg.LLMModel),
		zap.Bool("llm_enabled", cfg.LLMAPIKey != ""),
	)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
