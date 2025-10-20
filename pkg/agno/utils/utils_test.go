package utils

import (
    "errors"
    "testing"
    "time"
)

func TestJSONPretty(t *testing.T) {
    s := JSONPretty(map[string]int{"a":1})
    if s == "" || s[0] != '{' { t.Fatalf("unexpected: %s", s) }
}

func TestRetry(t *testing.T) {
    attempts := 0
    err := Retry(3, 1*time.Millisecond, func() error { attempts++; if attempts < 2 { return errors.New("x") }; return nil })
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if attempts != 2 { t.Fatalf("attempts = %d, want 2", attempts) }
}

