# Sentinel Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the scheduler/concurrency infrastructure and replace it with a flat serial pipeline, adding interfaces for testability and sentinel errors for clean failure signaling.

**Architecture:** `handler` calls `scraper.Scrape` then `email.Send` in sequence. The `scheduler` package and both `tasks.go` files are deleted. The scraper's HTML parsing is separated from HTTP fetching behind a `Fetcher` interface; email sending is separated from SMTP transport behind a `MailSender` interface.

**Tech Stack:** Go 1.23, AWS Lambda Go runtime, `github.com/PuerkitoBio/goquery` (already an indirect dep), `net/smtp`, AWS Secrets Manager.

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `src/pkg/config/errors.go` | `ErrConfig` sentinel |
| Create | `src/pkg/scraper/errors.go` | `ErrScrape`, `ErrParse` sentinels |
| Create | `src/pkg/email/errors.go` | `ErrEmail`, `ErrBodyGenerate` sentinels |
| Rewrite | `src/pkg/scraper/scraper.go` | `Fetcher` interface, `HTTPFetcher`, `Scrape`, `parseEvents` |
| Rewrite | `src/pkg/scraper/scraper_test.go` | Unit tests for `parseEvents` and `Scrape` |
| Rewrite | `src/pkg/email/email.go` | `MailSender` interface, `SMTPSender`, `Send`, `generateBody` |
| Rewrite | `src/pkg/email/email_test.go` | Unit tests for `generateBody` and `Send` |
| Rewrite | `src/cmd/lambda/main.go` | Flat serial pipeline, no goroutines |
| Update | `src/pkg/scraper/tasks.go` | Update `Scrape` call to new signature (deleted in Task 6) |
| Delete | `src/pkg/scheduler/scheduler.go` | Entire package removed in Task 6 |
| Delete | `src/pkg/scraper/tasks.go` | Removed in Task 6 |
| Delete | `src/pkg/email/tasks.go` | Removed in Task 6 |
| Create | `src/pkg/scraper/scraper_integration_test.go` | Real HTTP test + benchmark, guarded by build tag |
| Create | `src/pkg/email/email_integration_test.go` | Real SMTP test, guarded by build tag |
| Update | `deploy/terraform/main.tf` | Lambda timeout 3 → 15 seconds |

---

### Task 1: Add sentinel error vars

**Files:**
- Create: `src/pkg/config/errors.go`
- Create: `src/pkg/scraper/errors.go`
- Create: `src/pkg/email/errors.go`

- [ ] **Step 1: Create `src/pkg/config/errors.go`**

```go
package config

import "errors"

var ErrConfig = errors.New("failed to load config")
```

- [ ] **Step 2: Create `src/pkg/scraper/errors.go`**

```go
package scraper

import "errors"

var (
	ErrScrape = errors.New("scrape failed")
	ErrParse  = errors.New("failed to parse events page")
)
```

- [ ] **Step 3: Create `src/pkg/email/errors.go`**

```go
package email

import "errors"

var (
	ErrEmail        = errors.New("email failed")
	ErrBodyGenerate = errors.New("failed to generate email body")
)
```

- [ ] **Step 4: Verify build is green**

```
cd src && go build ./...
```
Expected: no output, exit 0.

- [ ] **Step 5: Commit**

```bash
git add src/pkg/config/errors.go src/pkg/scraper/errors.go src/pkg/email/errors.go
git commit -m "feat: add sentinel error vars to config, scraper, and email packages"
```

---

### Task 2: TDD parseEvents — replace colly with goquery

**Files:**
- Rewrite: `src/pkg/scraper/scraper.go`
- Rewrite: `src/pkg/scraper/scraper_test.go`

- [ ] **Step 1: Rewrite `src/pkg/scraper/scraper_test.go` with failing tests**

```go
package scraper

import (
	"strings"
	"testing"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/event"
)

var target = time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)

const fixtureWithEvent = `<!DOCTYPE html>
<html><body>
<div class="info clearfix">
  <div class="date">Apr 9, 2026 - 7:30PM</div>
  <h3><a href="/events/detail/test">Dallas Stars vs. Colorado Avalanche</a></h3>
