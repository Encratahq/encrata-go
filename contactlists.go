package encrata

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type ContactListRequest struct {
	Name    string
	Type    string
	Targets []string
	Emails  []string
}

// ListContactLists returns all contact lists.
func (c *Client) ListContactLists(ctx context.Context) ([]ContactList, error) {
	return c.ListContactListsByType(ctx, "")
}

// ListContactListsByType returns contact lists, optionally filtered by type.
func (c *Client) ListContactListsByType(ctx context.Context, listType string) ([]ContactList, error) {
	var q url.Values
	if listType != "" {
		q = url.Values{}
		q.Set("type", listType)
	}
	var raw json.RawMessage
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/lists", q, nil, &raw); err != nil {
		return nil, err
	}

	var lists []ContactList
	if err := json.Unmarshal(raw, &lists); err == nil {
		return lists, nil
	}
	var wrapper struct {
		Lists []ContactList `json:"lists"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, &APIError{apiBase{Message: "decoding contact lists: " + err.Error()}}
	}
	return wrapper.Lists, nil
}

// CreateContactList creates a new contact list, optionally seeded with emails.
func (c *Client) CreateContactList(ctx context.Context, name string, emails ...string) (*ContactList, error) {
	return c.CreateContactListWithRequest(ctx, ContactListRequest{Name: name, Emails: emails})
}

// CreateContactListWithRequest creates a contact list for email, phone, domain,
// ip, company, or darkweb targets. Emails is a legacy alias for Targets.
func (c *Client) CreateContactListWithRequest(ctx context.Context, req ContactListRequest) (*ContactList, error) {
	body := map[string]any{"name": req.Name}
	if req.Type != "" {
		body["type"] = req.Type
	}
	if len(req.Targets) > 0 {
		body["targets"] = req.Targets
	}
	if len(req.Emails) > 0 {
		body["emails"] = req.Emails
	}
	var cl ContactList
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/lists", nil, body, &cl); err != nil {
		return nil, err
	}
	return &cl, nil
}

// GetContactList returns a contact list by ID.
func (c *Client) GetContactList(ctx context.Context, listID string) (*ContactList, error) {
	var cl ContactList
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/lists/"+url.PathEscape(listID), nil, nil, &cl); err != nil {
		return nil, err
	}
	return &cl, nil
}

// DeleteContactList deletes a contact list.
func (c *Client) DeleteContactList(ctx context.Context, listID string) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/agent/lists/"+url.PathEscape(listID), nil, nil, nil)
}

// ListContactListEmails returns every email in a contact list.
func (c *Client) ListContactListEmails(ctx context.Context, listID string) ([]string, error) {
	var raw json.RawMessage
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/lists/"+url.PathEscape(listID)+"/emails", nil, nil, &raw); err != nil {
		return nil, err
	}

	var plain []string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return plain, nil
	}
	var objects []struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(raw, &objects); err == nil {
		emails := make([]string, len(objects))
		for i, o := range objects {
			emails[i] = o.Email
		}
		return emails, nil
	}
	var wrapper struct {
		Emails []string `json:"emails"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, &APIError{apiBase{Message: "decoding emails: " + err.Error()}}
	}
	return wrapper.Emails, nil
}

// AddContactListEmails adds emails to a contact list and returns how many were
// added.
func (c *Client) AddContactListEmails(ctx context.Context, listID string, emails []string) (int, error) {
	var resp struct {
		Added int `json:"added"`
	}
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/lists/"+url.PathEscape(listID)+"/emails", nil, map[string]any{"emails": emails}, &resp); err != nil {
		return 0, err
	}
	return resp.Added, nil
}

// DeleteContactListEmails removes emails from a contact list and returns how
// many were removed.
func (c *Client) DeleteContactListEmails(ctx context.Context, listID string, emails []string) (int, error) {
	var resp struct {
		Deleted int `json:"deleted"`
	}
	if err := c.doRequest(ctx, http.MethodDelete, "/api/agent/lists/"+url.PathEscape(listID)+"/emails", nil, map[string]any{"emails": emails}, &resp); err != nil {
		return 0, err
	}
	return resp.Deleted, nil
}
