# Deviations Log - Go vs Python AgentOS

记录 Go 版与 Python 版行为差异、原因与补偿措施。每条差异需注明影响的 provider/接口与对应的 fixture。

## 基线治具生成（2025-11-22 via `go run ./go/scripts/gen_provider_baseline`）
- **已生成基线**：
  - Gemini：chat + embedding（`gemini-2.5-flash` / `text-embedding-004`）。
  - GLM4：chat 基线（`GLM-4-Flash-250414`）。
  - Groq：chat 基线（`llama-3.3-70b-versatile`）。
  - SiliconFlow：chat + embedding（`Qwen/Qwen2.5-7B-Instruct` / `BAAI/bge-large-zh-v1.5`）。
- **未生成/占位原因**：
  - Groq embedding：官方未提供 embedding 模型，永久跳过（保留占位）。
  - GLM4 embedding：接口返回 429（无免费额度），待充值后生成。
  - OpenAI、OpenRouter：缺失 API key（status=not-configured），保留占位治具。
  - Cerebras：chat 401 / embedding 404（鉴权失败），保留占位。
  - ModelScope：调用 EOF（远端不可达/未确认免费额度），保留占位。
  - Ollama：已拉取 `qwen3:4b` 与 `nomic-embed-text`，但 /api/chat 返回空内容且 openai 兼容格式不匹配，需适配后再生成，当前保留占位。

## 影响与补偿
- 契约/供应商测试：
  - 已生成基线的 provider 将进行内容/向量对比。
  - 未生成的 provider 将因 status=not-configured 或运行时错误被测试跳过（见 `specs/001-go-agno-rewrite/artifacts/coverage/providers.log`），待补齐 key/模型后重新生成并替换治具。
- 后续动作：补齐缺失 key 或更新模型 ID 后，可重新运行 `go run ./go/scripts/gen_provider_baseline` 生成完整基线，再执行 `make providers-test` 更新报告。
