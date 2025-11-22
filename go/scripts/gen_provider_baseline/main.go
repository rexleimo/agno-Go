package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	runtimeconfig "github.com/rexleimo/agno-go/internal/runtime/config"
	"github.com/rexleimo/agno-go/pkg/providers/cerebras"
	"github.com/rexleimo/agno-go/pkg/providers/gemini"
	"github.com/rexleimo/agno-go/pkg/providers/glm4"
	"github.com/rexleimo/agno-go/pkg/providers/groq"
	"github.com/rexleimo/agno-go/pkg/providers/modelscope"
	"github.com/rexleimo/agno-go/pkg/providers/ollama"
	"github.com/rexleimo/agno-go/pkg/providers/openai"
	"github.com/rexleimo/agno-go/pkg/providers/openrouter"
	"github.com/rexleimo/agno-go/pkg/providers/siliconflow"
)

// providerFixture mirrors the contract fixture shape we consume in tests.
type providerFixture struct {
	FixtureID string `json:"fixtureId" yaml:"fixtureId"`
	Provider  string `json:"provider" yaml:"provider"`
	Type      string `json:"type" yaml:"type"` // chat|embedding
	Input     struct {
		Messages []struct {
			Role    string `json:"role" yaml:"role"`
			Content string `json:"content" yaml:"content"`
		} `json:"messages,omitempty" yaml:"messages,omitempty"`
		Text    []string `json:"text,omitempty" yaml:"text,omitempty"`
		ModelID string   `json:"modelId" yaml:"modelId"`
	} `json:"input" yaml:"input"`
	Expected struct {
		Contains  string      `json:"contains,omitempty" yaml:"contains,omitempty"`
		MinTokens int         `json:"minTokens,omitempty" yaml:"minTokens,omitempty"`
		Vectors   [][]float64 `json:"vectors,omitempty" yaml:"vectors,omitempty"`
	} `json:"expected" yaml:"expected"`
	Tolerance struct {
		TokenTolerance  int     `json:"tokenTolerance" yaml:"tokenTolerance"`
		EmbeddingCosine float64 `json:"embeddingCosine" yaml:"embeddingCosine"`
	} `json:"tolerance" yaml:"tolerance"`
	SourceCommit string `json:"sourceCommit,omitempty" yaml:"sourceCommit,omitempty"`
	Notes        string `json:"notes,omitempty" yaml:"notes,omitempty"`
}

var (
	chatModels = map[agent.Provider]string{
		agent.ProviderOpenAI:      "gpt-4.1-mini",
		agent.ProviderGemini:      "gemini-2.5-flash",
		agent.ProviderGLM4:        "GLM-4-Flash-250414",
		agent.ProviderOpenRouter:  "tngtech/deepseek-r1t2-chimera:free",
		agent.ProviderSiliconFlow: "Qwen/Qwen2.5-7B-Instruct",
		agent.ProviderCerebras:    "llama3.1-8b",
		agent.ProviderModelScope:  "qwen2-7b-instruct",
		agent.ProviderGroq:        "llama-3.3-70b-versatile",
		agent.ProviderOllama:      "qwen3:4b",
	}
	embedModels = map[agent.Provider]string{
		agent.ProviderOpenAI:      "text-embedding-3-small",
		agent.ProviderGemini:      "text-embedding-004",
		agent.ProviderOpenRouter:  "text-embedding-3-small",
		agent.ProviderGroq:        "",            // TODO: Groq 官方未提供 embedding 模型，保持占位并跳过
		agent.ProviderGLM4:        "embedding-2", // TODO: 需付费/配额，当前账户 429，无免费额度
		agent.ProviderSiliconFlow: "BAAI/bge-large-zh-v1.5",
		agent.ProviderCerebras:    "llama3.1-8b",           // TODO: 401/404 鉴权失败，待有效 key 后生成
		agent.ProviderModelScope:  "BAAI/bge-base-en-v1.5", // TODO: 远端 EOF，不确定免费额度
		agent.ProviderOllama:      "nomic-embed-text",      // TODO: 本地 /api/embeddings 空返回，需适配或换 openai 兼容路径
	}
)

