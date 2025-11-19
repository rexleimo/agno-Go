package agent

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	if exportPath := os.Getenv("US1_BENCHMARK_EXPORT"); exportPath != "" {
		if err := benchStore.FlushTo(exportPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to export benchmark metrics: %v\n", err)
		}
	}
	os.Exit(code)
}
