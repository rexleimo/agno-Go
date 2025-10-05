# MCP Integration

## What is MCP?

The **Model Context Protocol (MCP)** is an open standard that enables seamless integration between LLM applications and external data sources and tools. Developed by Anthropic, MCP provides a universal protocol for connecting AI models with various services through a standardized interface.

**模型上下文协议 (MCP)** 是一个开放标准,能够在 LLM 应用程序和外部数据源及工具之间实现无缝集成。由 Anthropic 开发,MCP 通过标准化接口为 AI 模型与各种服务的连接提供了通用协议。

## Why Use MCP with Agno-Go?

- **🔌 Extensibility** - Connect your agents to any MCP-compatible server
  - **可扩展性** - 将您的 agent 连接到任何兼容 MCP 的服务器
- **🔒 Security** - Built-in command validation and shell injection protection
  - **安全性** - 内置命令验证和 shell 注入保护
- **🚀 Performance** - Fast initialization (<100μs) and low memory footprint (<10KB)
  - **性能** - 快速初始化 (<100μs) 和低内存占用 (<10KB)
- **📦 Reusability** - Leverage existing MCP servers without reinventing the wheel
  - **可重用性** - 利用现有的 MCP 服务器,无需重新造轮子

## Architecture

Agno-Go's MCP implementation consists of several key components:

Agno-Go 的 MCP 实现由几个关键组件组成:

```
pkg/agno/mcp/
├── protocol/       # JSON-RPC 2.0 and MCP message types | JSON-RPC 2.0 和 MCP 消息类型
├── client/         # MCP client core and transports | MCP 客户端核心和传输
├── security/       # Command validation and security | 命令验证和安全
├── content/        # Content type handling | 内容类型处理
└── toolkit/        # Integration with agno toolkit system | 与 agno 工具包系统集成
```

## Quick Start

### Prerequisites | 前置要求

- Go 1.21 or later | Go 1.21 或更高版本
- An MCP server (e.g., calculator, filesystem, git)
  - 一个 MCP 服务器 (例如: calculator, filesystem, git)

### Installation | 安装

```bash
# Install uvx for managing MCP servers
# 安装 uvx 以管理 MCP 服务器
pip install uvx

# Install a sample MCP server
# 安装示例 MCP 服务器
uvx mcp install @modelcontextprotocol/server-calculator
```

### Basic Usage | 基本用法

#### 1. Create Security Validator | 创建安全验证器

```go
import "github.com/rexleimo/agno-go/pkg/agno/mcp/security"

// Create validator with default safe commands
// 使用默认安全命令创建验证器
validator := security.NewCommandValidator()

// Validate command before use
// 使用前验证命令
if err := validator.Validate("python", []string{"-m", "mcp_server"}); err != nil {
    log.Fatal(err)
}
```

#### 2. Setup Transport | 设置传输

```go
import "github.com/rexleimo/agno-go/pkg/agno/mcp/client"

// Create stdio transport for subprocess communication
// 创建 stdio 传输以进行子进程通信
transport, err := client.NewStdioTransport(client.StdioConfig{
    Command: "python",
    Args:    []string{"-m", "mcp_server_calculator"},
})
if err != nil {
    log.Fatal(err)
}
```

#### 3. Connect to MCP Server | 连接到 MCP 服务器

```go
// Create MCP client
// 创建 MCP 客户端
mcpClient, err := client.New(transport, client.Config{
    ClientName:    "my-agent",
    ClientVersion: "1.0.0",
})
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
if err := mcpClient.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer mcpClient.Disconnect()

// Get server information
// 获取服务器信息
serverInfo := mcpClient.GetServerInfo()
fmt.Printf("Connected to: %s v%s\n", serverInfo.Name, serverInfo.Version)
```

#### 4. Create MCP Toolkit for Agents | 为 Agent 创建 MCP 工具包

