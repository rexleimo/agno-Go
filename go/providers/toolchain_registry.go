package providers

import (
	"fmt"
	"strings"

	internalerrors "github.com/agno-agi/agno-go/go/internal/errors"
)

const (
	metadataFallbackProvider = "fallback_provider"
	metadataMigrationDoc     = "migration_doc"
)

// Toolkit describes a collection of providers and guardrails that power a set
// of tools in the runtime.
type Toolkit interface {
	ID() string
	Description() string
	Providers() []Provider
	Guardrails() []Guardrail
	Metadata() map[string]string
}

// Guardrail is a placeholder interface for guardrail implementations.
type Guardrail interface {
	ID() string
	Enforce(payload map[string]interface{}) error
}

// KnowledgeSource represents a retriever/dataset binding.
type KnowledgeSource interface {
	ID() string
	Provider() Provider
	Description() string
	Filters() map[string]string
}

// MemoryBinding wires a memory provider into the runtime.
type MemoryBinding interface {
	ID() string
	Provider() Provider
	Description() string
	Config() map[string]string
}

// GuardrailFactory builds concrete guardrail implementations based on the
// manifest spec. Projects can register custom factories for specific types.
type GuardrailFactory func(spec GuardrailSpec) (Guardrail, error)

// Registry manages the relationship between manifests and runtime bindings.
type Registry struct {
	providers          map[ID]Provider
	toolkits           map[string]*toolkitBinding
	knowledge          map[string]*knowledgeBinding
	memories           map[string]*memoryBinding
	guardrails         map[string]Guardrail
	guardrailFactories map[string]GuardrailFactory
}

// NewRegistry constructs a registry with the supplied base providers.
func NewRegistry(base []Provider) *Registry {
	r := &Registry{
		providers:          map[ID]Provider{},
		toolkits:           map[string]*toolkitBinding{},
		knowledge:          map[string]*knowledgeBinding{},
		memories:           map[string]*memoryBinding{},
		guardrails:         map[string]Guardrail{},
		guardrailFactories: map[string]GuardrailFactory{},
	}
	for _, p := range base {
		r.providers[p.ID] = p
	}
	// default guardrail factory simply stores the spec for later execution.
	r.guardrailFactories["default"] = func(spec GuardrailSpec) (Guardrail, error) {
		return staticGuardrail{spec: spec}, nil
	}
	return r
}

// RegisterGuardrailFactory allows consumers to override guardrail types.
func (r *Registry) RegisterGuardrailFactory(kind string, factory GuardrailFactory) {
	if factory == nil {
		return
	}
	r.guardrailFactories[strings.ToLower(kind)] = factory
}

// ApplyManifest loads providers, toolkits, knowledge sources and memories from
// the supplied manifest into the registry.
func (r *Registry) ApplyManifest(manifest *Manifest) error {
	if manifest == nil {
		return nil
	}
	for _, p := range manifest.ProviderConfigs() {
		r.providers[p.ID] = p
	}
	if err := r.loadGuardrails(manifest.Guardrails); err != nil {
		return err
	}
	if err := r.loadToolkits(manifest.Toolkits); err != nil {
		return err
	}
	if err := r.loadKnowledge(manifest.Knowledge); err != nil {
		return err
	}
	if err := r.loadMemories(manifest.Memories); err != nil {
		return err
	}
	return nil
}

// Toolkit retrieves a toolkit by ID.
func (r *Registry) Toolkit(id string) (Toolkit, bool) {
	tk, ok := r.toolkits[id]
	if !ok {
		return nil, false
	}
	return tk, true
}

// Knowledge retrieves a knowledge source by ID.
func (r *Registry) Knowledge(id string) (KnowledgeSource, bool) {
	ks, ok := r.knowledge[id]
	if !ok {
		return nil, false
	}
	return ks, true
}

// Memory retrieves a memory binding by ID.
func (r *Registry) Memory(id string) (MemoryBinding, bool) {
	mb, ok := r.memories[id]
	if !ok {
		return nil, false
	}
	return mb, true
}

func (r *Registry) loadGuardrails(specs []GuardrailSpec) error {
	for _, spec := range specs {
		typeKey := strings.ToLower(spec.Type)
		factory, ok := r.guardrailFactories[typeKey]
		if !ok {
			factory = r.guardrailFactories["default"]
		}
		guardrail, err := factory(spec)
		if err != nil {
			return err
		}
		r.guardrails[spec.ID] = guardrail
	}
	return nil
}

