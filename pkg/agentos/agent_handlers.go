package agentos

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/agno-go/pkg/agno/agent"
	"github.com/yourusername/agno-go/pkg/agno/session"
)

// AgentRunRequest represents a request to run an agent
type AgentRunRequest struct {
	Input     string `json:"input" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

// AgentRunResponse represents the response from running an agent
type AgentRunResponse struct {
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
			Error:   "agent ID is required",
			Code:    "INVALID_REQUEST",
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

	// TODO: Get agent from registry
	// For now, return a placeholder response
	s.logger.Info("agent run requested",
		"agent_id", agentID,
		"input", req.Input,
		"session_id", req.SessionID,
	)

	// If session_id provided, update the session
	var sess *session.Session
	if req.SessionID != "" {
		var err error
		sess, err = s.sessionStorage.Get(c.Request.Context(), req.SessionID)
		if err != nil {
			s.logger.Warn("failed to get session", "error", err, "session_id", req.SessionID)
		}
	}

	// Placeholder response (actual agent execution to be implemented)
	response := AgentRunResponse{
		Content:   "This is a placeholder response. Agent execution will be implemented next.",
		SessionID: req.SessionID,
		Metadata: map[string]interface{}{
			"agent_id": agentID,
			"input":    req.Input,
		},
	}

	// If we have a session, add the run to it
	if sess != nil {
		run := &agent.RunOutput{
			Content: response.Content,
			Metadata: map[string]interface{}{
				"agent_id": agentID,
			},
		}
		sess.AddRun(run)

		if err := s.sessionStorage.Update(c.Request.Context(), sess); err != nil {
			s.logger.Warn("failed to update session with run", "error", err, "session_id", req.SessionID)
		}
	}

	c.JSON(http.StatusOK, response)
}
