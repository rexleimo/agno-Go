# MCP Demo Example

## Overview | 概述

This example demonstrates how to connect to an MCP (Model Context Protocol) server and use its tools through the Agno-Go MCP client. You'll learn how to set up security validation, create transports, connect to MCP servers, and integrate MCP tools with your Agno agents.

本示例演示如何连接到 MCP (模型上下文协议) 服务器并通过 Agno-Go MCP 客户端使用其工具。您将学习如何设置安全验证、创建传输、连接到 MCP 服务器以及将 MCP 工具与您的 Agno agents 集成。

## You Will Learn | 你将学到

- How to create and configure security validation for MCP commands
  - 如何为 MCP 命令创建和配置安全验证
- How to set up stdio transport for subprocess communication
  - 如何设置用于子进程通信的 stdio 传输
- How to connect to an MCP server and discover available tools
  - 如何连接到 MCP 服务器并发现可用工具
- How to create an MCP toolkit for use with Agno agents
  - 如何创建用于 Agno agents 的 MCP 工具包
- How to call MCP tools directly
  - 如何直接调用 MCP 工具

## Prerequisites | 前置要求

- Go 1.21 or later | Go 1.21 或更高版本
- An MCP server installed (e.g., calculator server)
  - 已安装的 MCP 服务器 (例如: calculator 服务器)

## Setup | 设置

### 1. Install an MCP Server | 安装 MCP 服务器

```bash
# Install uvx package manager
# 安装 uvx 包管理器
pip install uvx

# Install the calculator MCP server
# 安装 calculator MCP 服务器
uvx mcp install @modelcontextprotocol/server-calculator

# Verify installation
# 验证安装
python -m mcp_server_calculator --help
```

### 2. Run the Example | 运行示例

```bash
# Navigate to example directory
# 进入示例目录
cd cmd/examples/mcp_demo

# Run directly
# 直接运行
go run main.go

# Or build and run
# 或构建并运行
go build -o mcp_demo
./mcp_demo
```

## Complete Code | 完整代码

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/rexleimo/agno-go/pkg/agno/mcp/client"
	"github.com/rexleimo/agno-go/pkg/agno/mcp/security"
	mcptoolkit "github.com/rexleimo/agno-go/pkg/agno/mcp/toolkit"
)

