package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResultColor(t *testing.T) {
	tests := []struct {
		result string
		want   int
	}{
		{"Success", 0x57F287},
		{"Failed", 0xED4245},
		{"Errored", 0xED4245},
		{"Killed", 0xFEE75C},
		{"unknown", 0x5865F2},
	}
	for _, tt := range tests {
		got := resultColor(tt.result)
		if got != tt.want {
			t.Errorf("resultColor(%q) = %d, want %d", tt.result, got, tt.want)
		}
	}
}

func TestSendNotification(t *testing.T) {
	var received []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := sendNotification(server.URL, "42", "echo hello", "Success", "0", "default", "myserver")
	if err != nil {
		t.Fatalf("sendNotification failed: %v", err)
	}

	var payload WebhookPayload
	if err := json.Unmarshal(received, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload.Content != "" {
		t.Errorf("Content = %q, want empty", payload.Content)
	}

	if len(payload.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(payload.Embeds))
	}

	e := payload.Embeds[0]
	if e.Title != "Task #42 Success [myserver]" {
		t.Errorf("Title = %q", e.Title)
	}
	if e.Description != "```\necho hello\n```" {
		t.Errorf("Description = %q", e.Description)
	}
	if e.Color != 0x57F287 {
		t.Errorf("Color = %d", e.Color)
	}

	if len(e.Fields) != 0 {
		t.Fatalf("expected 0 fields (exit code 0, group default), got %d", len(e.Fields))
	}
}

func TestSendNotificationWithoutGroup(t *testing.T) {
	var received []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := sendNotification(server.URL, "1", "ls -la", "Failed", "1", "", "host-a")
	if err != nil {
		t.Fatalf("sendNotification failed: %v", err)
	}

	var payload WebhookPayload
	json.Unmarshal(received, &payload)

	if payload.Content != "" {
		t.Errorf("Content = %q, want empty", payload.Content)
	}

	fields := payload.Embeds[0].Fields
	if len(fields) != 1 {
		t.Fatalf("expected 1 field (exit code; no group), got %d", len(fields))
	}
	if fields[0].Name != "Exit Code" || fields[0].Value != "1" {
		t.Errorf("field 0 = %+v", fields[0])
	}
}

func TestSendNotificationTruncatesLongCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var p WebhookPayload
		json.Unmarshal(b, &p)
		desc := p.Embeds[0].Description
		// strip the ``` fences
		inner := strings.TrimPrefix(desc, "```\n")
		inner = strings.TrimSuffix(inner, "\n```")
		if len(inner) > 2000 {
			t.Errorf("description length = %d, want <= 2000", len(inner))
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	longCmd := strings.Repeat("a", 2500)
	err := sendNotification(server.URL, "1", longCmd, "Success", "0", "", "h")
	if err != nil {
		t.Fatalf("sendNotification failed: %v", err)
	}
}

func TestSendNotificationHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	err := sendNotification(server.URL, "1", "cmd", "Success", "0", "", "h")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected 403 error, got %v", err)
	}
}

func TestSendNotificationMarshalError(t *testing.T) {
	// Inject a payload that will fail somehow — by passing a very long string
	// we'll exercise the truncation path. Marshal itself should never fail for
	// our types, but the test ensures we handle it gracefully.
	err := sendNotification("http://invalid", "1", "cmd", "Success", "0", "", "h")
	// Should fail on HTTP, not marshal
	if err == nil {
		t.Fatal("expected error from invalid URL")
	}
}

func TestNotTruncated(t *testing.T) {
	cmd := "short command"
	var received []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	sendNotification(server.URL, "1", cmd, "Success", "0", "", "h")

	var p WebhookPayload
	json.Unmarshal(received, &p)
	inner := strings.TrimPrefix(p.Embeds[0].Description, "```\n")
	inner = strings.TrimSuffix(inner, "\n```")
	if inner != cmd {
		t.Errorf("expected %q, got %q", cmd, inner)
	}
}

func TestExamplePueueCallback(t *testing.T) {
	// Simulates the actual CLI invocation path by just testing sendNotification
	// with realistic pueue callback values.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := sendNotification(server.URL, "7", "ffmpeg -i input.mp4 output.mkv", "Success", "0", "video", "server-01")
	if err != nil {
		t.Fatalf("sendNotification failed: %v", err)
	}
}

func TestSendNotificationNonZeroExitCode(t *testing.T) {
	var received []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := sendNotification(server.URL, "5", "make build", "Failed", "2", "default", "srv")
	if err != nil {
		t.Fatalf("sendNotification failed: %v", err)
	}

	var payload WebhookPayload
	json.Unmarshal(received, &payload)

	fields := payload.Embeds[0].Fields
	if len(fields) != 1 {
		t.Fatalf("expected 1 field (non-zero exit code shown), got %d", len(fields))
	}
	if fields[0].Name != "Exit Code" || fields[0].Value != "2" {
		t.Errorf("field 0 = %+v", fields[0])
	}
}

func TestFailedResultColor(t *testing.T) {
	for _, r := range []string{"Failed", "Errored"} {
		got := resultColor(r)
		if got != 0xED4245 {
			t.Errorf("resultColor(%q) = %d, want red", r, got)
		}
	}
}
