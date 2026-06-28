package testsuite

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	encrata "github.com/Encratahq/encrata-go"
)

func newClient(t *testing.T, h http.Handler) *encrata.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	client, err := encrata.New(
		"test-key",
		encrata.WithBaseURL(srv.URL),
		encrata.WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func TestMonitorMethods(t *testing.T) {
	ctx := context.Background()

	t.Run("list monitors", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/monitors")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"monitors": []map[string]any{
					{"id": "mon_123", "name": "Sales VIPs", "email_count": 2},
				},
			})
		}))

		monitors, err := client.ListMonitors(ctx)
		if err != nil {
			t.Fatalf("ListMonitors: %v", err)
		}
		if len(monitors) != 1 || monitors[0].ID != "mon_123" || monitors[0].EmailCount != 2 {
			t.Fatalf("monitors = %+v", monitors)
		}
	})

	t.Run("create monitor with emails", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodPost, "/api/agent/monitors")

			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["name"] != "Sales Leads" {
				t.Errorf("name = %v", body["name"])
			}
			if body["frequency"] != "weekly" {
				t.Errorf("frequency = %v", body["frequency"])
			}
			if body["change_detection"] != "full_refresh" {
				t.Errorf("change_detection = %v", body["change_detection"])
			}
			emails, ok := body["emails"].([]any)
			if !ok || len(emails) != 2 || emails[0] != "a@x.com" || emails[1] != "b@x.com" {
				t.Errorf("emails = %#v", body["emails"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "mon_123", "name": "Sales Leads"})
		}))

		monitor, err := client.CreateMonitor(
			ctx,
			"Sales Leads",
			encrata.WithMonitorEmails("a@x.com", "b@x.com"),
			encrata.WithFrequency("weekly"),
			encrata.WithChangeDetection("full_refresh"),
		)
		if err != nil {
			t.Fatalf("CreateMonitor: %v", err)
		}
		if monitor.ID != "mon_123" || monitor.Name != "Sales Leads" {
			t.Fatalf("monitor = %+v", monitor)
		}
	})

	t.Run("create monitor with contact list source", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodPost, "/api/agent/monitors")

			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["data_source_type"] != "list" {
				t.Errorf("data_source_type = %v", body["data_source_type"])
			}
			if body["data_source_ref"] != "list_456" {
				t.Errorf("data_source_ref = %v", body["data_source_ref"])
			}
			if _, ok := body["emails"]; ok {
				t.Errorf("emails should be omitted: %#v", body["emails"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "mon_123"})
		}))

		if _, err := client.CreateMonitor(ctx, "From List", encrata.WithListID("list_456")); err != nil {
			t.Fatalf("CreateMonitor: %v", err)
		}
	})

	t.Run("get monitor", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/monitors/mon_123")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "mon_123", "status": "active"})
		}))

		monitor, err := client.GetMonitor(ctx, "mon_123")
		if err != nil {
			t.Fatalf("GetMonitor: %v", err)
		}
		if monitor.ID != "mon_123" || monitor.Status != "active" {
			t.Fatalf("monitor = %+v", monitor)
		}
	})

	t.Run("trigger run", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodPost, "/api/agent/monitors/mon_123/run")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"run_id": "run_789",
				"status": "running",
			})
		}))

		run, err := client.TriggerRun(ctx, "mon_123")
		if err != nil {
			t.Fatalf("TriggerRun: %v", err)
		}
		if run.RunID != "run_789" || run.Status != "running" {
			t.Fatalf("run = %+v", run)
		}
	})
}

