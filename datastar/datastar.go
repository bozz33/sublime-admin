// Package datastar provides server-side helpers for the Datastar framework.
//
// Datastar (https://data-star.dev) is a lightweight (11KB) library that combines
// HTMX-style server-driven UI with Alpine.js-style client signals. It replaces
// both HTMX and Alpine.js in SublimeGo's architecture.
//
// # How it works
//
//  1. The browser sends a GET/POST/PATCH/DELETE request
//  2. The Go handler responds with SSE (Server-Sent Events)
//  3. Datastar reads the SSE events and updates the DOM automatically
//
// # SSE Event Types
//
//   - datastar-merge-fragments: merge HTML into the DOM (like hx-swap)
//   - datastar-merge-signals:   update client-side signal store
//   - datastar-remove-fragments: remove DOM elements
//   - datastar-redirect:        navigate to a new URL
//   - datastar-execute-script:  run JavaScript on the client
//
// # Usage
//
//	func MyHandler(w http.ResponseWriter, r *http.Request) {
//	    sse := datastar.NewSSE(w)
//	    sse.MergeFragment(`<div id="result">Hello!</div>`)
//	    sse.MergeSignals(map[string]any{"loading": false})
//	}
package datastar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
)

// SSE wraps an http.ResponseWriter and provides methods to send Datastar SSE events.
type SSE struct {
	w http.ResponseWriter
}

// NewSSE prepares the ResponseWriter for SSE streaming and returns an SSE helper.
// Call this at the start of any handler that will use Datastar server-push.
func NewSSE(w http.ResponseWriter) *SSE {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	return &SSE{w: w}
}

// MergeFragment sends a datastar-merge-fragments event.
// The HTML fragment must contain an element with an id attribute that
// matches an existing element in the DOM (for morphing/merging).
//
//	sse.MergeFragment(`<div id="user-card"><p>Updated name</p></div>`)
func (s *SSE) MergeFragment(html string) {
	s.writeEvent("datastar-merge-fragments", "fragments "+html)
}

// MergeFragmentTempl renders a templ component and sends it as a merge-fragments event.
func (s *SSE) MergeFragmentTempl(ctx context.Context, c templ.Component) error {
	var buf bytes.Buffer
	if err := c.Render(ctx, &buf); err != nil {
		return fmt.Errorf("datastar: render component: %w", err)
	}
	s.MergeFragment(buf.String())
	return nil
}

// AppendFragment sends a datastar-merge-fragments event with mergeMode=append.
// The fragment is appended inside the target element instead of replacing it.
func (s *SSE) AppendFragment(selector, html string) {
	s.writeEvent("datastar-merge-fragments",
		fmt.Sprintf("selector %s\nmergeMode append\nfragments %s", selector, html))
}

// PrependFragment sends a datastar-merge-fragments event with mergeMode=prepend.
func (s *SSE) PrependFragment(selector, html string) {
	s.writeEvent("datastar-merge-fragments",
		fmt.Sprintf("selector %s\nmergeMode prepend\nfragments %s", selector, html))
}

// RemoveFragments sends a datastar-remove-fragments event for the given CSS selector.
//
//	sse.RemoveFragments("#toast-1")
func (s *SSE) RemoveFragments(selector string) {
	s.writeEvent("datastar-remove-fragments", "selector "+selector)
}

// MergeSignals sends a datastar-merge-signals event to update client signal values.
// Signals are merged into the existing signal store (not replaced).
//
//	sse.MergeSignals(map[string]any{"loading": false, "count": 42})
func (s *SSE) MergeSignals(signals map[string]any) {
	data, _ := json.Marshal(signals)
	s.writeEvent("datastar-merge-signals", "signals "+string(data))
}

// Redirect sends a datastar-redirect event causing the browser to navigate.
//
//	sse.Redirect("/admin/users")
func (s *SSE) Redirect(url string) {
	s.writeEvent("datastar-redirect", "url "+url)
}

// ExecuteScript sends a datastar-execute-script event to run JavaScript on the client.
//
//	sse.ExecuteScript("console.log('hello from server')")
func (s *SSE) ExecuteScript(script string) {
	s.writeEvent("datastar-execute-script", "script "+script)
}

// Toast sends a signal update that triggers the toast system in app.js.
// The toast type can be "success", "error", "warning", or "info".
func (s *SSE) Toast(message, toastType string) {
	s.MergeSignals(map[string]any{
		"toastMessage": message,
		"toastType":    toastType,
		"toastVisible": true,
	})
}

// writeEvent writes a raw SSE event.
func (s *SSE) writeEvent(eventType, data string) {
	lines := strings.Split(data, "\n")
	fmt.Fprintf(s.w, "event: %s\n", eventType)
	for _, line := range lines {
		fmt.Fprintf(s.w, "data: %s\n", line)
	}
	fmt.Fprintf(s.w, "\n")
	if f, ok := s.w.(http.Flusher); ok {
		f.Flush()
	}
}

// ---------------------------------------------------------------------------
// Convenience standalone functions (for simple one-shot responses).
// ---------------------------------------------------------------------------

// MergeFragment is a convenience function that sets SSE headers and sends a single fragment.
// Use NewSSE() for handlers that send multiple events.
func MergeFragment(w http.ResponseWriter, html string) {
	NewSSE(w).MergeFragment(html)
}

// MergeSignals is a convenience function that sets SSE headers and sends signal updates.
func MergeSignals(w http.ResponseWriter, signals map[string]any) {
	NewSSE(w).MergeSignals(signals)
}

// Redirect is a convenience function that sets SSE headers and sends a redirect.
func Redirect(w http.ResponseWriter, url string) {
	NewSSE(w).Redirect(url)
}

// RemoveFragments is a convenience function that sets SSE headers and sends a remove event.
func RemoveFragments(w http.ResponseWriter, selector string) {
	NewSSE(w).RemoveFragments(selector)
}

// ---------------------------------------------------------------------------
// Request helpers — reading signals sent by the client.
// ---------------------------------------------------------------------------

// ReadSignals parses Datastar signals from the request body (for POST/PATCH requests).
// Returns an empty map if the body is not valid JSON or not present.
func ReadSignals(r *http.Request) map[string]any {
	signals := make(map[string]any)
	if r.Body == nil {
		return signals
	}
	dec := json.NewDecoder(r.Body)
	_ = dec.Decode(&signals)
	return signals
}

// Signal returns a specific signal value from the request, or the defaultValue if absent.
func Signal(r *http.Request, key string, defaultValue any) any {
	signals := ReadSignals(r)
	if v, ok := signals[key]; ok {
		return v
	}
	return defaultValue
}
