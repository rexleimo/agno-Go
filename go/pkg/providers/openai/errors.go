package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	// ErrUnauthorized indicates an auth failure.
	ErrUnauthorized = fmt.Errorf("openai unauthorized")
	// ErrRateLimited indicates rate limiting.
	ErrRateLimited = fmt.Errorf("openai rate limited")
	// ErrServerError covers 5xx responses.
	ErrServerError = fmt.Errorf("openai server error")
)

type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func mapHTTPError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", ErrUnauthorized, string(body))
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", ErrRateLimited, string(body))
	default:
		var apiErr apiError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
			return fmt.Errorf("openai error: %s (%v)", apiErr.Error.Message, apiErr.Error.Code)
		}
		if resp.StatusCode >= 500 {
			return fmt.Errorf("%w: %s", ErrServerError, string(body))
		}
		return fmt.Errorf("openai error: %s", resp.Status)
	}
}
