package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLogLevelParsingAndString(t *testing.T) {
	t.Parallel()

	if ParseLogLevel("debug") != DEBUG || ParseLogLevel("warn") != WARN || ParseLogLevel("error") != ERROR {
		t.Fatal("expected explicit log levels to parse")
	}
	if ParseLogLevel("unknown") != INFO {
		t.Fatal("unknown log levels should default to info")
	}
	if got := LogLevel(99).String(); got != "UNKNOWN" {
		t.Fatalf("unexpected fallback log level string %q", got)
	}
}

func TestLoggerFiltersAndMergesFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(&buf, "info").WithFields(map[string]interface{}{
		"component": "api",
		"shared":    "base",
	})

	logger.Debug("hidden")
	logger.InfoWithFields("served", map[string]interface{}{
		"shared": "request",
		"status": 200,
	})

	if strings.Contains(buf.String(), "hidden") {
		t.Fatalf("debug log should be filtered: %s", buf.String())
	}

	var entry LogEntry
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry.Level != "INFO" || entry.Message != "served" {
		t.Fatalf("unexpected log entry %#v", entry)
	}
	if entry.Fields["component"] != "api" || entry.Fields["shared"] != "request" || entry.Fields["status"].(float64) != 200 {
		t.Fatalf("unexpected merged fields %#v", entry.Fields)
	}
}

func TestLoggerSetLevelRequestIDAndFormattingHelpers(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := New(&buf, "error")
	logger.Warnf("hidden %d", 1)
	logger.SetLevel("debug")
	logger.Debugf("visible %s", "debug")
	logger.WithField("scope", "request").WithRequestID("req-1").Error("failed")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two emitted log lines, got %d: %s", len(lines), buf.String())
	}
	if !strings.Contains(lines[0], `"message":"visible debug"`) {
		t.Fatalf("missing debugf line: %s", lines[0])
	}
	if !strings.Contains(lines[1], `"request_id":"req-1"`) || !strings.Contains(lines[1], `"scope":"request"`) {
		t.Fatalf("missing request fields: %s", lines[1])
	}
}
