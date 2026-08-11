package run

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultE2BAPIBaseURL      = "https://api.e2b.app"
	defaultE2BSandboxBaseURL  = "https://sandbox.e2b.app"
	e2bSandboxPort            = "49983"
	e2bCleanupTimeout         = 5 * time.Second
	connectEnvelopeHeaderSize = 5
	connectFlagCompressed     = 0x01
	connectFlagEndStream      = 0x02
	maxConnectMessageBytes    = 8 << 20
)

var e2bExitStatus = regexp.MustCompile(`exit status (-?\d+)`)

// e2bExecutor is a thin adapter over E2B Cloud's public REST and Connect
// APIs. E2B currently publishes JavaScript and Python SDKs; keeping this
// adapter small and protocol-bound avoids depending on an unofficial Go SDK.
type e2bExecutor struct {
	apiKey         string
	templateID     string
	apiBaseURL     string
	sandboxBaseURL string
	httpClient     *http.Client
	timeout        time.Duration
	memoryLimit    int64
	pidLimit       int64
	outputLimit    int64
}

// NewE2BExecutor creates an explicit E2B Cloud executor. It never falls back
// to a local process or container provider. APIKey defaults to E2B_API_KEY;
// TemplateID must be supplied because it defines the E2B VM image and limits.
func NewE2BExecutor(config E2BConfig) (Executor, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("E2B_API_KEY")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("e2b api key is required")
	}
	if strings.TrimSpace(config.TemplateID) == "" {
		return nil, fmt.Errorf("e2b template ID is required")
	}
	apiBaseURL, err := validBaseURL(config.APIBaseURL, defaultE2BAPIBaseURL)
	if err != nil {
		return nil, fmt.Errorf("e2b api base URL: %w", err)
	}
	sandboxBaseURL, err := validBaseURL(config.SandboxBaseURL, defaultE2BSandboxBaseURL)
	if err != nil {
		return nil, fmt.Errorf("e2b sandbox base URL: %w", err)
	}
	client := config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return &e2bExecutor{
		apiKey:         apiKey,
		templateID:     config.TemplateID,
		apiBaseURL:     apiBaseURL,
		sandboxBaseURL: sandboxBaseURL,
		httpClient:     client,
		timeout:        DefaultTimeout,
		memoryLimit:    DefaultMemoryLimit,
		pidLimit:       DefaultPidLimit,
		outputLimit:    DefaultOutputLimit,
	}, nil
}

// Run creates a secure, offline E2B sandbox, uploads the code as a temporary
// file, executes it, and destroys the sandbox even after a timeout or failure.
func (e *e2bExecutor) Run(ctx context.Context, spec Spec) (result Result, err error) {
	if err := validateSpec(spec); err != nil {
		return Result{}, err
	}
	if spec.Workspace != "" {
		return Result{}, fmt.Errorf("e2b sandbox workspace mounts are not implemented")
	}
	spec = e.defaults(spec)
	template := runtimeTemplates[spec.Runtime]

	started := time.Now()
	defer func() { result.Duration = time.Since(started) }()
	runCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()

	sandbox, err := e.createSandbox(runCtx, spec)
	if err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			result.TimedOut = true
			return result, fmt.Errorf("%w after %s", ErrTimeout, spec.Timeout)
		}
		return result, err
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), e2bCleanupTimeout)
		defer cleanupCancel()
		if cleanupErr := e.deleteSandbox(cleanupCtx, sandbox.ID); cleanupErr != nil {
			err = errors.Join(err, cleanupErr)
		}
	}()
	if sandbox.AccessToken == "" {
		return result, fmt.Errorf("e2b sandbox response missing secure sandbox credentials")
	}

	payloadPath := "/tmp/agno-payload" + template.FileExtension
	if err := e.uploadPayload(runCtx, sandbox, payloadPath, spec.Code); err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			result.TimedOut = true
			return result, fmt.Errorf("%w after %s", ErrTimeout, spec.Timeout)
		}
		return result, err
	}

	result, err = e.startProcess(runCtx, sandbox, template, payloadPath, spec)
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		return result, fmt.Errorf("%w after %s", ErrTimeout, spec.Timeout)
	}
	return result, err
}

// Close releases no resources. Every E2B sandbox is deleted by Run.
func (e *e2bExecutor) Close() error { return nil }

func (e *e2bExecutor) defaults(spec Spec) Spec {
	if spec.Timeout == 0 {
		spec.Timeout = e.timeout
	}
	if spec.MemoryLimit == 0 {
		spec.MemoryLimit = e.memoryLimit
	}
	if spec.PidLimit == 0 {
		spec.PidLimit = e.pidLimit
	}
	if spec.OutputLimit == 0 {
		spec.OutputLimit = e.outputLimit
	}
	return spec
}

