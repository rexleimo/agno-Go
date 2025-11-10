package video

import (
    "context"
    "net/http"
    "strings"
    "time"

    evolinkp "github.com/rexleimo/agno-go/pkg/agno/providers/evolink"
    provvid "github.com/rexleimo/agno-go/pkg/agno/providers/evolink/video"
    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

// Config for Evolink video model
type Config struct {
    APIKey     string
    BaseURL    string
    Timeout    time.Duration
    HTTPClient *http.Client

    // Defaults
    AspectRatio     string
    DurationSeconds int
    RemoveWatermark bool
}

// Video implements models.Model for EvoLink video generations
type Video struct {
    models.BaseModel
    client *evolinkp.Client
    config Config
}

// New creates a new Evolink video model
func New(modelID string, cfg Config) (*Video, error) {
    c, err := evolinkp.NewClient(evolinkp.Config{APIKey: cfg.APIKey, BaseURL: cfg.BaseURL, Timeout: cfg.Timeout, HTTPClient: cfg.HTTPClient})
    if err != nil { return nil, err }
    if cfg.AspectRatio == "" { cfg.AspectRatio = "16:9" }
    if cfg.DurationSeconds == 0 { cfg.DurationSeconds = 10 }
    return &Video{ BaseModel: models.BaseModel{ID: modelID, Provider: "evolink"}, client: c, config: cfg }, nil
}

// Invoke triggers a video generation task
func (v *Video) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
    prompt := lastUserPrompt(req)
    ar := stringFromExtra(req, "aspect_ratio", v.config.AspectRatio)
    dur := intFromExtra(req, "duration", v.config.DurationSeconds)
    ref := stringFromExtra(req, "reference", "")
    rmw := boolFromExtra(req, "remove_watermark", v.config.RemoveWatermark)
    cb := stringFromExtra(req, "callback_url", "")

    resp, err := provvid.Generate(ctx, v.client, provvid.Options{Prompt: prompt, AspectRatio: ar, DurationSeconds: dur, ReferenceURL: ref, RemoveWatermark: rmw, CallbackURL: cb})
    if err != nil { return nil, err }
    return &types.ModelResponse{
        ID:      resp.TaskID,
        Content: "video task completed",
        Model:   v.ID,
        Metadata: types.Metadata{ Extra: map[string]interface{}{"status": resp.Status, "data": resp.Data} },
    }, nil
}

// InvokeStream is not supported for video tasks
func (v *Video) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
    return nil, types.NewAPIError("streaming not supported", nil)
}

func lastUserPrompt(req *models.InvokeRequest) string {
    if req == nil || len(req.Messages) == 0 { return "" }
    for i := len(req.Messages)-1; i >= 0; i-- {
        m := req.Messages[i]
        if string(m.Role) == string(types.RoleUser) && strings.TrimSpace(m.Content) != "" {
            return m.Content
        }
    }
    return req.Messages[len(req.Messages)-1].Content
}

func stringFromExtra(req *models.InvokeRequest, key, def string) string {
    if req != nil && req.Extra != nil {
        if v, ok := req.Extra[key]; ok {
            if s, ok := v.(string); ok && s != "" { return s }
        }
    }
    return def
}
func intFromExtra(req *models.InvokeRequest, key string, def int) int {
    if req != nil && req.Extra != nil {
        if v, ok := req.Extra[key]; ok {
            switch t := v.(type) {
            case int: if t>0 { return t }
            case float64: if int(t) > 0 { return int(t) }
            }
        }
    }
    return def
}
func boolFromExtra(req *models.InvokeRequest, key string, def bool) bool {
    if req != nil && req.Extra != nil {
        if v, ok := req.Extra[key]; ok {
            if b, ok := v.(bool); ok { return b }
        }
    }
    return def
}