</div>
<div class="info clearfix">
  <div class="date">Apr 10, 2026 - 7PM</div>
  <h3><a href="/events/detail/test2">Mavericks vs. Warriors</a></h3>
</div>
</body></html>`

const fixtureNoEvents = `<!DOCTYPE html>
<html><body>
<div class="info clearfix">
  <div class="date">Apr 10, 2026 - 7PM</div>
  <h3><a href="/events/detail/test">Mavericks vs. Warriors</a></h3>
</div>
</body></html>`

const fixtureMalformedDate = `<!DOCTYPE html>
<html><body>
<div class="info clearfix">
  <div class="date">not a valid date</div>
  <h3><a href="/events/detail/test">Bad Event</a></h3>
</div>
<div class="info clearfix">
  <div class="date">Apr 9, 2026 - 7:30PM</div>
  <h3><a href="/events/detail/test2">Valid Event</a></h3>
</div>
</body></html>`

func TestParseEvents(t *testing.T) {
	tests := []struct {
		name string
		html string
		want []event.Event
	}{
		{
			name: "returns event matching target date",
			html: fixtureWithEvent,
			want: []event.Event{
				{Title: "Dallas Stars vs. Colorado Avalanche", When: "2026-04-09 7:30PM"},
			},
		},
		{
			name: "returns nil when no events on target date",
			html: fixtureNoEvents,
			want: nil,
		},
		{
			name: "skips malformed date, returns valid event",
			html: fixtureMalformedDate,
			want: []event.Event{
				{Title: "Valid Event", When: "2026-04-09 7:30PM"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseEvents(strings.NewReader(tt.html), target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d events, want %d: %v", len(got), len(tt.want), got)
			}
			for i, e := range got {
				if e != tt.want[i] {
					t.Errorf("event[%d] = %+v, want %+v", i, e, tt.want[i])
				}
			}
		})
	}
}
```

- [ ] **Step 2: Run tests — expect failure (parseEvents undefined)**

```
cd src && go test ./pkg/scraper/...
```
Expected: compile error `undefined: parseEvents`.

- [ ] **Step 3: Rewrite `src/pkg/scraper/scraper.go`**

Replace the entire file. This removes colly and introduces `parseEvents` and a direct `net/http` call:

```go
package scraper

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cjohnhelms/sentinel/pkg/event"
)

// Scrape fetches the AAC events page and returns events matching target date.
// Signature will gain a Fetcher parameter in Task 3; for now uses net/http directly.
func Scrape(target time.Time) ([]event.Event, error) {
	resp, err := http.Get("https://www.americanairlinescenter.com/events")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScrape, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: unexpected status %d", ErrScrape, resp.StatusCode)
	}
	return parseEvents(resp.Body, target)
}

func parseEvents(r io.Reader, target time.Time) ([]event.Event, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}

	layouts := []string{
		"Jan 2, 2006 3:04PM",
		"Jan 2, 2006 3PM",
	}

	var events []event.Event
	doc.Find("div.info.clearfix").Each(func(_ int, s *goquery.Selection) {
		dt := strings.TrimSpace(s.Find("div.date").Text())
		title := strings.TrimSpace(s.Find("h3 a").Text())

		parts := strings.SplitN(dt, "-", 2)
		if len(parts) != 2 {
			slog.Warn("unexpected date format", "date", dt)
			return
		}
		combined := strings.TrimSpace(parts[0]) + " " + strings.TrimSpace(parts[1])

		var eventTime time.Time
		var parseErr error
		for _, layout := range layouts {
			eventTime, parseErr = time.Parse(layout, combined)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			slog.Warn("failed to parse event date", "combined", combined)
			return
		}

		if eventTime.Year() == target.Year() &&
			eventTime.Month() == target.Month() &&
			eventTime.Day() == target.Day() {
			events = append(events, event.Event{
				Title: title,
				When:  eventTime.Format("2006-01-02 3:04PM"),
			})
		}
	})

	return events, nil
}
```

