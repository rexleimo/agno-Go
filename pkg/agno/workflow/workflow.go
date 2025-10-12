package workflow

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/rexleimo/agno-go/pkg/agno/types"
)

// Workflow represents a multi-step process
// Workflow 代表多步骤流程
type Workflow struct {
	ID     string
	Name   string
	Steps  []Node
	logger *slog.Logger

	// History-related fields
	// 历史相关字段
	enableHistory     bool
	historyStore      WorkflowStorage
	numHistoryRuns    int
	addHistoryToSteps bool
}

// Node represents a node in the workflow graph
type Node interface {
	Execute(ctx context.Context, input *ExecutionContext) (*ExecutionContext, error)
	GetID() string
	GetType() NodeType
}

// NodeType represents the type of workflow node
type NodeType string

const (
	NodeTypeStep      NodeType = "step"
	NodeTypeCondition NodeType = "condition"
	NodeTypeLoop      NodeType = "loop"
	NodeTypeParallel  NodeType = "parallel"
	NodeTypeRouter    NodeType = "router"
)

// Config contains workflow configuration
// Config 包含工作流配置
type Config struct {
	ID     string
	Name   string
	Steps  []Node
	Logger *slog.Logger

	// EnableHistory enables workflow history tracking
	// EnableHistory 启用工作流历史跟踪
	EnableHistory bool `json:"enable_history"`

	// HistoryStore is the storage backend for workflow history
	// HistoryStore 是工作流历史的存储后端
	HistoryStore WorkflowStorage `json:"-"`

	// NumHistoryRuns is the number of recent runs to include in history context
	// NumHistoryRuns 是历史上下文中包含的最近运行数量
	NumHistoryRuns int `json:"num_history_runs"`

	// AddHistoryToSteps automatically adds history context to all steps
	// AddHistoryToSteps 自动将历史上下文添加到所有步骤
	AddHistoryToSteps bool `json:"add_history_to_steps"`
}

