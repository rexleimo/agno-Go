package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rexleimo/agno-go/pkg/agno/models"
	"github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		config    Config
		wantErr   bool
		errString string
	}{
		{
			name:    "valid config",
			modelID: "gemini-pro",
			config: Config{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			modelID: "gemini-pro",
			config:  Config{},
			wantErr: true,
		},
		{
			name:    "custom base URL",
			modelID: "gemini-pro",
			config: Config{
				APIKey:  "test-key",
				BaseURL: "https://custom.api.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := New(tt.modelID, tt.config)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if model.GetID() != tt.modelID {
				t.Errorf("expected model ID %s, got %s", tt.modelID, model.GetID())
			}

			if model.GetProvider() != "gemini" {
				t.Errorf("expected provider 'gemini', got %s", model.GetProvider())
			}
		})
	}
}

func TestInvoke(t *testing.T) {
	tests := []struct {
		name           string
		modelID        string
		request        *models.InvokeRequest
		serverResponse GeminiResponse
		wantContent    string
		wantErr        bool
	}{
		{
			name:    "simple text response",
			modelID: "gemini-pro",
			request: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "Hello"},
				},
			},
			serverResponse: GeminiResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: "Hello! How can I help you?"},
							},
						},
						FinishReason: "STOP",
					},
				},
				UsageMetadata: UsageMetadata{
					PromptTokenCount:     5,
					CandidatesTokenCount: 10,
					TotalTokenCount:      15,
				},
			},
			wantContent: "Hello! How can I help you?",
			wantErr:     false,
		},
		{
			name:    "with system message",
			modelID: "gemini-pro",
			request: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleSystem, Content: "You are a helpful assistant"},
					{Role: types.RoleUser, Content: "Hi"},
				},
			},
			serverResponse: GeminiResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: "Hi there!"},
							},
						},
						FinishReason: "STOP",
					},
				},
				UsageMetadata: UsageMetadata{
					TotalTokenCount: 10,
				},
			},
			wantContent: "Hi there!",
			wantErr:     false,
		},
		{
			name:    "with tool call",
			modelID: "gemini-pro",
			request: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "What's the weather?"},
				},
				Tools: []models.ToolDefinition{
					{
						Type: "function",
						Function: models.FunctionSchema{
							Name:        "get_weather",
							Description: "Get weather information",
							Parameters: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"location": map[string]interface{}{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
			serverResponse: GeminiResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{
									FunctionCall: &FunctionCall{
										Name: "get_weather",
										Args: map[string]interface{}{
											"location": "San Francisco",
										},
									},
								},
							},
						},
						FinishReason: "STOP",
					},
				},
				UsageMetadata: UsageMetadata{
					TotalTokenCount: 20,
				},
			},
			wantContent: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("expected POST request, got %s", r.Method)
				}

				// Verify request body
				var req GeminiRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("failed to decode request: %v", err)
				}

				// Send response
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			// Create Gemini model with test server URL
			model, err := New(tt.modelID, Config{
				APIKey:  "test-key",
				BaseURL: server.URL,
			})
			if err != nil {
				t.Fatalf("failed to create model: %v", err)
			}

			// Invoke model
			resp, err := model.Invoke(context.Background(), tt.request)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Content != tt.wantContent {
				t.Errorf("expected content %q, got %q", tt.wantContent, resp.Content)
			}

			if resp.Usage.TotalTokens != tt.serverResponse.UsageMetadata.TotalTokenCount {
				t.Errorf("expected total tokens %d, got %d",
					tt.serverResponse.UsageMetadata.TotalTokenCount,
					resp.Usage.TotalTokens)
			}
		})
	}
}

