package encrata

import (
	"context"
	"net/http"
)

// FaceSearch searches a face against your watchlist by public image URL.
func (c *Client) FaceSearch(ctx context.Context, imageURL string, threshold *float64) (*FaceSearch, error) {
	var out FaceSearch
	body := map[string]any{"image_url": imageURL}
	if threshold != nil {
		body["threshold"] = *threshold
	}
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/face", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
