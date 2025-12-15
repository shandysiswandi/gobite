package pkgconfig

import "io"

// Config defines a set of methods for retrieving configuration values of various types.
// Implementations of this interface should handle the retrieval and type conversion
// of configuration data, providing default behaviors or error handling as necessary.
type Config interface {
	io.Closer

	// GetInt retrieves the configuration value associated with the given key as an int64.
	// If the key does not exist or the value cannot be converted to an integer,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetInt(key string) int64

	// GetBool retrieves the configuration value associated with the given key as a bool.
	// If the key does not exist or the value cannot be converted to a boolean,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetBool(key string) bool

	// GetFloat retrieves the configuration value associated with the given key as a float64.
	// If the key does not exist or the value cannot be converted to a float,
	// the implementation should handle it accordingly (e.g., return a default value).
	GetFloat(key string) float64

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
