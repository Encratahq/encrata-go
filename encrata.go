package encrata

import (
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

type Option func(*Client)

func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(url, "/") }
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the per-request timeout. Default is 30s.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithMaxRetries sets how many times a failed request is retried. Default is 3.
func WithMaxRetries(n int) Option {
	return func(c *Client) { c.maxRetries = n }
}

func New(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, &AuthenticationError{apiBase{
			StatusCode: 0,
			Message:    "API key is required",
		}}
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
		maxRetries: defaultMaxRetries,
	}

	// Apply caller options on top of the defaults.
	for _, opt := range opts {
		opt(c)
	}

	if c.baseURL == "" {
		return nil, &InvalidRequestError{apiBase{Message: "base URL cannot be empty"}}
	}
	if c.maxRetries < 0 {
		return nil, &InvalidRequestError{apiBase{Message: "max retries cannot be negative"}}
	}
	if c.httpClient == nil {
		return nil, &InvalidRequestError{apiBase{Message: "HTTP client cannot be nil"}}
	}
	if c.httpClient.Timeout < 0 {
		return nil, &InvalidRequestError{apiBase{Message: "timeout cannot be negative"}}
	}
	return c, nil
}
