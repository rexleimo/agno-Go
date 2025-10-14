package agent

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/rexleimo/agno-go/pkg/agno/hooks"
	"github.com/rexleimo/agno-go/pkg/agno/memory"
	"github.com/rexleimo/agno-go/pkg/agno/models"
	"github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
	"github.com/rexleimo/agno-go/pkg/agno/types"
)

// Agent represents an AI agent
type Agent struct {
	ID           string
	Name         string
	Model        models.Model
	Toolkits     []toolkit.Toolkit
	Memory       memory.Memory
	Instructions string
	MaxLoops     int          // Maximum tool calling loops
	UserID       string       // User ID for multi-tenant memory isolation / 多租户内存隔离的用户ID
	PreHooks     []hooks.Hook // Hooks executed before processing input
	PostHooks    []hooks.Hook // Hooks executed after generating output
	logger       *slog.Logger

	// Storage control / 存储控制
	storeToolMessages    bool // Whether to store tool messages in RunOutput / 是否在 RunOutput 中存储工具消息
	storeHistoryMessages bool // Whether to store history messages in RunOutput / 是否在 RunOutput 中存储历史消息

	// Temporary instructions support for workflow history injection
	// 临时 instructions 支持,用于工作流历史注入
	tempInstructions string       // Temporary instructions (single execution only) / 临时指令（仅单次执行）
	instructionsMu   sync.RWMutex // Protects instructions modification / 保护指令修改
}

// Config contains agent configuration
type Config struct {
	ID           string
	Name         string
	Model        models.Model
	Toolkits     []toolkit.Toolkit
	Memory       memory.Memory
	Instructions string
	MaxLoops     int
	UserID       string       // User ID for multi-tenant scenarios / 多租户场景的用户ID
	PreHooks     []hooks.Hook // Hooks to execute before processing input
	PostHooks    []hooks.Hook // Hooks to execute after generating output
	Logger       *slog.Logger

	// Storage control flags (nil means use default: true) / 存储控制标志 (nil 表示使用默认值: true)
	// StoreToolMessages controls whether tool-related messages are included in RunOutput.
	// When false, tool messages and tool-related fields are filtered from output.
	// StoreToolMessages 控制是否在 RunOutput 中包含工具相关消息
	// 当为 false 时，工具消息和工具相关字段会从输出中过滤
	StoreToolMessages *bool

	// StoreHistoryMessages controls whether historical messages (from Memory) are included in RunOutput.
	// When false, only messages generated during the current Run are included.
	// StoreHistoryMessages 控制是否在 RunOutput 中包含历史消息(来自 Memory)
	// 当为 false 时，仅包含当前 Run 生成的消息
	StoreHistoryMessages *bool
}

