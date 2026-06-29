package encrata

import (
	"context"
	"net/http"
	"net/url"
)

// ListKeys lists all API keys for the authenticated account.
func (c *Client) ListKeys(ctx context.Context) ([]APIKey, error) {
	var resp struct {
		Keys []APIKey `json:"keys"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/keys", nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Keys, nil
}

// CreateKey creates a new API key. The full key is only returned once.
func (c *Client) CreateKey(ctx context.Context, name string) (*APIKey, error) {
	var key APIKey
	if err := c.doRequest(ctx, http.MethodPost, "/api/keys", nil, map[string]string{"name": name}, &key); err != nil {
		return nil, err
	}
	return &key, nil
}

// RevokeKey revokes an API key. Set permanent to true to permanently delete it.
func (c *Client) RevokeKey(ctx context.Context, keyID string, permanent bool) (RawObject, error) {
	q := url.Values{}
	q.Set("id", keyID)
	if permanent {
		q.Set("permanent", "true")
	}
	var resp RawObject
	if err := c.doRequest(ctx, http.MethodDelete, "/api/keys", q, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
