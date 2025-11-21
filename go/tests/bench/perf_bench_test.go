package bench

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/internal/runtime"
	runtimeconfig "github.com/rexleimo/agno-go/internal/runtime/config"
	"github.com/rexleimo/agno-go/pkg/memory"
	"github.com/rexleimo/agno-go/pkg/providers/stub"
)

// BenchmarkChatStream simulates the P2 perf scenario (100 concurrency, 128 token input, 10m duration by default).
// It uses the stub provider to avoid external calls; duration can be shortened via BENCH_DURATION env.
func BenchmarkChatStream(b *testing.B) {
	cfg, err := runtimeconfig.LoadWithEnv(filepath.Join(repoRoot(b), "config", "default.yaml"), "")
	if err != nil {
		b.Fatalf("load config: %v", err)
	}
	benchCfg := cfg.Bench
	if env := os.Getenv("BENCH_DURATION"); env != "" {
		if d, err := time.ParseDuration(env); err == nil {
			benchCfg.Duration = d
		}
	}
	if benchCfg.Concurrency <= 0 {
		benchCfg.Concurrency = 1
	}
	if benchCfg.InputTokens <= 0 {
		benchCfg.InputTokens = 16
	}

	ctx, cancel := context.WithTimeout(context.Background(), benchCfg.Duration)
	defer cancel()

	router := model.NewRouter()
	router.RegisterChatProvider(stub.New(agent.ProviderOpenAI, model.ProviderAvailable, nil))
	store := memory.NewInMemoryStore()
	svc := runtime.NewService(store, router)
	agentID, err := svc.CreateAgent(ctx, agent.Agent{
		Name: "bench-agent",
		Model: agent.ModelConfig{
			Provider: agent.ProviderOpenAI,
			ModelID:  "stub-bench",
			Stream:   true,
		},
		Memory: agent.MemoryConfig{
			TokenWindow: benchCfg.InputTokens * 2,
		},
	})
	if err != nil {
		b.Fatalf("create agent: %v", err)
	}

	payload := strings.Repeat("x", benchCfg.InputTokens)
	var ops uint64
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	b.ReportAllocs()
	b.SetBytes(int64(benchCfg.InputTokens))
	b.ResetTimer()

	for i := 0; i < benchCfg.Concurrency; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			session, err := svc.CreateSession(ctx, agentID, fmt.Sprintf("user-%d", worker), nil)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			req := runtime.MessageRequest{
				Messages: []agent.Message{
					{Role: agent.RoleUser, Content: payload},
				},
				Stream: true,
			}
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				err := svc.StreamMessage(ctx, agentID, session.ID, req, func(ev model.ChatStreamEvent) error { return nil })
				if err != nil && ctx.Err() == nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				atomic.AddUint64(&ops, 1)
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case err := <-errCh:
		b.Fatalf("benchmark error: %v", err)
	case <-ctx.Done():
		<-done
	}

	b.StopTimer()
	total := atomic.LoadUint64(&ops)
	if total == 0 {
		b.Fatalf("no operations recorded in duration %s", benchCfg.Duration)
	}
	b.ReportMetric(float64(total)/benchCfg.Duration.Seconds(), "ops/sec")
}

func repoRoot(tb testing.TB) string {
	tb.Helper()
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		tb.Fatalf("cannot resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}