type e2bSandbox struct {
	ID          string `json:"sandboxID"`
	AccessToken string `json:"envdAccessToken"`
}

func (e *e2bExecutor) createSandbox(ctx context.Context, spec Spec) (e2bSandbox, error) {
	ttlSeconds := int(math.Ceil(spec.Timeout.Seconds())) + int(e2bCleanupTimeout.Seconds())
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}
	envs, err := environmentMap(spec.Env)
	if err != nil {
		return e2bSandbox{}, err
	}
	body, err := json.Marshal(struct {
		TemplateID          string            `json:"templateID"`
		Timeout             int               `json:"timeout"`
		Secure              bool              `json:"secure"`
		AllowInternetAccess bool              `json:"allow_internet_access"`
		EnvVars             map[string]string `json:"envVars,omitempty"`
	}{
		TemplateID:          e.templateID,
		Timeout:             ttlSeconds,
		Secure:              true,
		AllowInternetAccess: false,
		EnvVars:             envs,
	})
	if err != nil {
		return e2bSandbox{}, fmt.Errorf("encode e2b create request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e.apiBaseURL+"/sandboxes", bytes.NewReader(body))
	if err != nil {
		return e2bSandbox{}, fmt.Errorf("create e2b sandbox request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", e.apiKey)

	response, err := e.httpClient.Do(request)
	if err != nil {
		return e2bSandbox{}, fmt.Errorf("create e2b sandbox: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return e2bSandbox{}, e2bResponseError("create e2b sandbox", response)
	}
	var sandbox e2bSandbox
	if err := json.NewDecoder(response.Body).Decode(&sandbox); err != nil {
		return e2bSandbox{}, fmt.Errorf("decode e2b sandbox: %w", err)
	}
	if sandbox.ID == "" {
		return e2bSandbox{}, fmt.Errorf("e2b sandbox response missing sandbox ID")
	}
	return sandbox, nil
}

func (e *e2bExecutor) uploadPayload(ctx context.Context, sandbox e2bSandbox, path, code string) error {
	endpoint := e.sandboxBaseURL + "/files?" + url.Values{"path": []string{path}}.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(code))
	if err != nil {
		return fmt.Errorf("create e2b upload request: %w", err)
	}
	e.sandboxHeaders(request, sandbox)
	request.Header.Set("Content-Type", "application/octet-stream")

	response, err := e.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("upload e2b payload: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return e2bResponseError("upload e2b payload", response)
	}
	return nil
}

func (e *e2bExecutor) startProcess(ctx context.Context, sandbox e2bSandbox, template runtimeTemplate, payloadPath string, spec Spec) (Result, error) {
	if len(template.FileEntrypoint) != 1 {
		return Result{}, fmt.Errorf("e2b runtime template requires one file entrypoint")
	}
	// E2B's template defines VM-level resource limits. This wrapper applies
	// per-run virtual-memory and process ceilings before replacing itself with
	// the runtime process.
	arguments := []string{
		"-c",
		`ulimit -v "$1" && ulimit -u "$2" && exec "$3" "$4"`,
		"--",
		strconv.FormatInt((spec.MemoryLimit+1023)/1024, 10),
		strconv.FormatInt(spec.PidLimit, 10),
		template.FileEntrypoint[0],
		payloadPath,
	}
	body, err := json.Marshal(struct {
		Process struct {
			Cmd  string   `json:"cmd"`
			Args []string `json:"args"`
		} `json:"process"`
		Stdin bool `json:"stdin"`
	}{
		Process: struct {
			Cmd  string   `json:"cmd"`
			Args []string `json:"args"`
		}{Cmd: "bash", Args: arguments},
		Stdin: false,
	})
	if err != nil {
		return Result{}, fmt.Errorf("encode e2b process request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e.sandboxBaseURL+"/process.Process/Start", bytes.NewReader(connectEnvelope(body)))
	if err != nil {
		return Result{}, fmt.Errorf("create e2b process request: %w", err)
	}
	e.sandboxHeaders(request, sandbox)
	request.Header.Set("Content-Type", "application/connect+json")
	request.Header.Set("Accept", "application/connect+json")
	request.Header.Set("Connect-Protocol-Version", "1")

	response, err := e.httpClient.Do(request)
	if err != nil {
		return Result{}, fmt.Errorf("start e2b process: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Result{}, e2bResponseError("start e2b process", response)
	}

	var stdout, stderr limitedBuffer
	stdout.max = int(spec.OutputLimit)
	stderr.max = int(spec.OutputLimit)
	result := Result{ExitCode: -1}
	for {
		payload, endStream, err := readConnectEnvelope(response.Body)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return result, fmt.Errorf("read e2b process event: %w", err)
		}
		if endStream {
			var streamError struct {
				Error *struct {
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(payload, &streamError); err == nil && streamError.Error != nil {
				return result, fmt.Errorf("e2b process stream: %s", streamError.Error.Message)
			}
			continue
		}
		var message e2bProcessEvent
		if err := json.Unmarshal(payload, &message); err != nil {
			return result, fmt.Errorf("decode e2b process event: %w", err)
		}
		if message.Event.Data != nil {
			if message.Event.Data.Stdout != "" {
				_, _ = stdout.Write(e2bEventBytes(message.Event.Data.Stdout))
			}
			if message.Event.Data.Stderr != "" {
				_, _ = stderr.Write(e2bEventBytes(message.Event.Data.Stderr))
			}
		}
		if message.Event.End != nil {
			result.ExitCode = parseE2BExitCode(message.Event.End.Status)
			if message.Event.End.Error != "" {
				result.Stdout, result.Stderr = stdout.String(), stderr.String()
				return result, fmt.Errorf("e2b process: %s", message.Event.End.Error)
			}
		}
	}
	result.Stdout, result.Stderr = stdout.String(), stderr.String()
	if result.ExitCode == -1 {
		return result, fmt.Errorf("e2b process stream ended without an exit status")
	}
	if result.ExitCode != 0 {
		return result, fmt.Errorf("e2b process exited with status %d", result.ExitCode)
	}
	return result, nil
}

type e2bProcessEvent struct {
	Event struct {
		Data *struct {
			Stdout string `json:"stdout"`
			Stderr string `json:"stderr"`
		} `json:"data"`
		End *struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"end"`
	} `json:"event"`
}

func (e *e2bExecutor) deleteSandbox(ctx context.Context, sandboxID string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, e.apiBaseURL+"/sandboxes/"+url.PathEscape(sandboxID), nil)
	if err != nil {
		return fmt.Errorf("create e2b delete request: %w", err)
	}
	request.Header.Set("X-API-Key", e.apiKey)
	response, err := e.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("delete e2b sandbox: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return e2bResponseError("delete e2b sandbox", response)
	}
	return nil
}

func (e *e2bExecutor) sandboxHeaders(request *http.Request, sandbox e2bSandbox) {
	request.Header.Set("X-Access-Token", sandbox.AccessToken)
	request.Header.Set("E2b-Sandbox-Id", sandbox.ID)
	request.Header.Set("E2b-Sandbox-Port", e2bSandboxPort)
}

func validBaseURL(value, fallback string) (string, error) {
	if value == "" {
		value = fallback
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("must be an absolute HTTP URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("must use http or https")
	}
	return strings.TrimRight(value, "/"), nil
}

func environmentMap(entries []string) (map[string]string, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	envs := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("sandbox env entry %q must be KEY=VALUE", entry)
		}
		envs[key] = value
	}
	return envs, nil
}

func e2bEventBytes(value string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err == nil && base64.StdEncoding.EncodeToString(decoded) == value {
		return decoded
	}
	return []byte(value)
}

func parseE2BExitCode(status string) int {
	matches := e2bExitStatus.FindStringSubmatch(status)
	if len(matches) != 2 {
		return -1
	}
	code, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1
	}
	return code
}

func e2bResponseError(operation string, response *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(response.Body, 8<<10))
	message := strings.TrimSpace(string(body))
	if message == "" {
		message = response.Status
	}
	return fmt.Errorf("%s: e2b returned HTTP %d: %s", operation, response.StatusCode, message)
}

func connectEnvelope(payload []byte) []byte {
	message := make([]byte, connectEnvelopeHeaderSize+len(payload))
	binary.BigEndian.PutUint32(message[1:connectEnvelopeHeaderSize], uint32(len(payload)))
	copy(message[connectEnvelopeHeaderSize:], payload)
	return message
}

func readConnectEnvelope(reader io.Reader) ([]byte, bool, error) {
	var header [connectEnvelopeHeaderSize]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, false, err
	}
	if header[0]&connectFlagCompressed != 0 {
		return nil, false, fmt.Errorf("compressed Connect envelopes are not supported")
	}
	size := binary.BigEndian.Uint32(header[1:])
	if size > maxConnectMessageBytes {
		return nil, false, fmt.Errorf("Connect envelope exceeds maximum size of %d bytes", maxConnectMessageBytes)
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, false, err
	}
	return payload, header[0]&connectFlagEndStream != 0, nil
}
