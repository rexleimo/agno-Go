package telemetry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewEventsHandlerFilters(t *testing.T) {
	recorder := NewInMemoryRecorder()
	recorder.Record(Event{ID: "1", Timestamp: time.Now(), Type: EventRunStarted, SessionID: "session-1"})
	recorder.Record(Event{ID: "2", Timestamp: time.Now(), Type: EventRunCompleted, SessionID: "session-2"})

	req := httptest.NewRequest(http.MethodGet, "/?sessionId=session-2", nil)
	resp := httptest.NewRecorder()
	NewEventsHandler(recorder).ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.Code)
	}
	var payload struct {
		Events []Event `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Events) != 1 || payload.Events[0].SessionID != "session-2" {
		t.Fatalf("unexpected events %+v", payload.Events)
	}
}
