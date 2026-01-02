// Package goroutine contains helpers for running goroutines safely.
//
// The Manager type limits concurrency, collects returned errors, and logs
// panics so that background work does not crash the process silently.
package goroutine
