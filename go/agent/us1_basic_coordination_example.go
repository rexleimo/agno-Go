package agent

// US1Input represents the input payload expected by the US1 parity scenario.
type US1Input struct {
	Query string `json:"query"`
}

// US1Output represents a normalized output structure for the US1 parity
// scenario. The internal shape can be refined later to more closely match
// the Python Team output.
type US1Output struct {
	Query  string                 `json:"query"`
	Result map[string]interface{} `json:"result,omitempty"`
}

// RunUS1Example is the Go-side entry point for the US1 parity scenario. It
// remains intentionally minimal at this stage and returns a deterministic
// placeholder result that mirrors the workflow metadata used in tests. The
// function will eventually be wired to the real Agent/Workflow execution path.
func RunUS1Example(input US1Input) (US1Output, error) {
	// Placeholder implementation: echo the query and attach a minimal Result
	// object. Later, this function will:
	//   - construct the appropriate Agents and Workflow
	//   - execute the Workflow with the given input
	//   - normalize the output into US1Output.Result
	out := US1Output{
		Query: input.Query,
		Result: map[string]interface{}{
			"workflow_id": "us1-basic-coordination",
			"status":      "placeholder",
		},
	}
	return out, nil
}
