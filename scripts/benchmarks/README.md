# Runtime Benchmark Artefacts

本目录统一存放 Go 与 Python 运行时的性能基线，便于 `scripts/go-ci.sh`、`scripts/ci/cross-language-parity.sh` 以及后续监控任务读取一致的数据格式。

## 目录结构

```text
scripts/benchmarks/
├── README.md
└── data/
    └── .gitkeep              # 数据文件示例：2025-11-19-us1-go.json
```

> 所有基线均使用 JSON，文件名建议为 `<date>-<scenario>-<runtime>.json`。

## 数据 schema

```json
{
  "scenario": "us1_basic_coordination",
  "runtime": "go",
  "commit": "00112233",
  "timestamp": "2025-11-19T12:00:00Z",
  "workload": {
    "type": "workflow",
    "concurrency": 100,
    "inputs": "specs/001-migrate-agno-core/fixtures/us1_basic_coordination.yaml"
  },
  "metrics": {
    "latency_ms": { "p50": 1200, "p95": 1700, "p99": 2100 },
    "tokens_per_second": 3200,
    "cpu_pct": 62.5,
    "rss_mb": 870,
    "tool_calls": 12
  },
  "python_baseline": {
    "latency_ms": { "p95": 2400 },
    "rss_mb": 1150,
    "tokens_per_second": 2850
  },
  "notes": [
    "runtime=go 标签写入 telemetry 成功",
    "偏差在 spec 容忍范围内"
  ]
}
```

字段约定：

- `scenario`：对齐 Cookbook/fixtures 命名。
- `runtime`：`python` 或 `go`。
- `commit`：产物对应的 Git commit（短 SHA）。
- `workload.concurrency`：并发数（需覆盖 1、10、100 三档）。
- `metrics`：关键指标，必须至少包含 `latency_ms`, `tokens_per_second`, `cpu_pct`, `rss_mb`。
- `python_baseline`：当 runtime=go 时必须存在，用于计算 “≤70% 延迟” 与 “RSS -25%” 目标。

## 采集命令示例

```bash
# Python baseline
python -m agno.tests.benchmarks.run \
  --scenario us1_basic_coordination \
  --output scripts/benchmarks/data/2025-11-19-us1-python.json

# Go benchmark
GO_TELEMETRY_EXPORTER=stdout \
go test ./go/agent -run TestUS1Bench -bench . -benchmem \
  | tee scripts/benchmarks/data/2025-11-19-us1-go.json
```

后续的 `scripts/benchmarks/collect_runtime_baselines.sh` 应该读取上述 JSON，将关键指标同步到监控系统或 PR 注释。

## 校验流程

1. 提交前运行 `jq -e ".'scenario'" <file>` 确保 JSON 有效。
2. 使用 `scripts/go-ci.sh` 聚合 `metrics` 并对比 `python_baseline`，若偏离 >10% 则阻断。
3. 在 PR 描述中附上 `latency_ms.p95`、`rss_mb` 与 Python 的差值，方便 reviewers 快速确认性能目标。
