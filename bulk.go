package encrata

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func (c *Client) BulkLookup(ctx context.Context, emails []string, fields ...string) (<-chan LookupResult, <-chan error) {
	results := make(chan LookupResult)
	errs := make(chan error, 1)

	go func() {
		defer close(results)
		defer close(errs)

		body := map[string]any{"emails": emails}
		query := ""
		if len(fields) > 0 {
			query = "?fields=" + strings.Join(fields, ",")
		}
		data, err := c.streamRequest(ctx, "/api/agent/bulk-lookup"+query, body)
		if err != nil {
			errs <- err
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			payload := parseSSEData(scanner.Text())
			if payload == "" {
				continue
			}
			if payload == "[DONE]" {
				return
			}
			var person Person
			if err := json.Unmarshal([]byte(payload), &person); err != nil {
				errs <- &APIError{apiBase{Message: "decoding bulk lookup result: " + err.Error()}}
				return
			}
			results <- LookupResult{Email: person.Email, Person: &person}
		}
		if err := scanner.Err(); err != nil {
			errs <- &APIConnectionError{Message: "reading bulk lookup stream", Err: err}
		}
	}()

	return results, errs
}

func (c *Client) BulkGoogleSearch(ctx context.Context, queries []string) (*BulkSearchResponse, error) {
	return c.bulkSearch(ctx, "/api/bulk-google-search", queries)
}

func (c *Client) BulkCompanySearch(ctx context.Context, queries []string) (*BulkSearchResponse, error) {
	return c.bulkSearch(ctx, "/api/bulk-company-search", queries)
}

func (c *Client) BulkDomainSearch(ctx context.Context, queries []string) (*BulkSearchResponse, error) {
	return c.bulkSearch(ctx, "/api/bulk-domain-search", queries)
}

func (c *Client) BulkIPSearch(ctx context.Context, queries []string) (*BulkSearchResponse, error) {
	return c.bulkSearch(ctx, "/api/bulk-ip-search", queries)
}

func (c *Client) bulkSearch(ctx context.Context, path string, queries []string) (*BulkSearchResponse, error) {
	data, err := c.streamRequest(ctx, path, map[string]any{"queries": queries})
	if err != nil {
		return nil, err
	}
	return parseBulkSearchStream(data)
}

func (c *Client) streamRequest(ctx context.Context, path string, body any) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, &InvalidRequestError{apiBase{Message: "encoding request body: " + err.Error()}}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, &APIConnectionError{Message: "building request", Err: err}
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream, application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &APIConnectionError{Message: "stream request failed", Err: err}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIConnectionError{Message: "reading stream response", Err: err}
	}
	if resp.StatusCode >= 400 {
		ra, _ := parseRetryAfter(resp.Header.Get("Retry-After"))
		return nil, newAPIError(resp.StatusCode, data, ra)
	}
	return data, nil
}

func parseBulkSearchStream(data []byte) (*BulkSearchResponse, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return &BulkSearchResponse{}, nil
	}
	if !bytes.Contains(data, []byte("data:")) {
		var resp BulkSearchResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, &APIError{apiBase{Message: "decoding bulk search response: " + err.Error()}}
		}
		return &resp, nil
	}

	resp := &BulkSearchResponse{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		payload := parseSSEData(scanner.Text())
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			break
		}
		var item RawObject
		if err := json.Unmarshal([]byte(payload), &item); err != nil {
			return nil, &APIError{apiBase{Message: "decoding bulk search result: " + err.Error()}}
		}
		if rawResults, ok := item["results"].([]any); ok {
			for _, raw := range rawResults {
				if obj, ok := raw.(map[string]any); ok {
					resp.Results = append(resp.Results, RawObject(obj))
				}
			}
		} else if _, onlyTotal := item["total"]; !onlyTotal || len(item) > 1 {
			resp.Results = append(resp.Results, item)
		}
		if credits, ok := item["credits_used"].(float64); ok {
			resp.CreditsUsed = int(credits)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, &APIConnectionError{Message: "reading bulk search stream", Err: err}
	}
	if resp.CreditsUsed == 0 {
		resp.CreditsUsed = len(resp.Results)
	}
	return resp, nil
}

func parseSSEData(line string) string {
	if !strings.HasPrefix(line, "data:") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(line, "data:"))
}
