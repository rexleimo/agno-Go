package shared

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
)

// OpenAICompatClient implements a minimal OpenAI-compatible REST/SSE client for providers
// that expose a /chat/completions and /embeddings surface.
type OpenAICompatClient struct {
	provider            agent.Provider
	endpoint            string
	apiKey              string
	status              model.ProviderStatus
	http                *http.Client
	chatPath            string
	embedPath           string
	authHeader          string
	extraHeaders        map[string]string
	parseUsage          bool
	parseToolArgs       bool
	errorParser         func(map[string]any) string
	usageExtractorChat  func(resp oaCompatChatResp) agent.Usage
	usageExtractorEmbed func(resp oaCompatEmbedResp) agent.Usage
}

// Config customizes OpenAI-compatible routes and headers.
type Config struct {
	Endpoint            string
	APIKey              string
	Status              model.ProviderStatus
	ChatPath            string
	EmbedPath           string
	AuthHeader          string
	ExtraHeaders        map[string]string
	Timeout             time.Duration
	ParseUsage          bool
	ParseToolArgs       bool
	ErrorParser         func(body map[string]any) string
	UsageExtractorChat  func(resp oaCompatChatResp) agent.Usage
	UsageExtractorEmbed func(resp oaCompatEmbedResp) agent.Usage
}

type oaCompatChatReq struct {
	Model       string        `json:"model"`
	Messages    []oaCompatMsg `json:"messages"`
	Stream      bool          `json:"stream,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type oaCompatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type oaCompatChatResp struct {
	Choices []struct {
		Delta struct {
			Content   string         `json:"content"`
			ToolCalls []oaCompatTool `json:"tool_calls"`
		} `json:"delta"`
		Message struct {
			Content   string         `json:"content"`
			ToolCalls []oaCompatTool `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type oaCompatTool struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type oaCompatEmbedReq struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type oaCompatEmbedResp struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OACompatChatResp is an exported alias for usage extractors in other packages.
type OACompatChatResp = oaCompatChatResp

// OACompatEmbedResp is an exported alias for usage extractors in other packages.
type OACompatEmbedResp = oaCompatEmbedResp

// NewOpenAICompat builds a client with the given provider name, endpoint, and API key.
func NewOpenAICompat(provider agent.Provider, cfg Config) *OpenAICompatClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	chatPath := cfg.ChatPath
	if chatPath == "" {
		chatPath = "/chat/completions"
	}
	embedPath := cfg.EmbedPath
	if embedPath == "" {
		embedPath = "/embeddings"
	}
	authHeader := cfg.AuthHeader
	if authHeader == "" {
		authHeader = "Authorization"
	}
	parseToolArgs := cfg.ParseToolArgs
	if !cfg.ParseToolArgs {
		parseToolArgs = true
	}
	return &OpenAICompatClient{
		provider:            provider,
		endpoint:            strings.TrimSuffix(cfg.Endpoint, "/"),
		apiKey:              cfg.APIKey,
		status:              cfg.Status,
		http:                &http.Client{Timeout: timeout},
		chatPath:            chatPath,
		embedPath:           embedPath,
		authHeader:          authHeader,
		extraHeaders:        cfg.ExtraHeaders,
		parseUsage:          cfg.ParseUsage,
		parseToolArgs:       parseToolArgs,
		errorParser:         cfg.ErrorParser,
		usageExtractorChat:  cfg.UsageExtractorChat,
		usageExtractorEmbed: cfg.UsageExtractorEmbed,
	}
}

func (c *OpenAICompatClient) Name() agent.Provider { return c.provider }

func (c *OpenAICompatClient) Status() model.ProviderStatus { return c.status }

func (c *OpenAICompatClient) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	if c.status.Status != model.ProviderAvailable {
		return nil, model.ErrProviderUnavailable
	}
	body := oaCompatChatReq{
		Model:       req.Model.ModelID,
		Messages:    toOAMessages(req.Messages),
		Stream:      false,
		MaxTokens:   req.Model.MaxTokens,
		Temperature: req.Model.Temperature,
	}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+c.chatPath, bytes.NewReader(payload))
	c.applyHeaders(httpReq)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out oaCompatChatResp
	if resp.StatusCode >= 400 {
		return nil, c.errorForStatus(resp, &out)
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Choices) == 0 {
		return nil, fmt.Errorf("%s empty response", c.provider)
	}
	choice := out.Choices[0]
	msg := agent.Message{
		Role:      agent.RoleAssistant,
		Content:   choice.Message.Content,
		ToolCalls: toToolCalls(choice.Message.ToolCalls, c.parseToolArgs),
	}
	respUsage := agent.Usage{}
	if c.parseUsage {
		switch {
		case c.usageExtractorChat != nil:
			respUsage = c.usageExtractorChat(out)
		default:
			respUsage = defaultUsageFromChat(out)
		}
	}
	return &model.ChatResponse{
		Message:      msg,
		Usage:        respUsage,
		FinishReason: choice.FinishReason,
	}, nil
}

