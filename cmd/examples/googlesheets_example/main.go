package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rexleimo/agno-go/pkg/agno/agent"
	"github.com/rexleimo/agno-go/pkg/agno/models/openai"
	"github.com/rexleimo/agno-go/pkg/agno/tools/googlesheets"
	"github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
)

// 本示例演示如何使用 Google Sheets 工具包
// This example demonstrates how to use the Google Sheets toolkit
func main() {
	// 1. 设置 API 密钥和凭证
	// 1. Set up API keys and credentials
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Google Sheets 凭证可以是:
	// Google Sheets credentials can be:
	// 1. 服务账号 JSON 文件路径
	//    Service account JSON file path
	// 2. JSON 字符串
	//    JSON string
	sheetsCredentials := os.Getenv("GOOGLE_SHEETS_CREDENTIALS")
	if sheetsCredentials == "" {
		log.Fatal("GOOGLE_SHEETS_CREDENTIALS environment variable is required (file path or JSON string)")
	}

	// 测试电子表格 ID (从 URL 中获取)
	// Test spreadsheet ID (get from URL)
	// 例如: https://docs.google.com/spreadsheets/d/SPREADSHEET_ID/edit
	// Example: https://docs.google.com/spreadsheets/d/SPREADSHEET_ID/edit
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	if spreadsheetID == "" {
		log.Println("SPREADSHEET_ID not set, using demo mode (will show available functions only)")
	}

	// 2. 创建 Google Sheets 工具包
	// 2. Create Google Sheets toolkit
	sheetsConfig := googlesheets.Config{
		CredentialsJSON: sheetsCredentials,
	}

	sheetsTool, err := googlesheets.New(sheetsConfig)
	if err != nil {
		log.Fatalf("Failed to create Google Sheets toolkit: %v", err)
	}

	fmt.Println("✅ Google Sheets toolkit initialized")

	// 3. 创建 OpenAI 模型
	// 3. Create OpenAI model
	modelConfig := openai.Config{
		APIKey: apiKey,
	}

	model, err := openai.New("gpt-4", modelConfig)
	if err != nil {
		log.Fatalf("Failed to create OpenAI model: %v", err)
	}

	// 4. 创建带有 Google Sheets 工具的 Agent
	// 4. Create agent with Google Sheets tools
	agentConfig := agent.Config{
		Name:  "sheets-agent",
		Model: model,
		Instructions: `你是一个 Google Sheets 助手。你可以使用以下工具操作电子表格：
- read_range: 读取指定范围的数据
- write_range: 写入数据到指定范围
- append_rows: 追加新行到表格

当用户要求操作电子表格时，使用相应的工具。`,
		Toolkits: []toolkit.Toolkit{sheetsTool},
	}

	ag, err := agent.New(agentConfig)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	fmt.Println("✅ Agent created with Google Sheets tools")

	// 5. 演示模式 - 显示可用的工具
	// 5. Demo mode - show available tools
	if spreadsheetID == "" {
		fmt.Println("\n📋 Available Google Sheets functions:")
		for _, fn := range sheetsTool.Functions() {
			fmt.Printf("\n🔧 %s\n", fn.Name)
			fmt.Printf("   描述: %s\n", fn.Description)
			fmt.Printf("   参数:\n")
			for paramName, param := range fn.Parameters {
				required := ""
				if param.Required {
					required = " (必需)"
				}
				fmt.Printf("     - %s: %s%s\n", paramName, param.Description, required)
			}
		}

		fmt.Println("\n💡 设置 SPREADSHEET_ID 环境变量来运行完整示例")
		return
	}

	// 6. 运行示例查询
	// 6. Run example queries
	ctx := context.Background()

	// 示例 1: 读取数据
	// Example 1: Read data
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("示例 1: 读取电子表格数据")
	fmt.Println(strings.Repeat("=", 50))

	query1 := fmt.Sprintf("请读取电子表格 %s 的 Sheet1!A1:B5 范围的数据", spreadsheetID)
	fmt.Printf("\n💬 User: %s\n", query1)

	output1, err := ag.Run(ctx, query1)
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		fmt.Printf("🤖 Assistant: %s\n", output1.Content)
	}

	// 示例 2: 写入数据
	// Example 2: Write data
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("示例 2: 写入数据到电子表格")
	fmt.Println(strings.Repeat("=", 50))

	query2 := fmt.Sprintf(`请将以下数据写入电子表格 %s 的 Sheet1!A1 位置:
第一行: Name, Score
第二行: Alice, 95
第三行: Bob, 87`, spreadsheetID)
	fmt.Printf("\n💬 User: %s\n", query2)

	output2, err := ag.Run(ctx, query2)
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		fmt.Printf("🤖 Assistant: %s\n", output2.Content)
	}

	// 示例 3: 追加数据
	// Example 3: Append data
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("示例 3: 追加新行到表格")
	fmt.Println(strings.Repeat("=", 50))

	query3 := fmt.Sprintf(`请在电子表格 %s 的 Sheet1 中追加一行:
Charlie, 92`, spreadsheetID)
	fmt.Printf("\n💬 User: %s\n", query3)

	output3, err := ag.Run(ctx, query3)
	if err != nil {
		log.Printf("❌ Error: %v", err)
	} else {
		fmt.Printf("🤖 Assistant: %s\n", output3.Content)
	}

	fmt.Println("\n✅ 示例完成!")
}
