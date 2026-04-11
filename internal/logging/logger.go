package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger is a structured logger
type Logger struct {
	mu     sync.Mutex
	output io.Writer
	level  LogLevel
	fields map[string]interface{}
}

// New creates a new Logger
func New(output io.Writer, level string) *Logger {
	return &Logger{
		output: output,
		level:  ParseLogLevel(level),
		fields: make(map[string]interface{}),
	}
}

// Default returns a default logger that writes to stdout
func Default() *Logger {
	return New(os.Stdout, "info")
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &Logger{
		output: l.output,
		level:  l.level,
		fields: newFields,
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		output: l.output,
		level:  l.level,
		fields: newFields,
	}
}

// log writes a log entry
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   msg,
		Fields:    fields,
	}

	// Merge logger fields with entry fields
	if len(l.fields) > 0 {
		if entry.Fields == nil {
			entry.Fields = make(map[string]interface{})
		}
		for k, v := range l.fields {
			if _, exists := entry.Fields[k]; !exists {
				entry.Fields[k] = v
			}
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, "failed to marshal log entry: %v\n", err)
		return
	}

	fmt.Fprintln(l.output, string(data))
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg, nil)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...), nil)
}

// DebugWithFields logs a debug message with fields
func (l *Logger) DebugWithFields(msg string, fields map[string]interface{}) {
	l.log(DEBUG, msg, fields)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(INFO, msg, nil)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...), nil)
}

// InfoWithFields logs an info message with fields
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	l.log(INFO, msg, fields)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(WARN, msg, nil)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...), nil)
}

// WarnWithFields logs a warning message with fields
func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	l.log(WARN, msg, fields)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(ERROR, msg, nil)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...), nil)
}

// ErrorWithFields logs an error message with fields
func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	l.log(ERROR, msg, fields)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = ParseLogLevel(level)
}

// RequestLogger creates a logger with request-specific fields
type RequestLogger struct {
	*Logger
	RequestID string
}

// WithRequestID creates a request logger with the given request ID
func (l *Logger) WithRequestID(requestID string) *RequestLogger {
	return &RequestLogger{
		Logger:    l.WithField("request_id", requestID),
		RequestID: requestID,
	}
}