func (c *OpenAICompatClient) Stream(ctx context.Context, req model.ChatRequest, fn model.StreamHandler) error {
	if c.status.Status != model.ProviderAvailable {
		return model.ErrProviderUnavailable
	}
	body := oaCompatChatReq{
		Model:       req.Model.ModelID,
		Messages:    toOAMessages(req.Messages),
		Stream:      true,
		MaxTokens:   req.Model.MaxTokens,
		Temperature: req.Model.Temperature,
	}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+c.chatPath, bytes.NewReader(payload))
	c.applyHeaders(httpReq)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.errorForStatus(resp, nil)
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk oaCompatChatResp
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return err
		}
		for _, choice := range chunk.Choices {
			if delta := strings.TrimSpace(choice.Delta.Content); delta != "" {
				if err := fn(model.ChatStreamEvent{Type: "token", Delta: delta}); err != nil {
					return err
				}
			}
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range toToolCalls(choice.Delta.ToolCalls, c.parseToolArgs) {
					tcCopy := tc
					if err := fn(model.ChatStreamEvent{Type: "tool_call", ToolCall: &tcCopy}); err != nil {
						return err
					}
				}
			}
			if choice.FinishReason != "" {
				if err := fn(model.ChatStreamEvent{Type: "end", Done: true, FinishReason: choice.FinishReason}); err != nil {
					return err
				}
				return nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return fn(model.ChatStreamEvent{Type: "end", Done: true})
}

func (c *OpenAICompatClient) Embed(ctx context.Context, req model.EmbeddingRequest) (*model.EmbeddingResponse, error) {
	if c.status.Status != model.ProviderAvailable {
		return nil, model.ErrProviderUnavailable
	}
	body := oaCompatEmbedReq{
		Model: req.Model.ModelID,
		Input: req.Input,
	}
	payload, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+c.embedPath, bytes.NewReader(payload))
	c.applyHeaders(httpReq)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out oaCompatEmbedResp
	if resp.StatusCode >= 400 {
		return nil, c.errorForStatus(resp, &out)
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	vectors := make([][]float64, len(out.Data))
	for i, d := range out.Data {
		vectors[i] = d.Embedding
	}
	usage := agent.Usage{}
	if c.parseUsage {
		switch {
		case c.usageExtractorEmbed != nil:
			usage = c.usageExtractorEmbed(out)
		default:
			usage = defaultUsageFromEmbed(out)
		}
	}
	return &model.EmbeddingResponse{Vectors: vectors, Usage: usage}, nil
}

func toOAMessages(msgs []agent.Message) []oaCompatMsg {
	out := make([]oaCompatMsg, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, oaCompatMsg{Role: string(m.Role), Content: m.Content})
	}
	return out
}

func toToolCalls(calls []oaCompatTool, parseArgs bool) []agent.ToolCall {
	out := make([]agent.ToolCall, 0, len(calls))
	for _, c := range calls {
		var args map[string]any
		if parseArgs {
			_ = json.Unmarshal([]byte(c.Function.Arguments), &args)
		}
		if len(args) == 0 {
			args = map[string]any{"arguments": c.Function.Arguments}
		}
		out = append(out, agent.ToolCall{
			ToolCallID: c.ID,
			Name:       c.Function.Name,
			Args:       args,
		})
	}
	return out
}

func defaultUsageFromChat(resp oaCompatChatResp) agent.Usage {
	return agent.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		LatencyMs:        0,
	}
}

func defaultUsageFromEmbed(resp oaCompatEmbedResp) agent.Usage {
	return agent.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		LatencyMs:        0,
	}
}

func (c *OpenAICompatClient) applyHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set(c.authHeader, "Bearer "+c.apiKey)
	}
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}
}

func (c *OpenAICompatClient) errorForStatus(resp *http.Response, fallback any) error {
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err == nil {
		if c.errorParser != nil {
			if msg := c.errorParser(body); msg != "" {
				return fmt.Errorf("%s error: %s (%s)", c.provider, resp.Status, msg)
			}
		}
		if msg, ok := body["error"].(string); ok && msg != "" {
			return fmt.Errorf("%s error: %s (%s)", c.provider, resp.Status, msg)
		}
		if errObj, ok := body["error"].(map[string]any); ok {
			if msg, ok := errObj["message"].(string); ok && msg != "" {
				return fmt.Errorf("%s error: %s (%s)", c.provider, resp.Status, msg)
			}
			if code, ok := errObj["code"].(string); ok && code != "" {
				return fmt.Errorf("%s error: %s (%s)", c.provider, resp.Status, code)
			}
			if bodyStr, ok := errObj["body"].(string); ok && bodyStr != "" {
				return fmt.Errorf("%s error: %s (%s)", c.provider, resp.Status, bodyStr)
			}
		}
	}
	// fallback: marshal known partial response if provided
	if fallback != nil {
		if b, err := json.Marshal(fallback); err == nil && len(b) > 0 {
			return fmt.Errorf("%s error: %s (%s)", c.provider, resp.Status, strings.TrimSpace(string(b)))
		}
	}
	return fmt.Errorf("%s error: %s", c.provider, resp.Status)
}