- [ ] **Step 4: Run tests — expect pass**

```
cd src && go test ./pkg/scraper/... -v -run TestParseEvents
```
Expected: `PASS` for all three subtests.

- [ ] **Step 5: Verify full build is green**

```
cd src && go build ./...
```
Expected: no output, exit 0. (`scraper/tasks.go` still calls `Scrape(target)` which still matches.)

- [ ] **Step 6: Commit**

```bash
git add src/pkg/scraper/scraper.go src/pkg/scraper/scraper_test.go
git commit -m "refactor: replace colly with goquery, extract testable parseEvents"
```

---

### Task 3: Add Fetcher interface and update Scrape signature

**Files:**
- Modify: `src/pkg/scraper/scraper.go`
- Modify: `src/pkg/scraper/scraper_test.go`
- Modify: `src/pkg/scraper/tasks.go`

- [ ] **Step 1: Add Fetcher tests to `src/pkg/scraper/scraper_test.go`**

Append these to the existing test file:

```go
import (
	// add these to the existing import block:
	"errors"
	"io"
)

type fakeFetcher struct {
	body string
	err  error
}

func (f *fakeFetcher) Fetch(_ string) (io.ReadCloser, error) {
	if f.err != nil {
		return nil, f.err
	}
	return io.NopCloser(strings.NewReader(f.body)), nil
}

func TestScrape_returnsTodayEvents(t *testing.T) {
	f := &fakeFetcher{body: fixtureWithEvent}
	got, err := Scrape(f, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
	if got[0].Title != "Dallas Stars vs. Colorado Avalanche" {
		t.Errorf("title = %q", got[0].Title)
	}
}

func TestScrape_fetchError(t *testing.T) {
	f := &fakeFetcher{err: errors.New("connection refused")}
	_, err := Scrape(f, target)
	if !errors.Is(err, ErrScrape) {
		t.Errorf("expected ErrScrape, got %v", err)
	}
}
```

- [ ] **Step 2: Run tests — expect failure (Scrape has wrong signature)**

```
cd src && go test ./pkg/scraper/... -run "TestScrape"
```
Expected: compile error — `Scrape` takes one arg, tests pass two.

- [ ] **Step 3: Add `Fetcher` interface and `HTTPFetcher`, update `Scrape` in `scraper.go`**

Replace the file:

```go
package scraper

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/cjohnhelms/sentinel/pkg/event"
)

// Fetcher abstracts the HTTP layer so Scrape can be tested without network calls.
type Fetcher interface {
	Fetch(url string) (io.ReadCloser, error)
}

// HTTPFetcher is the production Fetcher backed by net/http.
type HTTPFetcher struct {
	client *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *HTTPFetcher) Fetch(url string) (io.ReadCloser, error) {
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// Scrape fetches the AAC events page using f and returns events on target date.
func Scrape(f Fetcher, target time.Time) ([]event.Event, error) {
	body, err := f.Fetch("https://www.americanairlinescenter.com/events")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScrape, err)
	}
	defer body.Close()
	return parseEvents(body, target)
}

func parseEvents(r io.Reader, target time.Time) ([]event.Event, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrParse, err)
	}

	layouts := []string{
		"Jan 2, 2006 3:04PM",
		"Jan 2, 2006 3PM",
	}

	var events []event.Event
	doc.Find("div.info.clearfix").Each(func(_ int, s *goquery.Selection) {
		dt := strings.TrimSpace(s.Find("div.date").Text())
		title := strings.TrimSpace(s.Find("h3 a").Text())

		parts := strings.SplitN(dt, "-", 2)
		if len(parts) != 2 {
			slog.Warn("unexpected date format", "date", dt)
			return
		}
		combined := strings.TrimSpace(parts[0]) + " " + strings.TrimSpace(parts[1])

		var eventTime time.Time
		var parseErr error
		for _, layout := range layouts {
			eventTime, parseErr = time.Parse(layout, combined)
			if parseErr == nil {
				break
			}
		}
		if parseErr != nil {
			slog.Warn("failed to parse event date", "combined", combined)
			return
		}

		if eventTime.Year() == target.Year() &&
			eventTime.Month() == target.Month() &&
			eventTime.Day() == target.Day() {
			events = append(events, event.Event{
				Title: title,
				When:  eventTime.Format("2006-01-02 3:04PM"),
			})
		}
	})

	return events, nil
}
```

