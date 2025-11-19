package workflow

import (
	"fmt"
	"time"

	internalerrors "github.com/agno-agi/agno-go/go/internal/errors"
	"github.com/agno-agi/agno-go/go/internal/telemetry"
)

// NewNotMigratedPatternError returns a classified error indicating that the
// given collaboration pattern or workflow has not yet been migrated to Go.
func NewNotMigratedPatternError(wfID ID, pattern PatternType) *internalerrors.Error {
	msg := fmt.Sprintf("workflow %q with pattern %q is not yet migrated to Go", wfID, pattern)
	return internalerrors.NewNotMigrated(msg)
}

// RecordNotMigratedWorkflow emits a telemetry event for an attempt to run a
// workflow pattern that has not yet been migrated to Go.
func RecordNotMigratedWorkflow(rec telemetry.Recorder, wfID ID, pattern PatternType) {
	rec.Record(telemetry.Event{
		ID:        "workflow-not-migrated",
		Timestamp: time.Now(),
		SessionID: "",
		Type:      telemetry.EventWorkflowStepTransition,
		Payload: map[string]any{
			"workflow_id": string(wfID),
			"pattern":     string(pattern),
			"error_code":  string(internalerrors.CodeNotMigrated),
		},
	})
}
