package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/agno-go/pkg/agentos"
)

func main() {
	fmt.Println("🚀 AgentOS Server Demo")
	fmt.Println("Starting server on http://localhost:8080")
	fmt.Println()
	fmt.Println("Available endpoints:")
	fmt.Println("  GET    /health")
	fmt.Println("  POST   /api/v1/sessions")
	fmt.Println("  GET    /api/v1/sessions/:id")
	fmt.Println("  PUT    /api/v1/sessions/:id")
	fmt.Println("  DELETE /api/v1/sessions/:id")
	fmt.Println("  GET    /api/v1/sessions")
	fmt.Println("  POST   /api/v1/agents/:id/run")
	fmt.Println()

	// Create server with default configuration
	server, err := agentos.NewServer(&agentos.Config{
		Address: ":8080",
		Debug:   true,
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	fmt.Println("✅ Server started successfully!")
	fmt.Println()
	fmt.Println("Try:")
	fmt.Println("  curl http://localhost:8080/health")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	fmt.Println("✅ Server stopped gracefully")
}
