package logging

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"
)

type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestSafeGo_Recovery(t *testing.T) {
	// 1. Capture output of Default Logger to assert structural log formatting using a concurrent-safe buffer
	var safeBuf safeBuffer
	
	Default().mu.Lock()
	originalOutput := Default().output
	Default().output = &safeBuf
	Default().mu.Unlock()

	defer func() {
		Default().mu.Lock()
		Default().output = originalOutput
		Default().mu.Unlock()
	}()

	// 2. Spawn a panic-prone background goroutine wrapped inside SafeGo
	SafeGo(func() {
		// Intentionally trigger a nil pointer dereference panic
		var p *int
		*p = 42
	})

	// 3. Give goroutine a few milliseconds to execute and trigger recover
	time.Sleep(20 * time.Millisecond)

	// 4. Assert that the program did NOT crash (we reached this line) and log was generated
	logOutput := safeBuf.String()
	if !strings.Contains(logOutput, "Background goroutine panic caught and recovered") {
		t.Errorf("Expected panic log, got empty or missing string. Full output: %q", logOutput)
	}

	// 5. Verify the log contains "panic" and "stack" trace details
	if !strings.Contains(logOutput, "invalid memory address") && !strings.Contains(logOutput, "nil pointer dereference") {
		t.Errorf("Expected nil dereference error details in log fields, got: %q", logOutput)
	}
	if !strings.Contains(logOutput, "stack") || !strings.Contains(logOutput, "safego.go") {
		t.Errorf("Expected stack trace containing file source, got: %q", logOutput)
	}
}