func TestMonitorPaginationMethods(t *testing.T) {
	ctx := context.Background()

	t.Run("list runs", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/monitors/mon_123/runs")
			assertQuery(t, r, "limit", "25")
			assertQuery(t, r, "offset", "50")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"runs":  []map[string]any{{"id": "run_1", "monitor_id": "mon_123"}},
				"total": 7,
			})
		}))

		runs, total, err := client.ListRuns(ctx, "mon_123", 25, 50)
		if err != nil {
			t.Fatalf("ListRuns: %v", err)
		}
		if total != 7 || len(runs) != 1 || runs[0].ID != "run_1" {
			t.Fatalf("runs=%+v total=%d", runs, total)
		}
	})

	t.Run("get run results", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/monitors/mon_123/runs/run_789/results")
			assertQuery(t, r, "changes_only", "true")
			assertQuery(t, r, "limit", "10")
			assertQuery(t, r, "offset", "20")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{"id": "snap_1", "email": "a@x.com", "has_changes": true}},
				"total":   3,
			})
		}))

		results, total, err := client.GetRunResults(ctx, "mon_123", "run_789", true, 10, 20)
		if err != nil {
			t.Fatalf("GetRunResults: %v", err)
		}
		if total != 3 || len(results) != 1 || !results[0].HasChanges {
			t.Fatalf("results=%+v total=%d", results, total)
		}
	})

	t.Run("list all runs", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/monitoring/runs")
			assertQuery(t, r, "limit", "5")
			assertQuery(t, r, "offset", "10")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"runs":  []map[string]any{{"id": "run_1"}, {"id": "run_2"}},
				"total": 2,
			})
		}))

		runs, total, err := client.ListAllRuns(ctx, 5, 10)
		if err != nil {
			t.Fatalf("ListAllRuns: %v", err)
		}
		if total != 2 || len(runs) != 2 {
			t.Fatalf("runs=%+v total=%d", runs, total)
		}
	})

	t.Run("list all results", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/monitoring/results")
			assertQuery(t, r, "changes_only", "true")
			assertQuery(t, r, "limit", "5")
			assertQuery(t, r, "offset", "10")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{"id": "snap_1", "email": "a@x.com"}},
				"total":   1,
			})
		}))

		results, total, err := client.ListAllResults(ctx, true, 5, 10)
		if err != nil {
			t.Fatalf("ListAllResults: %v", err)
		}
		if total != 1 || len(results) != 1 || results[0].Email != "a@x.com" {
			t.Fatalf("results=%+v total=%d", results, total)
		}
	})
}

func TestMonitorIterators(t *testing.T) {
	ctx := context.Background()

	t.Run("iter run results", func(t *testing.T) {
		client := newClient(t, pagedResultsHandler(t, "/api/agent/monitors/mon_123/runs/run_789/results", "results"))

		var ids []string
		for snap, err := range client.IterRunResults(ctx, "mon_123", "run_789", false) {
			if err != nil {
				t.Fatalf("IterRunResults: %v", err)
			}
			ids = append(ids, snap.ID)
		}
		if strings.Join(ids, ",") != "item_1,item_2,item_3" {
			t.Fatalf("ids = %v", ids)
		}
	})

	t.Run("iter all runs", func(t *testing.T) {
		client := newClient(t, pagedResultsHandler(t, "/api/agent/monitoring/runs", "runs"))

		var ids []string
		for run, err := range client.IterAllRuns(ctx) {
			if err != nil {
				t.Fatalf("IterAllRuns: %v", err)
			}
			ids = append(ids, run.ID)
		}
		if strings.Join(ids, ",") != "item_1,item_2,item_3" {
			t.Fatalf("ids = %v", ids)
		}
	})

	t.Run("iter all results", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertQuery(t, r, "changes_only", "true")
			pagedResultsHandler(t, "/api/agent/monitoring/results", "results").ServeHTTP(w, r)
		}))

		var ids []string
		for snap, err := range client.IterAllResults(ctx, true) {
			if err != nil {
				t.Fatalf("IterAllResults: %v", err)
			}
			ids = append(ids, snap.ID)
		}
		if strings.Join(ids, ",") != "item_1,item_2,item_3" {
			t.Fatalf("ids = %v", ids)
		}
	})
}

func TestContactListMethods(t *testing.T) {
	ctx := context.Background()

	t.Run("list contact lists array response", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/lists")
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": "list_123", "name": "Sales", "email_count": 2},
			})
		}))

		lists, err := client.ListContactLists(ctx)
		if err != nil {
			t.Fatalf("ListContactLists: %v", err)
		}
		if len(lists) != 1 || lists[0].ID != "list_123" || lists[0].EmailCount != 2 {
			t.Fatalf("lists = %+v", lists)
		}
	})

	t.Run("list contact lists wrapped response", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/lists")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"lists": []map[string]any{{"id": "list_123", "name": "Sales"}},
			})
		}))

		lists, err := client.ListContactLists(ctx)
		if err != nil {
			t.Fatalf("ListContactLists: %v", err)
		}
		if len(lists) != 1 || lists[0].Name != "Sales" {
			t.Fatalf("lists = %+v", lists)
		}
	})

	t.Run("create contact list", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodPost, "/api/agent/lists")

			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["name"] != "Engineering" {
				t.Errorf("name = %v", body["name"])
			}
			emails, ok := body["emails"].([]any)
			if !ok || len(emails) != 2 || emails[0] != "a@x.com" || emails[1] != "b@x.com" {
				t.Errorf("emails = %#v", body["emails"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "list_123", "name": "Engineering"})
		}))

		list, err := client.CreateContactList(ctx, "Engineering", "a@x.com", "b@x.com")
		if err != nil {
			t.Fatalf("CreateContactList: %v", err)
		}
		if list.ID != "list_123" || list.Name != "Engineering" {
			t.Fatalf("list = %+v", list)
		}
	})

	t.Run("get contact list", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/lists/list_123")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "list_123", "name": "Sales"})
		}))

		list, err := client.GetContactList(ctx, "list_123")
		if err != nil {
			t.Fatalf("GetContactList: %v", err)
		}
		if list.ID != "list_123" || list.Name != "Sales" {
			t.Fatalf("list = %+v", list)
		}
	})

	t.Run("delete contact list", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodDelete, "/api/agent/lists/list_123")
			w.WriteHeader(http.StatusNoContent)
		}))

		if err := client.DeleteContactList(ctx, "list_123"); err != nil {
			t.Fatalf("DeleteContactList: %v", err)
		}
	})
}