// New creates a new agent
func New(config Config) (*Agent, error) {
	if config.Model == nil {
		return nil, types.NewInvalidConfigError("model is required", nil)
	}

	if config.ID == "" {
		config.ID = fmt.Sprintf("agent-%s", config.Model.GetID())
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	if config.Memory == nil {
		config.Memory = memory.NewInMemory(100)
	}

	if config.MaxLoops <= 0 {
		config.MaxLoops = 10
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	// Helper function to get bool value or default / 辅助函数获取布尔值或默认值
	boolOrDefault := func(ptr *bool, defaultVal bool) bool {
		if ptr == nil {
			return defaultVal
		}
		return *ptr
	}

	agent := &Agent{
		ID:           config.ID,
		Name:         config.Name,
		Model:        config.Model,
		Toolkits:     config.Toolkits,
		Memory:       config.Memory,
		Instructions: config.Instructions,
		MaxLoops:     config.MaxLoops,
		UserID:       config.UserID,
		PreHooks:     config.PreHooks,
		PostHooks:    config.PostHooks,
		logger:       config.Logger,

		// Storage control (default to true for backward compatibility) / 存储控制 (默认为 true 以保持向后兼容)
		storeToolMessages:    boolOrDefault(config.StoreToolMessages, true),
		storeHistoryMessages: boolOrDefault(config.StoreHistoryMessages, true),
	}

	// Add system message if instructions provided
	// 如果提供了指令则添加系统消息
	if config.Instructions != "" {
		agent.Memory.Add(types.NewSystemMessage(config.Instructions), config.UserID)
	}

	return agent, nil
}

// RunOutput contains the result of agent execution
type RunOutput struct {
	Content  string                 `json:"content"`
	Messages []*types.Message       `json:"messages"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Run executes the agent with the given input
func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error) {
	// Ensure temporary instructions are cleared after execution (even on early return)
	// 确保执行完成后清除临时指令（即使提前返回）
	defer a.ClearTempInstructions()

	if input == "" {
		return nil, types.NewInvalidInputError("input cannot be empty", nil)
	}

	// Get current instructions at execution start (temporary or permanent)
	// 在执行开始时获取当前指令（临时或永久）
	currentInstructions := a.GetInstructions()

	a.logger.Info("agent run started", "agent_id", a.ID, "input", input)

	// Record initial message count for history filtering / 记录初始消息数量用于历史过滤
	initialMessageCount := len(a.Memory.GetMessages(a.UserID))

	// Execute pre-hooks
	if len(a.PreHooks) > 0 {
		a.logger.Debug("executing pre-hooks", "count", len(a.PreHooks))
		hookInput := hooks.NewHookInput(input).
			WithAgentID(a.ID).
			WithMessages([]interface{}{})

		if err := hooks.ExecuteHooks(ctx, a.PreHooks, hookInput); err != nil {
			a.logger.Error("pre-hook failed", "error", err)
			return nil, types.NewInputCheckError("pre-hook validation failed", err)
		}
	}

	// Add user message
	// 添加用户消息
	userMsg := types.NewUserMessage(input)
	a.Memory.Add(userMsg, a.UserID)

	var finalResponse *types.ModelResponse
	loopCount := 0

	// Tool calling loop
	for loopCount < a.MaxLoops {
		loopCount++

		// Prepare request with current instructions
		// 使用当前指令准备请求
		messages := a.Memory.GetMessages(a.UserID)

		// If using temporary instructions, update system message
		// 如果使用临时指令，更新系统消息
		if currentInstructions != a.Instructions && currentInstructions != "" {
			// Replace or prepend system message with current instructions
			// 用当前指令替换或添加系统消息
			messages = a.updateSystemMessage(messages, currentInstructions)
		}

		req := &models.InvokeRequest{
			Messages: messages,
		}

		// Add tools if available
		if len(a.Toolkits) > 0 {
			req.Tools = toolkit.ToModelToolDefinitions(a.Toolkits)
		}

		// Call model
		resp, err := a.Model.Invoke(ctx, req)
		if err != nil {
			a.logger.Error("model invocation failed", "error", err)
			return nil, types.NewAPIError("model invocation failed", err)
		}

		// Store assistant response
		// 存储助手响应
		assistantMsg := &types.Message{
			Role:      types.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		a.Memory.Add(assistantMsg, a.UserID)

		// Check if there are tool calls
		if !resp.HasToolCalls() {
			finalResponse = resp
			break
		}

		// Execute tool calls
		a.logger.Info("executing tool calls", "count", len(resp.ToolCalls))
		if err := a.executeToolCalls(ctx, resp.ToolCalls); err != nil {
			a.logger.Error("tool execution failed", "error", err)
			return nil, types.NewToolExecutionError("tool execution failed", err)
		}

		// Continue loop to process tool results
	}

	if loopCount >= a.MaxLoops {
		a.logger.Warn("max loops reached", "max_loops", a.MaxLoops)
		return nil, types.NewError(types.ErrCodeUnknown, "max tool calling loops reached", nil)
	}

	if finalResponse == nil {
		return nil, types.NewError(types.ErrCodeUnknown, "no response from model", nil)
	}

	// Execute post-hooks
	if len(a.PostHooks) > 0 {
		a.logger.Debug("executing post-hooks", "count", len(a.PostHooks))
		hookInput := hooks.NewHookInput(input).
			WithOutput(finalResponse.Content).
			WithAgentID(a.ID).
			WithMessages([]interface{}{})

		if err := hooks.ExecuteHooks(ctx, a.PostHooks, hookInput); err != nil {
			a.logger.Error("post-hook failed", "error", err)
			return nil, types.NewOutputCheckError("post-hook validation failed", err)
		}
	}

	a.logger.Info("agent run completed", "agent_id", a.ID)

	// Build output / 构建输出
	output := &RunOutput{
		Content:  finalResponse.Content,
		Messages: a.Memory.GetMessages(a.UserID),
		Metadata: map[string]interface{}{
			"loops": loopCount,
			"usage": finalResponse.Usage,
		},
	}

	// Apply storage filters / 应用存储过滤器
	a.scrubRunOutputWithContext(output, initialMessageCount)

	return output, nil
}

// executeToolCalls executes all tool calls and adds results to memory
func (a *Agent) executeToolCalls(ctx context.Context, toolCalls []types.ToolCall) error {
	for _, tc := range toolCalls {
		// Find the toolkit that has this function
		var targetToolkit toolkit.Toolkit
		for _, tk := range a.Toolkits {
			if _, exists := tk.Functions()[tc.Function.Name]; exists {
				targetToolkit = tk
				break
			}
		}

		if targetToolkit == nil {
			errMsg := fmt.Sprintf("function %s not found in any toolkit", tc.Function.Name)
			a.logger.Warn("tool not found", "function", tc.Function.Name)
			a.Memory.Add(types.NewToolMessage(tc.ID, errMsg), a.UserID)
			continue
		}

		// Parse arguments
		args, err := toolkit.ParseArguments(tc.Function.Arguments)
		if err != nil {
			errMsg := fmt.Sprintf("failed to parse arguments: %v", err)
			a.logger.Error("argument parsing failed", "error", err)
			a.Memory.Add(types.NewToolMessage(tc.ID, errMsg), a.UserID)
			continue
		}

		// Execute tool
		a.logger.Info("executing tool", "function", tc.Function.Name, "args", args)

		// Get the function and execute it directly
		fn := targetToolkit.Functions()[tc.Function.Name]
		if fn == nil {
			errMsg := fmt.Sprintf("function %s not found", tc.Function.Name)
			a.logger.Error("function not found", "function", tc.Function.Name)
			a.Memory.Add(types.NewToolMessage(tc.ID, errMsg), a.UserID)
			continue
		}

		result, err := fn.Handler(ctx, args)
		if err != nil {
			errMsg := fmt.Sprintf("tool execution error: %v", err)
			a.logger.Error("tool execution failed", "function", tc.Function.Name, "error", err)
			a.Memory.Add(types.NewToolMessage(tc.ID, errMsg), a.UserID)
			continue
		}

		// Format and store result
		resultStr, err := toolkit.FormatResult(result)
		if err != nil {
			resultStr = fmt.Sprintf("%v", result)
		}

		a.logger.Info("tool executed successfully", "function", tc.Function.Name)
		a.Memory.Add(types.NewToolMessage(tc.ID, resultStr), a.UserID)
	}

	return nil
}

// ClearMemory clears the agent's conversation history for this user
// ClearMemory 清除此用户的Agent对话历史
func (a *Agent) ClearMemory() {
	a.Memory.Clear(a.UserID)
	// Re-add system message
	// 重新添加系统消息
	if a.Instructions != "" {
		a.Memory.Add(types.NewSystemMessage(a.Instructions), a.UserID)
	}
}

// GetID returns the agent ID
// GetID 返回 agent ID
func (a *Agent) GetID() string {
	return a.ID
}

// GetInstructions returns the current instructions (temporary or permanent)
// GetInstructions 返回当前指令（临时或永久）
func (a *Agent) GetInstructions() string {
	a.instructionsMu.RLock()
	defer a.instructionsMu.RUnlock()

	// Temporary instructions take precedence
	// 临时指令优先
	if a.tempInstructions != "" {
		return a.tempInstructions
	}
	return a.Instructions
}

// SetInstructions permanently sets the agent's instructions
// SetInstructions 永久设置 agent 的指令
func (a *Agent) SetInstructions(instructions string) {
	a.instructionsMu.Lock()
	defer a.instructionsMu.Unlock()

	a.Instructions = instructions
}

// SetTempInstructions temporarily sets instructions (only affects next Run)
// SetTempInstructions 临时设置指令（仅影响下一次 Run）
func (a *Agent) SetTempInstructions(instructions string) {
	a.instructionsMu.Lock()
	defer a.instructionsMu.Unlock()

	a.tempInstructions = instructions
}

// ClearTempInstructions clears temporary instructions
// ClearTempInstructions 清除临时指令
func (a *Agent) ClearTempInstructions() {
	a.instructionsMu.Lock()
	defer a.instructionsMu.Unlock()

	a.tempInstructions = ""
}

// updateSystemMessage updates or adds system message with new instructions
// updateSystemMessage 更新或添加带有新指令的系统消息
func (a *Agent) updateSystemMessage(messages []*types.Message, instructions string) []*types.Message {
	if len(messages) == 0 {
		return []*types.Message{types.NewSystemMessage(instructions)}
	}

	// Create a copy to avoid modifying the original
	// 创建副本以避免修改原始消息
	result := make([]*types.Message, 0, len(messages)+1)

	// Check if first message is system message
	// 检查第一条消息是否为系统消息
	systemMessageFound := false
	for i, msg := range messages {
		if i == 0 && msg.Role == types.RoleSystem {
			// Replace first system message
			// 替换第一条系统消息
			result = append(result, types.NewSystemMessage(instructions))
			systemMessageFound = true
		} else {
			result = append(result, msg)
		}
	}

	// If no system message found, prepend one
	// 如果没有找到系统消息，添加一个到开头
	if !systemMessageFound {
		result = append([]*types.Message{types.NewSystemMessage(instructions)}, result...)
	}

	return result
}

// filterToolMessages removes tool-related messages from the slice.
// It filters out messages with Role == RoleTool and clears tool-related fields from other messages.
// filterToolMessages 从切片中移除工具相关消息
// 它会过滤掉 Role == RoleTool 的消息，并清除其他消息中的工具相关字段
func (a *Agent) filterToolMessages(messages []*types.Message) []*types.Message {
	if len(messages) == 0 {
		return messages
	}

	// Pre-allocate with same capacity for efficiency / 预分配相同容量以提高效率
	filtered := make([]*types.Message, 0, len(messages))

	for _, msg := range messages {
		// Skip tool response messages entirely / 完全跳过工具响应消息
		if msg.Role == types.RoleTool {
			continue
		}

		// For other messages, clear tool-related fields / 对于其他消息，清除工具相关字段
		if len(msg.ToolCalls) > 0 || msg.ToolCallID != "" {
			// Create a shallow copy to avoid modifying the original message in Memory
			// 创建浅拷贝以避免修改 Memory 中的原始消息
			msgCopy := *msg
			msgCopy.ToolCalls = nil
			msgCopy.ToolCallID = ""
			filtered = append(filtered, &msgCopy)
		} else {
			// No tool data, can use original message / 没有工具数据，可以使用原始消息
			filtered = append(filtered, msg)
		}
	}

	return filtered
}

// filterHistoryMessages removes messages that existed before the current Run.
// It uses the initialCount to determine which messages are historical.
// filterHistoryMessages 移除当前 Run 之前就存在的消息
// 它使用 initialCount 来确定哪些消息是历史消息
func (a *Agent) filterHistoryMessages(messages []*types.Message, initialCount int) []*types.Message {
	// Defensive: handle nil / 防御性: 处理 nil
	if messages == nil {
		return nil
	}

	// Defensive: handle empty / 防御性: 处理空
	if len(messages) == 0 {
		return messages
	}

	// Defensive: handle negative count / 防御性: 处理负数
	if initialCount < 0 {
		initialCount = 0
	}

	// All messages are new / 所有消息都是新的
	if initialCount == 0 {
		return messages
	}

	// All messages are historical / 所有消息都是历史的
	if initialCount >= len(messages) {
		return []*types.Message{}
	}

	// Return new messages (after initialCount) / 返回新消息（initialCount 之后的）
	return messages[initialCount:]
}

// scrubRunOutputWithContext applies filters to RunOutput based on storage configuration.
// It modifies the output in place for performance.
// scrubRunOutputWithContext 根据存储配置对 RunOutput 应用过滤器
// 为了性能考虑，会原地修改输出
func (a *Agent) scrubRunOutputWithContext(output *RunOutput, initialMessageCount int) {
	if output == nil || output.Messages == nil {
		return
	}

	initialCount := len(output.Messages)

	// Filter tool messages first (order matters!) / 先过滤工具消息（顺序很重要！）
	if !a.storeToolMessages {
		// Count tool messages in history (before filtering) to adjust initialMessageCount
		// 计算历史消息中的工具消息数量（在过滤之前），以便调整 initialMessageCount
		toolMessagesInHistory := 0
		if initialMessageCount > 0 && initialMessageCount <= len(output.Messages) {
			for i := 0; i < initialMessageCount; i++ {
				if output.Messages[i].Role == types.RoleTool {
					toolMessagesInHistory++
				}
			}
		}

		output.Messages = a.filterToolMessages(output.Messages)

		// Adjust initialMessageCount by removing count of tool messages that were in history
		// 调整 initialMessageCount，减去历史中的工具消息数量
		if toolMessagesInHistory > 0 {
			initialMessageCount -= toolMessagesInHistory
			if initialMessageCount < 0 {
				initialMessageCount = 0
			}
		}

		initialCount = len(output.Messages)
	}

	// Then filter history messages / 然后过滤历史消息
	if !a.storeHistoryMessages {
		output.Messages = a.filterHistoryMessages(output.Messages, initialMessageCount)
	}

	// Log if messages were filtered / 如果消息被过滤则记录日志
	if len(output.Messages) < initialCount {
		a.logger.Debug("filtered messages from output",
			"original_count", initialCount,
			"filtered_count", len(output.Messages),
			"store_tool_messages", a.storeToolMessages,
			"store_history_messages", a.storeHistoryMessages,
		)
	}
}
