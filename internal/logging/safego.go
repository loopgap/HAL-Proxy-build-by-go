package logging

import (
	"fmt"
	"runtime/debug"
)

// SafeGo starts a background goroutine and protects the process from crashing
// if the goroutine panics. Any captured panic is logged with its stack trace.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Capture detailed stack trace
				stack := string(debug.Stack())

				// Output structured log message with details of the panic and stack trace
				Default().ErrorWithFields("Background goroutine panic caught and recovered", map[string]interface{}{
					"panic": fmt.Sprintf("%v", r),
					"stack": stack,
				})
			}
		}()
		fn()
	}()
}
