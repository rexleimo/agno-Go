package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rexleimo/agno-go/pkg/agentos"
	"github.com/rexleimo/agno-go/pkg/agno/agent"
	"github.com/rexleimo/agno-go/pkg/agno/models/openai"
	"github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
	"github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
)

// This example demonstrates running multiple AgentOS instances with different route prefixes
// 此示例演示使用不同路由前缀运行多个 AgentOS 实例
//
// This allows you to:
// - Host multiple agent services on the same port
// - Organize different agent types under different paths
// - Implement multi-tenant agent systems
//
// 这允许你:
// - 在同一端口上托管多个 agent 服务
// - 在不同路径下组织不同类型的 agent
// - 实现多租户 agent 系统

func main() {
	fmt.Println("🚀 AgentOS Multi-Instance Demo")
	fmt.Println("================================")
	fmt.Println()

	// Get API key from environment
	// 从环境变量获取 API 密钥
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create different agents for different purposes
	// 为不同目的创建不同的 agent

	// 1. Math Agent (for calculation tasks)
	// 1. 数学 Agent (用于计算任务)
	mathModel, err := openai.New("gpt-4o-mini", openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create math model: %v", err)
	}

	mathAgent, err := agent.New(agent.Config{
		Name:         "Math Assistant",
		Model:        mathModel,
		Toolkits:     []toolkit.Toolkit{calculator.New()},
		Instructions: "You are a math assistant. Help users with calculations.",
	})
	if err != nil {
		log.Fatalf("Failed to create math agent: %v", err)
	}

	// 2. Chat Agent (for general conversation)
	// 2. 聊天 Agent (用于一般对话)
	chatModel, err := openai.New("gpt-4o-mini", openai.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	chatAgent, err := agent.New(agent.Config{
		Name:         "Chat Assistant",
		Model:        chatModel,
		Instructions: "You are a friendly chat assistant. Engage in helpful conversations.",
	})
	if err != nil {
		log.Fatalf("Failed to create chat agent: %v", err)
	}

	// Create Server 1: Math service with prefix "/math"
	// 创建服务器 1: 带 "/math" 前缀的数学服务
	mathServer, err := agentos.NewServer(&agentos.Config{
		Address: ":8080",
		Prefix:  "/math", // All routes will be under /math prefix
		Debug:   true,
	})
	if err != nil {
		log.Fatalf("Failed to create math server: %v", err)
	}

	if err := mathServer.RegisterAgent("default", mathAgent); err != nil {
		log.Fatalf("Failed to register math agent: %v", err)
	}

	// Create Server 2: Chat service with prefix "/chat"
	// 创建服务器 2: 带 "/chat" 前缀的聊天服务
	chatServer, err := agentos.NewServer(&agentos.Config{
		Address: ":8081",
		Prefix:  "/chat", // All routes will be under /chat prefix
		Debug:   true,
	})
	if err != nil {
		log.Fatalf("Failed to create chat server: %v", err)
	}

	if err := chatServer.RegisterAgent("default", chatAgent); err != nil {
		log.Fatalf("Failed to register chat agent: %v", err)
	}

	// Create Server 3: Combined service (both agents on same port)
	// 创建服务器 3: 组合服务 (两个 agent 在同一端口)
	// Note: This demonstrates the power of route prefixes - multiple services on one port
	// 注意: 这演示了路由前缀的强大功能 - 一个端口上的多个服务

	fmt.Println("🎯 Starting AgentOS instances...")
	fmt.Println()
	fmt.Println("📍 Math Service (Port 8080):")
	fmt.Println("   Health:  http://localhost:8080/health")
	fmt.Println("   API:     http://localhost:8080/math/api/v1/agents")
	fmt.Println("   Sessions: http://localhost:8080/math/api/v1/sessions")
	fmt.Println()
	fmt.Println("📍 Chat Service (Port 8081):")
	fmt.Println("   Health:  http://localhost:8081/health")
	fmt.Println("   API:     http://localhost:8081/chat/api/v1/agents")
	fmt.Println("   Sessions: http://localhost:8081/chat/api/v1/sessions")
	fmt.Println()
	fmt.Println("💡 Example requests:")
	fmt.Println()
	fmt.Println("   # List math agents")
	fmt.Println("   curl http://localhost:8080/math/api/v1/agents")
	fmt.Println()
	fmt.Println("   # Create math session")
	fmt.Println("   curl -X POST http://localhost:8080/math/api/v1/sessions \\")
	fmt.Println("        -H 'Content-Type: application/json' \\")
	fmt.Println("        -d '{\"agent_id\": \"default\"}'")
	fmt.Println()
	fmt.Println("   # Run math agent")
	fmt.Println("   curl -X POST http://localhost:8080/math/api/v1/agents/default/run \\")
	fmt.Println("        -H 'Content-Type: application/json' \\")
	fmt.Println("        -d '{\"input\": \"What is 25 * 4 + 10?\"}'")
	fmt.Println()
	fmt.Println("   # List chat agents")
	fmt.Println("   curl http://localhost:8081/chat/api/v1/agents")
	fmt.Println()
	fmt.Println("   # Run chat agent")
	fmt.Println("   curl -X POST http://localhost:8081/chat/api/v1/agents/default/run \\")
	fmt.Println("        -H 'Content-Type: application/json' \\")
	fmt.Println("        -d '{\"input\": \"Hello, how are you?\"}'")
	fmt.Println()

	// Start servers in goroutines
	// 在 goroutine 中启动服务器
	go func() {
		if err := mathServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Math server error: %v", err)
		}
	}()

	go func() {
		if err := chatServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("Chat server error: %v", err)
		}
	}()

	// Give servers time to start
	// 给服务器启动时间
	time.Sleep(500 * time.Millisecond)

	fmt.Println("✅ All servers started successfully!")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop all servers")
	fmt.Println()

	// Wait for interrupt signal
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down all servers...")

	// Graceful shutdown
	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown both servers
	// 关闭两个服务器
	errChan := make(chan error, 2)

	go func() {
		errChan <- mathServer.Shutdown(ctx)
	}()

	go func() {
		errChan <- chatServer.Shutdown(ctx)
	}()

	// Wait for both shutdowns
	// 等待两个服务器关闭
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}

	fmt.Println("✅ All servers stopped gracefully")
}
