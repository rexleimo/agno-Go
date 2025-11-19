package providers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"sigs.k8s.io/yaml"
)

// Manifest represents a declarative provider registry describing toolkits,
// knowledge bases, memories and guardrails. It is intentionally schema-light
// so teams can version their manifests alongside fixtures.
type Manifest struct {
	Providers  []ProviderSpec  `json:"providers" yaml:"providers"`
	Toolkits   []ToolkitSpec   `json:"toolkits" yaml:"toolkits"`
	Knowledge  []KnowledgeSpec `json:"knowledge" yaml:"knowledge"`
	Memories   []MemorySpec    `json:"memories" yaml:"memories"`
	Guardrails []GuardrailSpec `json:"guardrails" yaml:"guardrails"`
}

// ProviderSpec maps 1:1 to the Provider struct exported by this package.
type ProviderSpec struct {
	ID           string                 `json:"id" yaml:"id"`
	Type         string                 `json:"type" yaml:"type"`
	DisplayName  string                 `json:"display_name" yaml:"display_name"`
	Capabilities []string               `json:"capabilities" yaml:"capabilities"`
	Config       map[string]interface{} `json:"config" yaml:"config"`
	Metadata     map[string]string      `json:"metadata" yaml:"metadata"`
}

// ToolkitSpec collects providers and metadata for a higher-level toolkit.
type ToolkitSpec struct {
	ID          string            `json:"id" yaml:"id"`
	Description string            `json:"description" yaml:"description"`
	Providers   []string          `json:"providers" yaml:"providers"`
	Guardrails  []string          `json:"guardrails" yaml:"guardrails"`
	Metadata    map[string]string `json:"metadata" yaml:"metadata"`
}

// KnowledgeSpec references retrievers or datasets exposed to workflows.
type KnowledgeSpec struct {
	ID          string            `json:"id" yaml:"id"`
	Description string            `json:"description" yaml:"description"`
	Provider    string            `json:"provider" yaml:"provider"`
	Filters     map[string]string `json:"filters" yaml:"filters"`
}

// MemorySpec describes how conversational memory should be sourced/stored.
type MemorySpec struct {
	ID          string            `json:"id" yaml:"id"`
	Description string            `json:"description" yaml:"description"`
	Provider    string            `json:"provider" yaml:"provider"`
	Config      map[string]string `json:"config" yaml:"config"`
}

// GuardrailSpec is a serialized representation of a guardrail entry. The
// actual enforcement contract is captured by the Guardrail interface below.
type GuardrailSpec struct {
	ID          string                 `json:"id" yaml:"id"`
	Type        string                 `json:"type" yaml:"type"`
	Description string                 `json:"description" yaml:"description"`
	Enforcement string                 `json:"enforcement" yaml:"enforcement"`
	Config      map[string]interface{} `json:"config" yaml:"config"`
}

// LoadManifest loads a manifest from a YAML or JSON file path.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	return DecodeManifest(data)
}

// DecodeManifest parses manifest bytes (JSON or YAML) into a Manifest struct.
func DecodeManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		converted, convErr := yaml.YAMLToJSON(data)
		if convErr != nil {
			return nil, fmt.Errorf("decode manifest: %w", err)
		}
		if err := json.Unmarshal(converted, &m); err != nil {
			return nil, fmt.Errorf("decode manifest: %w", err)
		}
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate enforces minimal structural rules on the manifest such as required
// IDs and duplicate detection.
func (m Manifest) Validate() error {
	var issues []string

	providerIDs := map[string]struct{}{}
	for i, p := range m.Providers {
		id := strings.TrimSpace(p.ID)
		if id == "" {
			issues = append(issues, fmt.Sprintf("providers[%d].id missing", i))
		} else {
			if _, exists := providerIDs[id]; exists {
				issues = append(issues, fmt.Sprintf("providers[%d].id duplicate", i))
			}
			providerIDs[id] = struct{}{}
		}
		if strings.TrimSpace(p.Type) == "" {
			issues = append(issues, fmt.Sprintf("providers[%d].type missing", i))
		}
	}

	for i, tk := range m.Toolkits {
		if strings.TrimSpace(tk.ID) == "" {
			issues = append(issues, fmt.Sprintf("toolkits[%d].id missing", i))
		}
	}

	for i, k := range m.Knowledge {
		if strings.TrimSpace(k.ID) == "" {
			issues = append(issues, fmt.Sprintf("knowledge[%d].id missing", i))
		}
	}

	for i, mem := range m.Memories {
		if strings.TrimSpace(mem.ID) == "" {
			issues = append(issues, fmt.Sprintf("memories[%d].id missing", i))
		}
	}

	for i, g := range m.Guardrails {
		if strings.TrimSpace(g.ID) == "" {
			issues = append(issues, fmt.Sprintf("guardrails[%d].id missing", i))
		}
		if strings.TrimSpace(g.Type) == "" {
			issues = append(issues, fmt.Sprintf("guardrails[%d].type missing", i))
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("invalid manifest: %s", strings.Join(issues, "; "))
	}
	return nil
}

// ProviderConfigs converts manifest provider specs into Provider structs.
func (m Manifest) ProviderConfigs() []Provider {
	out := make([]Provider, 0, len(m.Providers))
	for _, spec := range m.Providers {
		out = append(out, Provider{
			ID:           ID(spec.ID),
			Type:         Type(spec.Type),
			DisplayName:  spec.DisplayName,
			Config:       cloneConfig(spec.Config),
			Capabilities: convertCapabilities(spec.Capabilities),
		})
	}
	return out
}

func convertCapabilities(values []string) []Capability {
	if len(values) == 0 {
		return nil
	}
	result := make([]Capability, 0, len(values))
	for _, v := range values {
		result = append(result, Capability(strings.ToLower(v)))
	}
	return result
}

func cloneConfig(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = cloneValue(v)
	}
	return dst
}

func cloneValue(v interface{}) interface{} {
	switch value := v.(type) {
	case map[string]interface{}:
		return cloneConfig(value)
	case []interface{}:
		result := make([]interface{}, len(value))
		for i, item := range value {
			result[i] = cloneValue(item)
		}
		return result
	default:
		return value
	}
}

// MergeProviders merges manifest providers with an existing catalog and
// returns a deduplicated list. Later entries overwrite earlier duplicates.
func MergeProviders(base []Provider, manifest []Provider) []Provider {
	index := make(map[ID]int, len(base)+len(manifest))
	out := make([]Provider, 0, len(base)+len(manifest))

	for _, p := range base {
		index[p.ID] = len(out)
		out = append(out, p)
	}

	for _, p := range manifest {
		if idx, exists := index[p.ID]; exists {
			out[idx] = p
			continue
		}
		index[p.ID] = len(out)
		out = append(out, p)
	}
	return out
}

// EqualManifests performs a deep-equality comparison, primarily for tests.
func EqualManifests(a, b *Manifest) bool {
	return reflect.DeepEqual(a, b)
}
