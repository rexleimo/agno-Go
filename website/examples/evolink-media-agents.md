# EvoLink Media Agents

## Overview
This example shows how to use the EvoLink provider inside Agno-Go to generate text, images, and video via `pkg/agno/models/evolink/*`.

## Setup
- Export env variables:
```bash
export EVO_API_KEY=sk-your-evolink-key
export EVO_BASE_URL=https://api.evolink.ai # optional, defaults to this value
```

## Text (Chat Completions)
```go
package main

import (
  "context"
  "fmt"
  "os"

  evoText "github.com/rexleimo/agno-go/pkg/agno/models/evolink/text"
  "github.com/rexleimo/agno-go/pkg/agno/models"
  "github.com/rexleimo/agno-go/pkg/agno/types"
)

func main() {
  m, _ := evoText.New("evo-gpt-4o", evoText.Config{APIKey: os.Getenv("EVO_API_KEY"), BaseURL: os.Getenv("EVO_BASE_URL"), Temperature: 0.6})
  resp, _ := m.Invoke(context.Background(), &models.InvokeRequest{Messages: []*types.Message{ types.NewUserMessage("Write a 3-shot underwater city storyboard") }})
  fmt.Println(resp.Content)
}
```

## Images (GPT-4O Images)
Allowed sizes: `1:1`, `2:3`, `3:2`, `1024x1024`, `1024x1536`, `1536x1024`. `n ∈ {1,2,4}`. Up to 5 references. If `mask_url` is provided, exactly one reference is required and the mask must be a PNG.

```go
package main

import (
  "context"
  "fmt"
  "os"

  evoImg "github.com/rexleimo/agno-go/pkg/agno/models/evolink/image"
  "github.com/rexleimo/agno-go/pkg/agno/models"
  "github.com/rexleimo/agno-go/pkg/agno/types"
)

func main() {
  m, _ := evoImg.New("evo-gpt-4o-images", evoImg.Config{APIKey: os.Getenv("EVO_API_KEY"), BaseURL: os.Getenv("EVO_BASE_URL"), Size: "1:1", N: 2})
  resp, _ := m.Invoke(context.Background(), &models.InvokeRequest{Messages: []*types.Message{ types.NewUserMessage("A futuristic city made of coral") }})
  fmt.Printf("task: %s, data: %#v\n", resp.ID, resp.Metadata.Extra["data"])
}
```

## Video (Sora-2)
Aspect ratios: `16:9` or `9:16`. Durations: 10 or 15 seconds. Optional single image `reference` and `remove_watermark` flag. Callbacks must use HTTPS.

```go
package main

import (
  "context"
  "fmt"
  "os"

  evoVid "github.com/rexleimo/agno-go/pkg/agno/models/evolink/video"
  "github.com/rexleimo/agno-go/pkg/agno/models"
  "github.com/rexleimo/agno-go/pkg/agno/types"
)

func main() {
  v, _ := evoVid.New("evo-sora-2", evoVid.Config{APIKey: os.Getenv("EVO_API_KEY"), BaseURL: os.Getenv("EVO_BASE_URL"), AspectRatio: "9:16", DurationSeconds: 15})
  req := &models.InvokeRequest{Messages: []*types.Message{ types.NewUserMessage("A marine biologist swims with manta rays at dawn") }}
  resp, _ := v.Invoke(context.Background(), req)
  fmt.Printf("task: %s, status: %v\n", resp.ID, resp.Metadata.Extra["status"])
}
```

## Notes
- Assets may expire after 24h; persist them if needed.
- EvoLink enforces strict moderation and HTTPS-only webhooks.
- Reference docs: GPT-4O images: https://docs.evolink.ai/en/api-manual/image-series/gpt-4o/gpt-4o-image-generation. Local Sora-2 HTML reference: `/home/rex/下载/Sora-2 视频生成 - EvoLink.html`.
