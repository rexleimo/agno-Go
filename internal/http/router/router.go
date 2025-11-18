package router

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rexleimo/agno-go/internal/session/dto"
	"github.com/rexleimo/agno-go/internal/session/service"
	"github.com/rexleimo/agno-go/internal/session/store"
)

// Handler wires HTTP routes to the session service.
type Handler struct {
	service *service.Service
}

// New constructs a chi router exposing the session HTTP API.
func New(service *service.Service) http.Handler {
	h := &Handler{service: service}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/healthz", h.handleHealth)

	r.Route("/sessions", func(r chi.Router) {
		r.Get("/", h.handleListSessions)
		r.Post("/", h.handleCreateSession)

		r.Route("/{sessionID}", func(r chi.Router) {
			r.Get("/", h.handleGetSession)
			r.Delete("/", h.handleDeleteSession)
			r.Get("/runs", h.handleGetSessionRuns)
			r.Post("/rename", h.handleRenameSession)
		})
	})

	return r
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleListSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionType, err := parseSessionType(r.URL.Query().Get("type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SESSION_TYPE", err.Error(), nil)
		return
	}

	limit, err := parseIntDefault(r.URL.Query().Get("limit"), 20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "limit must be an integer", nil)
		return
	}
	page, err := parseIntDefault(r.URL.Query().Get("page"), 1)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "page must be an integer", nil)
		return
	}

	output, err := h.service.ListSessions(ctx, service.ListSessionsInput{
		SessionType: sessionType,
		ComponentID: r.URL.Query().Get("component_id"),
		UserID:      r.URL.Query().Get("user_id"),
		SessionName: r.URL.Query().Get("session_name"),
		SortBy:      r.URL.Query().Get("sort_by"),
		SortOrder:   r.URL.Query().Get("sort_order"),
		Limit:       limit,
		Page:        page,
		DatabaseID:  r.URL.Query().Get("db_id"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, output)
}

func (h *Handler) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionType, err := parseSessionType(r.URL.Query().Get("type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SESSION_TYPE", err.Error(), nil)
		return
	}

	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON payload", nil)
		return
	}

	input := service.CreateSessionInput{
		SessionID:    req.SessionID,
		SessionType:  sessionType,
		SessionName:  req.SessionName,
		SessionState: req.SessionState,
		Metadata:     req.Metadata,
		UserID:       req.UserID,
		AgentID:      req.AgentID,
		TeamID:       req.TeamID,
		WorkflowID:   req.WorkflowID,
		AgentData:    req.AgentData,
		TeamData:     req.TeamData,
		WorkflowData: req.WorkflowData,
		Runs:         req.Runs,
		Summary:      req.Summary,
		DatabaseID:   r.URL.Query().Get("db_id"),
	}

	detail, err := h.service.CreateSession(ctx, input)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, detail.Data)
}

func (h *Handler) handleGetSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionType, err := parseSessionType(r.URL.Query().Get("type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SESSION_TYPE", err.Error(), nil)
		return
	}

	detail, err := h.service.GetSession(ctx, service.GetSessionInput{
		SessionID:   chi.URLParam(r, "sessionID"),
		SessionType: sessionType,
		DatabaseID:  r.URL.Query().Get("db_id"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, detail.Data)
}

func (h *Handler) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionType, err := parseSessionType(r.URL.Query().Get("type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SESSION_TYPE", err.Error(), nil)
		return
	}

	if err := h.service.DeleteSession(ctx, service.DeleteSessionInput{
		SessionID:   chi.URLParam(r, "sessionID"),
		SessionType: sessionType,
		DatabaseID:  r.URL.Query().Get("db_id"),
	}); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleRenameSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionType, err := parseSessionType(r.URL.Query().Get("type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SESSION_TYPE", err.Error(), nil)
		return
	}

	var req renameSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON payload", nil)
		return
	}
	if req.SessionName == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "session_name is required", nil)
		return
	}

	detail, err := h.service.RenameSession(ctx, service.RenameSessionInput{
		SessionID:   chi.URLParam(r, "sessionID"),
		SessionType: sessionType,
		SessionName: req.SessionName,
		DatabaseID:  r.URL.Query().Get("db_id"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, detail.Data)
}

func (h *Handler) handleGetSessionRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionType, err := parseSessionType(r.URL.Query().Get("type"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SESSION_TYPE", err.Error(), nil)
		return
	}

	runs, err := h.service.GetSessionRuns(ctx, service.GetSessionInput{
		SessionID:   chi.URLParam(r, "sessionID"),
		SessionType: sessionType,
		DatabaseID:  r.URL.Query().Get("db_id"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

func parseSessionType(value string) (dto.SessionType, error) {
	return dto.ParseSessionType(value)
}

func parseIntDefault(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
	case errors.Is(err, service.ErrDatabaseRequired):
		writeError(w, http.StatusBadRequest, "DATABASE_REQUIRED", err.Error(), nil)
	case errors.Is(err, service.ErrDatabaseNotFound):
		writeError(w, http.StatusNotFound, "DATABASE_NOT_FOUND", err.Error(), nil)
	case errors.Is(err, dto.ErrInvalidSessionType):
		writeError(w, http.StatusBadRequest, "INVALID_SESSION_TYPE", err.Error(), nil)
	default:
		// Hide internal details from the top-level message but preserve them in details
		writeError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error", map[string]any{
			"error": err.Error(),
		})
	}
}

func writeError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	payload := map[string]any{
		"status":  "error",
		"code":    code,
		"message": message,
	}
	if details != nil && len(details) > 0 {
		payload["details"] = details
	}
	writeJSON(w, status, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

type createSessionRequest struct {
	SessionID    string           `json:"session_id"`
	SessionName  string           `json:"session_name"`
	SessionState map[string]any   `json:"session_state"`
	Metadata     map[string]any   `json:"metadata"`
	UserID       string           `json:"user_id"`
	AgentID      string           `json:"agent_id"`
	TeamID       string           `json:"team_id"`
	WorkflowID   string           `json:"workflow_id"`
	AgentData    map[string]any   `json:"agent_data"`
	TeamData     map[string]any   `json:"team_data"`
	WorkflowData map[string]any   `json:"workflow_data"`
	Runs         []map[string]any `json:"runs"`
	Summary      map[string]any   `json:"summary"`
}

type renameSessionRequest struct {
	SessionName string `json:"session_name"`
}
