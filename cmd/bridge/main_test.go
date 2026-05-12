package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDBPathPrecedence(t *testing.T) {
	t.Parallel()

	if got := dbPath(mapEnv(nil)); got != "bridgeos.db" {
		t.Fatalf("default db path = %q, want bridgeos.db", got)
	}
	if got := dbPath(mapEnv(map[string]string{"HAL_PROXY_DB": "legacy.db"})); got != "legacy.db" {
		t.Fatalf("legacy db path = %q, want legacy.db", got)
	}
	if got := dbPath(mapEnv(map[string]string{"HAL_PROXY_DB": "legacy.db", "BRIDGEOS_DB": "preferred.db"})); got != "preferred.db" {
		t.Fatalf("preferred db path = %q, want preferred.db", got)
	}
}

func TestActorValidation(t *testing.T) {
	t.Parallel()

	valid := []string{"cli", "owner_1", "service-agent", strings.Repeat("a", 64)}
	for _, actor := range valid {
		if !isValidActor(actor) {
			t.Fatalf("actor %q should be valid", actor)
		}
	}

	invalid := []string{"", "bad actor", "bad/slash", strings.Repeat("a", 65)}
	for _, actor := range invalid {
		if isValidActor(actor) {
			t.Fatalf("actor %q should be invalid", actor)
		}
	}
}

