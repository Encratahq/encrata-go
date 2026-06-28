package encrata

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/sync/errgroup"
)

type lookupConfig struct {
	fields      []string
	nocache     bool
	concurrency int
}

// LookupOption customizes a lookup. The same options apply to Lookup and
// LookupMany.
type LookupOption func(*lookupConfig)

// WithFields restricts the lookup to the named enrichment fields.
func WithFields(fields ...string) LookupOption {
	return func(cfg *lookupConfig) { cfg.fields = fields }
}

// WithNoCache forces a fresh lookup, bypassing any cached result.
func WithNoCache() LookupOption {
	return func(cfg *lookupConfig) { cfg.nocache = true }
}

// WithConcurrency caps the number of in-flight lookups in LookupMany. It has no
// effect on a single Lookup. The default is 10.
func WithConcurrency(n int) LookupOption {
	return func(cfg *lookupConfig) { cfg.concurrency = n }
}

// Lookup enriches a single email address into a Person.
func (c *Client) Lookup(ctx context.Context, email string, opts ...LookupOption) (*Person, error) {
	var cfg lookupConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	query := url.Values{}
	if len(cfg.fields) > 0 {
		query.Set("fields", strings.Join(cfg.fields, ","))
	}
	if cfg.nocache {
		query.Set("nocache", "1")
	}

	var person Person
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/lookup", query, map[string]string{"email": email}, &person); err != nil {
		return nil, err
	}
	return &person, nil
}

// Validate checks whether an email is deliverable. It does not consume credits.
func (c *Client) Validate(ctx context.Context, email string) (*Validation, error) {
	var v Validation
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/validate", nil, map[string]string{"email": email}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// Breaches reports data-breach exposure for an email. It does not consume credits.
func (c *Client) Breaches(ctx context.Context, email string) (*BreachReport, error) {
	var r BreachReport
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/breaches", nil, map[string]string{"email": email}, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// LookupResult pairs an input email with its lookup outcome. Exactly one of
// Person or Err is set.
type LookupResult struct {
	Email  string
	Person *Person
	Err    error
}

// LookupMany enriches many emails concurrently, preserving input order. A
// failure on one email is reported in that element's Err field rather than
// aborting the batch, so you keep the results you already paid for. The
// returned error is non-nil only when ctx is cancelled. Concurrency defaults to
// 10 (override with WithConcurrency).
func (c *Client) LookupMany(ctx context.Context, emails []string, opts ...LookupOption) ([]LookupResult, error) {
	results := make([]LookupResult, len(emails))
	if len(emails) == 0 {
		return results, nil
	}

	var cfg lookupConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	limit := cfg.concurrency
	if limit <= 0 {
		limit = 10
	}
	if limit > len(emails) {
		limit = len(emails)
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(limit)

	for i, email := range emails {
		results[i].Email = email
		g.Go(func() error {
			person, err := c.Lookup(gctx, email, opts...)
			results[i].Person = person
			results[i].Err = err
			return nil
		})
	}
	_ = g.Wait()

	return results, ctx.Err()
}
