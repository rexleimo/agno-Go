package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// ConcurrencyLimiter restricts the number of concurrent requests and returns a
// backpressure response when capacity is exhausted.
type ConcurrencyLimiter struct {
	limit        int
	queueDepth   int
	retryAfter   time.Duration
	queueTimeout time.Duration

	slots    chan struct{}
	queue    chan struct{}
	inFlight int32
}

// BackpressureResponse describes the throttling payload returned to clients.
type BackpressureResponse struct {
	Error        string `json:"error"`
	RetryAfterMs int    `json:"retryAfterMs,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	InFlight     int    `json:"inFlight,omitempty"`
}

// NewConcurrencyLimiter constructs a limiter. queueDepth controls how many
// requests may wait for a slot; 0 disables waiting and triggers an immediate
// backpressure response when limit is reached.
func NewConcurrencyLimiter(limit, queueDepth int, retryAfter time.Duration) *ConcurrencyLimiter {
	if limit <= 0 {
		limit = 64
	}
	if retryAfter <= 0 {
		retryAfter = 200 * time.Millisecond
	}
	cfg := &ConcurrencyLimiter{
		limit:        limit,
		queueDepth:   queueDepth,
		retryAfter:   retryAfter,
		queueTimeout: retryAfter,
		slots:        make(chan struct{}, limit),
	}
	if queueDepth > 0 {
		cfg.queue = make(chan struct{}, queueDepth)
	}
	return cfg
}

// Middleware wraps an http.Handler with concurrency limiting and backpressure.
func (l *ConcurrencyLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release, ok := l.acquire(r.Context())
		if !ok {
			l.writeBackpressure(w)
			return
		}
		defer release()
		next.ServeHTTP(w, r)
	})
}

// Stats exposes current limiter state for debugging.
func (l *ConcurrencyLimiter) Stats() BackpressureResponse {
	return BackpressureResponse{
		Limit:    l.limit,
		InFlight: int(atomic.LoadInt32(&l.inFlight)),
	}
}

func (l *ConcurrencyLimiter) acquire(ctx context.Context) (func(), bool) {
	if l.queue == nil {
		select {
		case l.slots <- struct{}{}:
			atomic.AddInt32(&l.inFlight, 1)
			return l.release, true
		default:
			return nil, false
		}
	}

	select {
	case l.queue <- struct{}{}:
		defer func() { <-l.queue }()
		select {
		case l.slots <- struct{}{}:
			atomic.AddInt32(&l.inFlight, 1)
			return l.release, true
		case <-ctx.Done():
			return nil, false
		case <-time.After(l.queueTimeout):
			return nil, false
		}
	default:
		return nil, false
	}
}

func (l *ConcurrencyLimiter) release() {
	<-l.slots
	atomic.AddInt32(&l.inFlight, -1)
}

func (l *ConcurrencyLimiter) writeBackpressure(w http.ResponseWriter) {
	retryAfter := l.retryAfter
	if retryAfter <= 0 {
		retryAfter = 200 * time.Millisecond
	}
	retryMs := int(retryAfter.Milliseconds())
	if retryMs <= 0 {
		retryMs = 1
	}
	w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(BackpressureResponse{
		Error:        "backpressure",
		RetryAfterMs: retryMs,
		Limit:        l.limit,
		InFlight:     int(atomic.LoadInt32(&l.inFlight)),
	})
}
