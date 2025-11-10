# EvoLink 媒体智能体（Media Agents）

## 概览
本文演示如何在 Agno-Go 中使用 EvoLink Provider，通过 `pkg/agno/models/evolink/*` 生成文本、图片与视频。

## 环境准备
- 设置环境变量：
```bash
export EVO_API_KEY=sk-your-evolink-key
export EVO_BASE_URL=https://api.evolink.ai # 可选，默认即为该值
```

## 文本（Chat Completions）
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
  resp, _ := m.Invoke(context.Background(), &models.InvokeRequest{Messages: []*types.Message{ types.NewUserMessage("写一段 3 个镜头的海底城市分镜") }})
  fmt.Println(resp.Content)
}
```

## 图片（GPT-4O Images）
支持尺寸：`1:1`、`2:3`、`3:2`、`1024x1024`、`1024x1536`、`1536x1024`。`n ∈ {1,2,4}`。最多 5 个参考图。当提供 `mask_url` 时，必须仅有 1 张参考图，且遮罩必须为 PNG。

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
  resp, _ := m.Invoke(context.Background(), &models.InvokeRequest{Messages: []*types.Message{ types.NewUserMessage("由珊瑚构成的未来主义城市") }})
  fmt.Printf("task: %s, data: %#v\n", resp.ID, resp.Metadata.Extra["data"])
}
```

## 视频（Sora-2）
宽高比：`16:9` 或 `9:16`；时长：10 或 15 秒。可选的单张 `reference` 参考图与 `remove_watermark` 去水印标志。回调地址必须为 HTTPS。

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
  req := &models.InvokeRequest{Messages: []*types.Message{ types.NewUserMessage("清晨海洋生物学家与蝠鲼同游") }}
  resp, _ := v.Invoke(context.Background(), req)
  fmt.Printf("task: %s, status: %v\n", resp.ID, resp.Metadata.Extra["status"])
}
```

## 注意事项
- 生成的资源可能在 24 小时后过期，必要时请自行持久化。
- EvoLink 有严格的内容审核；Webhook 回调仅支持 HTTPS。
- 参考文档：GPT-4O 图片：https://docs.evolink.ai/en/api-manual/image-series/gpt-4o/gpt-4o-image-generation。本地 Sora-2 HTML 参考：`/home/rex/下载/Sora-2 视频生成 - EvoLink.html`。
