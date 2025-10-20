package utils

import (
    "encoding/json"
    "time"
)

// JSONPretty returns pretty-printed JSON or empty string on error
func JSONPretty(v interface{}) string {
    b, err := json.MarshalIndent(v, "", "  ")
    if err != nil { return "" }
    return string(b)
}

// Retry retries fn up to attempts with fixed delay
func Retry(attempts int, delay time.Duration, fn func() error) error {
    var err error
    for i := 0; i < attempts; i++ {
        if err = fn(); err == nil { return nil }
        if i < attempts-1 { time.Sleep(delay) }
    }
    return err
}

