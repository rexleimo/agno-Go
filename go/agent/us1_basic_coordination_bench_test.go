package agent

import "testing"

// BenchmarkRunUS1Example measures the overhead of the Go-side US1 entrypoint.
// At this stage the implementation is intentionally lightweight; the benchmark
// serves as a baseline and will become more representative as the workflow
// execution logic is wired in.
func BenchmarkRunUS1Example(b *testing.B) {
	input := US1Input{
		Query: "Write an article about the top 2 stories on hackernews",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RunUS1Example(input)
		if err != nil {
			b.Fatalf("unexpected error in RunUS1Example: %v", err)
		}
	}
}
