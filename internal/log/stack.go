package log

import (
	"runtime/debug"
)

var (
	// StackTraceFieldName is the field name used for stack trace.
	StackTraceFieldName = "stacktrace"

	// StackTraceFunc is the function used for stack trace.
	StackTraceFunc func(skip int) string = func(skip int) string {
		return string(debug.Stack())
	}
)

// Stack returns [StackTraceFieldName] and stacktrace obtained by calling [StackTraceFunc].
func Stack(skip int) (string, string) {
	return StackTraceFieldName, StackTraceFunc(skip + 1)
}
