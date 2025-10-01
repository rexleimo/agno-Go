package workflow

import (
	"context"
	"fmt"
	"sync"
)

// Parallel represents a node that executes multiple nodes concurrently
type Parallel struct {
	ID    string
	Name  string
	Nodes []Node
}

// ParallelConfig contains parallel configuration
type ParallelConfig struct {
	ID    string
	Name  string
	Nodes []Node
}

// NewParallel creates a new parallel node
func NewParallel(config ParallelConfig) (*Parallel, error) {
	if len(config.Nodes) == 0 {
		return nil, fmt.Errorf("parallel node requires at least one child node")
	}

	if config.ID == "" {
		config.ID = fmt.Sprintf("parallel-%s", config.Name)
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	return &Parallel{
		ID:    config.ID,
		Name:  config.Name,
		Nodes: config.Nodes,
	}, nil
}

// Execute runs all child nodes in parallel
func (p *Parallel) Execute(ctx context.Context, execCtx *ExecutionContext) (*ExecutionContext, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)
	results := make([]*ExecutionContext, len(p.Nodes))

	for i, node := range p.Nodes {
		wg.Add(1)
		go func(idx int, n Node) {
			defer wg.Done()

			// Create a copy of the execution context for each parallel branch
			branchCtx := &ExecutionContext{
				Input:    execCtx.Input,
				Output:   execCtx.Output,
				Data:     make(map[string]interface{}),
				Metadata: make(map[string]interface{}),
			}

			// Copy data
			for k, v := range execCtx.Data {
				branchCtx.Data[k] = v
			}

			// Execute node
			result, err := n.Execute(ctx, branchCtx)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, node)
	}

	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		return nil, fmt.Errorf("parallel execution failed: %w", errors[0])
	}

	// Merge results back into main execution context
	// Store each branch result in the context
	for i, result := range results {
		if result != nil {
			execCtx.Set(fmt.Sprintf("parallel_%s_branch_%d_output", p.ID, i), result.Output)
			// Merge data with prefix to avoid conflicts
			for k, v := range result.Data {
				execCtx.Set(fmt.Sprintf("parallel_%s_branch_%d_%s", p.ID, i, k), v)
			}
		}
	}

	// Use the last result's output as the main output
	if len(results) > 0 && results[len(results)-1] != nil {
		execCtx.Output = results[len(results)-1].Output
	}

	return execCtx, nil
}

// GetID returns the parallel node ID
func (p *Parallel) GetID() string {
	return p.ID
}

// GetType returns the node type
func (p *Parallel) GetType() NodeType {
	return NodeTypeParallel
}