```go
import (
    "github.com/rexleimo/agno-go/pkg/agno/agent"
    mcptoolkit "github.com/rexleimo/agno-go/pkg/agno/mcp/toolkit"
)

// Create toolkit from MCP client
// 从 MCP 客户端创建工具包
toolkit, err := mcptoolkit.New(ctx, mcptoolkit.Config{
    Client: mcpClient,
    Name:   "calculator-tools",
    // Optional: filter specific tools
    // 可选: 过滤特定工具
    IncludeTools: []string{"add", "subtract", "multiply"},
})
if err != nil {
    log.Fatal(err)
}
defer toolkit.Close()

// Use with agno agent
// 与 agno agent 一起使用
ag, err := agent.New(agent.Config{
    Name:     "Math Assistant",
    Model:    yourModel,
    Toolkits: []toolkit.Toolkit{toolkit},
})
```

## Security Features | 安全功能

Agno-Go's MCP implementation prioritizes security:

Agno-Go 的 MCP 实现将安全放在首位:

### Command Whitelist | 命令白名单

Only specific commands are allowed by default:
默认只允许特定命令:

- `python`, `python3`
- `node`, `npm`, `npx`
- `uvx`
- `docker`

### Shell Injection Protection | Shell 注入保护

All command arguments are validated to prevent shell injection:
所有命令参数都经过验证以防止 shell 注入:

**Blocked characters | 阻止的字符:**
- `;` (command separator | 命令分隔符)
- `|` (pipe | 管道)
- `&` (background execution | 后台执行)
- `` ` `` (command substitution | 命令替换)
- `$` (variable expansion | 变量扩展)
- `>`, `<` (redirection | 重定向)

### Custom Security Policies | 自定义安全策略

```go
// Create custom validator with specific allowed commands
// 使用特定允许的命令创建自定义验证器
validator := security.NewCustomCommandValidator(
    []string{"go", "rust"},     // allowed commands | 允许的命令
    []string{";", "|", "&"},    // blocked chars | 阻止的字符
)

// Add or remove commands dynamically
// 动态添加或删除命令
validator.AddAllowedCommand("ruby")
validator.RemoveAllowedCommand("go")
```

## Tool Filtering | 工具过滤

You can selectively include or exclude tools from MCP servers:

您可以选择性地从 MCP 服务器包含或排除工具:

```go
// Include only specific tools
// 仅包含特定工具
toolkit, err := mcptoolkit.New(ctx, mcptoolkit.Config{
    Client:       mcpClient,
    IncludeTools: []string{"add", "subtract", "multiply"},
})

// Or exclude certain tools
// 或排除某些工具
toolkit, err := mcptoolkit.New(ctx, mcptoolkit.Config{
    Client:       mcpClient,
    ExcludeTools: []string{"divide"},  // exclude division | 排除除法
})
```

## Content Handling | 内容处理

MCP supports different content types:

MCP 支持不同的内容类型:

```go
import "github.com/rexleimo/agno-go/pkg/agno/mcp/content"

handler := content.New()

// Extract text from content
// 从内容中提取文本
text := handler.ExtractText(contents)

// Extract images
// 提取图像
images, err := handler.ExtractImages(contents)
for _, img := range images {
    fmt.Printf("Image: %s (%d bytes)\n", img.MimeType, len(img.Data))
}

