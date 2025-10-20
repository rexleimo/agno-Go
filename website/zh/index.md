---
layout: home

hero:
  name: "Agno-Go"
  text: "高性能多智能体框架"
  tagline: "比 Python 快 16 倍 | 180ns 实例化 | 每个智能体仅 1.2KB 内存"
  image:
    src: /logo.png
    alt: Agno-Go
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/quick-start
    - theme: alt
      text: 在 GitHub 上查看
      link: https://github.com/rexleimo/agno-Go

features:
  - icon: 🚀
    title: 极致性能
    details: 智能体实例化仅需 ~180ns,比 Python 版本快 16 倍。每个智能体内存占用仅 1.2KB,原生支持 Go 并发。

  - icon: 🤖
    title: 生产就绪
    details: AgentOS HTTP 服务器提供 RESTful API、会话管理、智能体注册表、健康监控和开箱即用的全面错误处理。

  - icon: 🧩
    title: 灵活架构
    details: 从 Agent(自主式)、Team(4 种协作模式)或 Workflow(5 种控制原语)中选择,构建您的多智能体系统。

  - icon: 🔌
    title: 多模型支持
    details: 内置支持 OpenAI(GPT-4)、Anthropic Claude、Ollama(本地模型)、DeepSeek、Google Gemini 和 ModelScope。

  - icon: 🔧
    title: 可扩展工具
    details: 易于扩展的工具包系统,内置计算器、HTTP 客户端、文件操作和 DuckDuckGo 搜索。几分钟内创建自定义工具。

  - icon: 💾
    title: RAG 与知识库
    details: ChromaDB 向量数据库集成,支持 OpenAI 嵌入。构建具有语义搜索和知识库的智能代理。

  - icon: ✅
    title: 完善测试
    details: 80.8% 测试覆盖率,85+ 个测试用例,100% 通过率。值得信赖的生产级代码。

  - icon: 📦
    title: 易于部署
    details: 包含 Docker、Docker Compose 和 Kubernetes 清单。几分钟内部署到任何云平台,提供完整部署指南。

  - icon: 📚
    title: 完整文档
    details: OpenAPI 3.0 规范、部署指南、架构文档、性能基准测试,以及每个功能的实际示例。
---

## 快速示例

仅需几行代码即可创建带工具的 AI 智能体:

```go
package main

import (
    "context"
    "fmt"
    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
)

func main() {
    // 创建模型
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: "your-api-key",
    })

    // 创建带工具的智能体
    ag, _ := agent.New(agent.Config{
        Name:     "数学助手",
        Model:    model,
        Toolkits: []toolkit.Toolkit{calculator.New()},
    })

    // 运行智能体
    output, _ := ag.Run(context.Background(), "25 * 4 + 15 等于多少?")
    fmt.Println(output.Content) // 输出: 115
}
```

## 性能对比

| 指标 | Python Agno | Agno-Go | 改进 |
|--------|-------------|---------|-------------|
| 智能体创建 | ~3μs | ~180ns | **快 16 倍** |
| 内存/智能体 | ~6.5KB | ~1.2KB | **减少 5.4 倍** |
| 并发性 | GIL 限制 | 原生 goroutine | **无限制** |

## 为什么选择 Agno-Go?

### 为生产而生

Agno-Go 不仅是一个框架——它是一个完整的生产系统。包含的 **AgentOS** 服务器提供:

- 带 OpenAPI 3.0 规范的 RESTful API
- 多轮对话的会话管理
- 线程安全的智能体注册表
- 健康监控和结构化日志
- CORS 支持和请求超时处理

### KISS 原则

遵循 **Keep It Simple, Stupid** 哲学:

- **3 个核心 LLM 提供商**(而非 45+) - OpenAI、Anthropic、Ollama
- **基础工具**(而非 115+) - 计算器、HTTP、文件、搜索
- **质量优于数量** - 专注于生产就绪的功能

### 开发者体验

- **类型安全**: Go 的强类型在编译时捕获错误
- **快速构建**: Go 的编译速度支持快速迭代
- **易于部署**: 单一二进制文件,无运行时依赖
- **优秀工具**: 内置测试、性能分析和竞态检测

## 5 分钟快速开始

```bash
# 克隆仓库
git clone https://github.com/rexleimo/agno-Go.git
cd agno-Go

# 设置 API 密钥
export OPENAI_API_KEY=sk-your-key-here

# 运行示例
go run cmd/examples/simple_agent/main.go

# 或启动 AgentOS 服务器
docker-compose up -d
curl http://localhost:8080/health
```

## 包含内容

- **核心框架**: Agent、Team(4 种模式)、Workflow(5 种原语)
- **模型**: OpenAI、Anthropic Claude、Ollama、DeepSeek、Gemini、ModelScope
- **工具**: Calculator(75.6%)、HTTP(88.9%)、File(76.2%)、Search(92.1%)
- **RAG**: ChromaDB 集成 + OpenAI 嵌入
- **AgentOS**: 生产级 HTTP 服务器(65.0% 覆盖率)
- **示例**: 6 个涵盖所有功能的实际示例
- **文档**: 完整指南、API 参考、部署说明

## 社区

- **GitHub**: [rexleimo/agno-Go](https://github.com/rexleimo/agno-Go)
- **Issues**: [报告问题和请求功能](https://github.com/rexleimo/agno-Go/issues)
- **Discussions**: [提问和分享想法](https://github.com/rexleimo/agno-Go/discussions)

## 许可证

Agno-Go 基于 [MIT 许可证](https://github.com/rexleimo/agno-Go/blob/main/LICENSE) 发布。

灵感来自 [Agno (Python)](https://github.com/agno-agi/agno) 框架。
