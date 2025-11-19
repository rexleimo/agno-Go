package workflow

// US1BasicCoordinationWorkflow constructs a workflow that mirrors the
// "basic coordination" team example in the Python cookbook.
func US1BasicCoordinationWorkflow() Workflow {
	steps := []Step{
		{
			ID:      StepID("search_hackernews"),
			AgentID: "hn_researcher",
			Name:    "Search HackerNews for relevant stories",
			Metadata: map[string]string{
				"phase": "research",
			},
		},
		{
			ID:      StepID("read_articles"),
			AgentID: "article_reader",
			Name:    "Read articles from URLs and summarize",
			Metadata: map[string]string{
				"phase": "synthesis",
			},
		},
	}

	rules := []RoutingRule{
		{
			From:      StepID("search_hackernews"),
			To:        StepID("read_articles"),
			Condition: "always",
		},
	}

	return Workflow{
		ID:          ID("us1-basic-coordination"),
		Name:        "US1 Basic Coordination",
		PatternType: PatternSequential,
		Steps:       steps,
		EntryPoints: []StepID{
			steps[0].ID,
		},
		TerminationCondition: TerminationCondition{
			MaxIterations: 1,
			OnError:       "fail-fast",
		},
		RoutingRules: rules,
	}
}