func main() {
	fmt.Println("=== Agno-Go MCP Demo ===")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 1: Create security validator
	// 步骤 1: 创建安全验证器
	fmt.Println("Step 1: Creating security validator...")
	validator := security.NewCommandValidator()

	command := "python"
	args := []string{"-m", "mcp_server_calculator"}

	if err := validator.Validate(command, args); err != nil {
		log.Fatalf("Command validation failed: %v", err)
	}
	fmt.Printf("✓ Command validated: %s %v\n", command, args)

	// Step 2: Create transport
	// 步骤 2: 创建传输
	fmt.Println("Step 2: Creating transport...")
	transport, err := client.NewStdioTransport(client.StdioConfig{
		Command: command,
		Args:    args,
	})
	if err != nil {
		log.Fatalf("Failed to create transport: %v", err)
	}
	fmt.Println("✓ Stdio transport created")

	// Step 3: Create MCP client
	// 步骤 3: 创建 MCP 客户端
	fmt.Println("Step 3: Creating MCP client...")
	mcpClient, err := client.New(transport, client.Config{
		ClientName:    "agno-go-demo",
		ClientVersion: "0.1.0",
	})
	if err != nil {
		log.Fatalf("Failed to create MCP client: %v", err)
	}
	fmt.Println("✓ MCP client created")

	// Step 4: Connect to server
	// 步骤 4: 连接到服务器
	fmt.Println("Step 4: Connecting to MCP server...")
	if err := mcpClient.Connect(ctx); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer mcpClient.Disconnect()

	fmt.Println("✓ Connected to MCP server")
	if serverInfo := mcpClient.GetServerInfo(); serverInfo != nil {
		fmt.Printf("  Server: %s v%s\n", serverInfo.Name, serverInfo.Version)
	}

	// Step 5: Discover tools
	// 步骤 5: 发现工具
	fmt.Println("Step 5: Discovering tools...")
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	fmt.Printf("✓ Found %d tools:\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	// Step 6: Create MCP toolkit
	// 步骤 6: 创建 MCP 工具包
	fmt.Println("Step 6: Creating MCP toolkit...")
	toolkit, err := mcptoolkit.New(ctx, mcptoolkit.Config{
		Client: mcpClient,
		Name:   "calculator-tools",
	})
	if err != nil {
		log.Fatalf("Failed to create toolkit: %v", err)
	}
	defer toolkit.Close()

	fmt.Println("✓ MCP toolkit created")
	fmt.Printf("  Toolkit name: %s\n", toolkit.Name())
	fmt.Printf("  Available functions: %d\n", len(toolkit.Functions()))

	// Step 7: Call a tool directly
	// 步骤 7: 直接调用工具
	fmt.Println("Step 7: Calling a tool...")
	result, err := mcpClient.CallTool(ctx, "add", map[string]interface{}{
		"a": 5,
		"b": 3,
	})
	if err != nil {
		log.Fatalf("Failed to call tool: %v", err)
	}

	fmt.Println("✓ Tool call successful")
	fmt.Printf("  Result: %v\n", result.Content)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("The MCP toolkit can now be passed to an agno Agent!")
}
```

## Code Explanation | 代码解释

### 1. Security Validation | 安全验证

```go
validator := security.NewCommandValidator()
if err := validator.Validate(command, args); err != nil {
    log.Fatalf("Command validation failed: %v", err)
}
```

- Creates a security validator with default whitelist | 使用默认白名单创建安全验证器
- Validates that the command is safe to execute | 验证命令是否安全可执行
- Blocks dangerous shell metacharacters | 阻止危险的 shell 元字符

### 2. Stdio Transport | Stdio 传输

```go
transport, err := client.NewStdioTransport(client.StdioConfig{
    Command: "python",
    Args:    []string{"-m", "mcp_server_calculator"},
})
```

- Creates a transport that communicates via stdin/stdout | 创建通过 stdin/stdout 通信的传输
- Spawns the MCP server as a subprocess | 将 MCP 服务器作为子进程生成
- Handles bidirectional JSON-RPC 2.0 messages | 处理双向 JSON-RPC 2.0 消息

### 3. MCP Client | MCP 客户端

```go
mcpClient, err := client.New(transport, client.Config{
    ClientName:    "agno-go-demo",
    ClientVersion: "0.1.0",
})
```

- Creates an MCP client with your application identity | 使用您的应用程序标识创建 MCP 客户端
- Manages the connection lifecycle | 管理连接生命周期
- Provides methods for tool discovery and invocation | 提供工具发现和调用的方法

### 4. Tool Discovery | 工具发现

```go
tools, err := mcpClient.ListTools(ctx)
for _, tool := range tools {
    fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
}
```

- Queries the MCP server for available tools | 查询 MCP 服务器的可用工具
- Returns tool metadata (name, description, parameters) | 返回工具元数据(名称、描述、参数)
- Used for dynamic tool discovery | 用于动态工具发现

### 5. MCP Toolkit Creation | MCP 工具包创建

```go
toolkit, err := mcptoolkit.New(ctx, mcptoolkit.Config{
    Client: mcpClient,
    Name:   "calculator-tools",
})
defer toolkit.Close()
```

- Converts MCP tools into Agno toolkit functions | 将 MCP 工具转换为 Agno 工具包函数
- Automatically generates function signatures from MCP schemas | 从 MCP schema 自动生成函数签名
- Compatible with `agent.Config.Toolkits` | 与 `agent.Config.Toolkits` 兼容

### 6. Direct Tool Call | 直接工具调用

```go
result, err := mcpClient.CallTool(ctx, "add", map[string]interface{}{
    "a": 5,
    "b": 3,
})
fmt.Printf("Result: %v\n", result.Content)
```

- Calls an MCP tool directly without an agent | 直接调用 MCP 工具,无需 agent
- Passes parameters as a map | 将参数作为 map 传递
- Returns the result content | 返回结果内容

## Expected Output | 预期输出

```
=== Agno-Go MCP Demo ===

Step 1: Creating security validator...
✓ Command validated: python [-m mcp_server_calculator]

Step 2: Creating transport...
✓ Stdio transport created

Step 3: Creating MCP client...
✓ MCP client created

Step 4: Connecting to MCP server...
✓ Connected to MCP server
  Server: calculator v0.1.0

Step 5: Discovering tools...
✓ Found 4 tools:
  - add: Add two numbers
  - subtract: Subtract two numbers
  - multiply: Multiply two numbers
  - divide: Divide two numbers

Step 6: Creating MCP toolkit...
✓ MCP toolkit created
  Toolkit name: calculator-tools
  Available functions: 4

Step 7: Calling a tool...
✓ Tool call successful
  Result: 8

=== Demo Complete ===
The MCP toolkit can now be passed to an agno Agent!
```

## Using with Agno Agents | 与 Agno Agents 一起使用

Once you have the MCP toolkit, you can use it with any Agno agent:

一旦您有了 MCP 工具包,您可以将其与任何 Agno agent 一起使用:

```go
import (
    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
)

// Create model
// 创建模型
model, _ := openai.New("gpt-4o-mini", openai.Config{
    APIKey: "your-api-key",
})

// Create agent with MCP toolkit
// 使用 MCP 工具包创建 agent
ag, _ := agent.New(agent.Config{
    Name:     "MCP Calculator Agent",
    Model:    model,
    Toolkits: []toolkit.Toolkit{toolkit},  // MCP toolkit here!
})

// Run agent
// 运行 agent
output, _ := ag.Run(context.Background(), "What is 25 * 4 + 15?")
fmt.Println(output.Content)
```

## Troubleshooting | 故障排除

**Error: "command not allowed"**
- Make sure your MCP server command is in the security whitelist
  - 确保您的 MCP 服务器命令在安全白名单中
- Add it with `validator.AddAllowedCommand("your-command")`

**Error: "failed to start process"**
- Verify the MCP server is installed: `python -m mcp_server_calculator --help`
  - 验证 MCP 服务器已安装
- Check that Python is in your PATH
  - 检查 Python 是否在您的 PATH 中

**Error: "connection timeout"**
- The MCP server may be taking too long to start
  - MCP 服务器可能启动时间过长
- Increase the context timeout: `context.WithTimeout(ctx, 60*time.Second)`
  - 增加上下文超时

**Tool calls return errors**
- Verify the tool exists: check `mcpClient.ListTools(ctx)`
  - 验证工具是否存在
- Ensure parameter types match the tool schema
  - 确保参数类型与工具 schema 匹配

## Next Steps | 下一步

- Read the [MCP Integration Guide](../guide/mcp.md) | 阅读 [MCP 集成指南](../guide/mcp.md)
- Try connecting to other MCP servers (filesystem, git, sqlite)
  - 尝试连接到其他 MCP 服务器 (filesystem, git, sqlite)
- Build a custom MCP server for your use case
  - 为您的用例构建自定义 MCP 服务器
- Combine MCP tools with built-in Agno tools
  - 将 MCP 工具与内置的 Agno 工具结合使用
