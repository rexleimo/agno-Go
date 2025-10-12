package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers A2A endpoints on the Gin router
// RegisterRoutes 在 Gin 路由器上注册 A2A 端点
func (a *A2AInterface) RegisterRoutes(router *gin.Engine) {
	group := router.Group(a.prefix)

	// POST /a2a/message/send - Non-streaming message endpoint
	// POST /a2a/message/send - 非流式消息端点
	group.POST("/message/send", a.HandleSendMessage)

	// POST /a2a/message/stream - Streaming message endpoint
	// POST /a2a/message/stream - 流式消息端点
	group.POST("/message/stream", a.HandleStreamMessage)
}

// HandleSendMessage handles non-streaming message requests
// HandleSendMessage 处理非流式消息请求
func (a *A2AInterface) HandleSendMessage(c *gin.Context) {
	var req JSONRPC2Request

	// Parse request body
	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("", ParseError, "Failed to parse JSON"))
		return
	}

	// Validate request
	// 验证请求
	if err := ValidateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(req.ID, InvalidRequest, err.Error()))
		return
	}

	// Find entity (Agent/Team/Workflow)
	// 查找实体（Agent/Team/Workflow）
	entity, err := a.FindEntity(req.Params.Message.AgentID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(req.ID, ServerError, err.Error()))
		return
	}

	// Map A2A request to RunInput
	// 将 A2A 请求映射到 RunInput
	runInput, err := MapA2ARequestToRunInput(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(req.ID, InvalidParams, err.Error()))
		return
	}

	// Execute entity
	// 执行实体
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	result, err := entity.Run(ctx, runInput.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(req.ID, InternalError, err.Error()))
		return
	}

	// Convert result to RunOutput
	// 将结果转换为 RunOutput
	runOutput, err := convertResultToRunOutput(result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(req.ID, InternalError,
			fmt.Sprintf("failed to convert result: %v", err)))
		return
	}

	// Map RunOutput to A2A Task
	// 将 RunOutput 映射到 A2A Task
	task := MapRunOutputToTask(runOutput, &req.Params.Message)

	// Send success response
	// 发送成功响应
	c.JSON(http.StatusOK, JSONRPC2Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: &Result{
			Task: *task,
		},
	})
}

// HandleStreamMessage handles streaming message requests using Server-Sent Events
// HandleStreamMessage 使用服务器发送事件处理流式消息请求
func (a *A2AInterface) HandleStreamMessage(c *gin.Context) {
	var req JSONRPC2Request

	// Parse request body
	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("", ParseError, "Failed to parse JSON"))
		return
	}

	// Validate request
	// 验证请求
	if err := ValidateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(req.ID, InvalidRequest, err.Error()))
		return
	}

	// Find entity
	// 查找实体
	entity, err := a.FindEntity(req.Params.Message.AgentID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse(req.ID, ServerError, err.Error()))
		return
	}

	// Map A2A request to RunInput
	// 将 A2A 请求映射到 RunInput
	runInput, err := MapA2ARequestToRunInput(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(req.ID, InvalidParams, err.Error()))
		return
	}

	// Set headers for SSE
	// 设置 SSE 头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Execute entity in background
	// 在后台执行实体
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	// For now, execute and send result (full streaming support requires event system)
	// 目前执行并发送结果（完整流式支持需要事件系统）
	result, err := entity.Run(ctx, runInput.Content)
	if err != nil {
		// Send error event
		// 发送错误事件
		errorEvent := TaskStatusUpdate{
			TaskID:    generateTaskID(),
			ContextID: req.Params.Message.ContextID,
			Status:    TaskStatusFailed,
			Error: &TaskError{
				Code:    "execution_error",
				Message: err.Error(),
			},
		}
		sendSSEEvent(c.Writer, "error", errorEvent)
		return
	}

	// Convert result and send success event
	// 转换结果并发送成功事件
	runOutput, err := convertResultToRunOutput(result)
	if err != nil {
		errorEvent := TaskStatusUpdate{
			TaskID:    generateTaskID(),
			ContextID: req.Params.Message.ContextID,
			Status:    TaskStatusFailed,
			Error: &TaskError{
				Code:    "conversion_error",
				Message: fmt.Sprintf("failed to convert result: %v", err),
			},
		}
		sendSSEEvent(c.Writer, "error", errorEvent)
		return
	}

	// Map to task and send completion event
	// 映射到任务并发送完成事件
	task := MapRunOutputToTask(runOutput, &req.Params.Message)

	completionEvent := TaskStatusUpdate{
		TaskID:    task.ID,
		ContextID: task.ContextID,
		Status:    TaskStatusCompleted,
		Message:   &task.History[len(task.History)-1], // Last message is agent's response
	}

	sendSSEEvent(c.Writer, "task_completed", completionEvent)
	c.Writer.Flush()
}

// Helper functions
// 辅助函数

// errorResponse creates a JSON-RPC 2.0 error response
// errorResponse 创建 JSON-RPC 2.0 错误响应
func errorResponse(id string, code int, message string) JSONRPC2Response {
	return JSONRPC2Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
}

// sendSSEEvent sends a Server-Sent Event
// sendSSEEvent 发送服务器发送事件
func sendSSEEvent(w http.ResponseWriter, event string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"failed to marshal data\"}\n\n")
		return
	}

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
}

// convertResultToRunOutput converts entity execution result to RunOutput
// convertResultToRunOutput 将实体执行结果转换为 RunOutput
func convertResultToRunOutput(result interface{}) (*RunOutput, error) {
	// Handle different result types
	// 处理不同的结果类型
	switch v := result.(type) {
	case *RunOutput:
		// Already in correct format
		// 已经是正确格式
		return v, nil

	case string:
		// Simple string response
		// 简单字符串响应
		return &RunOutput{
			Content:  v,
			Metadata: make(map[string]interface{}),
		}, nil

	case map[string]interface{}:
		// Try to extract content from map
		// 尝试从 map 中提取内容
		content, ok := v["content"].(string)
		if !ok {
			// If no content field, convert entire map to string
			// 如果没有 content 字段，将整个 map 转换为字符串
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}
			content = string(jsonBytes)
		}

		return &RunOutput{
			Content:  content,
			Metadata: v,
		}, nil

	default:
		// Try to convert to JSON string
		// 尝试转换为 JSON 字符串
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("unsupported result type: %T", result)
		}

		return &RunOutput{
			Content:  string(jsonBytes),
			Metadata: make(map[string]interface{}),
		}, nil
	}
}
