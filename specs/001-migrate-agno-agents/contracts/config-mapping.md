# Config Mapping: Python → Go（Agents / Providers / Workflows / Sessions）

**Feature**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/spec.md  
**Plan**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/plan.md  
**Date**: 2025-11-19

本文档定义从 Python 版 agno 配置到 Go 版 agno-Go 配置/构造代码的映射规则，并通过 US1（`01_basic_coordination.py`）示例给出一个完整的映射对。

---

## 1. 映射原则总览

1. **语义优先**：保持字段含义一致，而不是逐字段机械翻译。
2. **结构对应**：Python 对象 → Go 结构体，Python 配置字典 → Go `map[string]any` 或结构体字段。
3. **显式标识**：涉及 Provider/Agent/Workflow/Session 的 ID 在 Python 与 Go 中保持一致或使用可追踪派生规则。
4. **行为等价**：重要配置（例如模型 ID、工具名称、协作模式类型）在两端必须具有同样的业务语义。

---

## 2. Provider 配置映射

### 2.1 通用规则

- Python：

  - LLM 供应商常以 `OpenAIChat("model-id")` 形式出现；
  - 工具类供应商通过 `Toolkit` 子类（如 `HackerNewsTools()`、`Newspaper4kTools()`）注入到 Agent。

- Go：

  - 使用 `providers.Provider` 结构体描述静态配置：

    ```go
    type Provider struct {
        ID           ID
        Type         Type
        DisplayName  string
        Config       Config        // map[string]any
        Capabilities []Capability  // e.g. "generate", "invoke_tool"
    }
    ```

- 映射：

  | Python                             | Go                                      | 说明                                    |
  |------------------------------------|-----------------------------------------|-----------------------------------------|
  | `OpenAIChat("gpt-5-mini")`        | Provider ID: `openai-chat-gpt-5-mini`   | `Config["model"] = "gpt-5-mini"`        |
  | `HackerNewsTools()`               | Provider ID: `hackernews-tools`         | Type: `tool-executor`、Capability: `invoke_tool` |
  | `Newspaper4kTools()`              | Provider ID: `newspaper4k-tools`        | Type: `tool-executor`、Capability: `invoke_tool` |

### 2.2 US1 示例映射

> Python：`agno/cookbook/teams/basic_flows/01_basic_coordination.py`

```python
hn_researcher = Agent(
    name="HackerNews Researcher",
    model=OpenAIChat("gpt-5-mini"),
    role="Gets top stories from hackernews.",
    tools=[HackerNewsTools()],
)

article_reader = Agent(
    name="Article Reader",
    model=OpenAIChat("gpt-5-mini"),
    role="Reads articles from URLs.",
    tools=[Newspaper4kTools()],
)
```

> Go：`go/providers/providers.go`（US1 部分）

```go
var (
    US1OpenAIChat = Provider{
        ID:          ID("openai-chat-gpt-5-mini"),
        Type:        TypeLLM,
        DisplayName: "OpenAI Chat gpt-5-mini",
        Config: providers.Config{
            "model": "gpt-5-mini",
        },
        Capabilities: []Capability{CapabilityGenerate},
    }

    US1HackerNewsTools = Provider{
        ID:          ID("hackernews-tools"),
        Type:        TypeToolExecutor,
        DisplayName: "HackerNews Tools",
        Config:      Config{},
        Capabilities: []Capability{
            CapabilityInvokeTool,
        },
    }

    US1Newspaper4kTools = Provider{
        ID:          ID("newspaper4k-tools"),
        Type:        TypeToolExecutor,
        DisplayName: "Newspaper4k Tools",
        Config:      Config{},
        Capabilities: []Capability{
            CapabilityInvokeTool,
        },
    }
)
```

---

## 3. Agent 配置映射

### 3.1 通用规则

- Python `Agent`（简化）：

  ```python
  Agent(
      name="...",
      role="...",
      model=OpenAIChat("..."),
      tools=[SomeTools()],
      # 其他字段：db、memory、add_history_to_context 等
  )
  ```

- Go `agent.Agent`：

  ```go
  type Agent struct {
      ID               ID
      Name             string
      Role             string
      Description      string
      AllowedProviders []ProviderID
      AllowedTools     []ToolID
      InputSchema      Schema
      OutputSchema     Schema
      MemoryPolicy     MemoryPolicy
  }
  ```

- 映射：

  | Python              | Go                             |
  |---------------------|--------------------------------|
  | `name`              | `Name`                         |
  | `role`              | `Role`                         |
  | `model`             | 通过 `AllowedProviders` 绑定 Provider ID |
  | `tools`             | 通过 `AllowedTools` 绑定 Tool ID        |
  | 输入/输出模型/模式   | `InputSchema` / `OutputSchema` |

### 3.2 US1 示例映射

> Python：同上 `hn_researcher` / `article_reader`

> Go：`go/agent/us1_basic_coordination_agents.go`

