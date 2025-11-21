package bench

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/perf/benchstat"
)

func TestBenchstatReport(t *testing.T) {
	output := firstNonEmpty(os.Getenv("BENCH_OUTPUT"), filepath.Join("..", "..", "specs", "001-go-agno-rewrite", "artifacts", "bench", "bench.txt"))
	if _, err := os.Stat(output); err != nil {
		t.Skipf("bench output not found, run make bench to generate: %v", err)
	}
	paths := []string{output}
	if base := os.Getenv("BENCH_BASELINE"); base != "" {
		if _, err := os.Stat(base); err == nil {
			paths = append([]string{base}, paths...)
		}
	}

	var c benchstat.Collection
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			t.Fatalf("open %s: %v", path, err)
		}
		if err := c.AddFile(filepath.Base(path), f); err != nil {
			t.Fatalf("add bench file %s: %v", path, err)
		}
		_ = f.Close()
	}

	tables := c.Tables()
	var buf bytes.Buffer
	benchstat.FormatText(&buf, tables)

	dest := filepath.Join("..", "..", "specs", "001-go-agno-rewrite", "artifacts", "bench", "benchstat.txt")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatalf("mkdir bench dir: %v", err)
	}
	if err := os.WriteFile(dest, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write benchstat report: %v", err)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