// Create different content types
// 创建不同的内容类型
textContent := handler.CreateTextContent("Hello, world!")
imageContent := handler.CreateImageContent(imageData, "image/png")
resourceContent := handler.CreateResourceContent("file:///path", "text/plain")
```

## Known MCP Servers | 已知的 MCP 服务器

Compatible MCP servers you can use with Agno-Go:

可与 Agno-Go 一起使用的兼容 MCP 服务器:

| Server | Description | Installation |
|--------|-------------|--------------|
| **@modelcontextprotocol/server-calculator** | Math operations | 数学运算 | `uvx mcp install @modelcontextprotocol/server-calculator` |
| **@modelcontextprotocol/server-filesystem** | File operations | 文件操作 | `uvx mcp install @modelcontextprotocol/server-filesystem` |
| **@modelcontextprotocol/server-git** | Git operations | Git 操作 | `uvx mcp install @modelcontextprotocol/server-git` |
| **@modelcontextprotocol/server-sqlite** | SQLite database | SQLite 数据库 | `uvx mcp install @modelcontextprotocol/server-sqlite` |

More servers available at [MCP Servers Registry](https://github.com/modelcontextprotocol/servers).

更多服务器请访问 [MCP 服务器注册表](https://github.com/modelcontextprotocol/servers)。

## Performance | 性能

Agno-Go's MCP implementation is highly optimized:

Agno-Go 的 MCP 实现经过高度优化:

- **MCP Client Init | MCP 客户端初始化**: <100μs
- **Tool Discovery | 工具发现**: <50μs per server | 每个服务器 <50μs
- **Memory | 内存**: <10KB per connection | 每个连接 <10KB
- **Test Coverage | 测试覆盖率**: >80%

## Limitations | 限制

Current implementation status:

当前实现状态:

- ✅ Stdio transport (implemented | 已实现)
- ⏳ SSE transport (planned | 计划中)
- ⏳ HTTP transport (planned | 计划中)
- ✅ Tools (implemented | 已实现)
- ✅ Resources (implemented | 已实现)
- ✅ Prompts (implemented | 已实现)

## Best Practices | 最佳实践

1. **Always use security validation** - Never bypass command validation
   - **始终使用安全验证** - 永不绕过命令验证

2. **Filter tools appropriately** - Only expose tools your agent needs
   - **适当过滤工具** - 仅公开您的 agent 需要的工具

3. **Handle errors gracefully** - MCP servers may fail or timeout
   - **优雅地处理错误** - MCP 服务器可能会失败或超时

4. **Close connections** - Always defer `toolkit.Close()` to clean up resources
   - **关闭连接** - 始终 defer `toolkit.Close()` 以清理资源

5. **Test with mock servers** - Use the testing utilities in `pkg/agno/mcp/client/testing.go`
   - **使用模拟服务器测试** - 使用 `pkg/agno/mcp/client/testing.go` 中的测试工具

## Next Steps | 下一步

- Try the [MCP Demo Example](../examples/mcp-demo.md) | 尝试 [MCP 演示示例](../examples/mcp-demo.md)
- Read the [MCP Implementation Guide](../../pkg/agno/mcp/IMPLEMENTATION.md) | 阅读 [MCP 实现指南](../../pkg/agno/mcp/IMPLEMENTATION.md)
- Explore the [MCP Protocol Specification](https://spec.modelcontextprotocol.io/) | 探索 [MCP 协议规范](https://spec.modelcontextprotocol.io/)
- Join discussions on [GitHub](https://github.com/rexleimo/agno-Go/discussions)

## Troubleshooting | 故障排除

**Error: "command not allowed"**
- Check that your command is in the whitelist | 检查您的命令是否在白名单中
- Use `validator.AddAllowedCommand()` to add custom commands | 使用 `validator.AddAllowedCommand()` 添加自定义命令

**Error: "shell metacharacters detected"**
- Your command arguments contain dangerous characters | 您的命令参数包含危险字符
- Ensure arguments don't contain `;`, `|`, `&`, etc. | 确保参数不包含 `;`, `|`, `&` 等

**Error: "failed to start MCP server"**
- Verify the MCP server is installed | 验证 MCP 服务器已安装
- Check that the command path is correct | 检查命令路径是否正确
- Ensure you have necessary permissions | 确保您具有必要的权限

**MCP server not responding**
- Check server logs for errors | 检查服务器日志中的错误
- Verify JSON-RPC messages are correctly formatted | 验证 JSON-RPC 消息格式是否正确
- Try reconnecting with `mcpClient.Connect(ctx)` | 尝试使用 `mcpClient.Connect(ctx)` 重新连接