- [ ] **Step 4: Update `src/pkg/scraper/tasks.go` to use new signature**

The only change is passing `NewHTTPFetcher()` as the first argument so the package still compiles. This file will be deleted in Task 6.

Find this line in `tasks.go`:
```go
todaysEvents, err := Scrape(time.Now().In(zone))
```
Replace with:
```go
todaysEvents, err := Scrape(NewHTTPFetcher(), time.Now().In(zone))
```

- [ ] **Step 5: Run tests — expect pass**

```
cd src && go test ./pkg/scraper/... -v
```
Expected: all four tests pass (`TestParseEvents/...`, `TestScrape_returnsTodayEvents`, `TestScrape_fetchError`).

- [ ] **Step 6: Verify full build is green**

```
cd src && go build ./...
```
Expected: no output, exit 0.

- [ ] **Step 7: Commit**

```bash
git add src/pkg/scraper/scraper.go src/pkg/scraper/scraper_test.go src/pkg/scraper/tasks.go
git commit -m "feat: add Fetcher interface to scraper, inject HTTPFetcher in production path"
```

---

### Task 4: TDD standalone generateBody in email package

**Files:**
- Modify: `src/pkg/email/email.go`
- Modify: `src/pkg/email/email_test.go`

- [ ] **Step 1: Add `TestGenerateBody` to `src/pkg/email/email_test.go`**

Replace the entire file:

```go
package email

import (
	"strings"
	"testing"

	"github.com/cjohnhelms/sentinel/pkg/event"
)

func TestGenerateBody(t *testing.T) {
	events := []event.Event{
		{Title: "Dallas Stars vs. Avalanche", When: "2026-04-09 7:30PM"},
		{Title: "Concert", When: "2026-04-09 9PM"},
	}
	body, err := generateBody(events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(body, "Dallas Stars vs. Avalanche") {
		t.Error("body missing first event title")
	}
	if !strings.Contains(body, "Concert") {
		t.Error("body missing second event title")
	}
	if !strings.Contains(body, "Consider alternate routes") {
		t.Error("body missing footer")
	}
}
```

- [ ] **Step 2: Run tests — expect failure**

```
cd src && go test ./pkg/email/... -run TestGenerateBody
```
Expected: compile error — `generateBody` is a method on `*Emails`, not a package-level function. The test calling it as a plain function will fail.

- [ ] **Step 3: Add standalone `generateBody` to `src/pkg/email/email.go`**

Append to the bottom of the existing `email.go` (keep the `Emails` struct and its methods for now):

```go
// generateBody renders the email body for the given events.
// This is the standalone version; the method on *Emails will be removed in Task 6.
func generateBody(events []event.Event) (string, error) {
	var message strings.Builder
	tmpl, err := template.New("emailTemplate").Parse(
		"AAC Events:\n\n {{range .}}{{.Title}} - {{.When}}\n{{ end}}\n\nConsider alternate routes.",
	)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrBodyGenerate, err)
	}
	if err := tmpl.Execute(&message, events); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBodyGenerate, err)
	}
	return message.String(), nil
}
```

Note: the method `(e *Emails) generateBody` and the package-level function `generateBody` do not conflict in Go.

- [ ] **Step 4: Run tests — expect pass**

```
cd src && go test ./pkg/email/... -run TestGenerateBody -v
```
Expected: `PASS`.

- [ ] **Step 5: Verify full build is green**

```
cd src && go build ./...
```
Expected: no output, exit 0.

- [ ] **Step 6: Commit**

```bash
git add src/pkg/email/email.go src/pkg/email/email_test.go
git commit -m "feat: add standalone generateBody to email package"
```

---

### Task 5: Add MailSender interface and standalone Send

**Files:**
- Modify: `src/pkg/email/email.go`
- Modify: `src/pkg/email/email_test.go`

