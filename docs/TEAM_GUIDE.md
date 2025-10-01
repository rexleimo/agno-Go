# 团队协作指南

## 工作流程

### 每日流程
1. **每日站会** (15分钟)
   - 昨天完成什么
   - 今天计划什么
   - 遇到什么阻碍

2. **编码时间** (集中工作)
   - 开发: 实现功能
   - 测试: 编写测试
   - 架构: 代码评审

3. **代码提交** (下班前)
   - 提交到功能分支
   - 触发 CI 检查

---

## 分工协作

### 架构师
**核心任务:**
- [ ] Week 1: 定义项目结构和接口规范
- [ ] Week 2-3: 评审核心代码实现
- [ ] Week 4-5: 评审 LLM 集成方案
- [ ] Week 6-7: 评审存储和 API 设计
- [ ] Week 8: 总结和文档完善

**每周产出:**
- 技术设计文档
- 代码评审报告
- 风险识别和解决方案

---

### Go 开发工程师
**核心任务:**
- [ ] Week 1: 搭建项目 + 核心数据结构
- [ ] Week 2: Agent 核心逻辑
- [ ] Week 3-4: Team/Workflow + LLM 集成
- [ ] Week 5-6: 工具系统 + 向量数据库
- [ ] Week 7: AgentOS API
- [ ] Week 8: 优化和 Bug 修复

**代码规范:**
```go
// 1. 每个包有 doc.go
// Package agent provides core agent functionality.
package agent

// 2. 公开函数必须有注释
// NewAgent creates a new agent with the given configuration.
func NewAgent(config *Config) (*Agent, error) {
    // ...
}

// 3. 错误处理
if err != nil {
    return nil, fmt.Errorf("failed to create agent: %w", err)
}

// 4. 使用 context
func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // ...
    }
}
```

---

### 测试工程师
**核心任务:**
- [ ] Week 1: CI/CD 搭建 + 测试框架
- [ ] Week 2-7: 跟进开发,编写测试
- [ ] Week 8: 完整回归 + 性能报告

**测试策略:**
1. **单元测试** (60%)
   - 每个功能模块
   - 边界条件
   - 错误处理

2. **集成测试** (30%)
   - Agent + Model
   - Agent + Tools
   - 多组件协作

3. **端到端测试** (10%)
   - 完整场景
   - API 测试

**测试模板:**
```go
func TestAgent_Run(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "simple query",
            input: "hello",
            want:  "hi there",
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            agent := NewAgent(mockConfig())
            got, err := agent.Run(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            // 更多断言...
        })
    }
}
```

---

## Git 工作流

### 分支策略
```
main (protected)
  ↓
develop
  ↓
  ├── feature/agent-core (开发)
  ├── feature/llm-openai (开发)
  └── feature/tools-http  (开发)
```

### 提交规范
```bash
# 格式
<type>(<scope>): <subject>

# 示例
feat(agent): add Run method
fix(models): fix openai timeout issue
test(agent): add unit tests for memory
docs(readme): update installation guide

# Type
feat:     新功能
fix:      Bug修复
test:     测试
docs:     文档
refactor: 重构
perf:     性能优化
```

### PR 流程
1. **创建 PR**
   ```bash
   git checkout -b feature/my-feature
   git commit -m "feat(scope): add feature"
   git push origin feature/my-feature
   # 在 GitHub 创建 PR
   ```

2. **PR 检查项**
   - [ ] CI 通过
   - [ ] 测试覆盖率不降低
   - [ ] 代码有注释
   - [ ] 架构师已评审

3. **合并**
   - 使用 Squash Merge
   - 删除功能分支

---

## 沟通机制

### 技术讨论
- **同步会议**: 复杂问题 (视频/面对面)
- **异步讨论**: 简单问题 (文档注释/Issue)

### 决策流程
1. 提出问题 (任何人)
2. 讨论方案 (团队)
3. 架构师决策
4. 记录到文档

### 文档更新
- 代码变更 → 更新文档
- 新功能 → 更新 README 和示例
- 架构变化 → 更新 ARCHITECTURE.md

---

## 开发环境

### 必装工具
```bash
# Go 1.21+
go version

# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 测试工具
go install github.com/stretchr/testify@latest
```

### 编辑器配置
```json
// VSCode settings.json
{
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "editor.formatOnSave": true,
    "go.testFlags": ["-v", "-race"]
}
```

### 常用命令
```bash
# 运行测试
make test

# 代码检查
make lint

# 构建
make build

# 查看覆盖率
make coverage
```

---

## 问题处理

### 遇到阻碍
1. **先自己尝试** (30分钟)
2. **查文档/代码** (30分钟)
3. **询问团队** (立即)

### Bug 处理
1. 创建 Issue
2. 分配责任人
3. 关联 PR
4. 添加测试防止回归

### 性能问题
1. 使用 pprof 分析
   ```bash
   go test -cpuprofile=cpu.prof -bench=.
   go tool pprof cpu.prof
   ```
2. 找到瓶颈
3. 优化并验证

---

## 质量检查清单

### 代码提交前
- [ ] 通过 `go test ./...`
- [ ] 通过 `golangci-lint run`
- [ ] 通过 `go vet ./...`
- [ ] 添加必要注释
- [ ] 更新相关文档

### PR 合并前
- [ ] CI 全部通过
- [ ] 至少一人代码评审
- [ ] 测试覆盖率 ≥70%
- [ ] 架构师批准 (重要功能)

### 发布前
- [ ] 所有测试通过
- [ ] 性能测试达标
- [ ] 文档完整
- [ ] CHANGELOG 更新

---

## 效率提升

### 复用代码
- 建立内部工具库
- 使用代码生成 (go generate)
- 共享测试工具

### 自动化
- CI/CD 自动测试
- 自动生成文档
- 自动发布版本

### 知识共享
- 每周技术分享 (15分钟)
- 记录重要决策
- 维护 FAQ 文档

---

**保持沟通,快速迭代**
