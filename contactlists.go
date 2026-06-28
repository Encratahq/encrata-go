package encrata

import (
	"context"
	"encoding/json"
	"net/http"
)

// ListContactLists returns all contact lists.
func (c *Client) ListContactLists(ctx context.Context) ([]ContactList, error) {
	var raw json.RawMessage
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/lists", nil, nil, &raw); err != nil {
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
	body := map[string]any{"name": name}
	if len(emails) > 0 {
		body["emails"] = emails
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
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/lists/"+listID, nil, nil, &cl); err != nil {
		return nil, err
	}
	return &cl, nil
}

// DeleteContactList deletes a contact list.
func (c *Client) DeleteContactList(ctx context.Context, listID string) error {
	return c.doRequest(ctx, http.MethodDelete, "/api/agent/lists/"+listID, nil, nil, nil)
}

// ListContactListEmails returns every email in a contact list.
func (c *Client) ListContactListEmails(ctx context.Context, listID string) ([]string, error) {
	var raw json.RawMessage
	if err := c.doRequest(ctx, http.MethodGet, "/api/agent/lists/"+listID+"/emails", nil, nil, &raw); err != nil {
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
	if err := c.doRequest(ctx, http.MethodPost, "/api/agent/lists/"+listID+"/emails", nil, map[string]any{"emails": emails}, &resp); err != nil {
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
	if err := c.doRequest(ctx, http.MethodDelete, "/api/agent/lists/"+listID+"/emails", nil, map[string]any{"emails": emails}, &resp); err != nil {
		return 0, err
	}
	return resp.Deleted, nil
}
