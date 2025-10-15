package agentos

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// handleAgentRunStream 处理流式 Agent 运行请求（SSE）
// handleAgentRunStream handles streaming agent run requests (SSE)
func (s *Server) handleAgentRunStream(c *gin.Context) {
	agentID := c.Param("id")

	// 解析请求
	// Parse request
	var req AgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// 解析事件类型过滤器
	// Parse event type filter
	typesParam := c.Query("types")
	var types []string
	if typesParam != "" {
		types = strings.Split(typesParam, ",")
	}
	filter := NewEventFilter(types)

	// 获取 Agent
	// Get agent
	ag, err := s.agentRegistry.Get(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "agent_not_found",
			"message": fmt.Sprintf("agent with ID '%s' not found", agentID),
		})
		return
	}

	// 设置 SSE 响应头
	// Set SSE response headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用 nginx 缓冲

	// 创建事件通道
	// Create event channel
	eventChan := make(chan *Event, 10)
	errChan := make(chan error, 1)

	// 创建上下文
	// Create context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	// 发送运行开始事件
	// Send run start event
	startEvent := NewEvent(EventRunStart, RunStartData{
		Input:     req.Input,
		SessionID: req.SessionID,
	})
	startEvent.AgentID = agentID
	startEvent.SessionID = req.SessionID

	if filter.ShouldSend(startEvent) {
		if err := s.sendSSE(c.Writer, startEvent); err != nil {
			s.logger.Error("failed to send SSE event", "error", err)
			return
		}
	}

	// 在 goroutine 中运行 Agent
	// Run agent in goroutine
	go func() {
		defer close(eventChan)
		defer close(errChan)

		startTime := time.Now()

		// 这里简化实现，实际应该集成真实的 Agent.Run 并捕获事件
		// Simplified implementation, should integrate with actual Agent.Run and capture events
		output, err := ag.Run(ctx, req.Input)
		if err != nil {
			errChan <- err
			return
		}

		// 发送完成事件
		// Send complete event
		completeEvent := NewEvent(EventComplete, CompleteData{
			Output:   output.Content,
			Duration: time.Since(startTime).Seconds(),
		})
		completeEvent.AgentID = agentID
		completeEvent.SessionID = req.SessionID

		eventChan <- completeEvent
	}()

	// 流式发送事件
	// Stream events
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		s.logger.Error("streaming not supported")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "streaming_not_supported",
			"message": "streaming is not supported",
		})
		return
	}

	for {
		select {
		case <-ctx.Done():
			// 上下文取消
			// Context canceled
			errorEvent := NewEvent(EventError, ErrorData{
				Error: "request canceled or timeout",
				Code:  "CONTEXT_CANCELED",
			})
			if filter.ShouldSend(errorEvent) {
				s.sendSSE(c.Writer, errorEvent)
			}
			return

		case err := <-errChan:
			// 发生错误
			// Error occurred
			if err != nil {
				errorEvent := NewEvent(EventError, ErrorData{
					Error: err.Error(),
					Code:  "AGENT_ERROR",
				})
				if filter.ShouldSend(errorEvent) {
					s.sendSSE(c.Writer, errorEvent)
				}
			}
			return

		case event, ok := <-eventChan:
			if !ok {
				// 通道关闭，所有事件已发送
				// Channel closed, all events sent
				return
			}

			// 应用过滤器
			// Apply filter
			if filter.ShouldSend(event) {
				if err := s.sendSSE(c.Writer, event); err != nil {
					s.logger.Error("failed to send SSE event", "error", err)
					return
				}
				flusher.Flush()
			}
		}
	}
}

// sendSSE 发送单个 SSE 事件
// sendSSE sends a single SSE event
func (s *Server) sendSSE(w io.Writer, event *Event) error {
	sseData := event.ToSSE()
	_, err := fmt.Fprint(w, sseData)
	return err
}