func TestRunVersionAndUsageErrors(t *testing.T) {
	t.Parallel()

	dbFile := filepath.Join(t.TempDir(), "bridgeos.db")

	code, stdout, stderr := runBridge(t, dbFile, "version")
	if code != 0 {
		t.Fatalf("version exit code = %d, stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"name"`) || !strings.Contains(stdout, `"version"`) {
		t.Fatalf("version output missing fields: %s", stdout)
	}

	code, _, stderr = runBridge(t, dbFile)
	if code != 1 || !strings.Contains(stderr, "usage: bridge") {
		t.Fatalf("empty args code=%d stderr=%q", code, stderr)
	}

	code, _, stderr = runBridge(t, dbFile, "unknown")
	if code != 1 || !strings.Contains(stderr, `unknown command "unknown"`) {
		t.Fatalf("unknown command code=%d stderr=%q", code, stderr)
	}
}

func TestRunCaseReportDeviceAndSessionFlow(t *testing.T) {
	t.Parallel()
	cleanupArtifactsIfCreated(t)

	dir := t.TempDir()
	dbFile := filepath.Join(dir, "bridgeos.db")
	specPath := filepath.Join(dir, "case.json")
	spec := `{"title":"smoke","commands":[{"name":"read","action":"read_mem","risk_class":"observe"}]}`
	if err := os.WriteFile(specPath, []byte(spec), 0o600); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runBridge(t, dbFile, "case", "new", "--spec", specPath, "--actor", "owner-1")
	if code != 0 {
		t.Fatalf("case new exit code=%d stderr=%s", code, stderr)
	}
	caseID := decodeStringField(t, stdout, "id")
	if caseID == "" {
		t.Fatalf("case new output missing id: %s", stdout)
	}

	code, stdout, stderr = runBridge(t, dbFile, "case", "show", "--id", caseID, "--actor", "owner-1")
	if code != 0 || !strings.Contains(stdout, caseID) {
		t.Fatalf("case show code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}

	code, stdout, stderr = runBridge(t, dbFile, "case", "run", "--id", caseID, "--actor", "owner-1")
	if code != 0 || !strings.Contains(stdout, `"status": "completed"`) {
		t.Fatalf("case run code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}

	code, stdout, stderr = runBridge(t, dbFile, "case", "events", "--id", caseID)
	if code != 0 || !strings.Contains(stdout, `"case_id": "`+caseID+`"`) {
		t.Fatalf("case events code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}

	code, stdout, stderr = runBridge(t, dbFile, "report", "build", "--id", caseID)
	if code != 0 || !strings.Contains(stdout, `"case_id": "`+caseID+`"`) {
		t.Fatalf("report build code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}

	code, stdout, stderr = runBridge(t, dbFile, "device", "ls")
	if code != 0 || !strings.HasPrefix(strings.TrimSpace(stdout), "[") {
		t.Fatalf("device ls code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}

	code, stdout, stderr = runBridge(t, dbFile, "session", "ls")
	if code != 0 || !strings.HasPrefix(strings.TrimSpace(stdout), "[") {
		t.Fatalf("session ls code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestRunApprovalFlowAndFailurePaths(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbFile := filepath.Join(dir, "bridgeos.db")
	specPath := filepath.Join(dir, "approval-case.json")
	spec := `{"title":"approval","commands":[{"name":"reset","action":"reset","risk_class":"destructive"}]}`
	if err := os.WriteFile(specPath, []byte(spec), 0o600); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runBridge(t, dbFile, "case", "new", "--spec", specPath, "--actor", "owner-2")
	if code != 0 {
		t.Fatalf("case new exit code=%d stderr=%s", code, stderr)
	}
	caseID := decodeStringField(t, stdout, "id")

	code, stdout, stderr = runBridge(t, dbFile, "case", "run", "--id", caseID, "--actor", "owner-2")
	if code != 0 || !strings.Contains(stdout, `"status": "awaiting_approval"`) {
		t.Fatalf("case run awaiting approval code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}

	code, stdout, stderr = runBridge(t, dbFile, "approval", "ls", "--case-id", caseID)
	if code != 0 {
		t.Fatalf("approval ls code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	approvalID := decodeApprovalID(t, stdout)

	code, stdout, stderr = runBridge(t, dbFile, "approval", "approve", "--id", approvalID, "--actor", "approver-1", "--reason", "ok")
	if code != 0 || !strings.Contains(stdout, `"status": "approved"`) {
		t.Fatalf("approval approve code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}

	code, _, stderr = runBridge(t, dbFile, "case", "new", "--spec", specPath, "--actor", "bad actor")
	if code != 1 || !strings.Contains(stderr, "invalid actor name") {
		t.Fatalf("invalid actor code=%d stderr=%q", code, stderr)
	}

	code, _, stderr = runBridge(t, dbFile, "report", "build")
	if code != 1 || !strings.Contains(stderr, "--id is required") {
		t.Fatalf("report missing id code=%d stderr=%q", code, stderr)
	}

	code, _, stderr = runBridge(t, dbFile, "approval", "reject")
	if code != 1 || !strings.Contains(stderr, "--id is required") {
		t.Fatalf("reject missing id code=%d stderr=%q", code, stderr)
	}
}

func runBridge(t *testing.T, dbFile string, args ...string) (int, string, string) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(args, mapEnv(map[string]string{"BRIDGEOS_DB": dbFile}), &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func mapEnv(values map[string]string) func(string) string {
	return func(key string) string {
		if values == nil {
			return ""
		}
		return values[key]
	}
}

func decodeStringField(t *testing.T, raw, field string) string {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode payload: %v\n%s", err, raw)
	}
	value, _ := payload[field].(string)
	return value
}

func decodeApprovalID(t *testing.T, raw string) string {
	t.Helper()

	var approvals []map[string]any
	if err := json.Unmarshal([]byte(raw), &approvals); err != nil {
		t.Fatalf("decode approvals: %v\n%s", err, raw)
	}
	if len(approvals) == 0 {
		t.Fatalf("expected approval output, got %s", raw)
	}
	id, _ := approvals[0]["id"].(string)
	if id == "" {
		t.Fatalf("approval output missing id: %s", raw)
	}
	return id
}

func cleanupArtifactsIfCreated(t *testing.T) {
	t.Helper()

	if _, err := os.Stat("artifacts"); os.IsNotExist(err) {
		t.Cleanup(func() {
			_ = os.RemoveAll("artifacts")
		})
	}
}
