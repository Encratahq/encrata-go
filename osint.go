package encrata

import (
	"context"
	"net/http"
)

// IP looks up geolocation, ASN, company, and threat data for an IP address.
func (c *Client) IP(ctx context.Context, ip string) (*IPInfo, error) {
	var out IPInfo
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/ip", nil, map[string]string{"ip": ip}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PhoneLookup looks up carrier, format, country, validation, risk, and breach data.
func (c *Client) PhoneLookup(ctx context.Context, query string) (*PhoneInfo, error) {
	var out PhoneInfo
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/phone", nil, map[string]string{"query": query}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DomainSearch looks up WHOIS, DNS, SSL, threat intel, and recon data.
func (c *Client) DomainSearch(ctx context.Context, query string) (*DomainInfo, error) {
	var out DomainInfo
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/domain", nil, map[string]string{"query": query}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CompanySearch finds people and a unified company profile by company name.
func (c *Client) CompanySearch(ctx context.Context, query string) (*CompanyInfo, error) {
	var out CompanyInfo
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/company", nil, map[string]string{"query": query}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GoogleSearch runs a Google dork query with OSINT enrichment.
func (c *Client) GoogleSearch(ctx context.Context, query string) (*GoogleSearch, error) {
	var out GoogleSearch
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/google", nil, map[string]string{"query": query}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DarkWebSearch searches dark web intelligence for leaks, forums, markets, and chat mentions.
func (c *Client) DarkWebSearch(ctx context.Context, query string, offset int) (*DarkWebSearch, error) {
	var out DarkWebSearch
	body := map[string]any{"query": query, "offset": offset}
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/darkweb", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