func TestInvokeError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "400 bad request",
			statusCode: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "401 unauthorized",
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "500 server error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": "test error"}`))
			}))
			defer server.Close()

			model, err := New("gemini-pro", Config{
				APIKey:  "test-key",
				BaseURL: server.URL,
			})
			if err != nil {
				t.Fatalf("failed to create model: %v", err)
			}

			req := &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "test"},
				},
			}

			_, err = model.Invoke(context.Background(), req)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestInvokeStream(t *testing.T) {
	modelID := "gemini-pro"
	expectedChunks := []string{"Hello", " there", "!"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("streaming not supported")
		}

		for _, chunk := range expectedChunks {
			resp := GeminiResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: chunk},
							},
						},
					},
				},
			}

			data, _ := json.Marshal(resp)
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}))
	defer server.Close()

	model, err := New(modelID, Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			{Role: types.RoleUser, Content: "test"},
		},
	}

	chunks, err := model.InvokeStream(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var receivedContent []string
	for chunk := range chunks {
		if chunk.Error != nil && chunk.Error.Error() != "EOF" {
			t.Fatalf("chunk error: %v", chunk.Error)
		}
		if chunk.Content != "" {
			receivedContent = append(receivedContent, chunk.Content)
		}
		if chunk.Done {
			break
		}
	}

	if len(receivedContent) < 1 {
		t.Errorf("expected at least 1 chunk, got %d", len(receivedContent))
	}

	for i, expected := range expectedChunks {
		if i >= len(receivedContent) {
			break
		}
		if receivedContent[i] != expected {
			t.Errorf("chunk %d: expected %q, got %q", i, expected, receivedContent[i])
		}
	}
}

func TestBuildGeminiRequest(t *testing.T) {
	model := &Gemini{
		BaseModel: models.BaseModel{
			ID:       "gemini-pro",
			Provider: "gemini",
		},
		config: Config{
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	}

	tests := []struct {
		name     string
		request  *models.InvokeRequest
		validate func(*testing.T, *GeminiRequest)
	}{
		{
			name: "simple message",
			request: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "Hello"},
				},
			},
			validate: func(t *testing.T, req *GeminiRequest) {
				if len(req.Contents) != 1 {
					t.Errorf("expected 1 content, got %d", len(req.Contents))
				}
				if req.Contents[0].Role != "user" {
					t.Errorf("expected role 'user', got %s", req.Contents[0].Role)
				}
			},
		},
		{
			name: "with system instruction",
			request: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleSystem, Content: "You are helpful"},
					{Role: types.RoleUser, Content: "Hi"},
				},
			},
			validate: func(t *testing.T, req *GeminiRequest) {
				if req.SystemInstruction == nil {
					t.Fatal("expected system instruction, got nil")
				}
				if len(req.SystemInstruction.Parts) == 0 {
					t.Fatal("expected system instruction parts")
				}
			},
		},
		{
			name: "with tools",
			request: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "Hello"},
				},
				Tools: []models.ToolDefinition{
					{
						Type: "function",
						Function: models.FunctionSchema{
							Name:        "test_func",
							Description: "Test function",
						},
					},
				},
			},
			validate: func(t *testing.T, req *GeminiRequest) {
				if len(req.Tools) != 1 {
					t.Fatalf("expected 1 tool, got %d", len(req.Tools))
				}
				if len(req.Tools[0].FunctionDeclarations) != 1 {
					t.Fatalf("expected 1 function declaration, got %d",
						len(req.Tools[0].FunctionDeclarations))
				}
				if req.Tools[0].FunctionDeclarations[0].Name != "test_func" {
					t.Errorf("expected function name 'test_func', got %s",
						req.Tools[0].FunctionDeclarations[0].Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geminiReq := model.buildGeminiRequest(tt.request)
			tt.validate(t, geminiReq)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			config:  Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConvertResponse(t *testing.T) {
	model := &Gemini{
		BaseModel: models.BaseModel{
			ID: "gemini-pro",
		},
	}

	tests := []struct {
		name     string
		response *GeminiResponse
		validate func(*testing.T, *types.ModelResponse)
	}{
		{
			name: "text response",
			response: &GeminiResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{Text: "Hello world"},
							},
						},
						FinishReason: "STOP",
					},
				},
				UsageMetadata: UsageMetadata{
					TotalTokenCount: 10,
				},
			},
			validate: func(t *testing.T, resp *types.ModelResponse) {
				if resp.Content != "Hello world" {
					t.Errorf("expected content 'Hello world', got %q", resp.Content)
				}
			},
		},
		{
			name: "function call response",
			response: &GeminiResponse{
				Candidates: []Candidate{
					{
						Content: Content{
							Parts: []Part{
								{
									FunctionCall: &FunctionCall{
										Name: "test_func",
										Args: map[string]interface{}{"key": "value"},
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, resp *types.ModelResponse) {
				if len(resp.ToolCalls) != 1 {
					t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
				}
				if resp.ToolCalls[0].Function.Name != "test_func" {
					t.Errorf("expected function name 'test_func', got %s",
						resp.ToolCalls[0].Function.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := model.convertResponse(tt.response)
			tt.validate(t, resp)
		})
	}
}

func TestSSEDecoder(t *testing.T) {
	data := "data: {\"test\": \"message1\"}\n\ndata: {\"test\": \"message2\"}\n\n"

	decoder := NewSSEDecoder(strings.NewReader(data))

	messages := []string{}
	for {
		msg, err := decoder.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("failed to read message: %v", err)
		}

		var result map[string]string
		if err := json.Unmarshal(msg, &result); err != nil {
			// Skip invalid JSON
			continue
		}

		if val, ok := result["test"]; ok {
			messages = append(messages, val)
		}
	}

	if len(messages) < 1 {
		t.Errorf("expected at least 1 message, got %d", len(messages))
	}

	// Check that we got message1
	found := false
	for _, msg := range messages {
		if msg == "message1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find 'message1' in messages, got %v", messages)
	}
}
