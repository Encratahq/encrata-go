package encrata

import (
	"context"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

type monitorConfig struct {
	emails          []string
	frequency       string
	changeDetection string
	listID          string
}

type CreateMonitorRequest struct {
	Name            string
	Emails          []string
	Frequency       string
	ChangeDetection string
	ListID          string
}

var validFrequencies = map[string]bool{
	"weekly":    true,
	"biweekly":  true,
	"monthly":   true,
	"quarterly": true,
}

var validChangeDetection = map[string]bool{
	"diff_only":    true,
	"full_refresh": true,
}

// MonitorOption customizes CreateMonitor.
type MonitorOption func(*monitorConfig)

// WithMonitorEmails sets the emails a monitor tracks.
func WithMonitorEmails(emails ...string) MonitorOption {
	return func(cfg *monitorConfig) { cfg.emails = emails }
}

// WithFrequency sets the run cadence: "weekly", "biweekly", "monthly", or
// "quarterly". The default is "monthly".
func WithFrequency(frequency string) MonitorOption {
	return func(cfg *monitorConfig) { cfg.frequency = frequency }
}

// WithChangeDetection sets the detection mode: "diff_only" or "full_refresh".
// The default is "diff_only".
func WithChangeDetection(mode string) MonitorOption {
	return func(cfg *monitorConfig) { cfg.changeDetection = mode }
}

// WithListID uses a contact list as the monitor's data source.
func WithListID(listID string) MonitorOption {
	return func(cfg *monitorConfig) { cfg.listID = listID }
}

// ListMonitors returns all monitors.
func (c *Client) ListMonitors(ctx context.Context) ([]Monitor, error) {
	var resp struct {
		Monitors []Monitor `json:"monitors"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/monitors", nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Monitors, nil
}

// CreateMonitor creates a new monitor.
func (c *Client) CreateMonitor(ctx context.Context, req CreateMonitorRequest, opts ...MonitorOption) (*Monitor, error) {
	cfg := monitorConfig{frequency: "monthly", changeDetection: "diff_only"}
	for _, opt := range opts {
		opt(&cfg)
	}

	if !validFrequencies[cfg.frequency] {
		return nil, &InvalidRequestError{apiBase{Message: "invalid frequency: " + cfg.frequency + " (want weekly, biweekly, monthly, or quarterly)"}}
	}
	if !validChangeDetection[cfg.changeDetection] {
		return nil, &InvalidRequestError{apiBase{Message: "invalid change detection: " + cfg.changeDetection + " (want diff_only or full_refresh)"}}
	}

	body := map[string]any{
		"name":             req.Name,
		"frequency":        cfg.frequency,
		"change_detection": cfg.changeDetection,
	}
	if cfg.listID != "" {
		body["data_source_type"] = "list"
		body["data_source_ref"] = cfg.listID
	}
	if len(cfg.emails) > 0 {
		body["emails"] = cfg.emails
	}

	var m Monitor
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/monitors", nil, body, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// GetMonitor returns a monitor by ID.
func (c *Client) GetMonitor(ctx context.Context, monitorID string) (*Monitor, error) {
	var m Monitor
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/monitors/"+url.PathEscape(monitorID), nil, nil, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// TriggerRun starts an immediate monitoring run.
func (c *Client) TriggerRun(ctx context.Context, monitorID string) (*RunTrigger, error) {
	var t RunTrigger
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/monitors/"+url.PathEscape(monitorID)+"/run", nil, map[string]any{}, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func pageQuery(limit, offset int) url.Values {
	q := url.Values{}
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	return q
}

// ListRuns returns one page of runs for a monitor along with the total count.
func (c *Client) ListRuns(ctx context.Context, monitorID string, limit, offset int) ([]MonitorRun, int, error) {
	var resp struct {
		Runs  []MonitorRun `json:"runs"`
		Total int          `json:"total"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/monitors/"+url.PathEscape(monitorID)+"/runs", pageQuery(limit, offset), nil, &resp); err != nil {
		return nil, 0, err
	}
	total := resp.Total
	if total == 0 {
		total = len(resp.Runs)
	}
	return resp.Runs, total, nil
}

// IterRuns iterates over every run of a monitor, fetching pages on demand:
//
//	for run, err := range client.IterRuns(ctx, monitorID) {
//	    if err != nil {
//	        return err
//	    }
//	    // use run
//	}
func (c *Client) IterRuns(ctx context.Context, monitorID string) iter.Seq2[MonitorRun, error] {
	return func(yield func(MonitorRun, error) bool) {
		offset := 0
		for {
			runs, total, err := c.ListRuns(ctx, monitorID, 100, offset)
			if err != nil {
				yield(MonitorRun{}, err)
				return
			}
			if len(runs) == 0 {
				return
			}
			for _, run := range runs {
				if !yield(run, nil) {
					return
				}
			}
			offset += len(runs)
			if offset >= total {
				return
			}
		}
	}
}

// GetRunResults returns one page of result snapshots for a run.
func (c *Client) GetRunResults(ctx context.Context, monitorID, runID string, changesOnly bool, limit, offset int) ([]MonitorSnapshot, int, error) {
	q := pageQuery(limit, offset)
	if changesOnly {
		q.Set("changes_only", "true")
	}
	var resp struct {
		Results []MonitorSnapshot `json:"results"`
		Total   int               `json:"total"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/monitors/"+url.PathEscape(monitorID)+"/runs/"+url.PathEscape(runID)+"/results", q, nil, &resp); err != nil {
		return nil, 0, err
	}
	total := resp.Total
	if total == 0 {
		total = len(resp.Results)
	}
	return resp.Results, total, nil
}

// IterRunResults iterates over every result snapshot of a run.
func (c *Client) IterRunResults(ctx context.Context, monitorID, runID string, changesOnly bool) iter.Seq2[MonitorSnapshot, error] {
	return func(yield func(MonitorSnapshot, error) bool) {
		offset := 0
		for {
			snaps, total, err := c.GetRunResults(ctx, monitorID, runID, changesOnly, 100, offset)
			if err != nil {
				yield(MonitorSnapshot{}, err)
				return
			}
			if len(snaps) == 0 {
				return
			}
			for _, s := range snaps {
				if !yield(s, nil) {
					return
				}
			}
			offset += len(snaps)
			if offset >= total {
				return
			}
		}
	}
}

// ListAllRuns returns one page of runs across all monitors.
func (c *Client) ListAllRuns(ctx context.Context, limit, offset int) ([]MonitorRun, int, error) {
	var resp struct {
		Runs  []MonitorRun `json:"runs"`
		Total int          `json:"total"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/monitoring/runs", pageQuery(limit, offset), nil, &resp); err != nil {
		return nil, 0, err
	}
	total := resp.Total
	if total == 0 {
		total = len(resp.Runs)
	}
	return resp.Runs, total, nil
}

// IterAllRuns iterates over every run across all monitors.
func (c *Client) IterAllRuns(ctx context.Context) iter.Seq2[MonitorRun, error] {
	return func(yield func(MonitorRun, error) bool) {
		offset := 0
		for {
			runs, total, err := c.ListAllRuns(ctx, 100, offset)
			if err != nil {
				yield(MonitorRun{}, err)
				return
			}
			if len(runs) == 0 {
				return
			}
			for _, run := range runs {
				if !yield(run, nil) {
					return
				}
			}
			offset += len(runs)
			if offset >= total {
				return
			}
		}
	}
}

// ListAllResults returns one page of enrichment results across all monitors.
func (c *Client) ListAllResults(ctx context.Context, changesOnly bool, limit, offset int) ([]MonitorSnapshot, int, error) {
	q := pageQuery(limit, offset)
	if changesOnly {
		q.Set("changes_only", "true")
	}
	var resp struct {
		Results []MonitorSnapshot `json:"results"`
		Total   int               `json:"total"`
	}
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/monitoring/results", q, nil, &resp); err != nil {
		return nil, 0, err
	}
	total := resp.Total
	if total == 0 {
		total = len(resp.Results)
	}
	return resp.Results, total, nil
}

// IterAllResults iterates over every enrichment result across all monitors.
func (c *Client) IterAllResults(ctx context.Context, changesOnly bool) iter.Seq2[MonitorSnapshot, error] {
	return func(yield func(MonitorSnapshot, error) bool) {
		offset := 0
		for {
			snaps, total, err := c.ListAllResults(ctx, changesOnly, 100, offset)
			if err != nil {
				yield(MonitorSnapshot{}, err)
				return
			}
			if len(snaps) == 0 {
				return
			}
			for _, s := range snaps {
				if !yield(s, nil) {
					return
				}
			}
			offset += len(snaps)
			if offset >= total {
				return
			}
		}
	}
}
