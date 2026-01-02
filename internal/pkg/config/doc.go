// Package config provides a small abstraction for reading configuration values.
//
// The application expects config values to come from a concrete implementation
// (for example Viper). Business code should depend on the Config interface so it
// stays easy to test and does not care where values come from (file, env, etc).
//
// This package focuses on convenience getters for common types and simple
// decoding rules (for example base64 for binary values).
package config
