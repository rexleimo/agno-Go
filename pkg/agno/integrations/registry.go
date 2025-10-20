package integrations

import (
    "context"
    "sync"
    "time"
)

// Integration represents a third-party integration with optional health check
type Integration struct {
    Name        string
    Description string
    // Health checks the integration and returns latency and error (if unhealthy)
    Health func(ctx context.Context) (latency time.Duration, err error)
}

// Registry stores registered integrations
type Registry struct {
    mu           sync.RWMutex
    integrations map[string]Integration
}

// NewRegistry creates a new integration registry
func NewRegistry() *Registry {
    return &Registry{ integrations: make(map[string]Integration) }
}

// Register adds or replaces an integration by name
func (r *Registry) Register(in Integration) {
    r.mu.Lock(); defer r.mu.Unlock()
    r.integrations[in.Name] = in
}

// List returns names of all registered integrations
func (r *Registry) List() []string {
    r.mu.RLock(); defer r.mu.RUnlock()
    out := make([]string, 0, len(r.integrations))
    for k := range r.integrations { out = append(out, k) }
    return out
}

// CheckHealth runs health check if provided; returns map[name]error (nil if healthy)
func (r *Registry) CheckHealth(ctx context.Context) map[string]error {
    r.mu.RLock(); defer r.mu.RUnlock()
    results := make(map[string]error, len(r.integrations))
    for name, in := range r.integrations {
        if in.Health == nil {
            results[name] = nil
            continue
        }
        _, err := in.Health(ctx)
        results[name] = err
    }
    return results
}

