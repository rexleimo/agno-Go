package agentos

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/agno-go/pkg/agno/session"
)

// CreateSessionRequest represents the request to create a new session
type CreateSessionRequest struct {
	AgentID    string                 `json:"agent_id" binding:"required"`
	UserID     string                 `json:"user_id,omitempty"`
	TeamID     string                 `json:"team_id,omitempty"`
	WorkflowID string                 `json:"workflow_id,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateSessionRequest represents the request to update a session
type UpdateSessionRequest struct {
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	State    map[string]interface{} `json:"state,omitempty"`
}

// SessionResponse represents the session API response
type SessionResponse struct {
	SessionID  string                 `json:"session_id"`
	AgentID    string                 `json:"agent_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	TeamID     string                 `json:"team_id,omitempty"`
	WorkflowID string                 `json:"workflow_id,omitempty"`
	Name       string                 `json:"name,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	State      map[string]interface{} `json:"state,omitempty"`
	RunCount   int                    `json:"run_count"`
	CreatedAt  int64                  `json:"created_at"`
	UpdatedAt  int64                  `json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// handleCreateSession creates a new session
// POST /api/v1/sessions
func (s *Server) handleCreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Generate session ID
	sessionID := uuid.New().String()

	// Create session
	sess := session.NewSession(sessionID, req.AgentID)
	sess.UserID = req.UserID
	sess.TeamID = req.TeamID
	sess.WorkflowID = req.WorkflowID
	sess.Name = req.Name

	if req.Metadata != nil {
		sess.Metadata = req.Metadata
	}

	// Store session
	if err := s.sessionStorage.Create(c.Request.Context(), sess); err != nil {
		s.logger.Error("failed to create session", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed to create session",
			Message: err.Error(),
			Code:    "STORAGE_ERROR",
		})
		return
	}

	s.logger.Info("session created", "session_id", sessionID, "agent_id", req.AgentID)

	c.JSON(http.StatusCreated, sessionToResponse(sess))
}

// handleGetSession retrieves a session by ID
// GET /api/v1/sessions/:id
func (s *Server) handleGetSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "session ID is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	sess, err := s.sessionStorage.Get(c.Request.Context(), sessionID)
	if err == session.ErrSessionNotFound {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "session not found",
			Code:  "SESSION_NOT_FOUND",
		})
		return
	}
	if err != nil {
		s.logger.Error("failed to get session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed to get session",
			Message: err.Error(),
			Code:    "STORAGE_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, sessionToResponse(sess))
}

// handleUpdateSession updates an existing session
// PUT /api/v1/sessions/:id
func (s *Server) handleUpdateSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "session ID is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	var req UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get existing session
	sess, err := s.sessionStorage.Get(c.Request.Context(), sessionID)
	if err == session.ErrSessionNotFound {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "session not found",
			Code:  "SESSION_NOT_FOUND",
		})
		return
	}
	if err != nil {
		s.logger.Error("failed to get session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed to get session",
			Message: err.Error(),
			Code:    "STORAGE_ERROR",
		})
		return
	}

	// Update fields
	if req.Name != "" {
		sess.Name = req.Name
	}
	if req.Metadata != nil {
		sess.Metadata = req.Metadata
	}
	if req.State != nil {
		sess.State = req.State
	}

	// Save updated session
	if err := s.sessionStorage.Update(c.Request.Context(), sess); err != nil {
		s.logger.Error("failed to update session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed to update session",
			Message: err.Error(),
			Code:    "STORAGE_ERROR",
		})
		return
	}

	s.logger.Info("session updated", "session_id", sessionID)

	c.JSON(http.StatusOK, sessionToResponse(sess))
}

// handleDeleteSession deletes a session
// DELETE /api/v1/sessions/:id
func (s *Server) handleDeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "session ID is required",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	err := s.sessionStorage.Delete(c.Request.Context(), sessionID)
	if err == session.ErrSessionNotFound {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "session not found",
			Code:  "SESSION_NOT_FOUND",
		})
		return
	}
	if err != nil {
		s.logger.Error("failed to delete session", "error", err, "session_id", sessionID)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed to delete session",
			Message: err.Error(),
			Code:    "STORAGE_ERROR",
		})
		return
	}

	s.logger.Info("session deleted", "session_id", sessionID)

	c.JSON(http.StatusOK, gin.H{
		"message": "session deleted successfully",
	})
}

// handleListSessions lists sessions with optional filters
// GET /api/v1/sessions?agent_id=xxx&user_id=yyy
func (s *Server) handleListSessions(c *gin.Context) {
	// Build filters from query parameters
	filters := make(map[string]interface{})

	if agentID := c.Query("agent_id"); agentID != "" {
		filters["agent_id"] = agentID
	}
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if teamID := c.Query("team_id"); teamID != "" {
		filters["team_id"] = teamID
	}
	if workflowID := c.Query("workflow_id"); workflowID != "" {
		filters["workflow_id"] = workflowID
	}

	sessions, err := s.sessionStorage.List(c.Request.Context(), filters)
	if err != nil {
		s.logger.Error("failed to list sessions", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "failed to list sessions",
			Message: err.Error(),
			Code:    "STORAGE_ERROR",
		})
		return
	}

	// Convert to response format
	responses := make([]SessionResponse, len(sessions))
	for i, sess := range sessions {
		responses[i] = sessionToResponse(sess)
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": responses,
		"count":    len(responses),
	})
}

// sessionToResponse converts a session to API response format
func sessionToResponse(sess *session.Session) SessionResponse {
	return SessionResponse{
		SessionID:  sess.SessionID,
		AgentID:    sess.AgentID,
		UserID:     sess.UserID,
		TeamID:     sess.TeamID,
		WorkflowID: sess.WorkflowID,
		Name:       sess.Name,
		Metadata:   sess.Metadata,
		State:      sess.State,
		RunCount:   sess.GetRunCount(),
		CreatedAt:  sess.CreatedAt.Unix(),
		UpdatedAt:  sess.UpdatedAt.Unix(),
	}
}
