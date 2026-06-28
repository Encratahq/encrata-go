# encrata-go

The official Go SDK for the [Encrata](https://encrata.com) email intelligence API.

```bash
go get github.com/Encratahq/encrata-go
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	encrata "github.com/Encratahq/encrata-go"
)

func main() {
	client, err := encrata.New("enc_live_...")
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

`New` takes the API key plus optional functional options:

```go
client, err := encrata.New("enc_live_...",
	encrata.WithBaseURL("https://api.encrata.com"),
	encrata.WithTimeout(15*time.Second),
	encrata.WithMaxRetries(5),
	encrata.WithHTTPClient(customClient),
)
```

Every request method takes a `context.Context` for cancellation and deadlines.

## Email intelligence

```go
person, err := client.Lookup(ctx, "satya@microsoft.com",
	encrata.WithFields("name", "company", "socials"),
	encrata.WithNoCache(),
)

validation, err := client.Validate(ctx, "sundar@google.com") // free, no credits
report, err := client.Breaches(ctx, "tim@apple.com")          // free, no credits
```

### Bulk lookups

`LookupMany` enriches many emails concurrently, preserving input order. A
failure on one email is reported inline so it never discards results you already
paid for.

```go
results, err := client.LookupMany(ctx, emails, encrata.WithConcurrency(10))
for _, r := range results {
	if r.Err != nil {
		fmt.Printf("%s failed: %v\n", r.Email, r.Err)
		continue
	}
	fmt.Println(r.Email, r.Person.Name)
}
```

## Monitors

```go
m, err := client.CreateMonitor(ctx, "VIP list",
	encrata.WithMonitorEmails("satya@microsoft.com", "jensen@nvidia.com"),
	encrata.WithFrequency("weekly"),
	encrata.WithChangeDetection("diff_only"),
)

run, err := client.TriggerRun(ctx, m.ID)
```

### Pagination

List methods return one page plus a total; `Iter*` methods stream every page on
demand using Go 1.23 range-over-func iterators:

```go
for run, err := range client.IterRuns(ctx, monitorID) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(run.ID, run.Status)
}

for snap, err := range client.IterAllResults(ctx, true /* changesOnly */) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(snap.Email, snap.HasChanges)
}
```

## Contact lists

```go
list, err := client.CreateContactList(ctx, "Newsletter", "satya@microsoft.com")
emails, err := client.ListContactListEmails(ctx, list.ID)
added, err := client.AddContactListEmails(ctx, list.ID, []string{"tim@apple.com"})
```

## Error handling

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

| Error type                  | When                                  |
| --------------------------- | ------------------------------------- |
| `AuthenticationError`       | HTTP 401 — bad or missing API key     |
| `InsufficientCreditsError`  | HTTP 402 — out of credits             |
| `InvalidRequestError`       | HTTP 400 — bad parameters             |
| `RateLimitError`            | HTTP 429 — carries `RetryAfter`       |
| `APIError`                  | any other 4xx/5xx                     |
| `APIConnectionError`        | network failure or timeout            |

Transient failures (429 and 5xx) are retried automatically with full-jitter
exponential backoff, honoring `Retry-After`.

## License

MIT
