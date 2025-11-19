package providers

import "strings"

// CustomDocument represents a simple document returned by the custom internal
// search provider.
type CustomDocument struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

// US3CustomInternalSearch is a minimal, in-memory implementation of a custom
// internal search provider used for the US3 parity scenario.
type US3CustomInternalSearch struct {
	Documents []CustomDocument
}

// DefaultUS3CustomInternalSearch constructs a provider instance with the same
// document set as the Python example.
func DefaultUS3CustomInternalSearch() US3CustomInternalSearch {
	return US3CustomInternalSearch{
		Documents: []CustomDocument{
			{
				ID:    "doc1",
				Title: "Internal API design guidelines",
				Tags:  []string{"internal", "api", "design"},
			},
			{
				ID:    "doc2",
				Title: "Billing service integration checklist",
				Tags:  []string{"billing", "integration"},
			},
			{
				ID:    "doc3",
				Title: "Search relevance tuning playbook",
				Tags:  []string{"search", "relevance"},
			},
		},
	}
}

// SearchDocuments performs a simple case-insensitive substring match on the
// document titles and returns all matches.
func (p US3CustomInternalSearch) SearchDocuments(query string) []CustomDocument {
	q := strings.ToLower(query)
	var out []CustomDocument
	for _, doc := range p.Documents {
		if strings.Contains(strings.ToLower(doc.Title), q) {
			out = append(out, doc)
		}
	}
	return out
}
