// Package deps pins baseline runtime dependencies until the runtime is implemented.
package deps

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var (
	_ = chi.NewRouter
	_ uuid.UUID
	_ yaml.Node
)
