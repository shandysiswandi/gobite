package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// ErrScanValueNotBytes indicates the database value is not a byte slice.
var ErrScanValueNotBytes = errors.New("valueobject: jsonmap scan value is not []byte")

// JSONMap stores arbitrary JSON object data.
// @swaggertype object
type JSONMap map[string]any

// ---------------------------------------------------------------------
// SQL INTERFACES
// ---------------------------------------------------------------------

// Value implements driver.Valuer for JSONMap.
func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements sql.Scanner for JSONMap.
func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}

	var bytes []byte

	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	case json.RawMessage:
		bytes = []byte(v)
	case map[string]any:
		// Some drivers may already decode JSON into a map
		*j = JSONMap(v)
		return nil
	default:
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