- [ ] **Step 1: Add `TestSend` and `TestSend_senderError` to `src/pkg/email/email_test.go`**

Append to the existing test file:

```go
import (
	// add to existing import block:
	"errors"

	"github.com/cjohnhelms/sentinel/pkg/config"
)

type fakeSender struct {
	capturedAddr string
	capturedFrom string
	capturedTo   []string
	capturedMsg  []byte
	err          error
}

func (f *fakeSender) SendMail(addr, from string, to []string, msg []byte) error {
	f.capturedAddr = addr
	f.capturedFrom = from
	f.capturedTo = to
	f.capturedMsg = msg
	return f.err
}

func TestSend(t *testing.T) {
	cfg := &config.Config{
		ServiceEmail:        "sentinel@example.com",
		EmailServer:         "smtp.example.com",
		RecipientEmails:     []string{"user@example.com"},
		EmailServerPassword: "secret",
	}
	events := []event.Event{
		{Title: "Test Event", When: "2026-04-09 7:30PM"},
	}
	sender := &fakeSender{}

	if err := Send(sender, cfg, events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.capturedAddr != "smtp.example.com:587" {
		t.Errorf("addr = %q, want smtp.example.com:587", sender.capturedAddr)
	}
	if sender.capturedFrom != "sentinel@example.com" {
		t.Errorf("from = %q, want sentinel@example.com", sender.capturedFrom)
	}
	if len(sender.capturedTo) != 1 || sender.capturedTo[0] != "user@example.com" {
		t.Errorf("to = %v, want [user@example.com]", sender.capturedTo)
	}
	if !strings.Contains(string(sender.capturedMsg), "Test Event") {
		t.Error("message does not contain event title")
	}
}

func TestSend_senderError(t *testing.T) {
	cfg := &config.Config{
		ServiceEmail:    "sentinel@example.com",
		EmailServer:     "smtp.example.com",
		RecipientEmails: []string{"user@example.com"},
	}
	events := []event.Event{{Title: "Test", When: "2026-04-09 7:30PM"}}
	sender := &fakeSender{err: errors.New("connection refused")}

	err := Send(sender, cfg, events)
	if !errors.Is(err, ErrEmail) {
		t.Errorf("expected ErrEmail, got %v", err)
	}
}
```

- [ ] **Step 2: Run tests — expect failure**

```
cd src && go test ./pkg/email/... -run "TestSend"
```
Expected: compile error — `Send` as a package-level function with these parameters is not defined.

- [ ] **Step 3: Add `MailSender`, `SMTPSender`, and standalone `Send` to `src/pkg/email/email.go`**

Append to the bottom of `email.go` (keep the `Emails` struct for now):

```go
// MailSender abstracts the SMTP transport layer.
type MailSender interface {
	SendMail(addr, from string, to []string, msg []byte) error
}

// SMTPSender is the production MailSender backed by net/smtp.
type SMTPSender struct {
	password string
}

func NewSMTPSender(password string) *SMTPSender {
	return &SMTPSender{password: password}
}

func (s *SMTPSender) SendMail(addr, from string, to []string, msg []byte) error {
	server := strings.Split(addr, ":")[0]
	auth := smtp.PlainAuth("", from, s.password, server)
	return smtp.SendMail(addr, auth, from, to, msg)
}

// Send renders the email body and sends it to all recipients.
func Send(s MailSender, cfg *config.Config, events []event.Event) error {
	body, err := generateBody(events)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:587", cfg.EmailServer)
	msg := []byte(fmt.Sprintf("From: Sentinel %s\nSubject: AAC Sentinel Notification\n\n%s",
		cfg.ServiceEmail, body))
	if err := s.SendMail(addr, cfg.ServiceEmail, cfg.RecipientEmails, msg); err != nil {
		return fmt.Errorf("%w: %v", ErrEmail, err)
	}
	return nil
}
```

- [ ] **Step 4: Run tests — expect pass**

```
cd src && go test ./pkg/email/... -v -run "TestGenerateBody|TestSend"
```
Expected: all three tests pass.

