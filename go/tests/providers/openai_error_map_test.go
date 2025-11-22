package providers

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/rexleimo/agno-go/pkg/providers/openai"
)

func TestOpenAIErrorMapping(t *testing.T) {
	cases := []struct {
		code int
		body string
		want string
	}{
		{http.StatusUnauthorized, `{"error":{"message":"bad key","code":"invalid_api_key"}}`, openai.ErrUnauthorized.Error()},
		{http.StatusTooManyRequests, `{"error":{"message":"slow down","code":"rate_limit"}}`, openai.ErrRateLimited.Error()},
		{http.StatusInternalServerError, `{"error":{"message":"server boom","code":"500"}}`, "server boom"},
		{http.StatusBadRequest, `{"error":{"message":"bad","code":"400"}}`, ""},
	}
	for _, tt := range cases {
		resp := &http.Response{
			StatusCode: tt.code,
			Body:       io.NopCloser(strings.NewReader(tt.body)),
		}
		err := openai.MapHTTPErrorForTest(resp)
		if tt.want == "" && err == nil {
			continue
		}
		if !strings.Contains(err.Error(), tt.want) {
			t.Fatalf("code %d: expected %v in %v", tt.code, tt.want, err)
		}
	}
}
