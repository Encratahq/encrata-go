package encrata

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type WorkflowRequest struct {
	Name        string
	Description string
	Status      string
	Trigger     RawObject
	Steps       []RawObject
	TemplateID  string
}

type WorkflowListOptions struct {
	Page   int
	Limit  int
	Status string
}

type WorkflowRunListOptions struct {
	Page       int
	Limit      int
	WorkflowID string
}

func workflowBody(req WorkflowRequest, includeStatus bool) map[string]any {
	body := map[string]any{}
	if req.Name != "" {
		body["name"] = req.Name
	}
	if req.Description != "" {
		body["description"] = req.Description
	}
	if includeStatus && req.Status != "" {
		body["status"] = req.Status
	}
	if req.Trigger != nil {
		body["trigger"] = req.Trigger
	}
	if req.Steps != nil {
		body["steps"] = req.Steps
	}
	if req.TemplateID != "" {
		body["template_id"] = req.TemplateID
	}
	return body
}

func (c *Client) ListWorkflows(ctx context.Context, opts WorkflowListOptions) ([]Workflow, int, error) {
	q := url.Values{}
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	if opts.Status != "" {
		q.Set("status", opts.Status)
	}
	var resp struct {
		Workflows []Workflow `json:"workflows"`
		Total     int        `json:"total"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/workflows", q, nil, &resp); err != nil {
		return nil, 0, err
	}
	if resp.Total == 0 {
		resp.Total = len(resp.Workflows)
	}
	return resp.Workflows, resp.Total, nil
}

func (c *Client) CreateWorkflow(ctx context.Context, req WorkflowRequest) (*Workflow, error) {
	var workflow Workflow
	if err := c.doRequest(ctx, http.MethodPost, "/api/workflows", nil, workflowBody(req, false), &workflow); err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (c *Client) GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {
	var workflow Workflow
	if err := c.doRequest(ctx, http.MethodGet, "/api/workflows/"+url.PathEscape(workflowID), nil, nil, &workflow); err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (c *Client) UpdateWorkflow(ctx context.Context, workflowID string, req WorkflowRequest) (*Workflow, error) {
	var workflow Workflow
	if err := c.doRequest(ctx, http.MethodPut, "/api/workflows/"+url.PathEscape(workflowID), nil, workflowBody(req, true), &workflow); err != nil {
		return nil, err
	}
	if workflow.ID == "" {
		workflow.ID = workflowID
	}
	return &workflow, nil
}

func (c *Client) ListWorkflowRuns(ctx context.Context, opts WorkflowRunListOptions) ([]WorkflowRun, int, error) {
	q := url.Values{}
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	if opts.WorkflowID != "" {
		q.Set("workflow_id", opts.WorkflowID)
	}
	var resp struct {
		Runs  []WorkflowRun `json:"runs"`
		Total int           `json:"total"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/workflows/runs", q, nil, &resp); err != nil {
		return nil, 0, err
	}
	if resp.Total == 0 {
		resp.Total = len(resp.Runs)
	}
	return resp.Runs, resp.Total, nil
}

func (c *Client) GetWorkflowRun(ctx context.Context, runID string) (*WorkflowRun, error) {
	var run WorkflowRun
	if err := c.doRequest(ctx, http.MethodGet, "/api/workflows/runs/"+url.PathEscape(runID), nil, nil, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (c *Client) ListWorkflowTemplates(ctx context.Context, category string) ([]WorkflowTemplate, error) {
	var q url.Values
	if category != "" {
		q = url.Values{}
		q.Set("category", category)
	}
	var resp struct {
		Templates []WorkflowTemplate `json:"templates"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/workflows/templates", q, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Templates, nil
}

func (c *Client) ListWorkflowSecrets(ctx context.Context) ([]WorkflowSecret, error) {
	var resp struct {
		Secrets []WorkflowSecret `json:"secrets"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/workflows/secrets", nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Secrets, nil
}

func (c *Client) CreateWorkflowSecret(ctx context.Context, name, value string) (RawObject, error) {
	var resp RawObject
	if err := c.doRequest(ctx, http.MethodPost, "/api/workflows/secrets", nil, map[string]string{"name": name, "value": value}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteWorkflowSecret(ctx context.Context, name string) (RawObject, error) {
	var resp RawObject
	if err := c.doRequest(ctx, http.MethodDelete, "/api/workflows/secrets", nil, map[string]string{"name": name}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
