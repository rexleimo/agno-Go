//go:build !logfire
// +build !logfire

package main

import "log"

func main() {
	log.Println("logfire_observability example requires build tag `logfire`. Run with `go run -tags logfire .` after installing the OpenTelemetry dependencies.")
}
