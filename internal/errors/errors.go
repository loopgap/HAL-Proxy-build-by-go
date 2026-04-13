package errors

import (
	"fmt"
)

// Error codes for HAL-Proxy
const (
	// General errors (1000-1999)
	ErrCodeGeneral      = 1000
	ErrCodeInternal     = 1001
	ErrCodeInvalidInput = 1002
	ErrCodeNotFound     = 1003
	ErrCodeUnauthorized = 1004
	ErrCodeForbidden    = 1005
	ErrCodeConflict     = 1006
	ErrCodeTimeout      = 1007

	// Case errors (2000-2999)
	ErrCodeCaseNotFound      = 2001
	ErrCodeCaseInvalidStatus = 2002
	ErrCodeCaseAlreadyExists = 2003
	ErrCodeCaseNotRunnable   = 2004

	// Approval errors (3000-3999)
	ErrCodeApprovalNotFound = 3001
	ErrCodeApprovalInvalid  = 3002
	ErrCodeApprovalExpired  = 3003

	// Report errors (4000-4999)
	ErrCodeReportNotFound   = 4001
	ErrCodeReportGeneration = 4002

	// Store errors (5000-5999)
	ErrCodeStoreInit      = 5001
	ErrCodeStoreOperation = 5002
	ErrCodeStoreNotFound  = 5003
)

// AppError represents an application error with code and message
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with code and message
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// WrapIf wraps an error only if it's not nil
func WrapIf(code int, message string, err error) error {
	if err == nil {
		return nil
	}
	return Wrap(code, message, err)
}

// General errors
var (
	ErrInternal = New(ErrCodeInternal, "internal server error")
)

// Case errors
func ErrCaseNotFound(id string) *AppError {
	return New(ErrCodeCaseNotFound, fmt.Sprintf("case not found: %s", id))
}

func ErrCaseInvalidStatus(current, expected string) *AppError {
	return New(ErrCodeCaseInvalidStatus, fmt.Sprintf("invalid status transition: current=%s, expected=%s", current, expected))
}

func ErrCaseAlreadyExists(id string) *AppError {
	return New(ErrCodeCaseAlreadyExists, fmt.Sprintf("case already exists: %s", id))
}

func ErrCaseNotRunnable(id string, reason string) *AppError {
	return New(ErrCodeCaseNotRunnable, fmt.Sprintf("case %s not runnable: %s", id, reason))
}

// Approval errors
func ErrApprovalNotFound(id string) *AppError {
	return New(ErrCodeApprovalNotFound, fmt.Sprintf("approval not found: %s", id))
}

func ErrApprovalInvalid(id string, reason string) *AppError {
	return New(ErrCodeApprovalInvalid, fmt.Sprintf("invalid approval %s: %s", id, reason))
}

// Report errors
func ErrReportNotFound(caseID string) *AppError {
	return New(ErrCodeReportNotFound, fmt.Sprintf("report not found for case: %s", caseID))
}

func ErrReportGeneration(err error) *AppError {
	return Wrap(ErrCodeReportGeneration, "failed to generate report", err)
}

// Store errors
func ErrStoreInit(err error) *AppError {
	return Wrap(ErrCodeStoreInit, "failed to initialize store", err)
}

func ErrStoreOperation(op string, err error) *AppError {
	return Wrap(ErrCodeStoreOperation, fmt.Sprintf("store operation failed: %s", op), err)
}

// Is checks if the error matches the given code
func Is(err error, code int) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// As attempts to convert an error to *AppError
func As(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