```go
var US1HackerNewsResearcher = Agent{
    ID:          ID("hn_researcher"),
    Name:        "HackerNews Researcher",
    Role:        "Gets top stories from hackernews.",
    Description: "Uses HackerNews API to find and analyze relevant posts.",
    AllowedProviders: []ProviderID{
        ProviderOpenAIUS,
    },
    AllowedTools: []ToolID{
        ToolHackerNews,
    },
    InputSchema: Schema{
        Name:        "US1Query",
        Description: "User query describing the topic to research on HackerNews.",
    },
    OutputSchema: Schema{
        Name:        "US1HackerNewsFindings",
        Description: "Summarized findings and URLs from HackerNews.",
    },
    MemoryPolicy: MemoryPolicy{
        Persist:            false,
        WindowSize:         0,
        SensitiveFiltering: true,
    },
}
```

---

## 4. Workflow 配置映射

### 4.1 通用规则

- Python：
  - 使用 `Team` + `instructions`、`members` 等描述协作模式；
  - 在更底层可能有 Workflow 配置。
- Go：
  - 使用 `workflow.Workflow`：

    ```go
    type Workflow struct {
        ID                   ID
        Name                 string
        PatternType          PatternType
        Steps                []Step
        EntryPoints          []StepID
        TerminationCondition TerminationCondition
        RoutingRules         []RoutingRule
    }
    ```

- 映射（US1 下的主要元素）：

  | Python（Team）                                                 | Go（Workflow）                                             |
  |----------------------------------------------------------------|------------------------------------------------------------|
  | `members=[hn_researcher, article_reader]`                      | `Steps` 中分别引用对应 AgentID                            |
  | `instructions` 描述执行顺序（先查 HN，再读文章，最后总结）     | `PatternType = sequential` + `RoutingRules` 对应顺序      |

### 4.2 US1 示例映射

> Python：`Team` 在 US1 示例中的成员与 instructions 描述顺序。

> Go：`go/workflow/us1_basic_coordination_workflow.go`

```go
func US1BasicCoordinationWorkflow() Workflow {
    steps := []Step{
        {
            ID:      StepID("search_hackernews"),
            AgentID: agent.ID("hn_researcher"),
            Name:    "Search HackerNews for relevant stories",
        },
        {
            ID:      StepID("read_articles"),
            AgentID: agent.ID("article_reader"),
            Name:    "Read articles from URLs and summarize",
        },
    }

    rules := []RoutingRule{
        {
            From:      StepID("search_hackernews"),
            To:        StepID("read_articles"),
            Condition: "always",
        },
    }

    return Workflow{
        ID:          ID("us1-basic-coordination"),
        Name:        "US1 Basic Coordination",
        PatternType: PatternSequential,
        Steps:       steps,
        EntryPoints: []StepID{steps[0].ID},
        TerminationCondition: TerminationCondition{
            MaxIterations: 1,
            OnError:       "fail-fast",
        },
        RoutingRules: rules,
    }
}
```

---

## 5. Session 配置映射

### 5.1 通用规则

- Python：
  - 使用 `TeamSession` / `WorkflowSession` 等类记录会话上下文、历史和结果。
- Go：
  - 使用 `session.Session`：

    ```go
    type Session struct {
        ID        ID
        Workflow  workflow.ID
        Context   UserContext
        History   []HistoryEntry
        Status    Status
        Result    *Result
        TraceID   string
        CreatedAt time.Time
        UpdatedAt time.Time
    }
    ```

- 映射：

  | Python                          | Go                           |
  |---------------------------------|------------------------------|
  | Session/Run ID                  | `Session.ID`                 |
  | workflow_id                     | `Session.Workflow`           |
  | 用户输入/上下文                 | `Session.Context.Payload`    |
  | 历史记录（messages / events）   | `Session.History`           |
  | 最终结果                        | `Session.Result`             |

### 5.2 US1 示例映射

> Go：`go/session/us1_basic_coordination_session.go`

```go
func RunUS1Session(input agent.US1Input, wf workflow.Workflow) Session {
    now := time.Now()

    s := Session{
        ID:       ID("us1-" + ID(now.Format("20060102150405"))),
        Workflow: wf.ID,
        Context: UserContext{
            Channel:   "us1-basic-coordination",
            Payload:   map[string]any{"query": input.Query},
            StartedAt: now,
        },
        Status: StatusCompleted,
        Result: &Result{
            Success: true,
            Reason:  "placeholder",
            Data: map[string]any{
                "query": input.Query,
            },
        },
        CreatedAt: now,
        UpdatedAt: now,
    }

    // History / Telemetry 略，详见源码
    return s
}
```

---

## 6. 小结：US1 完整映射对

- Providers：从 Python 的 `OpenAIChat/HackerNewsTools/Newspaper4kTools` 映射到 `US1OpenAIChat`、`US1HackerNewsTools`、`US1Newspaper4kTools`。
- Agents：从 Python `hn_researcher` / `article_reader` 映射到 Go 的 `US1HackerNewsResearcher` / `US1ArticleReader`。
- Workflow/Team：从 Python Team 的成员与 instructions 映射到 `US1BasicCoordinationWorkflow` 的 steps 和 routing rules。
- Session：从 Python 的会话/运行概念映射到 Go 的 `Session` 结构及 `RunUS1Session` 入口。

在后续迁移中，只要遵循上述映射规则，其他 Python 场景（包括不同 providers、协作模式和 session 管理策略）都可以采用类似方式迁移到 Go。*** End Patch