// New creates a new workflow
// New 创建新的工作流
func New(config Config) (*Workflow, error) {
	if config.ID == "" {
		config.ID = fmt.Sprintf("workflow-%s", config.Name)
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	// 历史配置验证和默认值
	// History configuration validation and defaults
	if config.EnableHistory && config.HistoryStore == nil {
		// 使用默认内存存储
		// Use default memory storage
		config.HistoryStore = NewMemoryStorage(100)
	}

	if config.NumHistoryRuns <= 0 {
		config.NumHistoryRuns = 3 // 默认包含最近 3 次运行
	}

	return &Workflow{
		ID:                config.ID,
		Name:              config.Name,
		Steps:             config.Steps,
		logger:            config.Logger,
		enableHistory:     config.EnableHistory,
		historyStore:      config.HistoryStore,
		numHistoryRuns:    config.NumHistoryRuns,
		addHistoryToSteps: config.AddHistoryToSteps,
	}, nil
}

// Run executes the workflow
// Run 执行工作流
// sessionID 参数可选,为空则自动生成
func (w *Workflow) Run(ctx context.Context, input string, sessionID string) (*ExecutionContext, error) {
	// 验证输入
	// Validate input
	if input == "" {
		return nil, types.NewInvalidInputError("input cannot be empty", nil)
	}

	// 生成 session ID（如果未提供）
	// Generate session ID if not provided
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	w.logger.Info("workflow started",
		"workflow_id", w.ID,
		"session_id", sessionID,
		"steps", len(w.Steps))

	// 创建执行上下文
	// Create execution context
	execCtx := NewExecutionContextWithSession(input, sessionID, "")

	// 加载历史（如果启用）
	// Load history (if enabled)
	if w.enableHistory && w.historyStore != nil {
		if err := w.loadHistory(ctx, execCtx); err != nil {
			w.logger.Error("failed to load history", "error", err)
			// 不阻止执行，仅记录错误
			// Don't block execution, just log error
		}
	}

	// 传递历史配置到执行上下文
	// Pass history configuration to execution context
	if w.enableHistory {
		execCtx.SetSessionState("workflow_history_config", &WorkflowHistoryConfig{
			AddHistoryToSteps: w.addHistoryToSteps,
			NumHistoryRuns:    w.numHistoryRuns,
		})
	}

	// 创建 WorkflowRun 记录
	// Create WorkflowRun record
	var workflowRun *WorkflowRun
	if w.enableHistory {
		runID := generateRunID()
		workflowRun = NewWorkflowRun(runID, sessionID, w.ID, input)
		workflowRun.MarkStarted()
	}

	// 执行步骤
	// Execute steps
	for i, step := range w.Steps {
		select {
		case <-ctx.Done():
			if workflowRun != nil {
				workflowRun.MarkCancelled()
				w.saveRun(ctx, sessionID, workflowRun)
			}
			return nil, ctx.Err()
		default:
		}

		w.logger.Info("executing step",
			"step_id", step.GetID(),
			"step_type", step.GetType(),
			"sequence", i+1)

		result, err := step.Execute(ctx, execCtx)
		if err != nil {
			w.logger.Error("step execution failed",
				"step_id", step.GetID(),
				"error", err)

			if workflowRun != nil {
				workflowRun.MarkFailed(err)
				w.saveRun(ctx, sessionID, workflowRun)
			}

			return nil, types.NewError(types.ErrCodeUnknown, fmt.Sprintf("step %s failed", step.GetID()), err)
		}

		execCtx = result
	}

	// 保存历史（如果启用）
	// Save history (if enabled)
	if workflowRun != nil {
		workflowRun.MarkCompleted(execCtx.Output)
		workflowRun.Messages = extractMessages(execCtx)
		w.saveRun(ctx, sessionID, workflowRun)
	}

	w.logger.Info("workflow completed",
		"workflow_id", w.ID,
		"session_id", sessionID)

	return execCtx, nil
}

// AddStep adds a step to the workflow
func (w *Workflow) AddStep(step Node) {
	w.Steps = append(w.Steps, step)
}

// loadHistory 从存储加载历史并添加到上下文
// loadHistory loads history from storage and adds to context
func (w *Workflow) loadHistory(ctx context.Context, execCtx *ExecutionContext) error {
	if w.historyStore == nil {
		return nil
	}

	// 获取或创建 session
	// Get or create session
	session, err := w.historyStore.GetSession(ctx, execCtx.SessionID)
	if err != nil {
		if err == ErrSessionNotFound {
			// 创建新 session
			// Create new session
			session, err = w.historyStore.CreateSession(
				ctx,
				execCtx.SessionID,
				w.ID,
				execCtx.UserID,
			)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get session: %w", err)
		}
	}

	// 获取历史记录
	// Get history
	history := session.GetHistory(w.numHistoryRuns)
	if len(history) == 0 {
		return nil
	}

	// 直接设置到 ExecutionContext
	// Set directly to ExecutionContext
	execCtx.SetWorkflowHistory(history)

	// 如果配置了添加历史到步骤
	// If configured to add history to steps
	if w.addHistoryToSteps {
		historyContext := session.GetHistoryContext(w.numHistoryRuns)
		execCtx.SetHistoryContext(historyContext)

		// 也保存到 session state (向后兼容)
		// Also save to session state (backward compatibility)
		execCtx.SetSessionState("workflow_history_context", historyContext)
	}

	// 将历史数据存储在上下文中（向后兼容）
	// Store history data in context (backward compatibility)
	execCtx.SetSessionState("workflow_history", history)
	execCtx.SetSessionState("workflow_session", session)

	w.logger.Debug("loaded history",
		"session_id", execCtx.SessionID,
		"history_count", len(history))

	return nil
}

// saveRun 保存运行记录到存储
// saveRun saves run record to storage
func (w *Workflow) saveRun(ctx context.Context, sessionID string, run *WorkflowRun) error {
	if w.historyStore == nil {
		return nil
	}

	// 获取 session
	// Get session
	session, err := w.historyStore.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// 添加运行记录
	// Add run record
	session.AddRun(run)

	// 更新 session
	// Update session
	if err := w.historyStore.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	w.logger.Debug("saved run",
		"session_id", sessionID,
		"run_id", run.RunID,
		"status", run.Status)

	return nil
}

// extractMessages 从执行上下文提取消息
// extractMessages extracts messages from execution context
func extractMessages(execCtx *ExecutionContext) []*types.Message {
	// 从 session state 中提取消息历史
	// Extract message history from session state
	if messages, ok := execCtx.GetSessionState("messages"); ok {
		if msgList, ok := messages.([]*types.Message); ok {
			return msgList
		}
	}

	// 如果没有消息历史，创建基本的输入/输出消息
	// If no message history, create basic input/output messages
	return []*types.Message{
		types.NewUserMessage(execCtx.Input),
		types.NewAssistantMessage(execCtx.Output),
	}
}

// generateSessionID 生成唯一的 session ID
// generateSessionID generates a unique session ID
func generateSessionID() string {
	return "session-" + uuid.New().String()
}

// generateRunID 生成唯一的 run ID
// generateRunID generates a unique run ID
func generateRunID() string {
	return "run-" + uuid.New().String()
}
