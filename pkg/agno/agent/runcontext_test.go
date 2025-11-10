package agent

import (
    "context"
    "testing"

    "github.com/rexleimo/agno-go/pkg/agno/hooks"
    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

// ctxModel is a tiny model that first asks to call a tool, then completes.
type ctxModel struct{ models.BaseModel; step int }

func (m *ctxModel) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
    if m.step == 0 {
        m.step++
        return &types.ModelResponse{ToolCalls: []types.ToolCall{
            {ID: "tc1", Function: types.ToolCallFunction{Name: "ctx_probe", Arguments: "{}"}},
        }}, nil
    }
    return &types.ModelResponse{Content: "done"}, nil
}

func (m *ctxModel) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
    ch := make(chan types.ResponseChunk)
    close(ch)
    return ch, nil
}

func TestRun_PropagatesRunContextToHooksAndTools(t *testing.T) {
    var seenPre, seenTool, seenPost string

    // Toolkit with a function that returns RunContextID
    tk := toolkit.NewBaseToolkit("ctx")
    tk.RegisterFunction(&toolkit.Function{
        Name:        "ctx_probe",
        Description: "echo run context id",
        Parameters:  map[string]toolkit.Parameter{},
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            if id, ok := RunContextID(ctx); ok { seenTool = id; return map[string]string{"run_context_id": id}, nil }
            return map[string]string{"run_context_id": ""}, nil
        },
    })

    pre := func(ctx context.Context, in *hooks.HookInput) error { if id, ok := RunContextID(ctx); ok { seenPre = id }; return nil }
    post := func(ctx context.Context, in *hooks.HookInput) error { if id, ok := RunContextID(ctx); ok { seenPost = id }; return nil }

    ag, err := New(Config{Model: &ctxModel{BaseModel: models.BaseModel{ID: "m", Provider: "mock"}}, Toolkits: []toolkit.Toolkit{tk}, PreHooks: []hooks.Hook{hooks.HookFunc(pre)}, PostHooks: []hooks.Hook{hooks.HookFunc(post)}})
    if err != nil { t.Fatalf("new agent: %v", err) }

    ctx := WithRunContext(context.Background(), "rc-test-1")
    out, err := ag.Run(ctx, "hello")
    if err != nil { t.Fatalf("run error: %v", err) }
    if out == nil || out.Content != "done" { t.Fatalf("unexpected output: %#v", out) }

    if seenPre != "rc-test-1" || seenTool != "rc-test-1" || seenPost != "rc-test-1" {
        t.Fatalf("run context not propagated, pre=%q tool=%q post=%q", seenPre, seenTool, seenPost)
    }
}
