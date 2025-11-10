package video

import "testing"

func TestOptionsValidate(t *testing.T) {
    v := Options{Prompt: "p", AspectRatio: "16:9", DurationSeconds: 10}
    if err := v.validate(); err != nil { t.Fatalf("valid: %v", err) }

    v2 := Options{Prompt: "p", AspectRatio: "4:3", DurationSeconds: 10}
    if err := v2.validate(); err == nil { t.Fatal("expected invalid aspect ratio") }

    v3 := Options{Prompt: "p", AspectRatio: "16:9", DurationSeconds: 12}
    if err := v3.validate(); err == nil { t.Fatal("expected invalid duration") }
}

