package agentos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rexleimo/agno-go/pkg/agno/agent"
	"github.com/rexleimo/agno-go/pkg/agno/media"
	"github.com/rexleimo/agno-go/pkg/agno/session"
	"github.com/rexleimo/agno-go/pkg/agno/types"
)

// AgentRunRequest represents a request to run an agent
type AgentRunRequest struct {
	Input     string      `json:"input"`
	SessionID string      `json:"session_id,omitempty"`
	Stream    bool        `json:"stream,omitempty"`
	Media     interface{} `json:"media,omitempty"`
}

var (
	errMissingRunInput     = errors.New("agent run requires input or media attachments")
	errInvalidMediaPayload = errors.New("invalid media payload")
)

func normalizeRunRequest(req *AgentRunRequest) ([]media.Attachment, error) {
	attachments, err := media.Normalize(req.Media)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errInvalidMediaPayload, err)
	}

	if strings.TrimSpace(req.Input) == "" {
		if len(attachments) == 0 {
			return nil, errMissingRunInput
		}
		req.Input = buildMediaPlaceholder(len(attachments))
	}

	return attachments, nil
}

func buildMediaPlaceholder(count int) string {
	if count <= 1 {
		return "Media request (1 attachment)"
	}
	return fmt.Sprintf("Media request (%d attachments)", count)
}

