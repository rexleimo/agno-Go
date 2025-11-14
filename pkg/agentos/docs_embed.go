package agentos

import _ "embed"

//go:embed openapi.yaml
var openAPISpec []byte

//go:embed docs.html
var docsHTML []byte
