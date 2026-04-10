# Sentinel Refactor Design

**Date:** 2026-04-09  
**Status:** Approved

## Summary

Refactor the Sentinel Lambda to remove unnecessary concurrency infrastructure and replace it with a simple, serial pipeline. The two tasks — scraping the AAC events page and sending an email — are inherently serial and have no need for goroutines, channels, or a scheduler. The goal is a reliable, testable codebase with clear error propagation and composable pure functions.

---

## Architecture

The Lambda handler becomes a flat serial pipeline:

1. Load config from AWS Secrets Manager
2. Scrape the AAC events page for today's events
3. If no events, log and return nil
4. Send email to recipients
5. Return any error to the Lambda runtime (signals failed invocation to CloudWatch)

EventBridge Scheduler (cron `0 9 * * ? *`, America/Chicago) continues to own the scheduling concern. The `scheduler` package is deleted — Lambda has no need to re-implement scheduling internally.

OS signal handling (`SIGINT`/`SIGTERM`) is removed. It serves no purpose in a Lambda execution environment.

---

## Package Structure

### Deleted
- `src/pkg/scheduler/` — entire package removed
- `src/pkg/scraper/tasks.go` — task wrapper removed
- `src/pkg/email/tasks.go` — task wrapper removed

### Remaining (refactored)
- `src/cmd/lambda/main.go` — flat serial pipeline, no goroutines
- `src/pkg/config/` — unchanged except called once
- `src/pkg/event/` — unchanged
- `src/pkg/scraper/` — fetch/parse split, `Fetcher` interface added
- `src/pkg/email/` — `MailSender` interface added

### Dependencies removed
- `github.com/google/uuid` — no longer needed (task IDs gone)
- `github.com/gorhill/cronexpr` — no longer needed (scheduler gone)
- `github.com/gocolly/colly` — replaced by `net/http` + `goquery` (already a transitive dep)

---

## Interface Design

### Scraper

The core problem with the current scraper is that colly bundles HTTP fetching and HTML parsing in callbacks, making the parsing logic impossible to unit test without a live network call. The fix is to separate fetch from parse.

```go
// Fetcher abstracts the HTTP layer
type Fetcher interface {
    Fetch(url string) (io.ReadCloser, error)
}

// HTTPFetcher is the real implementation
type HTTPFetcher struct{ client *http.Client }

func NewHTTPFetcher() *HTTPFetcher {
    return &HTTPFetcher{client: &http.Client{Timeout: 10 * time.Second}}
}

// Scrape is the public entry point — Fetcher is injected
func Scrape(f Fetcher, target time.Time) ([]event.Event, error)

// parseEvents is unexported — pure function, takes io.Reader
func parseEvents(r io.Reader, target time.Time) ([]event.Event, error)
```

### Email

```go
// MailSender abstracts the SMTP layer
type MailSender interface {
    SendMail(addr, from string, to []string, msg []byte) error
}

// SMTPSender is the real implementation wrapping net/smtp
type SMTPSender struct{}

// Send is the public entry point — MailSender is injected
func Send(s MailSender, cfg *config.Config, events []event.Event) error

// generateBody is unexported — pure function
func generateBody(events []event.Event) (string, error)
```

---

## Error Handling

Each package defines sentinel errors for its domain. Errors are wrapped at the call site in `handler` so CloudWatch logs contain both the sentinel and the underlying cause. Callers can use `errors.Is` to distinguish failure modes.

```go
// scraper/errors.go
var ErrScrape = errors.New("scrape failed")
var ErrParse  = errors.New("failed to parse events page")

// email/errors.go
var ErrEmail        = errors.New("email failed")
var ErrBodyGenerate = errors.New("failed to generate email body")

// config/errors.go
var ErrConfig = errors.New("failed to load config")
```

`handler` returns errors to the Lambda runtime. A non-nil return causes Lambda to record a failed invocation in CloudWatch metrics, which is the correct signal for internal application failures. EventBridge Scheduler does not auto-retry cron invocations by default, so there is no retry risk.

---

## Testing

### Unit tests (no network, no AWS)

**Scraper:** `parseEvents` is a pure function taking `io.Reader`. Tests feed fixture HTML strings directly. Table-driven, covering: event on target date, no events, malformed date string, multiple events.

**Email:** `generateBody` is a pure function taking `[]event.Event`. Tests verify rendered output. `Send` is tested with a `fakeSender` that captures the message bytes, verifying the correct address and content are passed.

### Integration tests (real network / real SMTP)

Existing `TestEmail_Send` (real SMTP via env vars) and `BenchmarkScrapeEvents` (real HTTP) are retained but guarded with a build tag (`//go:build integration`) so they are excluded from standard `go test ./...` runs in CI.

---

## Lambda Timeout

The Terraform config sets the Lambda timeout to 3 seconds. This is likely too tight for a real HTTP scrape plus an SMTP round-trip. Recommend increasing to 15–30 seconds.
