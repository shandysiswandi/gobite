// Package uid provides helpers for generating unique identifiers.
//
// The codebase uses these interfaces to avoid hard-coding a specific UID
// strategy. Depending on the use case you can generate:
//   - String IDs (for example UUIDs).
//   - Numeric IDs (for example Snowflake-style IDs).
package uid
