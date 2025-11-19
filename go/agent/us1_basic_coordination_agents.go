package agent

// US1 tool identifiers used by the "basic coordination" scenario.
const (
	ToolHackerNews   ToolID     = "hackernews-tools"
	ToolNewspaper4k  ToolID     = "newspaper4k-tools"
	ProviderOpenAIUS ProviderID = "openai-chat-gpt-5-mini"
)

// US1HackerNewsResearcher defines the Agent configuration for the
// HackerNews Researcher in the US1 scenario.
var US1HackerNewsResearcher = Agent{
	ID:          ID("hn_researcher"),
	Name:        "HackerNews Researcher",
	Role:        "Gets top stories from hackernews.",
	Description: "Uses HackerNews API to find and analyze relevant posts.",
	AllowedProviders: []ProviderID{
		ProviderOpenAIUS,
	},
	AllowedTools: []ToolID{
		ToolHackerNews,
	},
	InputSchema: Schema{
		Name:        "US1Query",
		Description: "User query describing the topic to research on HackerNews.",
	},
	OutputSchema: Schema{
		Name:        "US1HackerNewsFindings",
		Description: "Summarized findings and URLs from HackerNews.",
	},
	MemoryPolicy: MemoryPolicy{
		Persist:            false,
		WindowSize:         0,
		SensitiveFiltering: true,
	},
}

// US1ArticleReader defines the Agent configuration for the Article Reader
// in the US1 scenario.
var US1ArticleReader = Agent{
	ID:          ID("article_reader"),
	Name:        "Article Reader",
	Role:        "Reads articles from URLs.",
	Description: "Reads article content from URLs and summarizes them.",
	AllowedProviders: []ProviderID{
		ProviderOpenAIUS,
	},
	AllowedTools: []ToolID{
		ToolNewspaper4k,
	},
	InputSchema: Schema{
		Name:        "US1ArticleLinks",
		Description: "List of article URLs to read and summarize.",
	},
	OutputSchema: Schema{
		Name:        "US1ArticleSummaries",
		Description: "Summaries of the articles with reference links.",
	},
	MemoryPolicy: MemoryPolicy{
		Persist:            false,
		WindowSize:         0,
		SensitiveFiltering: true,
	},
}

// US1Agents returns the Agents participating in the US1 scenario.
func US1Agents() []Agent {
	return []Agent{
		US1HackerNewsResearcher,
		US1ArticleReader,
	}
}
