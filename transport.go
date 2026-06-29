package encrata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// doRequest performs an HTTP request against the Encrata API. Transient
// failures (connection errors and the retryable status codes) are retried with
// full-jitter exponential backoff, honoring a Retry-After header when present.
// On a 2xx response the body is decoded into out when out is non-nil.
func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body, out any) error {
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var payload []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return &InvalidRequestError{apiBase{Message: fmt.Sprintf("encoding request body: %v", err)}}
		}
		payload = b
	}

	for attempt := 0; ; attempt++ {
		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
		if err != nil {
			return &APIConnectionError{Message: "building request", Err: err}
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent)
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return &APIConnectionError{Message: "request cancelled", Err: ctx.Err()}
			}
			if attempt < c.maxRetries {
				if werr := c.waitBackoff(ctx, attempt, 0); werr != nil {
					return &APIConnectionError{Message: "request cancelled", Err: werr}
				}
				continue
			}
			return &APIConnectionError{Message: "request failed", Err: err}
		}

		data, status, retryAfter, rerr := drainResponse(resp)
		if rerr != nil {
			if attempt < c.maxRetries {
				if werr := c.waitBackoff(ctx, attempt, 0); werr != nil {
					return &APIConnectionError{Message: "request cancelled", Err: werr}
				}
				continue
			}
			return &APIConnectionError{Message: "reading response body", Err: rerr}
		}

		if retryableStatus[status] && attempt < c.maxRetries {
			if werr := c.waitBackoff(ctx, attempt, retryAfter); werr != nil {
				return &APIConnectionError{Message: "request cancelled", Err: werr}
			}
			continue
		}

		if status >= 400 {
			return newAPIError(status, data, retryAfter)
		}

		if out != nil && len(data) > 0 {
			if err := json.Unmarshal(data, out); err != nil {
				return &APIError{apiBase{StatusCode: status, Message: fmt.Sprintf("decoding response: %v", err)}}
			}
		}
		return nil
	}
}

// drainResponse reads and closes the response body, returning the bytes, the
// status code, and any Retry-After hint.
func drainResponse(resp *http.Response) (body []byte, status int, retryAfter time.Duration, err error) {
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, 0, err
	}
	ra, _ := parseRetryAfter(resp.Header.Get("Retry-After"))
	return b, resp.StatusCode, ra, nil
}

// waitBackoff sleeps before the next retry. A Retry-After hint takes priority
// (capped at maxBackoff); otherwise it uses full-jitter exponential backoff.
// It returns ctx.Err() if the context is cancelled while waiting.
func (c *Client) waitBackoff(ctx context.Context, attempt int, retryAfter time.Duration) error {
	delay := retryAfter
	if delay > maxBackoff {
		delay = maxBackoff
	}
	if delay <= 0 {
		delay = backoffDelay(attempt)
	}
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// backoffDelay returns a random delay in [0, min(initial*factor^attempt, max)]
// - the AWS "full jitter" strategy, so many clients retrying at once spread
// their load instead of spiking together.
func backoffDelay(attempt int) time.Duration {
	ceiling := float64(initialBackoff) * math.Pow(backoffFactor, float64(attempt))
	if ceiling > float64(maxBackoff) {
		ceiling = float64(maxBackoff)
	}
	return time.Duration(rand.Float64() * ceiling)
}

// parseRetryAfter understands both forms allowed by the HTTP spec: a number of
// seconds ("5" or "1.5") or an HTTP-date. The returned duration is never
// negative; ok is false when the header is missing or unparseable.
func parseRetryAfter(value string) (delay time.Duration, ok bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if secs, err := strconv.ParseFloat(value, 64); err == nil {
		if secs < 0 {
			secs = 0
		}
		return time.Duration(secs * float64(time.Second)), true
	}
	if when, err := http.ParseTime(value); err == nil {
		d := time.Until(when)
		if d < 0 {
			d = 0
		}
		return d, true
	}
	return 0, false
}

// newAPIError maps an HTTP status code and error body to the matching typed
// error.
func newAPIError(status int, body []byte, retryAfter time.Duration) error {
	var parsed struct {
		Message string `json:"message"`
		Error   string `json:"error"`
		Code    string `json:"code"`
	}
	_ = json.Unmarshal(body, &parsed)

	msg := parsed.Message
	if msg == "" {
		msg = parsed.Error
	}
	if msg == "" {
		msg = http.StatusText(status)
	}
	base := apiBase{StatusCode: status, Code: parsed.Code, Message: msg}

	switch status {
	case http.StatusUnauthorized:
		return &AuthenticationError{base}
	case http.StatusPaymentRequired:
		return &InsufficientCreditsError{base}
	case http.StatusBadRequest:
		return &InvalidRequestError{base}
	case http.StatusTooManyRequests:
		return &RateLimitError{apiBase: base, RetryAfter: retryAfter}
	default:
		return &APIError{base}
	}
}
