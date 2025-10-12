package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rexleimo/agno-go/pkg/agentos/a2a"
)

type MockAgent struct {
	ID   string
	Name string
}

func (m *MockAgent) Run(ctx context.Context, input string) (interface{}, error) {
	response := fmt.Sprintf("你好！我是 %s。你说: %s", m.Name, input)
	return &a2a.RunOutput{
		Content:  response,
		Metadata: map[string]interface{}{"agent_id": m.ID},
	}, nil
}

func (m *MockAgent) GetID() string   { return m.ID }
func (m *MockAgent) GetName() string { return m.Name }

func main() {
	agent := &MockAgent{ID: "demo-agent", Name: "演示助手"}
	a2aInterface, _ := a2a.New(a2a.Config{Agents: []a2a.Entity{agent}})

	router := gin.Default()
	a2aInterface.RegisterRoutes(router)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "agents": a2aInterface.ListEntities()})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "7777"
	}

	fmt.Printf("\n🚀 A2A Server 运行在 http://localhost:%s\n", port)
	log.Fatal(router.Run(":" + port))
}
