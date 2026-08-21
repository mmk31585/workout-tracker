package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	_ = w.Close()
	os.Stdout = old
	<-done
	_ = r.Close()

	return buf.String()
}

func TestNew_ProductionReturnsJSONHandler(t *testing.T) {
	output := captureStdout(t, func() {
		log := New("production", "info", "")
		log.Info("test message", "key", "value")
	})

	var m map[string]any
	if err := json.Unmarshal([]byte(output), &m); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v (body: %q)", err, output)
	}
	if m["msg"] != "test message" {
		t.Errorf("expected msg 'test message', got %v", m["msg"])
	}
	if m["key"] != "value" {
		t.Errorf("expected key 'value', got %v", m["key"])
	}
	if _, ok := m["source"]; !ok {
		t.Error("expected source field in production logger output")
	}
}

func TestNew_DevelopmentReturnsTextHandler(t *testing.T) {
	output := captureStdout(t, func() {
		log := New("development", "debug", "")
		log.Debug("debug message", "key", "value")
	})

	if !strings.Contains(output, "debug message") {
		t.Errorf("expected text output to contain 'debug message', got %q", output)
	}
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Errorf("expected text format output, got JSON: %q", output)
	}
}

func TestNew_UnknownEnvFallsBackToTextHandler(t *testing.T) {
	log := New("staging", "info", "")
	if log == nil {
		t.Fatal("expected non-nil logger for unknown env")
	}
	log.Info("info message", "key", "value")
}

func TestNew_EmptyEnvFallsBackToTextHandler(t *testing.T) {
	log := New("", "info", "")
	if log == nil {
		t.Fatal("expected non-nil logger for empty env")
	}
	log.Info("info message")
}

func TestNew_ReturnsNonNilLogger(t *testing.T) {
	for _, env := range []string{"production", "development", "test", "staging", ""} {
		log := New(env, "info", "")
		if log == nil {
			t.Errorf("expected non-nil logger for env %q", env)
		}
	}
}

func TestNew_LoggerCanLogAllLevels(t *testing.T) {
	output := captureStdout(t, func() {
		log := New("development", "debug", "json")
		log.Debug("debug message")
		log.Info("info message", "request_id", "abc-123")
		log.Warn("warn message", "shortcode", "abc123de")
		log.Error("error message", "code", "NOT_FOUND")
	})

	for _, msg := range []string{"debug message", "info message", "warn message", "error message"} {
		if !strings.Contains(output, msg) {
			t.Errorf("expected output to contain %q", msg)
		}
	}
}

func TestNew_ProductionLoggerDoesNotOutputDebug(t *testing.T) {
	output := captureStdout(t, func() {
		log := New("production", "info", "")
		log.Debug("this should not appear in production")
		log.Info("this should appear in production")
	})

	if strings.Contains(output, "this should not appear") {
		t.Error("debug message should not appear in production output")
	}
	if !strings.Contains(output, "this should appear") {
		t.Error("info message should appear in production output")
	}
}

func TestNew_CanCreateChildLogger(t *testing.T) {
	log := New("development", "debug", "")
	child := log.With("request_id", "550e8400-e29b-41d4-a716-446655440000")
	child.Info("request processed")
}

func TestNew_HasAddSourceEnabledProduction(t *testing.T) {
	output := captureStdout(t, func() {
		log := New("production", "info", "")
		log.Info("source test", "key", "value")
	})

	var m map[string]any
	if err := json.Unmarshal([]byte(output), &m); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v (body: %q)", err, output)
	}
	if _, ok := m["source"]; !ok {
		t.Error("expected source field in production logger output")
	}
}

func TestNew_HasAddSourceEnabledDevelopment(t *testing.T) {
	output := captureStdout(t, func() {
		log := New("development", "debug", "json")
		log.Info("source test", "key", "value")
	})

	var m map[string]any
	if err := json.Unmarshal([]byte(output), &m); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v (body: %q)", err, output)
	}
	if _, ok := m["source"]; !ok {
		t.Error("expected source field in development logger output")
	}
}
