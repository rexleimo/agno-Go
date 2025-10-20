package integrations

import (
    "context"
    "errors"
    "testing"
    "time"
)

func TestRegistry_Basic(t *testing.T) {
    reg := NewRegistry()
    reg.Register(Integration{Name: "zendesk"})
    reg.Register(Integration{Name: "zoom", Health: func(ctx context.Context) (time.Duration, error) { return 1 * time.Millisecond, nil }})
    reg.Register(Integration{Name: "broken", Health: func(ctx context.Context) (time.Duration, error) { return 0, errors.New("down") }})

    names := reg.List()
    if len(names) != 3 { t.Fatalf("expected 3 integrations, got %d", len(names)) }

    res := reg.CheckHealth(context.Background())
    if res["zoom"] != nil && res["zendesk"] != nil { // zendesk has no health and should be nil
        t.Fatalf("unexpected health results: %#v", res)
    }
    if res["broken"] == nil { t.Fatalf("broken should report error") }
}

