package run

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewE2BExecutorValidation(t *testing.T) {
	t.Setenv("E2B_API_KEY", "")
	tests := []struct {
		name    string
		config  E2BConfig
		wantErr string
	}{
		{name: "missing key", config: E2BConfig{TemplateID: "template"}, wantErr: "api key"},
		{name: "missing template", config: E2BConfig{APIKey: "key"}, wantErr: "template ID"},
		{name: "missing resource shell", config: E2BConfig{APIKey: "key", TemplateID: "template"}, wantErr: "resource shell"},
		{name: "bad api URL", config: E2BConfig{APIKey: "key", TemplateID: "template", ResourceShell: "bash", APIBaseURL: "://bad"}, wantErr: "api base URL"},
		{name: "bad sandbox URL", config: E2BConfig{APIKey: "key", TemplateID: "template", ResourceShell: "bash", SandboxBaseURL: "ftp://example.test"}, wantErr: "sandbox base URL"},
		{name: "shell has arguments", config: E2BConfig{APIKey: "key", TemplateID: "template", ResourceShell: "bash -c"}, wantErr: "without arguments"},
		{name: "valid", config: E2BConfig{APIKey: "key", TemplateID: "template", ResourceShell: "bash"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewE2BExecutor(tt.config)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("NewE2BExecutor() error = %v", err)
				}
				if _, ok := executor.(*e2bExecutor); !ok {
					t.Fatalf("NewE2BExecutor() type = %T, want *e2bExecutor", executor)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NewE2BExecutor() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestNewExecutorE2BIsExplicit(t *testing.T) {
	executor, err := NewExecutor(Config{
		Backend: "e2b",
		E2B:     &E2BConfig{APIKey: "key", TemplateID: "template", ResourceShell: "bash"},
	})
	if err != nil {
		t.Fatalf("NewExecutor() error = %v", err)
	}
	if _, ok := executor.(*e2bExecutor); !ok {
		t.Fatalf("NewExecutor() type = %T, want *e2bExecutor", executor)
	}

	_, err = NewExecutor(Config{Backend: "e2b"})
	if err == nil || !strings.Contains(err.Error(), "e2b configuration is required") {
		t.Fatalf("NewExecutor() error = %v, want e2b configuration error", err)
	}
}

func TestE2BRunLifecycle(t *testing.T) {
	var mu sync.Mutex
	var created, uploaded, started, deleted bool
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-API-Key") != "test-key" && request.URL.Path != "/files" && request.URL.Path != "/process.Process/Start" {
			t.Errorf("X-API-Key = %q, want test-key", request.Header.Get("X-API-Key"))
		}
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/sandboxes":
			var body struct {
				TemplateID          string            `json:"templateID"`
				Secure              bool              `json:"secure"`
				AllowInternetAccess bool              `json:"allow_internet_access"`
				EnvVars             map[string]string `json:"envVars"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode create request: %v", err)
			}
			if body.TemplateID != "hno-template" || !body.Secure || body.AllowInternetAccess {
				t.Fatalf("unsafe create request: %#v", body)
			}
			if body.EnvVars["SAFE"] != "value" {
				t.Fatalf("EnvVars = %#v, want SAFE=value", body.EnvVars)
			}
			mu.Lock()
			created = true
			mu.Unlock()
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"sandboxID":"sandbox-1","envdAccessToken":"access-token"}`))
		case request.Method == http.MethodPost && request.URL.Path == "/files":
			if request.Header.Get("X-Access-Token") != "access-token" || request.Header.Get("E2b-Sandbox-Id") != "sandbox-1" {
				t.Fatal("upload missing sandbox credentials")
			}
			if request.URL.Query().Get("path") != "/tmp/agno-payload.py" {
				t.Fatalf("upload path = %q", request.URL.Query().Get("path"))
			}
			payload, _ := io.ReadAll(request.Body)
			if string(payload) != "print('safe')" {
				t.Fatalf("payload = %q", payload)
			}
			mu.Lock()
			uploaded = true
			mu.Unlock()
			writer.WriteHeader(http.StatusOK)
		case request.Method == http.MethodPost && request.URL.Path == "/process.Process/Start":
			var body struct {
				Process struct {
					Cmd  string   `json:"cmd"`
					Args []string `json:"args"`
				} `json:"process"`
			}
			payload, _, err := readConnectEnvelope(request.Body)
			if err != nil {
				t.Fatalf("read process envelope: %v", err)
			}
			if err := json.Unmarshal(payload, &body); err != nil {
				t.Fatalf("decode process request: %v", err)
			}
			if body.Process.Cmd != "bash" || !strings.Contains(strings.Join(body.Process.Args, " "), "/tmp/agno-payload.py") {
				t.Fatalf("process = %#v, want bash wrapper with uploaded payload", body.Process)
			}
			mu.Lock()
			started = true
			mu.Unlock()
			writer.Header().Set("Content-Type", "application/connect+json")
			stdout := base64.StdEncoding.EncodeToString([]byte("safe output\n"))
			stderr := base64.StdEncoding.EncodeToString([]byte("safe error\n"))
			_, _ = writer.Write(connectEnvelope([]byte(`{"event":{"data":{"stdout":"` + stdout + `"}}}`)))
			_, _ = writer.Write(connectEnvelope([]byte(`{"event":{"data":{"stderr":"` + stderr + `"}}}`)))
			_, _ = writer.Write(connectEnvelope([]byte(`{"event":{"end":{"status":"exit status 0"}}}`)))
		case request.Method == http.MethodDelete && request.URL.Path == "/sandboxes/sandbox-1":
			mu.Lock()
			deleted = true
			mu.Unlock()
			writer.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	executor, err := NewE2BExecutor(E2BConfig{
		APIKey:         "test-key",
		TemplateID:     "hno-template",
		ResourceShell:  "bash",
		APIBaseURL:     server.URL,
		SandboxBaseURL: server.URL,
		HTTPClient:     server.Client(),
	})
	if err != nil {
		t.Fatalf("NewE2BExecutor() error = %v", err)
	}
	result, err := executor.Run(context.Background(), Spec{
		Runtime: "python",
		Code:    "print('safe')",
		Env:     []string{"SAFE=value"},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ExitCode != 0 || result.Stdout != "safe output\n" || result.Stderr != "safe error\n" {
		t.Fatalf("Run() result = %#v", result)
	}
	mu.Lock()
	defer mu.Unlock()
	if !created || !uploaded || !started || !deleted {
		t.Fatalf("lifecycle = create:%t upload:%t start:%t delete:%t", created, uploaded, started, deleted)
	}
}

func TestE2BRunRejectsWorkspace(t *testing.T) {
	executor, err := NewE2BExecutor(E2BConfig{APIKey: "key", TemplateID: "template", ResourceShell: "bash"})
	if err != nil {
		t.Fatalf("NewE2BExecutor() error = %v", err)
	}
	_, err = executor.Run(context.Background(), Spec{Runtime: "python", Code: "print(1)", Workspace: "."})
	if err == nil || !strings.Contains(err.Error(), "workspace mounts") {
		t.Fatalf("Run() error = %v, want workspace rejection", err)
	}
}

func TestE2BRunCleansSandboxWhenSecureTokenMissing(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method + " " + request.URL.Path {
		case http.MethodPost + " /sandboxes":
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusCreated)
			_, _ = writer.Write([]byte(`{"sandboxID":"sandbox-without-token","envdAccessToken":null}`))
		case http.MethodDelete + " /sandboxes/sandbox-without-token":
			deleted = true
			writer.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	executor, err := NewE2BExecutor(E2BConfig{
		APIKey:         "key",
		TemplateID:     "template",
		ResourceShell:  "bash",
		APIBaseURL:     server.URL,
		SandboxBaseURL: server.URL,
		HTTPClient:     server.Client(),
	})
	if err != nil {
		t.Fatalf("NewE2BExecutor() error = %v", err)
	}
	_, err = executor.Run(context.Background(), Spec{Runtime: "python", Code: "print(1)"})
	if err == nil || !strings.Contains(err.Error(), "secure sandbox credentials") {
		t.Fatalf("Run() error = %v, want secure credential error", err)
	}
	if !deleted {
		t.Fatal("Run() did not delete sandbox after missing access token")
	}
}

func TestE2BRunCleansSandboxAfterPartialCreateResponse(t *testing.T) {
	deleted := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method + " " + request.URL.Path {
		case http.MethodPost + " /sandboxes":
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusCreated)
			// The ID is valid but the access token has an invalid type. The
			// provider must still delete the already-created sandbox.
			_, _ = writer.Write([]byte(`{"sandboxID":"partially-decoded","envdAccessToken":123}`))
		case http.MethodDelete + " /sandboxes/partially-decoded":
			deleted = true
			writer.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
	}))
	defer server.Close()

	executor, err := NewE2BExecutor(E2BConfig{
		APIKey:         "key",
		TemplateID:     "template",
		ResourceShell:  "bash",
		APIBaseURL:     server.URL,
		SandboxBaseURL: server.URL,
		HTTPClient:     server.Client(),
	})
	if err != nil {
		t.Fatalf("NewE2BExecutor() error = %v", err)
	}
	_, err = executor.Run(context.Background(), Spec{Runtime: "python", Code: "print(1)"})
	if err == nil || !strings.Contains(err.Error(), "decode e2b sandbox") {
		t.Fatalf("Run() error = %v, want decode error", err)
	}
	if !deleted {
		t.Fatal("Run() did not delete sandbox after partial create response")
	}
}

func TestE2BEventBytesAndExitCode(t *testing.T) {
	if got := string(e2bEventBytes(base64.StdEncoding.EncodeToString([]byte("hello")))); got != "hello" {
		t.Fatalf("e2bEventBytes() = %q, want hello", got)
	}
	if got := string(e2bEventBytes("not-base64")); got != "not-base64" {
		t.Fatalf("e2bEventBytes() = %q, want raw fallback", got)
	}
	if got := parseE2BExitCode("exit status 17"); got != 17 {
		t.Fatalf("parseE2BExitCode() = %d, want 17", got)
	}
	if got := parseE2BExitCode("unknown"); got != -1 {
		t.Fatalf("parseE2BExitCode() = %d, want -1", got)
	}
}
