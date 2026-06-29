# encrata-go

The official Go SDK for the [Encrata](https://encrata.com) email intelligence API.

Use it to look up people from email addresses, validate emails, check breach
exposure, manage monitors, and work with contact lists from Go.

Requires Go 1.25.6 or newer.

## Features

- Typed request and response models
- Context-aware methods for cancellation and deadlines
- Automatic retries for rate limits and transient server errors
- Typed errors for authentication, credits, validation, rate limits, and network failures
- Bulk lookup helpers with concurrency control
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

Create a client:

```go
ctx := context.Background()

client, err := encrata.New(os.Getenv("ENCRATA_API_KEY"))
if err != nil {
	log.Fatal(err)
}
```

Look up one email:

```go
person, err := client.Lookup(ctx, "elon@tesla.com")
if err != nil {
	log.Fatal(err)
}

fmt.Println(person.Name, person.Company)
```

Complete example:

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

	ctx := context.Background()

	person, err := client.Lookup(ctx, "elon@tesla.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(person.Name, person.Company)
}
```

## Common examples

```go
person, err := client.Lookup(ctx, "satya@microsoft.com")
validation, err := client.Validate(ctx, "sundar@google.com")
breaches, err := client.Breaches(ctx, "tim@apple.com")
monitors, err := client.ListMonitors(ctx)
lists, err := client.ListContactLists(ctx)
```

## Configuration

Use the defaults for most apps:

```go
client, err := encrata.New(os.Getenv("ENCRATA_API_KEY"))
```

The default client uses:

| Option      | Default                    |
| ----------- | -------------------------- |
| Base URL    | `https://api.encrata.com`  |
| Timeout     | 30 seconds                 |
| Max retries | 3                          |

Add options only when you need them:

```go
client, err := encrata.New("enc_live_...",
	encrata.WithTimeout(15*time.Second),
	encrata.WithMaxRetries(5),
)
```

Use a custom API URL or HTTP client for tests, proxies, or internal gateways:

```go
client, err := encrata.New("enc_live_...",
	encrata.WithBaseURL("https://api.encrata.com"),
	encrata.WithHTTPClient(customClient),
)
```

Every request method takes a `context.Context` for cancellation and deadlines.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

person, err := client.Lookup(ctx, "satya@microsoft.com")
```

## Email intelligence

Basic lookup:

```go
person, err := client.Lookup(ctx, "satya@microsoft.com")
```

Request only specific fields:

```go
person, err := client.Lookup(ctx, "satya@microsoft.com",
	encrata.WithFields("name", "company", "socials"),
)
```

Force a fresh lookup:

```go
person, err := client.Lookup(ctx, "satya@microsoft.com",
	encrata.WithNoCache(),
)
```

Free checks:

```go
validation, err := client.Validate(ctx, "sundar@google.com")
report, err := client.Breaches(ctx, "tim@apple.com")
```

`Validate` and `Breaches` do not consume credits.

### Bulk lookups

`LookupMany` enriches many emails concurrently, preserving input order. A
failure on one email is reported inline so it never discards results you already
paid for.

```go
emails := []string{
	"satya@microsoft.com",
	"sundar@google.com",
	"tim@apple.com",
}

results, err := client.LookupMany(ctx, emails, encrata.WithConcurrency(10))
if err != nil {
	log.Fatal(err)
}

for _, r := range results {
	if r.Err != nil {
		fmt.Printf("%s failed: %v\n", r.Email, r.Err)
		continue
	}
	fmt.Println(r.Email, r.Person.Name)
}
```

## Monitors

Create a monitor from emails:

```go
m, err := client.CreateMonitor(ctx, "VIP list",
	encrata.WithMonitorEmails("satya@microsoft.com", "jensen@nvidia.com"),
	encrata.WithFrequency("weekly"),
)
```

Create a monitor from a contact list:

```go
list, err := client.CreateContactList(ctx, "Newsletter", "satya@microsoft.com")
if err != nil {
	log.Fatal(err)
}

m, err := client.CreateMonitor(ctx, "Newsletter",
	encrata.WithListID(list.ID),
)
```

Trigger a run:

```go
run, err := client.TriggerRun(ctx, m.ID)
```

### Pagination

List one page:

```go
const monitorID = "mon_..."

runs, total, err := client.ListRuns(ctx, monitorID, 20, 0)
```

Stream every page with Go 1.23 range-over-func iterators:

```go
for run, err := range client.IterRuns(ctx, monitorID) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(run.ID, run.Status)
}
```

Stream changed results only:

```go
for snap, err := range client.IterAllResults(ctx, true /* changesOnly */) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(snap.Email, snap.HasChanges)
}
```

## Contact lists

Create a list:

```go
list, err := client.CreateContactList(ctx, "Newsletter", "satya@microsoft.com")
```

Read emails from a list:

```go
emails, err := client.ListContactListEmails(ctx, list.ID)
```

Add emails:

```go
added, err := client.AddContactListEmails(ctx, list.ID, []string{"tim@apple.com"})
```

Remove emails:

```go
deleted, err := client.DeleteContactListEmails(ctx, list.ID, []string{"tim@apple.com"})
```

Delete a list:

```go
err := client.DeleteContactList(ctx, list.ID)
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
| `AuthenticationError`       | HTTP 401 - bad or missing API key     |
| `InsufficientCreditsError`  | HTTP 402 - out of credits             |
| `InvalidRequestError`       | HTTP 400 - bad parameters             |
| `RateLimitError`            | HTTP 429 - carries `RetryAfter`       |
| `APIError`                  | any other 4xx/5xx                     |
| `APIConnectionError`        | network failure or timeout            |

Transient failures (429 and 5xx) are retried automatically with full-jitter
exponential backoff, honoring `Retry-After`.

## License

MIT