func TestContactListEmailMethods(t *testing.T) {
	ctx := context.Background()

	t.Run("list emails plain array", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/lists/list_123/emails")
			_ = json.NewEncoder(w).Encode([]string{"a@x.com", "b@x.com"})
		}))

		emails, err := client.ListContactListEmails(ctx, "list_123")
		if err != nil {
			t.Fatalf("ListContactListEmails: %v", err)
		}
		if strings.Join(emails, ",") != "a@x.com,b@x.com" {
			t.Fatalf("emails = %v", emails)
		}
	})

	t.Run("list emails object array", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/lists/list_123/emails")
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"email": "a@x.com"},
				{"email": "b@x.com"},
			})
		}))

		emails, err := client.ListContactListEmails(ctx, "list_123")
		if err != nil {
			t.Fatalf("ListContactListEmails: %v", err)
		}
		if strings.Join(emails, ",") != "a@x.com,b@x.com" {
			t.Fatalf("emails = %v", emails)
		}
	})

	t.Run("list emails wrapped response", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodGet, "/api/agent/lists/list_123/emails")
			_ = json.NewEncoder(w).Encode(map[string]any{"emails": []string{"a@x.com", "b@x.com"}})
		}))

		emails, err := client.ListContactListEmails(ctx, "list_123")
		if err != nil {
			t.Fatalf("ListContactListEmails: %v", err)
		}
		if strings.Join(emails, ",") != "a@x.com,b@x.com" {
			t.Fatalf("emails = %v", emails)
		}
	})

	t.Run("add emails", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodPost, "/api/agent/lists/list_123/emails")
			assertEmailBody(t, r, "a@x.com,b@x.com")
			_ = json.NewEncoder(w).Encode(map[string]any{"added": 2})
		}))

		added, err := client.AddContactListEmails(ctx, "list_123", []string{"a@x.com", "b@x.com"})
		if err != nil {
			t.Fatalf("AddContactListEmails: %v", err)
		}
		if added != 2 {
			t.Fatalf("added = %d", added)
		}
	})

	t.Run("delete emails", func(t *testing.T) {
		client := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertRequest(t, r, http.MethodDelete, "/api/agent/lists/list_123/emails")
			assertEmailBody(t, r, "a@x.com")
			_ = json.NewEncoder(w).Encode(map[string]any{"deleted": 1})
		}))

		deleted, err := client.DeleteContactListEmails(ctx, "list_123", []string{"a@x.com"})
		if err != nil {
			t.Fatalf("DeleteContactListEmails: %v", err)
		}
		if deleted != 1 {
			t.Fatalf("deleted = %d", deleted)
		}
	})
}

func assertRequest(t *testing.T, r *http.Request, method, path string) {
	t.Helper()
	if r.Method != method {
		t.Errorf("method = %s, want %s", r.Method, method)
	}
	if r.URL.Path != path {
		t.Errorf("path = %q, want %q", r.URL.Path, path)
	}
	if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
		t.Errorf("Authorization = %q", got)
	}
}

func assertQuery(t *testing.T, r *http.Request, key, want string) {
	t.Helper()
	if got := r.URL.Query().Get(key); got != want {
		t.Errorf("%s = %q, want %q; raw query %q", key, got, want, r.URL.RawQuery)
	}
}

func assertEmailBody(t *testing.T, r *http.Request, want string) {
	t.Helper()
	var body map[string][]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got := strings.Join(body["emails"], ","); got != want {
		t.Errorf("emails = %q, want %q", got, want)
	}
}

func pagedResultsHandler(t *testing.T, path, responseKey string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, http.MethodGet, path)
		switch r.URL.Query().Get("offset") {
		case "0":
			_ = json.NewEncoder(w).Encode(map[string]any{
				responseKey: []map[string]any{{"id": "item_1"}, {"id": "item_2"}},
				"total":     3,
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{
				responseKey: []map[string]any{{"id": "item_3"}},
				"total":     3,
			})
		}
	}
}
