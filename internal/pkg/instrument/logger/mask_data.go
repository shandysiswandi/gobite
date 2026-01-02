package logger

import (
	"encoding/json"
	"strings"
)

func maskAny(val any, maskKeys map[string]struct{}) (any, bool) {
	switch v := val.(type) {
	case map[string]any:
		return maskData(v, maskKeys), true
	case map[string]string:
		converted := make(map[string]any, len(v))
		for k, v2 := range v {
			converted[k] = v2
		}
		return maskData(converted, maskKeys), true
	case []any:
		return maskData(v, maskKeys), true
	default:
		return nil, false
	}
}

func maskJSONString(payload string, maskKeys map[string]struct{}) (string, bool) {
	payload = strings.TrimLeft(payload, " \t\r\n")
	if payload == "" || (payload[0] != '{' && payload[0] != '[') {
		return "", false
	}
	var jsonBody any
	if err := json.Unmarshal([]byte(payload), &jsonBody); err != nil {
		return "", false
	}
	masked := maskData(jsonBody, maskKeys)
	if maskedBytes, err := json.Marshal(masked); err == nil {
		return string(maskedBytes), true
	}
	return "", false
}

func maskJSONBytes(payload []byte, maskKeys map[string]struct{}) (string, bool) {
	if len(payload) == 0 {
		return "", false
	}
	var jsonBody any
	if err := json.Unmarshal(payload, &jsonBody); err != nil {
		return "", false
	}
	masked := maskData(jsonBody, maskKeys)
	if maskedBytes, err := json.Marshal(masked); err == nil {
		return string(maskedBytes), true
	}
	return "", false
}

func maskData(v any, maskKeys map[string]struct{}) any {
	switch val := v.(type) {
	case map[string]any:
		masked := make(map[string]any, len(val))
		for k, v2 := range val {
			if _, found := maskKeys[strings.ToLower(k)]; found {
				masked[k] = "***"
			} else {
				masked[k] = maskData(v2, maskKeys)
			}
		}
		return masked
	case []any:
		res := make([]any, len(val))
		for i, v2 := range val {
			res[i] = maskData(v2, maskKeys)
		}
		return res
	default:
		return v
	}
}