// AgentRunResponse represents the response from running an agent
type AgentRunResponse struct {
	RunID     string                 `json:"run_id,omitempty"`
	Status    agent.RunStatus        `json:"status,omitempty"`
	Content   string                 `json:"content"`
	SessionID string                 `json:"session_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// handleAgentRun runs an agent with the given input
// POST /api/v1/agents/:id/run
func (s *Server) handleAgentRun(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "agent ID is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	var req AgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	attachments, err := normalizeRunRequest(&req)
	if err != nil {
		switch {
		case errors.Is(err, errInvalidMediaPayload):
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid media payload",
				Message: err.Error(),
				Code:    "INVALID_MEDIA",
			})
		case errors.Is(err, errMissingRunInput):
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "input or media payload is required",
				Code:  "INVALID_REQUEST",
			})
		default:
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid request",
				Message: err.Error(),
				Code:    "INVALID_REQUEST",
			})
		}
		return
	}

	// Get agent from registry
	ag, err := s.agentRegistry.Get(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "agent not found",
			Message: err.Error(),
			Code:    "AGENT_NOT_FOUND",
		})
		return
	}

	s.logger.Info("agent run requested",
		"agent_id", agentID,
		"input", req.Input,
		"session_id", req.SessionID,
		"media_count", len(attachments),
	)

	if shouldStreamRequest(c, req.Stream) {
		s.streamAgentRun(c, agentID, ag, req, attachments)
		return
	}

	// Get session if provided
	var sess *session.Session
	if req.SessionID != "" {
		sess, err = s.sessionStorage.Get(c.Request.Context(), req.SessionID)
		if err == session.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "session not found",
				Code:  "SESSION_NOT_FOUND",
			})
			return
		}
		if err != nil {
			s.logger.Error("failed to get session", "error", err, "session_id", req.SessionID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "failed to get session",
				Message: err.Error(),
				Code:    "STORAGE_ERROR",
			})
			return
		}
	}

	// Run the agent
	output, err := ag.Run(c.Request.Context(), req.Input)

	if err != nil {
		s.logger.Error("agent run failed", "error", err, "agent_id", agentID)
		status := http.StatusInternalServerError
		errorCode := "EXECUTION_ERROR"
		if agnoErr, ok := err.(*types.AgnoError); ok && agnoErr.Code == types.ErrCodeCancelled {
			status = http.StatusRequestTimeout
			errorCode = string(types.ErrCodeCancelled)
		}
		c.JSON(status, ErrorResponse{
			Error:   "agent execution failed",
			Message: err.Error(),
			Code:    errorCode,
		})
		return
	}

	if len(attachments) > 0 {
		if output.Metadata == nil {
			output.Metadata = make(map[string]interface{})
		}
		output.Metadata["media"] = attachments
	}

	if sess != nil && output != nil {
		sess.AddRun(output)
		if updateErr := s.sessionStorage.Update(c.Request.Context(), sess); updateErr != nil {
			s.logger.Warn("failed to update session with run", "error", updateErr, "session_id", req.SessionID)
		}
	}

	metadata := map[string]interface{}{
		"agent_id": agentID,
	}
	if output.Metadata != nil {
		for k, v := range output.Metadata {
			metadata[k] = v
		}
	}
	if len(attachments) > 0 {
		metadata["media"] = attachments
	}

	response := AgentRunResponse{
		RunID:     output.RunID,
		Status:    output.Status,
		Content:   output.Content,
		SessionID: req.SessionID,
		Metadata:  metadata,
	}

	s.logger.Info("agent run completed", "agent_id", agentID, "run_id", output.RunID)

	c.JSON(http.StatusOK, response)
}

func shouldStreamRequest(c *gin.Context, bodyFlag bool) bool {
	if bodyFlag {
		return true
	}
	query := c.Query("stream_events")
	if query == "" {
		return false
	}
	val, err := strconv.ParseBool(query)
	if err != nil {
		return false
	}
	return val
}

func (s *Server) streamAgentRun(c *gin.Context, agentID string, ag *agent.Agent, req AgentRunRequest, attachments []media.Attachment) {
	filter := NewEventFilter(splitCommaQuery(c.Query("types")))

	var sess *session.Session
	if req.SessionID != "" {
		stored, err := s.sessionStorage.Get(c.Request.Context(), req.SessionID)
		if err == session.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "session not found",
				Code:  "SESSION_NOT_FOUND",
			})
			return
		}
		if err != nil {
			s.logger.Error("failed to get session", "error", err, "session_id", req.SessionID)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "failed to get session",
				Message: err.Error(),
				Code:    "STORAGE_ERROR",
			})
			return
		}
		sess = stored
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "streaming_not_supported",
			"message": "streaming is not supported",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	startEvent := NewEvent(EventRunStart, RunStartData{
		Input:     req.Input,
		SessionID: req.SessionID,
		Media:     attachments,
	})
	startEvent.AgentID = agentID
	startEvent.SessionID = req.SessionID
	if filter.ShouldSend(startEvent) {
		s.sendSSE(c.Writer, startEvent)
		flusher.Flush()
	}

	resultChan := make(chan struct {
		output *agent.RunOutput
		err    error
	}, 1)

	go func() {
		output, err := ag.Run(ctx, req.Input)
		if output != nil && len(attachments) > 0 {
			if output.Metadata == nil {
				output.Metadata = make(map[string]interface{})
			}
			output.Metadata["media"] = attachments
		}
		if sess != nil && output != nil {
			sess.AddRun(output)
			updateCtx, cancelUpdate := context.WithTimeout(context.Background(), 2*time.Second)
			if updateErr := s.sessionStorage.Update(updateCtx, sess); updateErr != nil {
				s.logger.Warn("failed to update session with run", "error", updateErr, "session_id", req.SessionID)
			}
			cancelUpdate()
		}
		resultChan <- struct {
			output *agent.RunOutput
			err    error
		}{output: output, err: err}
		close(resultChan)
	}()

	for {
		select {
		case <-ctx.Done():
			errorEvent := NewEvent(EventError, ErrorData{
				Error: ctx.Err().Error(),
				Code:  "CONTEXT_CANCELED",
			})
			if filter.ShouldSend(errorEvent) {
				s.sendSSE(c.Writer, errorEvent)
				flusher.Flush()
			}
			return

		case res, ok := <-resultChan:
			if !ok {
				return
			}

			output := res.output
			err := res.err

			if err != nil {
				code := "AGENT_ERROR"
				if agnoErr, ok := err.(*types.AgnoError); ok {
					code = string(agnoErr.Code)
				}
				errorEvent := NewEvent(EventError, ErrorData{
					Error: err.Error(),
					Code:  code,
				})
				if filter.ShouldSend(errorEvent) {
					s.sendSSE(c.Writer, errorEvent)
					flusher.Flush()
				}
			}

			if output != nil {
				s.emitRunEvents(c.Writer, flusher, filter, agentID, req.SessionID, output, ag)
			}
			return
		}
	}
}

func splitCommaQuery(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// handleListAgents lists all registered agents
// GET /api/v1/agents
func (s *Server) handleListAgents(c *gin.Context) {
	agents := s.agentRegistry.List()

	// Convert to response format
	agentList := make([]map[string]interface{}, 0, len(agents))
	for id, ag := range agents {
		agentList = append(agentList, map[string]interface{}{
			"id":   id,
			"name": ag.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agentList,
		"count":  len(agentList),
	})
}
