# encrata-go

The official Go SDK for the [Encrata](https://encrata.com) email intelligence
API.

Use it to look up people from email addresses, validate emails, check breach
exposure, enrich OSINT targets, scrape/extract pages, run bulk searches, manage
monitors, contact lists, workflows, webhooks, and API keys from Go.

Requires Go 1.25.6 or newer.

## Features

- Typed request and response models
- Context-aware methods for cancellation and deadlines
- Automatic retries for rate limits and transient server errors
- Typed errors for authentication, credits, validation, rate limits, and network failures
- Email intelligence, validation, and breach checks
- OSINT lookups for IPs, phones, domains, companies, Google dorks, and dark web search
- Web tools for scrape, extract, screenshot, and face search
- Bulk email lookup and bulk OSINT searches
- Monitoring, contact lists, workflows, webhooks, and API key management
- Pagination helpers for monitor runs and results

## Install

```bash
go get github.com/Encratahq/encrata-go
```

## Quickstart

Set your API key:

```bash
export ENCRATA_API_KEY="enc_live_..."
```

Create a client and look up one email:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	encrata "github.com/Encratahq/encrata-go"
)

func main() {
	client, err := encrata.New(os.Getenv("ENCRATA_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	person, err := client.Lookup(context.Background(), "elon@tesla.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(person.Name, person.Company)
}
```

## Configuration

Use defaults for most apps:

```go
client, err := encrata.New(os.Getenv("ENCRATA_API_KEY"))
```

The default client uses:

| Option      | Default                   |
| ----------- | ------------------------- |
| Base URL    | `https://api.encrata.com` |
| Timeout     | 30 seconds                |
| Max retries | 3                         |

Customize when needed:

```go
client, err := encrata.New("enc_live_...",
	encrata.WithTimeout(15*time.Second),
	encrata.WithMaxRetries(5),
	encrata.WithBaseURL("https://api.encrata.com"),
	encrata.WithHTTPClient(customClient),
)
```

Every request method takes a `context.Context`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

person, err := client.Lookup(ctx, "satya@microsoft.com")
```

## Email Intelligence

Look up a person by email:

```go
person, err := client.Lookup(ctx, "elon@tesla.com")
fmt.Println(person.Name, person.Company, person.Role)
```

Select specific fields:

```go
person, err := client.Lookup(ctx, "satya@microsoft.com",
	encrata.WithFields("name", "company", "role", "socials"),
)
```

Force a fresh lookup:

```go
person, err := client.Lookup(ctx, "elon@tesla.com", encrata.WithNoCache())
```

Validate an email address. This does not consume credits:

```go
validation, err := client.Validate(ctx, "satya@microsoft.com")
fmt.Println(validation.Validity, validation.Message)
```

Check data breaches. This does not consume credits:

```go
report, err := client.Breaches(ctx, "sundar@google.com")
fmt.Println(report.Count, report.Services)
```

## OSINT Lookups

Enrich IPs, phone numbers, domains, companies, Google dork queries, and dark web
mentions. Each lookup costs credits according to your API plan.

```go
ip, err := client.IP(ctx, "8.8.8.8")
fmt.Println(ip.Location["country"], ip.Company["name"])

phone, err := client.PhoneLookup(ctx, "+14155552671")
fmt.Println(phone.Country["name"], phone.Carrier["name"])

domain, err := client.DomainSearch(ctx, "tesla.com")
fmt.Println(domain.Whois["registrar"], domain.ThreatIntel["malicious"])

company, err := client.CompanySearch(ctx, "Tesla")
fmt.Println(company.Profile["name"], company.Total)

search, err := client.GoogleSearch(ctx, "site:example.com filetype:pdf")
for _, result := range search.Results {
	fmt.Println(result.Title, result.URL)
}

darkweb, err := client.DarkWebSearch(ctx, "user@example.com", 0)
fmt.Println(darkweb.Total)
```

## Web Tools

Scrape a page:

```go
page, err := client.Scrape(ctx, "https://example.com/pricing", true)
fmt.Println(page.StatusCode, page.Content)
```

Extract structured data:

```go
extracted, err := client.Extract(ctx, encrata.ExtractRequest{
	URL:           "https://example.com/product/123",
	Mode:          "selectors",
	Selectors:     map[string]string{"title": "h1", "price": ".price"},
	RenderJS:      true,
	BlockAds:      true,
	BlockTrackers: true,
})
fmt.Println(string(extracted.Extracted))
```

Capture a screenshot:

```go
shot, err := client.Screenshot(ctx, encrata.ScreenshotRequest{
	URL:           "https://example.com",
	FullPage:      true,
	Format:        "png",
	RenderJS:      true,
	BlockAds:      true,
	BlockTrackers: true,
})
fmt.Println(shot.Screenshot) // base64 image data
```

Run a face search:

```go
threshold := 0.9
face, err := client.FaceSearch(ctx, "https://example.com/photo.jpg", &threshold)
fmt.Println(face.Matched, face.FacesDetected)
```

## Bulk Operations

Bulk endpoints stream results over Server-Sent Events.

Bulk email lookup:

```go
results, errs := client.BulkLookup(ctx, []string{
	"elon@tesla.com",
	"satya@microsoft.com",
}, "name", "company")

for result := range results {
	if result.Err != nil {
		fmt.Println(result.Email, result.Err)
		continue
	}
	fmt.Println(result.Person.Name, result.Person.Company)
}
if err := <-errs; err != nil {
	log.Fatal(err)
}
```

Bulk OSINT searches collect the stream into a single response:

```go
res, err := client.BulkIPSearch(ctx, []string{"8.8.8.8", "1.1.1.1"})
fmt.Println(res.CreditsUsed)
for _, item := range res.Results {
	fmt.Println(item["query"], item["ip"])
}

client.BulkGoogleSearch(ctx, []string{"site:tesla.com", "site:microsoft.com"})
client.BulkCompanySearch(ctx, []string{"Tesla", "Microsoft"})
client.BulkDomainSearch(ctx, []string{"tesla.com", "microsoft.com"})
```

For in-process concurrent lookups that preserve input order, use `LookupMany`:

```go
results, err := client.LookupMany(ctx, emails, encrata.WithConcurrency(10))
```

## Monitoring

Create a monitor from emails:

```go
monitor, err := client.CreateMonitor(ctx, "Sales Leads",
	encrata.WithMonitorEmails("satya@microsoft.com", "jensen@nvidia.com"),
	encrata.WithFrequency("weekly"),
)
```

Create a monitor from a contact list:

```go
list, err := client.CreateContactList(ctx, "Newsletter", "satya@microsoft.com")
monitor, err := client.CreateMonitor(ctx, "Team Monitor", encrata.WithListID(list.ID))
```

List monitors and trigger a run:

```go
monitors, err := client.ListMonitors(ctx)
run, err := client.TriggerRun(ctx, monitors[0].ID)
fmt.Println(run.RunID, run.Status)
```

Get run results:

```go
runs, total, err := client.ListRuns(ctx, monitor.ID, 20, 0)
snapshots, total, err := client.GetRunResults(ctx, monitor.ID, runs[0].ID, true, 100, 0)
```

## Contact Lists

Create an email list:

```go
list, err := client.CreateContactList(ctx, "Engineering Team",
	"satya@microsoft.com",
	"sundar@google.com",
)
```

Create a non-email list:

```go
domains, err := client.CreateContactListWithRequest(ctx, encrata.ContactListRequest{
	Name:    "Competitor Domains",
	Type:    "domain",
	Targets: []string{"competitor1.com", "competitor2.io"},
})
```

Manage targets:

```go
lists, err := client.ListContactLists(ctx)
domainLists, err := client.ListContactListsByType(ctx, "domain")

added, err := client.AddContactListEmails(ctx, list.ID, []string{"tim@apple.com"})
emails, err := client.ListContactListEmails(ctx, list.ID)
deleted, err := client.DeleteContactListEmails(ctx, list.ID, []string{"sundar@google.com"})
err = client.DeleteContactList(ctx, list.ID)
```

## Workflows

Automate multi-step OSINT pipelines with triggers and steps.

```go
workflow, err := client.CreateWorkflow(ctx, encrata.WorkflowRequest{
	Name:    "Lead enrichment",
	Trigger: encrata.RawObject{"type": "webhook"},
	Steps: []encrata.RawObject{
		{"id": "step1", "type": "email_lookup", "config": map[string]any{"field": "email"}},
		{"id": "step2", "type": "company_lookup", "config": map[string]any{"field": "step1.company"}},
	},
})

workflows, total, err := client.ListWorkflows(ctx, encrata.WorkflowListOptions{
	Status: "active",
})

workflow, err = client.GetWorkflow(ctx, workflow.ID)
workflow, err = client.UpdateWorkflow(ctx, workflow.ID, encrata.WorkflowRequest{
	Name:   "Renamed",
	Status: "paused",
})
```

Runs, templates, and secrets:

```go
runs, total, err := client.ListWorkflowRuns(ctx, encrata.WorkflowRunListOptions{
	WorkflowID: workflow.ID,
})
run, err := client.GetWorkflowRun(ctx, runs[0].ID)

templates, err := client.ListWorkflowTemplates(ctx, "enrichment")

secrets, err := client.ListWorkflowSecrets(ctx)
_, err = client.CreateWorkflowSecret(ctx, "SLACK_WEBHOOK_URL", "https://hooks.slack.com/...")
_, err = client.DeleteWorkflowSecret(ctx, "SLACK_WEBHOOK_URL")
```

## Webhooks

Register endpoints, test delivery, and inspect recent attempts:

```go
webhook, err := client.CreateWebhook(ctx,
	"https://example.com/encrata/webhook",
	[]string{"monitor.run.completed"},
	"Production webhook",
)

webhooks, err := client.ListWebhooks(ctx)
_, err = client.TestWebhook(ctx, webhook.ID)
deliveries, err := client.ListWebhookDeliveries(ctx, webhook.ID)

active := false
_, err = client.UpdateWebhook(ctx, encrata.WebhookUpdateRequest{
	ID:       webhook.ID,
	URL:      webhook.URL,
	IsActive: &active,
})

_, err = client.DeleteWebhook(ctx, webhook.ID)
```

## API Keys

Manage account API keys. The full key is only returned once at creation.

```go
keys, err := client.ListKeys(ctx)

newKey, err := client.CreateKey(ctx, "Production")
fmt.Println(newKey.Key)

_, err = client.RevokeKey(ctx, newKey.ID, false) // soft disable
_, err = client.RevokeKey(ctx, newKey.ID, true)  // permanent delete
```

## Automatic Pagination

List endpoints return one page plus a total. Iterator helpers fetch monitor
pages on demand:

```go
for run, err := range client.IterAllRuns(ctx) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(run.ID, run.Status)
}

for snapshot, err := range client.IterAllResults(ctx, true) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(snapshot.Email, snapshot.HasChanges)
}