func (r *Registry) loadToolkits(specs []ToolkitSpec) error {
	for _, spec := range specs {
		providers, err := r.providersForIDs(spec.Providers, spec.Metadata)
		if err != nil {
			return err
		}
		guardrails := make([]Guardrail, 0, len(spec.Guardrails))
		for _, guardrailID := range spec.Guardrails {
			guardrail, ok := r.guardrails[guardrailID]
			if !ok {
				return newNotMigratedError(fmt.Sprintf("guardrail %s", guardrailID), spec.Metadata)
			}
			guardrails = append(guardrails, guardrail)
		}
		r.toolkits[spec.ID] = &toolkitBinding{
			id:          spec.ID,
			description: spec.Description,
			providers:   providers,
			guardrails:  guardrails,
			metadata:    cloneMetadata(spec.Metadata),
		}
	}
	return nil
}

func (r *Registry) loadKnowledge(specs []KnowledgeSpec) error {
	for _, spec := range specs {
		provider, err := r.providerForID(spec.Provider, spec.Filters)
		if err != nil {
			return err
		}
		r.knowledge[spec.ID] = &knowledgeBinding{
			id:          spec.ID,
			description: spec.Description,
			provider:    provider,
			filters:     cloneMetadata(spec.Filters),
		}
	}
	return nil
}

func (r *Registry) loadMemories(specs []MemorySpec) error {
	for _, spec := range specs {
		provider, err := r.providerForID(spec.Provider, spec.Config)
		if err != nil {
			return err
		}
		r.memories[spec.ID] = &memoryBinding{
			id:          spec.ID,
			description: spec.Description,
			provider:    provider,
			config:      cloneMetadata(spec.Config),
		}
	}
	return nil
}

func (r *Registry) providersForIDs(ids []string, metadata map[string]string) ([]Provider, error) {
	providers := make([]Provider, 0, len(ids))
	for _, id := range ids {
		provider, ok := r.providers[ID(id)]
		if !ok {
			return nil, newNotMigratedError(fmt.Sprintf("provider %s", id), metadata)
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func (r *Registry) providerForID(id string, metadata map[string]string) (Provider, error) {
	provider, ok := r.providers[ID(id)]
	if !ok {
		return Provider{}, newNotMigratedError(fmt.Sprintf("provider %s", id), metadata)
	}
	return provider, nil
}

type toolkitBinding struct {
	id          string
	description string
	providers   []Provider
	guardrails  []Guardrail
	metadata    map[string]string
}

func (t *toolkitBinding) ID() string                  { return t.id }
func (t *toolkitBinding) Description() string         { return t.description }
func (t *toolkitBinding) Providers() []Provider       { return append([]Provider(nil), t.providers...) }
func (t *toolkitBinding) Guardrails() []Guardrail     { return append([]Guardrail(nil), t.guardrails...) }
func (t *toolkitBinding) Metadata() map[string]string { return cloneMetadata(t.metadata) }

type knowledgeBinding struct {
	id          string
	description string
	provider    Provider
	filters     map[string]string
}

func (k *knowledgeBinding) ID() string                 { return k.id }
func (k *knowledgeBinding) Description() string        { return k.description }
func (k *knowledgeBinding) Provider() Provider         { return k.provider }
func (k *knowledgeBinding) Filters() map[string]string { return cloneMetadata(k.filters) }

type memoryBinding struct {
	id          string
	description string
	provider    Provider
	config      map[string]string
}

func (m *memoryBinding) ID() string                { return m.id }
func (m *memoryBinding) Description() string       { return m.description }
func (m *memoryBinding) Provider() Provider        { return m.provider }
func (m *memoryBinding) Config() map[string]string { return cloneMetadata(m.config) }

type staticGuardrail struct {
	spec GuardrailSpec
}

func (s staticGuardrail) ID() string { return s.spec.ID }

func (s staticGuardrail) Enforce(payload map[string]interface{}) error {
	return nil
}

// NotMigratedError augments the classified error with fallback information.
type NotMigratedError struct {
	Err      *internalerrors.Error
	Fallback string
	DocURL   string
}

func (e NotMigratedError) Error() string {
	msg := e.Err.Error()
	if e.Fallback != "" || e.DocURL != "" {
		msg = fmt.Sprintf("%s (fallback=%s, docs=%s)", msg, e.Fallback, e.DocURL)
	}
	return msg
}

func (e NotMigratedError) Unwrap() error {
	return e.Err
}

func newNotMigratedError(target string, metadata map[string]string) NotMigratedError {
	fallback := metadata[metadataFallbackProvider]
	doc := metadata[metadataMigrationDoc]
	return NotMigratedError{
		Err:      internalerrors.NewNotMigrated(fmt.Sprintf("%s is not migrated", target)),
		Fallback: fallback,
		DocURL:   doc,
	}
}

func cloneMetadata(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	clone := make(map[string]string, len(src))
	for k, v := range src {
		clone[k] = v
	}
	return clone
}
