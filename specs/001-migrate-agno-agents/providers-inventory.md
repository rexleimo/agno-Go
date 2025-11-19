# Providers Inventory: agno 核心 agents 能力迁移

**Feature**: /Users/rex/cool.cnb/agno-Go/specs/001-migrate-agno-agents/spec.md  
**Date**: 2025-11-19

本清单用于跟踪 `agno/cookbook` 中出现的 agents 供应商及其在 Go 侧的迁移状态。当前版本重点列出已在本功能中实现或显式使用的供应商；完整覆盖 cookbook 中所有供应商将作为后续迭代扩展。

## 首批供应商（示例场景相关）

| Provider ID              | Type           | Source Examples                                                        | Used In Scenarios                    | Batch         | Go Adapter Status                                | Notes |
|--------------------------|----------------|-------------------------------------------------------------------------|--------------------------------------|--------------|--------------------------------------------------|-------|
| openai-chat-gpt-5-mini   | llm            | agno/cookbook/teams/basic_flows/01_basic_coordination.py               | US1: teams-basic-coordination-us1    | 首批必须迁移 | Implemented as `US1OpenAIChat` in go/providers   | 模型配置采用 `model=gpt-5-mini`，用于团队内所有文本生成 |
| hackernews-tools         | tool-executor  | agno/cookbook/teams/basic_flows/01_basic_coordination.py               | US1: teams-basic-coordination-us1    | 首批必须迁移 | Implemented as `US1HackerNewsTools` in go/providers | 提供对 HackerNews 的检索能力，在 US1 中作为 HN Researcher 的工具 |
| newspaper4k-tools        | tool-executor  | agno/cookbook/teams/basic_flows/01_basic_coordination.py               | US1: teams-basic-coordination-us1    | 首批必须迁移 | Implemented as `US1Newspaper4kTools` in go/providers | 用于读取新闻文章内容，在 US1 中作为 Article Reader 的工具 |
| custom_internal_search   | tool-executor  | agno/libs/agno/agno/tools/custom_internal_search.py; agno/cookbook/scripts/us3_custom_provider_parity.py | US3: custom-internal-search-us3      | 首批必须迁移 | Implemented as `US3CustomInternalSearch` in go/providers | 纯内存实现的内部搜索示例，用于自定义 Provider 协议与 parity 演示 |

> 注：上述供应商均已在 Go 侧具备最小可用适配层，并与至少一个 ParityTestScenario 绑定（US1 或 US3）。其它 cookbook 中的供应商将在后续迭代中补充到本清单，并按“首批必须迁移 / 后续批次 / 不再支持”分类。

