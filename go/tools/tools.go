//go:build tools

// Package tools tracks developer tooling dependencies.
package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "golang.org/x/perf/cmd/benchstat"
	_ "mvdan.cc/gofumpt"
)
