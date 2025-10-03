package agentos

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rexleimo/agno-go/pkg/agno/session"
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
	)

	// Get session if provided
	var sess *session.Session
	if req.SessionID != "" {
		sess, err = s.sessionStorage.Get(c.Request.Context(), req.SessionID)
		if err != nil {
			s.logger.Warn("failed to get session", "error", err, "session_id", req.SessionID)
		}
	}

	// Run the agent
	output, err := ag.Run(c.Request.Context(), req.Input)
	if err != nil {
		s.logger.Error("agent run failed", "error", err, "agent_id", agentID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "agent execution failed",
			Message: err.Error(),
			Code:    "EXECUTION_ERROR",
		})
		return
	}

	response := AgentRunResponse{
		Content:   output.Content,
		SessionID: req.SessionID,
		Metadata: map[string]interface{}{
			"agent_id": agentID,
		},
	}

	// If we have a session, add the run to it
	if sess != nil {
		sess.AddRun(output)

		if err := s.sessionStorage.Update(c.Request.Context(), sess); err != nil {
			s.logger.Warn("failed to update session with run", "error", err, "session_id", req.SessionID)
		}
	}

	s.logger.Info("agent run completed", "agent_id", agentID, "content_length", len(output.Content))

	c.JSON(http.StatusOK, response)
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
