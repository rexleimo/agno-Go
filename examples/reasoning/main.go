package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rexleimo/agno-go/pkg/agno/agent"
	openaiModel "github.com/rexleimo/agno-go/pkg/agno/models/openai"
)

func main() {
	// 创建 OpenAI o1 模型 / Create OpenAI o1 model
	model, err := openaiModel.New("o1-preview", openaiModel.Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// 创建 Agent / Create Agent
	ag, err := agent.New(agent.Config{
		Name:         "ReasoningAgent",
		Model:        model,
		Instructions: "You are a helpful assistant that uses reasoning to solve complex problems.",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 运行 Agent / Run Agent
	ctx := context.Background()
	input := "Solve this complex problem: What is the relationship between quantum entanglement and information theory?"

	fmt.Println("🤖 ReasoningAgent is thinking...")
	fmt.Println("📝 Input:", input)
	fmt.Println()

	output, err := ag.Run(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	// 输出结果 / Print result
	fmt.Println("💬 Response:")
	fmt.Println(output.Content)
	fmt.Println()

	// 检查是否有推理内容 / Check if there's reasoning content
	// Note: 推理内容会自动提取并存储在消息中
	// Note: Reasoning content is automatically extracted and stored in messages
	for _, msg := range output.Messages {
		if msg.ReasoningContent != nil {
			fmt.Println("🧠 Reasoning Process:")
			fmt.Println(msg.ReasoningContent.Content)
			fmt.Println()

			// 如果有 token 统计 / If token count is available
			if msg.ReasoningContent.TokenCount != nil {
				fmt.Printf("📊 Reasoning Tokens: %d\n", *msg.ReasoningContent.TokenCount)
			}

			// 如果有脱敏内容 / If redacted content is available
			if msg.ReasoningContent.RedactedContent != nil {
				fmt.Println("🔒 Redacted Reasoning:")
				fmt.Println(*msg.ReasoningContent.RedactedContent)
			}
			break
		}
	}
}
