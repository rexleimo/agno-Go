package agent

import (
	"context"

	"github.com/rexleimo/agno-go/pkg/agno/run"
)

// ctxKey is a private type to avoid key collisions in context
type ctxKey string

const ctxKeyRunContextID ctxKey = "agno.run_context_id"

// WithRunContext returns a child context carrying a run-context identifier
func WithRunContext(ctx context.Context, id string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if id == "" {
		return ctx
	}
	rc, _ := run.FromContext(ctx)
	if rc == nil {
		rc = run.NewContext()
	} else {
		rc = rc.Clone()
	}
	rc.RunID = id
	ctx = run.WithContext(ctx, rc)
	return context.WithValue(ctx, ctxKeyRunContextID, id)
}

// RunContextID retrieves the run-context identifier from context, if present
func RunContextID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	if rc, ok := run.FromContext(ctx); ok && rc != nil && rc.RunID != "" {
		return rc.RunID, true
	}
	v := ctx.Value(ctxKeyRunContextID)
	if s, ok := v.(string); ok && s != "" {
		return s, true
	}
	return "", false
}
