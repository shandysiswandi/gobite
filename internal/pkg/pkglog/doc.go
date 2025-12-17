// Package pkglog contains logging helpers used across the application.
//
// It is built around slog and keeps logs consistent by:
//   - Initializing a JSON handler with stable keys.
//   - Attaching request correlation IDs (when present) to each log record.
package pkglog