for run, err := range client.IterRuns(ctx, monitor.ID) {
	fmt.Println(run.ID)
}

for snapshot, err := range client.IterRunResults(ctx, monitor.ID, run.RunID, false) {
	fmt.Println(snapshot.Email)
}
```

## Error Handling

All API failures are typed. Use `errors.As` to inspect them:

```go
person, err := client.Lookup(ctx, "satya@microsoft.com")
if err != nil {
	var rl *encrata.RateLimitError
	switch {
	case errors.As(err, &rl):
		time.Sleep(rl.RetryAfter)
	case errors.As(err, new(*encrata.AuthenticationError)):
		log.Fatal("check your API key")
	case errors.As(err, new(*encrata.InsufficientCreditsError)):
		log.Fatal("out of credits")
	default:
		log.Fatal(err)
	}
}
```

| Error type                 | When                              |
| -------------------------- | --------------------------------- |
| `AuthenticationError`      | HTTP 401 - bad or missing API key |
| `InsufficientCreditsError` | HTTP 402 - out of credits         |
| `InvalidRequestError`      | HTTP 400 - bad parameters         |
| `RateLimitError`           | HTTP 429 - carries `RetryAfter`   |
| `APIError`                 | any other 4xx/5xx                 |
| `APIConnectionError`       | network failure or timeout        |

Transient failures (429 and 5xx) are retried automatically with full-jitter
exponential backoff, honoring `Retry-After`.

## MCP

Encrata also provides an MCP server for AI agent frameworks:

```json
{
  "mcpServers": {
    "encrata": {
      "url": "https://api.encrata.com/mcp",
      "headers": {
        "Authorization": "Bearer enc_live_..."
      }
    }
  }
}
```

## Support

- Documentation: [docs.encrata.com](https://docs.encrata.com)
- Dashboard: [encrata.com](https://encrata.com)
- Email: support@encrata.com

## License

MIT
