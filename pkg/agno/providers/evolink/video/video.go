package video

import (
    "context"
    "fmt"
    "net/url"
    "strings"
    "time"

    evolink "github.com/rexleimo/agno-go/pkg/agno/providers/evolink"
)

// Options defines Sora-2 video generation parameters
type Options struct {
    Prompt           string
    AspectRatio      string // 16:9 or 9:16
    DurationSeconds  int    // 10 or 15
    ReferenceURL     string // optional, <=1 image
    RemoveWatermark  bool
    CallbackURL      string // optional HTTPS only
}

func (o *Options) validate() error {
    if strings.TrimSpace(o.Prompt) == "" {
        return fmt.Errorf("prompt is required")
    }
    ar := o.AspectRatio
    if ar != "16:9" && ar != "9:16" {
        return fmt.Errorf("invalid aspect_ratio: %s", ar)
    }
    if o.DurationSeconds != 10 && o.DurationSeconds != 15 {
        return fmt.Errorf("invalid duration, must be 10 or 15")
    }
    if o.ReferenceURL != "" {
        if _, err := url.ParseRequestURI(o.ReferenceURL); err != nil {
            return fmt.Errorf("invalid reference url: %w", err)
        }
    }
    if o.CallbackURL != "" {
        u, err := url.ParseRequestURI(o.CallbackURL)
        if err != nil {
            return fmt.Errorf("invalid callback url: %w", err)
        }
        if u.Scheme != "https" {
            return fmt.Errorf("callback url must be https")
        }
    }
    return nil
}

// Response is a minimal video task completion payload
type Response struct {
    TaskID string                 `json:"task_id"`
    Status string                 `json:"status"`
    Data   map[string]interface{} `json:"data"`
}

// Generate creates a video generation task and polls until completion
func Generate(ctx context.Context, c *evolink.Client, opts Options) (*Response, error) {
    if err := opts.validate(); err != nil {
        return nil, err
    }
    payload := map[string]interface{}{
        "prompt":          opts.Prompt,
        "aspect_ratio":    opts.AspectRatio,
        "duration_seconds": opts.DurationSeconds,
        "remove_watermark": opts.RemoveWatermark,
    }
    if opts.ReferenceURL != "" {
        payload["reference_url"] = opts.ReferenceURL
    }
    if opts.CallbackURL != "" {
        payload["callback_url"] = opts.CallbackURL
    }

    var createResp struct{ TaskID string `json:"task_id"` }
    if err := c.PostJSON(ctx, "/v1/videos/generations", payload, &createResp); err != nil {
        return nil, err
    }
    tr, err := c.PollTask(ctx, createResp.TaskID, 2*time.Second)
    if err != nil {
        return nil, err
    }
    return &Response{TaskID: tr.ID, Status: tr.Status, Data: tr.Data}, nil
}
