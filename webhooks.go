package encrata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type WebhookUpdateRequest struct {
	ID          string
	URL         string
	Events      []string
	Description string
	IsActive    *bool
}

func (c *Client) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	var raw json.RawMessage
	if err := c.doRequest(ctx, http.MethodGet, "/api/webhooks", nil, nil, &raw); err != nil {
		return nil, err
	}
	var list []Webhook
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var wrapper struct {
		Webhooks []Webhook `json:"webhooks"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, &APIError{apiBase{Message: "decoding webhooks: " + err.Error()}}
	}
	return wrapper.Webhooks, nil
}

func (c *Client) CreateWebhook(ctx context.Context, endpoint string, events []string, description string) (*Webhook, error) {
	body := map[string]any{"url": endpoint, "events": events}
	if description != "" {
		body["description"] = description
	}
	var webhook Webhook
	if err := c.doRequest(ctx, http.MethodPost, "/api/webhooks", nil, body, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (c *Client) UpdateWebhook(ctx context.Context, req WebhookUpdateRequest) (RawObject, error) {
	body := map[string]any{"id": req.ID, "url": req.URL}
	if req.Events != nil {
		body["events"] = req.Events
	}
	if req.Description != "" {
		body["description"] = req.Description
	}
	if req.IsActive != nil {
		body["is_active"] = *req.IsActive
	}
	var resp RawObject
	if err := c.doRequest(ctx, http.MethodPut, "/api/webhooks", nil, body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, webhookID string) (RawObject, error) {
	var resp RawObject
	if err := c.doRequest(ctx, http.MethodDelete, "/api/webhooks", nil, map[string]string{"id": webhookID}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) TestWebhook(ctx context.Context, webhookID string) (RawObject, error) {
	var resp RawObject
	if err := c.doRequest(ctx, http.MethodPost, "/api/webhooks/test", nil, map[string]string{"id": webhookID}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) ListWebhookDeliveries(ctx context.Context, webhookID string) ([]WebhookDelivery, error) {
	q := url.Values{}
	q.Set("webhook_id", webhookID)
	var raw json.RawMessage
	if err := c.doRequest(ctx, http.MethodGet, "/api/webhooks/deliveries", q, nil, &raw); err != nil {
		return nil, err
	}
	var list []WebhookDelivery
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var wrapper struct {
		Deliveries []WebhookDelivery `json:"deliveries"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, &APIError{apiBase{Message: "decoding webhook deliveries: " + err.Error()}}
	}
	return wrapper.Deliveries, nil
}
