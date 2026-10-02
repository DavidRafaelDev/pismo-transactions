package securelog

import (
	"bytes"
	"encoding/json"
)

// sensitiveFields lists JSON field names whose string values are redacted
// in logs. Keep keys exactly as they appear on the wire (snake_case here).
// Add new entries as new sensitive fields enter the API contract.
var sensitiveFields = map[string]bool{
	"document_number": true,
}

// MaskJSONBody takes a raw body (expected JSON) and returns a copy with
// sensitive fields masked. The second return value signals whether the
// input parsed as JSON, so the caller can decide whether to emit the body
// inline as JSON or as a plain string in logs.
//
// If the body is not valid JSON (or is empty), it is returned unchanged
// and isJSON is false. Masking never panics; worst case the body passes
// through unchanged.
func MaskJSONBody(raw []byte) (masked []byte, isJSON bool) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return raw, false
	}
	var v any
	if err := json.Unmarshal(trimmed, &v); err != nil {
		return raw, false
	}
	out, err := json.Marshal(maskWalk(v))
	if err != nil {
		return raw, false
	}
	return out, true
}

// maskWalk recursively traverses a decoded JSON value, redacting any string
// whose containing field name is in sensitiveFields.
func maskWalk(v any) any {
	switch node := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(node))
		for k, val := range node {
			if sensitiveFields[k] {
				if s, ok := val.(string); ok {
					out[k] = MaskDocument(s)
					continue
				}
			}
			out[k] = maskWalk(val)
		}
		return out
	case []any:
		out := make([]any, len(node))
		for i, val := range node {
			out[i] = maskWalk(val)
		}
		return out
	default:
		return v
	}
}