- [ ] **Step 5: Verify full build is green**

```
cd src && go build ./...
```
Expected: no output, exit 0.

- [ ] **Step 6: Commit**

```bash
git add src/pkg/email/email.go src/pkg/email/email_test.go
git commit -m "feat: add MailSender interface and standalone Send to email package"
```

---

### Task 6: Rewire main.go and delete dead code

**Files:**
- Rewrite: `src/cmd/lambda/main.go`
- Rewrite: `src/pkg/email/email.go` (remove `Emails` struct and old methods)
- Delete: `src/pkg/scheduler/scheduler.go`
- Delete: `src/pkg/scraper/tasks.go`
- Delete: `src/pkg/email/tasks.go`

- [ ] **Step 1: Rewrite `src/cmd/lambda/main.go`**

```go
package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/email"
	"github.com/cjohnhelms/sentinel/pkg/scraper"
)

func main() {
	lambda.Start(handler)
}

func handler() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("%w: %v", config.ErrConfig, err)
	}
	config.InitLogger(cfg.Log)

	zone, err := time.LoadLocation("America/Chicago")
	if err != nil {
		slog.Warn("failed to load timezone, using UTC")
		zone = time.UTC
	}

	events, err := scraper.Scrape(scraper.NewHTTPFetcher(), time.Now().In(zone))
	if err != nil {
		return err
	}
	if len(events) == 0 {
		slog.Info("no events today")
		return nil
	}

	return email.Send(email.NewSMTPSender(cfg.EmailServerPassword), cfg, events)
}
```

- [ ] **Step 2: Rewrite `src/pkg/email/email.go`** — remove `Emails` struct and old methods, keep only what's needed

```go
package email

import (
	"fmt"
	"net/smtp"
	"strings"
	"text/template"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/event"
)

// MailSender abstracts the SMTP transport layer.
type MailSender interface {
	SendMail(addr, from string, to []string, msg []byte) error
}

// SMTPSender is the production MailSender backed by net/smtp.
type SMTPSender struct {
	password string
}

func NewSMTPSender(password string) *SMTPSender {
	return &SMTPSender{password: password}
}

func (s *SMTPSender) SendMail(addr, from string, to []string, msg []byte) error {
	server := strings.Split(addr, ":")[0]
	auth := smtp.PlainAuth("", from, s.password, server)
	return smtp.SendMail(addr, auth, from, to, msg)
}

// Send renders the email body and sends it to all recipients.
func Send(s MailSender, cfg *config.Config, events []event.Event) error {
	body, err := generateBody(events)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("%s:587", cfg.EmailServer)
	msg := []byte(fmt.Sprintf("From: Sentinel %s\nSubject: AAC Sentinel Notification\n\n%s",
		cfg.ServiceEmail, body))
	if err := s.SendMail(addr, cfg.ServiceEmail, cfg.RecipientEmails, msg); err != nil {
		return fmt.Errorf("%w: %v", ErrEmail, err)
	}
	return nil
}

func generateBody(events []event.Event) (string, error) {
	var message strings.Builder
	tmpl, err := template.New("emailTemplate").Parse(
		"AAC Events:\n\n {{range .}}{{.Title}} - {{.When}}\n{{ end}}\n\nConsider alternate routes.",
	)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrBodyGenerate, err)
	}
	if err := tmpl.Execute(&message, events); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBodyGenerate, err)
	}
	return message.String(), nil
}
```

- [ ] **Step 3: Delete dead files**

```bash
rm src/pkg/scheduler/scheduler.go
rmdir src/pkg/scheduler
rm src/pkg/scraper/tasks.go
rm src/pkg/email/tasks.go
```

- [ ] **Step 4: Run all tests and verify build**

