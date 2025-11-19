package telemetry

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// NewEventsHandler returns an HTTP handler aligned with the TelemetryEvents
// contract. It supports filtering by session/workflow/runtime/query params.
func NewEventsHandler(store Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter := Filter{
			SessionID:     r.URL.Query().Get("sessionId"),
			WorkflowRunID: r.URL.Query().Get("workflowRunId"),
			AgentID:       r.URL.Query().Get("agentId"),
			Runtime:       r.URL.Query().Get("runtime"),
		}
		if eventType := r.URL.Query().Get("eventType"); eventType != "" {
			filter.EventType = EventType(eventType)
		}
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				filter.Limit = limit
			}
		}
		if filter.Runtime == "" {
			filter.Runtime = RuntimeGo
		}

		events := store.List(filter)
		resp := map[string]any{
			"events": events,
			"stats":  store.Stats(),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})
}
