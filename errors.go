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
	StatusCode int    // HTTP status code, e.g. 401
	Code       string // machine-readable error code from the API body, if any
	Message    string // human-readable message
}

func (e *apiBase) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("encrata: HTTP %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("encrata: HTTP %d: %s", e.StatusCode, e.Message)
}

func (e *apiBase) isEncrataError() {} // satisfies EncrataError for every embedder

// AuthenticationError => HTTP 401 (bad or missing API key).
type AuthenticationError struct{ apiBase }

// InsufficientCreditsError => HTTP 402 (out of credits).
type InsufficientCreditsError struct{ apiBase }

// InvalidRequestError => HTTP 400 (bad parameters).
type InvalidRequestError struct{ apiBase }

// RateLimitError => HTTP 429. Carries how long the server asked us to wait.
type RateLimitError struct {
	apiBase
	RetryAfter time.Duration // 0 if the server didn't say
}

// APIError => any other 4xx/5xx not covered above.
type APIError struct{ apiBase }

// APIConnectionError => network failure or timeout (no HTTP response at all).
type APIConnectionError struct {
	Message string
	Err     error // the underlying net/url error, for errors.Is/As unwrapping
}

func (e *APIConnectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("encrata: connection error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("encrata: connection error: %s", e.Message)
}
func (e *APIConnectionError) Unwrap() error   { return e.Err } // lets errors.Is reach the cause
func (e *APIConnectionError) isEncrataError() {}

// Compile-time checks: fail to build if any type stops satisfying EncrataError.
var (
	_ EncrataError = (*AuthenticationError)(nil)
	_ EncrataError = (*InsufficientCreditsError)(nil)
	_ EncrataError = (*InvalidRequestError)(nil)
	_ EncrataError = (*RateLimitError)(nil)
	_ EncrataError = (*APIError)(nil)
	_ EncrataError = (*APIConnectionError)(nil)
)
