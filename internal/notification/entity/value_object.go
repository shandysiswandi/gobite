package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

var ErrScanValueNotBytes = errors.New("notification: jsonmap scan value is not []byte")

type JSONMap map[string]any

// ---------------------------------------------------------------------
// SQL INTERFACES
// ---------------------------------------------------------------------

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return ErrScanValueNotBytes
	}

	var result JSONMap
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// ---------------------------------------------------------------------
// SETTERS
// ---------------------------------------------------------------------

// Set adds or updates a key-value pair.
func (j JSONMap) Set(key string, value any) {
	j[key] = value
}

// SetIfAbsent only sets the value if the key does not exist.
func (j JSONMap) SetIfAbsent(key string, value any) {
	if _, exists := j[key]; !exists {
		j[key] = value
	}
}

// ---------------------------------------------------------------------
// GETTERS (Type-Safe Helpers)
// ---------------------------------------------------------------------
// Get returns the raw value or nil.
func (j JSONMap) Get(key string) any {
	return j[key]
}

// Has checks if a key exists.
func (j JSONMap) Has(key string) bool {
	_, ok := j[key]
	return ok
}

// GetString safely returns a string value. Returns "" if missing or wrong type.
func (j JSONMap) GetString(key string) string {
	if v, ok := j[key].(string); ok {
		return v
	}
	return ""
}

// GetInt safely returns an int. Returns 0 if missing or wrong type.
// Note: JSON numbers often unmarshal as float64, so we handle that too.
func (j JSONMap) GetInt(key string) int {
	if v, ok := j[key].(int); ok {
		return v
	}
	if v, ok := j[key].(float64); ok {
		return int(v)
	}
	return 0
}

// GetInt64 safely returns an int64.
func (j JSONMap) GetInt64(key string) int64 {
	if v, ok := j[key].(int64); ok {
		return v
	}
	if v, ok := j[key].(float64); ok {
		return int64(v)
	}
	return 0
}

// GetBool safely returns a boolean. Returns false if missing or wrong type.
func (j JSONMap) GetBool(key string) bool {
	if v, ok := j[key].(bool); ok {
		return v
	}
	return false
}
