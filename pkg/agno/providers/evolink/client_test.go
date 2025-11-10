package evolink

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "sync/atomic"
    "testing"
    "time"
)

func TestPollTask_Completes(t *testing.T) {
    var calls int32
    mux := http.NewServeMux()
    mux.HandleFunc("/v1/tasks/t1", func(w http.ResponseWriter, r *http.Request) {
        c := atomic.AddInt32(&calls, 1)
        w.Header().Set("content-type", "application/json")
        if c < 2 {
            json.NewEncoder(w).Encode(map[string]any{"id":"t1","status":"processing"})
            return
        }
        json.NewEncoder(w).Encode(map[string]any{"id":"t1","status":"completed","data":map[string]any{"url":"https://example.com/video.mp4"}})
    })
    srv := httptest.NewServer(mux)
    defer srv.Close()

    c, err := NewClient(Config{APIKey: "k", BaseURL: srv.URL, Timeout: 5*time.Second})
    if err != nil { t.Fatalf("client: %v", err) }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    tr, err := c.PollTask(ctx, "t1", 50*time.Millisecond)
    if err != nil { t.Fatalf("poll: %v", err) }
    if tr.Status != "completed" { t.Fatalf("want completed, got %s", tr.Status) }
    if tr.Data["url"] == "" { t.Fatalf("missing data url") }
}