func main() {
	cfgPath := flag.String("config", filepath.FromSlash("../config/default.yaml"), "path to config YAML")
	envPath := flag.String("env", filepath.FromSlash("../.env"), "path to .env file")
	destDir := flag.String("dest", filepath.FromSlash("../specs/001-go-agno-rewrite/contracts/fixtures"), "destination fixtures directory")
	flag.Parse()

	cfg, err := runtimeconfig.LoadWithEnv(*cfgPath, *envPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := os.MkdirAll(*destDir, 0o755); err != nil {
		log.Fatalf("mkdir dest: %v", err)
	}

	statuses := cfg.ProviderStatuses()
	configs := cfg.ProviderConfigs()

	now := time.Now().UTC().Format(time.RFC3339)
	var generated int

	for _, st := range statuses {
		if st.Status != model.ProviderAvailable {
			log.Printf("skip %s: %s (missing=%v)", st.Provider, st.Status, st.MissingEnv)
			continue
		}
		prov := st.Provider
		cfgEntry := configs[prov]
		chatClient, embedClient := buildClients(prov, st, cfgEntry)
		if chatClient != nil {
			if err := writeChatFixture(*destDir, prov, chatClient, now); err != nil {
				log.Printf("chat %s: %v", prov, err)
			} else {
				generated++
			}
		}
		if embedClient != nil {
			if err := writeEmbedFixture(*destDir, prov, embedClient, now); err != nil {
				log.Printf("embed %s: %v", prov, err)
			} else {
				generated++
			}
		}
	}

	log.Printf("fixtures generated/updated: %d -> %s", generated, *destDir)
}

func buildClients(p agent.Provider, st model.ProviderStatus, cfg runtimeconfig.ProviderConfig) (model.ChatProvider, model.EmbeddingProvider) {
	switch p {
	case agent.ProviderOpenAI:
		return openai.New(cfg.Endpoint, cfg.APIKey, st.MissingEnv), openai.New(cfg.Endpoint, cfg.APIKey, st.MissingEnv)
	case agent.ProviderGemini:
		c := gemini.New(cfg.Endpoint, cfg.APIKey, st.MissingEnv)
		return c, c
	case agent.ProviderGLM4:
		return glm4.New(st, cfg.Endpoint, cfg.APIKey), glm4.NewEmbed(st, cfg.Endpoint, cfg.APIKey)
	case agent.ProviderOpenRouter:
		headers := map[string]string{
			"HTTP-Referer": "https://local.agno",
			"X-Title":      "Go-Agno",
		}
		return openrouter.New(st, cfg.Endpoint, cfg.APIKey, headers), openrouter.NewEmbed(st, cfg.Endpoint, cfg.APIKey, headers)
	case agent.ProviderSiliconFlow:
		return siliconflow.New(st, cfg.Endpoint, cfg.APIKey), siliconflow.NewEmbed(st, cfg.Endpoint, cfg.APIKey)
	case agent.ProviderCerebras:
		return cerebras.New(st, cfg.Endpoint, cfg.APIKey), cerebras.NewEmbed(st, cfg.Endpoint, cfg.APIKey)
	case agent.ProviderModelScope:
		return modelscope.New(st, cfg.Endpoint, cfg.APIKey), modelscope.NewEmbed(st, cfg.Endpoint, cfg.APIKey)
	case agent.ProviderGroq:
		return groq.New(st, cfg.Endpoint, cfg.APIKey), groq.NewEmbed(st, cfg.Endpoint, cfg.APIKey)
	case agent.ProviderOllama:
		return ollama.New(st, cfg.Endpoint, cfg.APIKey), ollama.NewEmbed(st, cfg.Endpoint, cfg.APIKey)
	default:
		return nil, nil
	}
}

func writeChatFixture(dest string, p agent.Provider, client model.ChatProvider, now string) error {
	modelID := chosenModel(p, false)
	if modelID == "" {
		return fmt.Errorf("no chat model for %s", p)
	}
	prompt := fmt.Sprintf("Respond ONLY with: BASELINE-%s", strings.ToUpper(string(p)))
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	_, err := client.Chat(ctx, model.ChatRequest{
		Model:    agent.ModelConfig{Provider: p, ModelID: modelID},
		Messages: []agent.Message{{Role: agent.RoleUser, Content: prompt}},
	})
	if err != nil {
		return err
	}
	fx := providerFixture{}
	fx.FixtureID = fmt.Sprintf("chat-%s-baseline", p)
	fx.Provider = string(p)
	fx.Type = "chat"
	fx.Input.ModelID = modelID
	fx.Input.Messages = []struct {
		Role    string `json:"role" yaml:"role"`
		Content string `json:"content" yaml:"content"`
	}{{Role: "user", Content: prompt}}
	fx.Expected.Contains = fmt.Sprintf("BASELINE-%s", strings.ToUpper(string(p)))
	fx.Expected.MinTokens = 0
	fx.Tolerance.TokenTolerance = 2
	fx.Tolerance.EmbeddingCosine = 0.98
	fx.SourceCommit = "go-generated"
	fx.Notes = fmt.Sprintf("Generated %s via live provider call", now)

	path := filepath.Join(dest, fmt.Sprintf("chat_%s.json", p))
	return writeFixtureFile(path, fx)
}

func writeEmbedFixture(dest string, p agent.Provider, client model.EmbeddingProvider, now string) error {
	modelID := chosenModel(p, true)
	if modelID == "" {
		return fmt.Errorf("no embed model for %s", p)
	}
	text := fmt.Sprintf("baseline embedding text for %s", p)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	resp, err := client.Embed(ctx, model.EmbeddingRequest{
		Model: agent.ModelConfig{Provider: p, ModelID: modelID},
		Input: []string{text},
	})
	if err != nil {
		return err
	}
	if len(resp.Vectors) == 0 {
		return fmt.Errorf("empty embedding from %s", p)
	}
	fx := providerFixture{}
	fx.FixtureID = fmt.Sprintf("embedding-%s-baseline", p)
	fx.Provider = string(p)
	fx.Type = "embedding"
	fx.Input.ModelID = modelID
	fx.Input.Text = []string{text}
	fx.Expected.Vectors = resp.Vectors
	fx.Tolerance.TokenTolerance = 2
	fx.Tolerance.EmbeddingCosine = 0.98
	fx.SourceCommit = "go-generated"
	fx.Notes = fmt.Sprintf("Generated %s via live provider call", now)

	path := filepath.Join(dest, fmt.Sprintf("embedding_%s.json", p))
	return writeFixtureFile(path, fx)
}

func writeFixtureFile(path string, fx providerFixture) error {
	data, err := json.MarshalIndent(fx, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return nil
}

func chosenModel(p agent.Provider, embedding bool) string {
	envKey := fmt.Sprintf("%s_MODEL", strings.ToUpper(string(p)))
	if embedding {
		envKey = fmt.Sprintf("%s_EMBED_MODEL", strings.ToUpper(string(p)))
	}
	if val := strings.TrimSpace(os.Getenv(envKey)); val != "" {
		return val
	}
	if embedding {
		return embedModels[p]
	}
	return chatModels[p]
}