```
cd src && go build ./... && go test ./...
```
Expected: build succeeds, all unit tests pass.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "refactor: flat serial pipeline in handler, delete scheduler and task wrappers"
```

---

### Task 7: Clean up go.mod

**Files:**
- Modify: `src/go.mod`
- Modify: `src/go.sum`

- [ ] **Step 1: Promote goquery to a direct dependency**

```
cd src && go get github.com/PuerkitoBio/goquery
```

- [ ] **Step 2: Remove unused direct dependencies and tidy**

```
cd src && go mod tidy
```

This removes `github.com/google/uuid`, `github.com/gorhill/cronexpr`, `github.com/gocolly/colly` and their unique transitive dependencies from `go.mod`.

- [ ] **Step 3: Verify build still passes**

```
cd src && go build ./... && go test ./...
```
Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add src/go.mod src/go.sum
git commit -m "chore: go mod tidy — remove colly, uuid, cronexpr; promote goquery to direct dep"
```

---

### Task 8: Guard integration tests with build tag

**Files:**
- Create: `src/pkg/scraper/scraper_integration_test.go`
- Create: `src/pkg/email/email_integration_test.go`

- [ ] **Step 1: Create `src/pkg/scraper/scraper_integration_test.go`**

```go
//go:build integration

package scraper

import (
	"testing"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/event"
)

var blackhole []event.Event

func TestScrape_integration(t *testing.T) {
	events, err := Scrape(NewHTTPFetcher(), time.Now())
	if err != nil {
		t.Fatalf("failed to scrape events: %v", err)
	}
	if len(events) == 0 {
		t.Log("scrape succeeded but no events found today")
	}
}

func BenchmarkScrapeEvents(b *testing.B) {
	b.ReportAllocs()
	f := NewHTTPFetcher()
	target := time.Date(2025, 7, 22, 0, 0, 0, 0, time.Local)
	for i := 0; i < b.N; i++ {
		events, err := Scrape(f, target)
		if err != nil {
			b.Fatalf("failed to scrape events: %v", err)
		}
		blackhole = events
	}
}
```

- [ ] **Step 2: Create `src/pkg/email/email_integration_test.go`**

```go
//go:build integration

package email

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cjohnhelms/sentinel/pkg/config"
	"github.com/cjohnhelms/sentinel/pkg/event"
)

func TestSend_integration(t *testing.T) {
	vars, err := loadEnv()
	if err != nil {
		t.Fatalf("missing env vars: %v", err)
	}
	cfg := &config.Config{
		RecipientEmails:     strings.Split(vars["email_recipients"], ","),
		ServiceEmail:        vars["service_email"],
		EmailServer:         vars["email_server"],
		EmailServerPassword: vars["email_server_password"],
	}
	events := []event.Event{
		{Title: "Test title 1", When: time.Now().Format("2006-01-02 3:04PM")},
		{Title: "Test title 2", When: time.Now().Format("2006-01-02 3:04PM")},
	}
	if err := Send(NewSMTPSender(cfg.EmailServerPassword), cfg, events); err != nil {
		t.Error(err)
	}
}

func loadEnv() (map[string]string, error) {
	keys := []string{"EMAIL_RECIPIENTS", "EMAIL_SERVER", "EMAIL_SERVER_PASSWORD", "SERVICE_EMAIL"}
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		if !ok {
			return nil, errors.New(k + " not set")
		}
		out[strings.ToLower(k)] = v
	}
	return out, nil
}
```

- [ ] **Step 3: Verify `go test ./...` passes without -tags integration**

```
cd src && go test ./...
```
Expected: all unit tests pass; integration tests not compiled.

- [ ] **Step 4: Verify integration tests compile with the tag**

```
cd src && go build -tags integration ./...
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add src/pkg/scraper/scraper_integration_test.go src/pkg/email/email_integration_test.go
git commit -m "test: move integration tests behind //go:build integration tag"
```

---

### Task 9: Bump Lambda timeout in Terraform

**Files:**
- Modify: `deploy/terraform/main.tf`

- [ ] **Step 1: Update timeout in `deploy/terraform/main.tf`**

Find:
```hcl
  timeout       = 3
```
Replace with:
```hcl
  timeout       = 15
```

- [ ] **Step 2: Commit**

```bash
git add deploy/terraform/main.tf
git commit -m "fix: increase Lambda timeout from 3s to 15s to accommodate scrape and SMTP latency"
```
