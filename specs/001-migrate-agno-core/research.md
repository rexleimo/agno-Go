# Phase 0 Research — Agno 核心模块 Go 迁移

## Session & Memory Persistence Compatibility
- **Decision**: 以 Python `agno.session.AgentSession`/`TeamSession` 的字段为基线，Go 侧暴露 `session.Store` 接口（CRUD + 搜索），由调用方注入 Sqlite/Postgres/Redis 驱动；默认提供内存实现，仅作为测试与示例使用。
- **Rationale**: Python 版本允许通过 `BaseDb` 实现自定义后端，因此 Go 需要相同的抽象以复用现有数据；接口驱动可避免在本迭代耦合具体数据库，还能在 parity 测试中注入 fake store 复现历史记录。
- **Alternatives considered**:
  1. **直接绑定 Sqlite**：实现简单，但会阻塞使用 Postgres/Redis 的团队，也与“运行期无 Python 依赖”目标不符，因为 Python 基线允许多种数据库。
  2. **引入 ORM（GORM/Ent）**：降低 CRUD 代码量，但增加依赖体积与学习成本，本阶段优先保持轻量接口。

## Telemetry Event Schema & Runtime Labeling
- **Decision**: 沿用 Python `agno.run.events` 中的事件类型（RunStarted、ReasoningStep、ToolCall、SessionSummary、RunCompleted），在 `go/internal/telemetry` 中新增结构化枚举，并强制 runtime 标签（`"runtime":"go"`），方便同一 sink 区分来源。
- **Rationale**: 规格要求 Telemetry 与 Python 等价。Go 侧已有 `telemetry.Event`，仅需扩展字段与 Recorder 接口就可满足需求，且 runtime 标签能帮助运营团队观察双栈行为。
- **Alternatives considered**:
  1. **新建 Go 专属事件名**：更贴合 Go 实现，但无法与 Python 仪表盘直接对齐。
  2. **依赖第三方 OTLP SDK**：标准化格式，但当前 repo 尚未锁定 OpenTelemetry 依赖，贸然引入将影响最小可行交付。

## Provider/Toolkit Capability Mapping
- **Decision**: 在 `go/providers` 中维持 provider registry（LLM、Tool、Retriever、Business API），并为 Toolkit/Knowledge/Memories 暴露 interface；使用 YAML/JSON manifest 表述 provider 能力，供 parity 测试加载。
- **Rationale**: Python 侧通过 `Toolkit`, `KnowledgeFilter`, `MemoryManager` 支持多种扩展，Go 需要对等接口以载入配置并在运行期注入依赖；manifest 便于对照测试与文档解释。
- **Alternatives considered**:
  1. **硬编码 providers**：减少抽象，但会阻碍第三方扩展和 cookbook 示例更新。
  2. **完全依赖用户实现接口**：灵活但缺少默认示例与 parity 保障；manifest 方案可在样板中提供起点并确保配置结构一致。

## Cross-language Parity Harness & Deterministic Runs
- **Decision**: `scripts/ci/cross-language-parity.sh` 将驱动 Python 与 Go 两个 CLI：Python 通过 `python -m agno.tests.contracts`, Go 通过 `go test ./go/... -run TestParity`；两端读取同一 YAML fixture（输入消息、工具响应、随机种子），输出标准 JSON，再由脚本比较。
- **Rationale**: Specs 要求 ≥99% 输出一致率。统一 fixture + JSON 序列化最易比较，也能在 CI 中复用。随机种子由 fixture 指定可避免 nondeterminism。
- **Alternatives considered**:
  1. **直接比较日志文本**：实现简单但容易被格式差异扰动。
  2. **引入 gRPC 桥接实时比较**：更实时但实现复杂、耦合更紧，不利于快速迭代。

