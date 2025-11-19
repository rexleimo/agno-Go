# Parity Fixtures 指南

本目录存放所有跨语言对照测试（ParityFixture）的声明式输入。文档基于 `data-model.md` 中的 `ParityFixture` 定义，帮助在 Go 与 Python 运行时之间共享完全一致的 Workflow 快照、用户输入、工具响应与断言。

## 目录结构

```text
fixtures/
├── README.md                     # 当前说明
├── <scenario>.yaml               # 具体场景（如 us1_basic_coordination）
├── assets/                       # 复用的清单与 Mock 数据
│   ├── workflow_templates/       # 归档 Python/Go 对齐的 Workflow/Agent manifest
│   ├── tool_responses/           # 结构化工具响应，供脚本注入
│   └── expected_assertions/      # 常用断言片段（JSON/YAML）
└── templates/                    # schema 或示例文件
```

> 目录中 `.gitkeep` 仅用于占位，后续可被真实数据替换。

## 字段映射

| 字段 | 说明 | 参考 |
|------|------|------|
| `fixture_id` | 唯一标识，通常对应 cookbook 场景 | spec.md / cookbook |
| `description` | 调试说明，支持多行 | spec.md |
| `workflow_template` | Workflow/Agent 快照，必须与 Python/Go 双方一致 | `assets/workflow_templates/` |
| `user_inputs[]` | 有序消息，包含 `role/content/timestamp/seed` | data-model.md |
| `tool_responses[]` | 预录的工具响应或 API 回放 | `assets/tool_responses/` |
| `expected_assertions[]` | equality/contains/tolerance 断言 | `assets/expected_assertions/` |

## 示例 YAML

```yaml
fixture_id: us1_basic_coordination
description: >
  Replays the cookbook coordination workflow with deterministic seeds.
workflow_template:
  id: us1-basic-coordination
  version: v1
  agents:
    - id: researcher
      runtime_ref: agent://researcher
    - id: strategist
      runtime_ref: agent://strategist
  entry_points: [researcher]
user_inputs:
  - role: user
    content: "请提供关于 AI 中心城市的最新动态，并给出策略建议。"
    seed: 42
    timestamp: "2025-11-19T00:00:00Z"
tool_responses:
  - tool_name: search_perplexity
    run_id: run-001
    outputs:
      value: "Perplexity 检索结果......"
expected_assertions:
  - type: equality
    path: outputs[0].content
    expected: "完成摘要与策略建议。"
  - type: tolerance
    path: metrics.tokens.prompt
    expected: 1200
    tolerance: 50
```

## 示例工具响应（JSON）

```json
{
  "tool_name": "knowledge_arxiv",
  "run_id": "mock-arxiv-001",
  "outputs": {
    "papers": [
      {"title": "LLM Orchestration", "year": 2024, "score": 0.91},
      {"title": "Go Runtime for AI Agents", "year": 2025, "score": 0.87}
    ]
  },
  "metadata": {
    "latency_ms": 512,
    "cached": true
  }
}
```

## 校验流程

1. 使用 `yq eval` 或 `python -m jsonschema` 对照 `templates/parity_fixture.schema.json`（待补充）进行 schema 验证。
2. 将 fixture 文件传入 `scripts/ci/cross-language-parity.sh --fixture <path>`，确保 Python 与 Go 输出的 `RunOutput`、`SessionRecord`、`Telemetry` 与断言匹配。
3. 若断言需要跨场景复用，可将公共片段拆分到 `assets/expected_assertions/`，并在 fixture 中通过 `!include` 或脚本支持的 `@` 语法引用。

## 命名规范

- 文件名使用 `<cookbook-scenario>_<variation>.yaml`；若包含多语言，可追加 `-en`/`-zh`。
- `tool_responses` 子目录中的 JSON 应与 `tool_name` 同名，便于脚本自动查找。
- 所有时间戳统一为 UTC ISO-8601，文本采用 UTF-8，无注释。

## 贡献流程

1. 在 `assets/workflow_templates/` 放置 Workflow/Agent 只读快照，文件名对齐 `workflow_template.id`。
2. 添加新的 fixture YAML，并在 PR 描述中说明对应任务/场景。
3. 更新 `tasks.md` 中关联的任务项（如 T014）并附带测试命令输出，确保可复现。
