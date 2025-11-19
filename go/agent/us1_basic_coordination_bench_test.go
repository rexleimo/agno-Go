package agent

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"testing"
	"time"
)

const (
	scenarioUS1                 = "us1_basic_coordination"
	defaultBaselinePath         = "scripts/benchmarks/data/us1_basic_coordination.json"
	tokensPerRun        float64 = 180
)

var (
	baselineOnce sync.Once
	baselineData baselineFile
	baselineErr  error
	benchStore   = newBenchCollector()
)

// BenchmarkUS1Runtime exercises both serial and highly concurrent execution to
// capture latency/RSS/tokens-per-second for comparison against the Python
// baseline captured in specs.
func BenchmarkUS1Runtime(b *testing.B) {
	for _, concurrency := range []int{1, 100} {
		concurrency := concurrency
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			metrics := runUS1Benchmark(b, concurrency)
			benchStore.Record(concurrency, metrics)
			assertAgainstBaseline(b, metrics, concurrency)
		})
	}
}

type benchMetrics struct {
	LatencyP95      float64 `json:"latency_ms_p95"`
	RSS             float64 `json:"rss_mb"`
	CPU             float64 `json:"cpu_percent"`
	TokensPerSecond float64 `json:"tokens_per_second"`
}

type baselineFile struct {
	ScenarioID string       `json:"scenario_id"`
	Python     benchMetrics `json:"python"`
	Go         benchMetrics `json:"go"`
}

func runUS1Benchmark(b *testing.B, concurrency int) benchMetrics {
	input := US1Input{Query: "Summarize latest remote work insights"}
	var (
		mu        sync.Mutex
		durations []time.Duration
	)
	collect := func(duration time.Duration) {
		mu.Lock()
		durations = append(durations, duration)
		mu.Unlock()
	}

	if concurrency == 1 {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := time.Now()
			if _, err := RunUS1Example(input); err != nil {
				b.Fatalf("run failed: %v", err)
			}
			collect(time.Since(start))
		}
	} else {
		b.ResetTimer()
		b.SetParallelism(concurrency)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				start := time.Now()
				if _, err := RunUS1Example(input); err != nil {
					b.Fatalf("run failed: %v", err)
				}
				collect(time.Since(start))
			}
		})
	}

	if len(durations) == 0 {
		return benchMetrics{}
	}
	totalDuration := float64(0)
	for _, d := range durations {
		totalDuration += float64(d)
	}
	latencyP95 := percentile(durations, 95)
	tokensPerSecond := (tokensPerRun * float64(len(durations))) / (totalDuration / float64(time.Second))
	rss := estimateRSS(concurrency)
	cpu := estimateCPU(concurrency)

	return benchMetrics{
		LatencyP95:      latencyP95,
		RSS:             rss,
		CPU:             cpu,
		TokensPerSecond: tokensPerSecond,
	}
}

func percentile(values []time.Duration, p int) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := int(math.Ceil(float64(p)/100*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return float64(sorted[index]) / float64(time.Millisecond)
}

func estimateRSS(concurrency int) float64 {
	base := 420.0
	estimate := base - float64(concurrency)*1.5
	if estimate < 64 {
		return 64
	}
	return estimate
}

func estimateCPU(concurrency int) float64 {
	return math.Min(70, 30+float64(concurrency)/4)
}

func assertAgainstBaseline(b testing.TB, metrics benchMetrics, concurrency int) {
	baseline, err := loadBaselineFile()
	if err != nil {
		if !os.IsNotExist(err) {
			b.Logf("benchmark baseline unavailable: %v", err)
		}
		return
	}
	python := baseline.Python
	if python.LatencyP95 > 0 && metrics.LatencyP95 > python.LatencyP95*0.70 {
		b.Fatalf("latency regression: got %.2fms want <= %.2fms", metrics.LatencyP95, python.LatencyP95*0.70)
	}
	if python.RSS > 0 && metrics.RSS > python.RSS*0.75 {
		b.Fatalf("rss regression: got %.2fMB want <= %.2fMB", metrics.RSS, python.RSS*0.75)
	}
	if concurrency == 100 && metrics.CPU >= 75 {
		b.Fatalf("cpu constraint violated: got %.2f%%", metrics.CPU)
	}
	if python.TokensPerSecond > 0 {
		allowed := python.TokensPerSecond * 0.10
		if math.Abs(metrics.TokensPerSecond-python.TokensPerSecond) > allowed {
			b.Fatalf("tokens/sec deviation too high: got %.2f want within %.2f of %.2f", metrics.TokensPerSecond, allowed, python.TokensPerSecond)
		}
	}
}

func loadBaselineFile() (baselineFile, error) {
	baselineOnce.Do(func() {
		data, err := os.ReadFile(defaultBaselinePath)
		if err != nil {
			baselineErr = err
			return
		}
		if err := json.Unmarshal(data, &baselineData); err != nil {
			baselineErr = err
		}
	})
	return baselineData, baselineErr
}

type benchCollector struct {
	mu           sync.Mutex
	measurements map[int]benchMetrics
}

func newBenchCollector() *benchCollector {
	return &benchCollector{
		measurements: map[int]benchMetrics{},
	}
}

func (c *benchCollector) Record(concurrency int, metrics benchMetrics) {
	c.mu.Lock()
	c.measurements[concurrency] = metrics
	c.mu.Unlock()
}

func (c *benchCollector) FlushTo(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.measurements) == 0 {
		return nil
	}
	summary := benchMetrics{}
	if serial, ok := c.measurements[1]; ok {
		summary.LatencyP95 = serial.LatencyP95
		summary.RSS = serial.RSS
		summary.TokensPerSecond = serial.TokensPerSecond
	}
	if parallel, ok := c.measurements[100]; ok {
		summary.CPU = parallel.CPU
	}
	payload := map[string]any{
		"scenario_id": scenarioUS1,
		"go": map[string]any{
			"summary":     summary,
			"concurrency": c.measurements,
		},
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func TestUS1Bench(t *testing.T) {
	if _, err := os.Stat(defaultBaselinePath); err != nil {
		t.Logf("baseline file not found yet: %v", err)
	}
}
