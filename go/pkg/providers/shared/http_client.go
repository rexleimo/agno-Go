package shared

import (
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	defaultTransport http.RoundTripper
	transportOnce    sync.Once
)

// DefaultHTTPClient returns a pooled HTTP client with sane defaults for provider calls.
func DefaultHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	transportOnce.Do(func() {
		defaultTransport = &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			DialContext:         (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:   true,
			MaxIdleConns:        256,
			MaxIdleConnsPerHost: 128,
			MaxConnsPerHost:     256,
			IdleConnTimeout:     90 * time.Second,
		}
	})
	return &http.Client{
		Timeout:   timeout,
		Transport: defaultTransport,
	}
}
