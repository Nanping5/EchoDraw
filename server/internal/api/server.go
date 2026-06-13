// Package api HTTP 路由: 把 voice text → engine/llm → Intent JSON 的管道串起来。
package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/echodraw/server/internal/engine"
	"github.com/echodraw/server/internal/llm"
	"github.com/echodraw/server/internal/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server 把 engine + llm 包装成 HTTP server
type Server struct {
	engine *engine.Engine
	llm    *llm.Client // 可能为 nil (LLM_API_KEY 未配置)
}

func New(e *engine.Engine, l *llm.Client) *Server {
	return &Server{engine: e, llm: l}
}

func (s *Server) Routes() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger())

	// CORS: 允许 Vite dev server (默认 5173) 和生产环境
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept"},
		MaxAge:          12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"ok":          true,
			"name":        "EchoDraw",
			"version":     "0.1.0",
			"llm_enabled": s.llm != nil,
			"time":        time.Now().UTC().Format(time.RFC3339),
		})
	})

	api := r.Group("/api")
	{
		api.POST("/understand", s.handleUnderstand)
	}

	return r
}

// handleUnderstand 核心端点: 接收 {text, context} → 返回 {intents, reply}
func (s *Server) handleUnderstand(c *gin.Context) {
	var req model.UnderstandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1) 规则引擎优先
	r := s.engine.Understand(req.Text, req.Context)
	if r.Hit {
		c.JSON(http.StatusOK, model.UnderstandResponse{
			Intents: r.Intents,
			Reply:   r.Reply,
		})
		return
	}

	// 2) 规则未命中, 走 LLM
	if s.llm == nil {
		c.JSON(http.StatusOK, model.UnderstandResponse{
			Intents: []model.Intent{{
				Cmd:      model.CmdAskBack,
				Question: "没听清，请换个说法 (LLM 未配置)",
				Raw:      req.Text,
			}},
			Reply: "没听清",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	llmResp, err := s.llm.Understand(ctx, req.Text, req.Context)
	if err != nil {
		log.Printf("llm error: %v", err)
		c.JSON(http.StatusOK, model.UnderstandResponse{
			Intents: []model.Intent{{
				Cmd:      model.CmdAskBack,
				Question: "LLM 调用失败: " + err.Error(),
				Raw:      req.Text,
			}},
			Reply: "LLM 错误",
		})
		return
	}
	c.JSON(http.StatusOK, llmResp)
}

// requestLogger 简单 access log
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s %d %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
