package image

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rexleimo/agno-go/pkg/agno/models"
	evolinkp "github.com/rexleimo/agno-go/pkg/agno/providers/evolink"
	provimg "github.com/rexleimo/agno-go/pkg/agno/providers/evolink/image"
	"github.com/rexleimo/agno-go/pkg/agno/types"
)

// ImageModel is an alias to the provider-level model identifiers for EvoLink images.
type ImageModel = provimg.Model

const (
	ModelGPT4O             ImageModel = provimg.ModelGPT4O
	ModelSeedream40        ImageModel = provimg.ModelSeedream40
	ModelNanoBanana        ImageModel = provimg.ModelNanoBanana
	ModelQwenImageEdit     ImageModel = provimg.ModelQwenImageEdit
	ModelWan25TextToImage  ImageModel = provimg.ModelWan25TextToImage
	ModelWan25ImageToImage ImageModel = provimg.ModelWan25ImageToImage
)

// Config for Evolink image model
type Config struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
	Model      ImageModel

	// Defaults
	Size string
	N    int
}

// Image implements models.Model for EvoLink image generations
type Image struct {
	models.BaseModel
	client *evolinkp.Client
	config Config
}

// New creates a new Evolink image model
func New(modelID string, cfg Config) (*Image, error) {
	c, err := evolinkp.NewClient(evolinkp.Config{APIKey: cfg.APIKey, BaseURL: cfg.BaseURL, Timeout: cfg.Timeout, HTTPClient: cfg.HTTPClient})
	if err != nil {
		return nil, err
	}
	if cfg.Size == "" {
		cfg.Size = "1:1"
	}
	if cfg.N == 0 {
		cfg.N = 1
	}
	if cfg.Model == "" {
		cfg.Model = ModelGPT4O
	}
	return &Image{BaseModel: models.BaseModel{ID: modelID, Provider: "evolink"}, client: c, config: cfg}, nil
}

// Invoke triggers an image generation task
func (im *Image) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
	prompt := lastUserPrompt(req)
	model := modelFromExtra(req, "model", im.config.Model)
	size := stringFromExtra(req, "size", im.config.Size)
	n := intFromExtra(req, "n", im.config.N)
	refs := stringSliceFromExtra(req, "references")
	mask := stringFromExtra(req, "mask_url", "")
	cb := stringFromExtra(req, "callback_url", "")

	resp, err := provimg.Generate(ctx, im.client, provimg.Options{Model: provimg.Model(model), Prompt: prompt, Size: size, N: n, References: refs, MaskURL: mask, CallbackURL: cb})
	if err != nil {
		return nil, err
	}

	return &types.ModelResponse{
		ID:       resp.TaskID,
		Content:  "image task completed",
		Model:    im.ID,
		Metadata: types.Metadata{Extra: map[string]interface{}{"status": resp.Status, "data": resp.Data}},
	}, nil
}

// InvokeStream is not supported for image tasks
func (im *Image) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
	return nil, types.NewAPIError("streaming not supported", nil)
}

func lastUserPrompt(req *models.InvokeRequest) string {
	if req == nil || len(req.Messages) == 0 {
		return ""
	}
	for i := len(req.Messages) - 1; i >= 0; i-- {
		m := req.Messages[i]
		if string(m.Role) == string(types.RoleUser) && strings.TrimSpace(m.Content) != "" {
			return m.Content
		}
	}
	// fallback
	return req.Messages[len(req.Messages)-1].Content
}

func stringFromExtra(req *models.InvokeRequest, key, def string) string {
	if req != nil && req.Extra != nil {
		if v, ok := req.Extra[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return def
}
func intFromExtra(req *models.InvokeRequest, key string, def int) int {
	if req != nil && req.Extra != nil {
		if v, ok := req.Extra[key]; ok {
			switch t := v.(type) {
			case int:
				if t > 0 {
					return t
				}
			case float64:
				if int(t) > 0 {
					return int(t)
				}
			}
		}
	}
	return def
}
func stringSliceFromExtra(req *models.InvokeRequest, key string) []string {
	if req != nil && req.Extra != nil {
		if v, ok := req.Extra[key]; ok {
			switch vv := v.(type) {
			case []string:
				return vv
			case []interface{}:
				var out []string
				for _, it := range vv {
					if s, ok := it.(string); ok {
						out = append(out, s)
					}
				}
				return out
			}
		}
	}
	return nil
}

func modelFromExtra(req *models.InvokeRequest, key string, def ImageModel) ImageModel {
	if req != nil && req.Extra != nil {
		if v, ok := req.Extra[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return ImageModel(s)
			}
		}
	}
	return def
}
