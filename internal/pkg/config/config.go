package config

import (
	"io"
	"time"
)

// TimeConfig defines helpers for retrieving time-based configuration values.
type TimeConfig interface {
	// GetSecond retrieves the configuration value associated with the given key as seconds.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetSecond(key string) time.Duration

	// GetMinute retrieves the configuration value associated with the given key as minutes.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetMinute(key string) time.Duration

	// GetHour retrieves the configuration value associated with the given key as hours.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetHour(key string) time.Duration

	// GetDay retrieves the configuration value associated with the given key as days (24h).
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetDay(key string) time.Duration
}

// SignedIntConfig defines helpers for retrieving signed integer configuration values.
type SignedIntConfig interface {
	// GetInt retrieves the configuration value associated with the given key as an int.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetInt(key string) int

	// GetInt32 retrieves the configuration value associated with the given key as an int32.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetInt32(key string) int32

	// GetInt64 retrieves the configuration value associated with the given key as an int64.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetInt64(key string) int64
}

// UnsignedIntConfig defines helpers for retrieving unsigned integer configuration values.
type UnsignedIntConfig interface {
	// GetUint retrieves the configuration value associated with the given key as a uint.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetUint(key string) uint

	// GetUint16 retrieves the configuration value associated with the given key as a uint16.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetUint16(key string) uint16

	// GetUint32 retrieves the configuration value associated with the given key as a uint32.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetUint32(key string) uint32

	// GetUint64 retrieves the configuration value associated with the given key as a uint64.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetUint64(key string) uint64
}

// FloatConfig defines helpers for retrieving floating-point configuration values.
type FloatConfig interface {
	// GetFloat32 retrieves the configuration value associated with the given key as a float32.
	// If the key does not exist or the value cannot be converted to a float,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetFloat32(key string) float32

	// GetFloat64 retrieves the configuration value associated with the given key as a float64.
	// If the key does not exist or the value cannot be converted to a float,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetFloat64(key string) float64
}

// Config defines a set of methods for retrieving configuration values of various types.
// Implementations of this interface should handle the retrieval and type conversion
// of configuration data, providing default behaviors or error handling as necessary.
type Config interface {
	io.Closer
	TimeConfig
	SignedIntConfig
	UnsignedIntConfig
	FloatConfig

	// GetBool retrieves the configuration value associated with the given key as a bool.
	// If the key does not exist or the value cannot be converted to a boolean,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetBool(key string) bool

	// GetString retrieves the configuration value associated with the given key as a string.
	// If the key does not exist, the implementation should handle it accordingly.
	GetString(key string) string

	// GetBinary retrieves the configuration value associated with the given key as a byte slice.
	// If the key does not exist or the value cannot be converted to binary,
	// the implementation should handle it accordingly (e.g., return a default value).
	// Configuration value is stored as base64 encoded.
	GetBinary(key string) []byte

	// GetArray retrieves the configuration value associated with the given key as a slice of strings.
	// If the key does not exist or the value cannot be converted to a string slice,
	// the implementation should handle it accordingly (e.g., return a default value).
	// Configuration value is stored with format <element1>,<element2>,...
	GetArray(key string) []string

	// GetMap retrieves the configuration value associated with the given key as a map of strings to strings.
	// If the key does not exist or the value cannot be converted to a map,
	// the implementation should handle it accordingly (e.g., return a default value).
	// Configuration value is stored with format <key1>:<value1>,<key2>:<value2>,...
	GetMap(key string) map[string]string
}
