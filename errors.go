package encrata

import (
	"fmt"
	"time"
)

type EncrataError interface {
	error
	isEncrataError()
}

type apiBase struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *apiBase) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("encrata: HTTP %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("encrata: HTTP %d: %s", e.StatusCode, e.Message)
}

func (e *apiBase) isEncrataError() {}

type AuthenticationError struct{ apiBase }

type InsufficientCreditsError struct{ apiBase }

type InvalidRequestError struct{ apiBase }

type RateLimitError struct {
	apiBase
	RetryAfter time.Duration
}

type APIError struct{ apiBase }

type APIConnectionError struct {
	Message string
	Err     error
}

func (e *APIConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("encrata: connection error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("encrata: connection error: %s", e.Message)
}
func (e *APIConnectionError) Unwrap() error   { return e.Err }
func (e *APIConnectionError) isEncrataError() {}

var (
	_ EncrataError = (*AuthenticationError)(nil)
	_ EncrataError = (*InsufficientCreditsError)(nil)
	_ EncrataError = (*InvalidRequestError)(nil)
	_ EncrataError = (*RateLimitError)(nil)
	_ EncrataError = (*APIError)(nil)
	_ EncrataError = (*APIConnectionError)(nil)
)
