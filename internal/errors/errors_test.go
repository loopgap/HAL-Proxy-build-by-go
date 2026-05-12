package errors

import (
	"errors"
	"testing"
)

func TestAppErrorFormattingAndWrapping(t *testing.T) {
	t.Parallel()

	base := errors.New("disk unavailable")
	wrapped := Wrap(ErrCodeStoreOperation, "store_operation_failed", "store failed", base)
	if got := wrapped.Error(); got != "[5002] store failed: disk unavailable" {
		t.Fatalf("wrapped error = %q", got)
	}
	if !errors.Is(wrapped, base) {
		t.Fatal("wrapped error should unwrap to base error")
	}

	plain := New(ErrCodeInvalidInput, "invalid_input", "bad payload")
	if got := plain.Error(); got != "[1002] bad payload" {
		t.Fatalf("plain error = %q", got)
	}
}

func TestConstructorsAndHelpers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  *AppError
		code int
	}{
		{"invalid", ErrInvalidInput("bad"), ErrCodeInvalidInput},
		{"notfound", ErrNotFound("case"), ErrCodeNotFound},
		{"timeout", ErrTimeout("slow"), ErrCodeTimeout},
		{"unauthorized", ErrUnauthorized("token"), ErrCodeUnauthorized},
		{"forbidden", ErrForbidden("role"), ErrCodeForbidden},
		{"conflict", ErrConflict("locked"), ErrCodeConflict},
		{"case-missing", ErrCaseNotFound("case-1"), ErrCodeCaseNotFound},
		{"case-status", ErrCaseInvalidStatus("ready", "running"), ErrCodeCaseInvalidStatus},
		{"case-exists", ErrCaseAlreadyExists("case-1"), ErrCodeCaseAlreadyExists},
		{"case-runnable", ErrCaseNotRunnable("case-1", "paused"), ErrCodeCaseNotRunnable},
		{"approval-missing", ErrApprovalNotFound("approval-1"), ErrCodeApprovalNotFound},
		{"approval-invalid", ErrApprovalInvalid("approval-1", "stale"), ErrCodeApprovalInvalid},
		{"approval-expired", ErrApprovalExpired("approval-1"), ErrCodeApprovalExpired},
		{"report-missing", ErrReportNotFound("report-1"), ErrCodeReportNotFound},
		{"report-content", ErrReportContentMissing("report-1"), ErrCodeReportContentMissing},
	}

	for _, tc := range cases {
		if tc.err.Code != tc.code {
			t.Fatalf("%s code=%d want=%d", tc.name, tc.err.Code, tc.code)
		}
		if !Is(tc.err, tc.code) {
			t.Fatalf("%s should match code %d", tc.name, tc.code)
		}
		if got, ok := As(tc.err); !ok || got != tc.err {
			t.Fatalf("%s should convert to AppError", tc.name)
		}
	}
}

func TestWrapIfAndWrappedConstructors(t *testing.T) {
	t.Parallel()

	if got := WrapIf(ErrCodeInternal, "internal", "ok", nil); got != nil {
		t.Fatalf("WrapIf(nil)=%v, want nil", got)
	}

	source := errors.New("boom")
	wrapped := []error{
		WrapIf(ErrCodeInternal, "internal", "failed", source),
		ErrReportGeneration(source),
		ErrStoreInit(source),
		ErrStoreOperation("insert", source),
	}
	for _, err := range wrapped {
		if err == nil || !errors.Is(err, source) {
			t.Fatalf("wrapped error should contain source: %v", err)
		}
	}

	if Is(source, ErrCodeInternal) {
		t.Fatal("plain error should not match application error codes")
	}
	if _, ok := As(source); ok {
		t.Fatal("plain error should not cast to AppError")
	}
}
