package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/agno-agi/agno-go/go/agent"
)

func main() {
	input := agent.US1Input{
		Query: "Write an article about the top 2 stories on hackernews",
	}

	out, err := agent.RunUS1Example(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running US1 example: %v\n", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding output: %v\n", err)
		os.Exit(1)
	}
}
