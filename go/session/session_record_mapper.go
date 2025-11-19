package session

import "time"

// RunMessage mirrors the contracts/RunMessage schema for parity purposes.
type RunMessage struct {
	Role        string         `json:"role"`
	Content     string         `json:"content"`
	References  []string       `json:"references,omitempty"`
	ToolResults []ToolResult   `json:"tool_results,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ToolResult serializes tool invocation results within run messages.
type ToolResult struct {
	ToolName string         `json:"tool_name"`
	Payload  map[string]any `json:"payload,omitempty"`
}

// SessionSummary captures textual session summaries shared with the Python
// runtime.
type SessionSummary struct {
	Text    string `json:"text"`
	Version string `json:"version"`
}

// CachePolicy mirrors the contracts schema and is included for completeness.
type CachePolicy struct {
	SearchSessionHistory bool `json:"search_session_history"`
	NumHistorySessions   int  `json:"num_history_sessions"`
}

// SessionRecord is a Go representation of the JSON schema shared with the
// Python runtime. It intentionally carries only Go-native types for easy
// serialization.
type SessionRecord struct {
	SessionID   string             `json:"session_id"`
	UserID      string             `json:"user_id,omitempty"`
	TeamID      string             `json:"team_id,omitempty"`
	StateBlob   map[string]any     `json:"state_blob,omitempty"`
	History     []RunMessage       `json:"history,omitempty"`
	Summary     SessionSummary     `json:"summary"`
	Metrics     map[string]float64 `json:"metrics,omitempty"`
	CachePolicy CachePolicy        `json:"cache_policy"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// NewSessionRecordFromSession converts an in-memory Session struct into the
// serialized SessionRecord representation defined in contracts/runtime-openapi.
func NewSessionRecordFromSession(s *Session) SessionRecord {
	if s == nil {
		return SessionRecord{}
	}
	record := SessionRecord{
		SessionID:   string(s.ID),
		UserID:      s.Context.UserID,
		StateBlob:   cloneMap(s.Context.Payload),
		History:     historyToMessages(s.History),
		Summary:     SessionSummary{Text: summaryText(s), Version: s.UpdatedAt.Format(time.RFC3339Nano)},
		Metrics:     extractMetrics(s),
		CachePolicy: CachePolicy{},
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
	return record
}

// ApplySessionRecord merges a SessionRecord back into a Session struct,
// returning a detached copy to avoid mutating the original input.
func ApplySessionRecord(record SessionRecord, base *Session) *Session {
	var target Session
	if base != nil {
		target = *base
	}
	target.ID = ID(record.SessionID)
	target.Context.UserID = record.UserID
	target.Context.Payload = cloneMap(record.StateBlob)
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	target.Context.StartedAt = record.CreatedAt
	target.History = messagesToHistory(record.History)
	target.CreatedAt = record.CreatedAt
	target.UpdatedAt = record.UpdatedAt
	if target.Result == nil {
		target.Result = &Result{Data: map[string]any{}}
	}
	if target.Result.Data == nil {
		target.Result.Data = map[string]any{}
	}
	target.Result.Reason = record.Summary.Text
	target.Result.Data["metrics"] = cloneMetrics(record.Metrics)
	return &target
}

func summaryText(s *Session) string {
	if s.Result == nil {
		return ""
	}
	if s.Result.Reason != "" {
		return s.Result.Reason
	}
	if val, ok := s.Result.Data["summary"].(string); ok {
		return val
	}
	return ""
}

func extractMetrics(s *Session) map[string]float64 {
	metrics := map[string]float64{}
	if s.Result != nil {
		if s.Result.Success {
			metrics["success"] = 1
		}
		for k, v := range s.Result.Data {
			if f, ok := toFloat64(v); ok {
				metrics[k] = f
			}
		}
	}
	return metrics
}

func historyToMessages(history []HistoryEntry) []RunMessage {
	if len(history) == 0 {
		return nil
	}
	messages := make([]RunMessage, len(history))
	for i, entry := range history {
		messages[i] = RunMessage{
			Role:      entry.Source,
			Content:   entry.Message,
			Timestamp: entry.Timestamp,
			Metadata:  cloneMap(entry.Metadata),
		}
	}
	return messages
}

func messagesToHistory(messages []RunMessage) []HistoryEntry {
	if len(messages) == 0 {
		return nil
	}
	history := make([]HistoryEntry, len(messages))
	for i, msg := range messages {
		history[i] = HistoryEntry{
			Timestamp: msg.Timestamp,
			Source:    msg.Role,
			Message:   msg.Content,
			Metadata:  cloneMap(msg.Metadata),
		}
	}
	return history
}

func cloneMetrics(src map[string]float64) map[string]float64 {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]float64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
