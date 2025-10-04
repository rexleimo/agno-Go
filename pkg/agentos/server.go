package agentos

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rexleimo/agno-go/pkg/agno/agent"
	"github.com/rexleimo/agno-go/pkg/agno/session"
)

// Server represents the AgentOS HTTP server
type Server struct {
	router         *gin.Engine
	config         *Config
	sessionStorage session.Storage
	agentRegistry  *AgentRegistry
	logger         *slog.Logger
	httpServer     *http.Server
}

// Config holds server configuration
// Config 持有服务器配置
type Config struct {
	// Server address (default: :8080)
	// 服务器地址 (默认: :8080)
	Address string

	// API route prefix (e.g., "/api/v1", "/chat", empty for no prefix)
	// API 路由前缀 (例如: "/api/v1", "/chat", 空字符串表示无前缀)
	Prefix string

	// Session storage
	// Session 存储
	SessionStorage session.Storage

	// Logger
	// 日志记录器
	Logger *slog.Logger

	// Enable debug mode
	// 启用调试模式
	Debug bool

	// CORS settings
	// CORS 设置
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string

	// Request timeout
	// 请求超时时间
	RequestTimeout time.Duration

	// Max request size (in bytes)
	// 最大请求大小 (字节)
	MaxRequestSize int64
}

// NewServer creates a new AgentOS server
func NewServer(config *Config) (*Server, error) {
	if config == nil {
		config = &Config{}
	}

	// Set defaults
	if config.Address == "" {
		config.Address = ":8080"
	}

	if config.SessionStorage == nil {
		config.SessionStorage = session.NewMemoryStorage()
	}

	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}

	if config.MaxRequestSize == 0 {
		config.MaxRequestSize = 10 * 1024 * 1024 // 10MB
	}

	// Set Gin mode
	if config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(loggerMiddleware(config.Logger))
	router.Use(corsMiddleware(config))
	router.Use(timeoutMiddleware(config.RequestTimeout))

	server := &Server{
		router:         router,
		config:         config,
		sessionStorage: config.SessionStorage,
		agentRegistry:  NewAgentRegistry(),
		logger:         config.Logger,
	}

	// Register routes
	server.registerRoutes()

	return server, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("starting AgentOS server", "address", s.config.Address)

	s.httpServer = &http.Server{
		Addr:         s.config.Address,
		Handler:      s.router,
		ReadTimeout:  s.config.RequestTimeout,
		WriteTimeout: s.config.RequestTimeout,
	}

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down AgentOS server")

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}

	// Close storage
	if s.sessionStorage != nil {
		if err := s.sessionStorage.Close(); err != nil {
			s.logger.Warn("failed to close session storage", "error", err)
		}
	}

	return nil
}

// RegisterAgent registers an agent with the server
func (s *Server) RegisterAgent(agentID string, ag *agent.Agent) error {
	return s.agentRegistry.Register(agentID, ag)
}

// GetAgentRegistry returns the agent registry
func (s *Server) GetAgentRegistry() *AgentRegistry {
	return s.agentRegistry
}

// registerRoutes registers all API routes
// registerRoutes 注册所有 API 路由
func (s *Server) registerRoutes() {
	// Create base group with prefix (if specified)
	// 使用前缀创建基础路由组 (如果指定了前缀)
	var baseGroup *gin.RouterGroup
	if s.config.Prefix != "" {
		baseGroup = s.router.Group(s.config.Prefix)
	} else {
		baseGroup = &s.router.RouterGroup
	}

	// Health check (always at root level)
	// 健康检查 (始终在根级别)
	s.router.GET("/health", s.handleHealth)

	// API v1 under the prefix
	// 前缀下的 API v1
	v1 := baseGroup.Group("/api/v1")
	{
		// Session endpoints
		// Session 端点
		sessions := v1.Group("/sessions")
		{
			sessions.POST("", s.handleCreateSession)
			sessions.GET("/:id", s.handleGetSession)
			sessions.PUT("/:id", s.handleUpdateSession)
			sessions.DELETE("/:id", s.handleDeleteSession)
			sessions.GET("", s.handleListSessions)
		}

		// Agent endpoints
		// Agent 端点
		agents := v1.Group("/agents")
		{
			agents.GET("", s.handleListAgents)
			agents.POST("/:id/run", s.handleAgentRun)
		}
	}
}

// Health check handler
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "agentos",
		"time":    time.Now().Unix(),
	})
}

// Middleware: Logger
func loggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		logger.Info("request",
			"method", method,
			"path", path,
			"status", status,
			"duration", duration.String(),
			"ip", c.ClientIP(),
		)
	}
}

// Middleware: CORS
func corsMiddleware(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(config.AllowOrigins) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Origin", config.AllowOrigins[0])
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		if len(config.AllowMethods) > 0 {
			methods := ""
			for i, m := range config.AllowMethods {
				if i > 0 {
					methods += ", "
				}
				methods += m
			}
			c.Writer.Header().Set("Access-Control-Allow-Methods", methods)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}

		if len(config.AllowHeaders) > 0 {
			headers := ""
			for i, h := range config.AllowHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += h
			}
			c.Writer.Header().Set("Access-Control-Allow-Headers", headers)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Middleware: Timeout
func timeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
