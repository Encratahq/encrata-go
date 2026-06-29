package encrata

import (
	"context"
	"net/http"
)

type ExtractRequest struct {
	URL           string
	Mode          string
	Selectors     map[string]string
	RenderJS      bool
	BlockAds      bool
	BlockTrackers bool
	WaitFor       string
	Timeout       int
	Headers       map[string]string
}

type ScreenshotRequest struct {
	URL           string
	FullPage      bool
	Format        string
	Selector      string
	RenderJS      bool
	BlockAds      bool
	BlockTrackers bool
	WaitFor       string
	Timeout       int
	Headers       map[string]string
}

// Scrape scrapes a URL and returns clean, LLM-ready markdown with metadata.
func (c *Client) Scrape(ctx context.Context, url string, renderJS bool) (*ScrapeResult, error) {
	var out ScrapeResult
	body := map[string]any{"url": url, "render_js": renderJS}
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/scrape", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Extract extracts structured data from a page as markdown or selector-keyed JSON.
func (c *Client) Extract(ctx context.Context, req ExtractRequest) (*ExtractResult, error) {
	var out ExtractResult
	mode := req.Mode
	if mode == "" {
		mode = "markdown"
	}
	body := map[string]any{
		"url":            req.URL,
		"mode":           mode,
		"render_js":      req.RenderJS,
		"block_ads":      req.BlockAds,
		"block_trackers": req.BlockTrackers,
	}
	if req.Selectors != nil {
		body["selectors"] = req.Selectors
	}
	if req.WaitFor != "" {
		body["wait_for"] = req.WaitFor
	}
	if req.Timeout > 0 {
		body["timeout"] = req.Timeout
	}
	if req.Headers != nil {
		body["headers"] = req.Headers
	}
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/extract", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Screenshot captures a page screenshot as base64 PNG or JPEG data.
func (c *Client) Screenshot(ctx context.Context, req ScreenshotRequest) (*ScreenshotResult, error) {
	var out ScreenshotResult
	format := req.Format
	if format == "" {
		format = "png"
	}
	body := map[string]any{
		"url":            req.URL,
		"full_page":      req.FullPage,
		"format":         format,
		"render_js":      req.RenderJS,
		"block_ads":      req.BlockAds,
		"block_trackers": req.BlockTrackers,
	}
	if req.Selector != "" {
		body["selector"] = req.Selector
	}
	if req.WaitFor != "" {
		body["wait_for"] = req.WaitFor
	}
	if req.Timeout > 0 {
		body["timeout"] = req.Timeout
	}
	if req.Headers != nil {
		body["headers"] = req.Headers
	}
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/screenshot", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
