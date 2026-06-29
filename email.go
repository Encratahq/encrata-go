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

type LookupOption func(*lookupConfig)

func WithFields(fields ...string) LookupOption {
	return func(cfg *lookupConfig) { cfg.fields = fields }
}

func WithNoCache() LookupOption {
	return func(cfg *lookupConfig) { cfg.nocache = true }
}

func WithConcurrency(n int) LookupOption {
	return func(cfg *lookupConfig) { cfg.concurrency = n }
}

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

func (c *Client) Validate(ctx context.Context, email string) (*Validation, error) {
	var v Validation
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/validate", nil, map[string]string{"email": email}, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) Breaches(ctx context.Context, email string) (*BreachReport, error) {
	var r BreachReport
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/breaches", nil, map[string]string{"email": email}, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

type LookupResult struct {
	Email  string
	Person *Person
	Err    error
}

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
