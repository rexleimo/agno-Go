package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConcurrencyLimiterBackpressure(t *testing.T) {
	limiter := NewConcurrencyLimiter(1, 0, 50*time.Millisecond)
	called := make(chan struct{})
	h := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called <- struct{}{}
	}))

	// First request acquires slot.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	go h.ServeHTTP(resp, req)
	<-called

	// Second request should hit backpressure.
	resp2 := httptest.NewRecorder()
	h.ServeHTTP(resp2, req.Clone(req.Context()))
	if resp2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp2.Code)
	}
}

func TestConcurrencyLimiterStats(t *testing.T) {
	limiter := NewConcurrencyLimiter(2, 1, 10*time.Millisecond)
	stats := limiter.Stats()
	if stats.Limit != 2 || stats.InFlight != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
