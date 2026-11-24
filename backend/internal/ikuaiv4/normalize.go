package ikuaiv4

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// DecodeAny decodes JSON preserving numeric precision via json.Number.
func DecodeAny(raw json.RawMessage) (any, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var out any
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

var objectWrapperKeys = []string{
	"data", "items", "sysinfo", "config", "info",
	"groups_data", "accounts_data", "static_data", "dir_data", "proto_data", "time_data",
	"backup_info", "auto_backup", "iface_data", "results",
}

// ExtractObject extracts a single JSON object from raw payload (unwrapping common wrappers and 1-element arrays).
func ExtractObject(raw json.RawMessage) map[string]any {
	return ExtractObjectByID(raw, "")
}

// ExtractObjectByID extracts a single JSON object from raw payload, preferring the item matching targetID when given.
func ExtractObjectByID(raw json.RawMessage, targetID string) map[string]any {
	v, err := DecodeAny(raw)
	if err != nil || v == nil {
		return map[string]any{}
	}
	if obj, ok := extractObjectFromAny(v, strings.TrimSpace(targetID), 0); ok {
		return obj
	}
	return map[string]any{}
}

func extractObjectFromAny(v any, targetID string, depth int) (map[string]any, bool) {
	if depth > 6 {
		return nil, false
	}
	switch typed := v.(type) {
	case []any:
		if len(typed) == 0 {
			return map[string]any{}, true
		}
		if targetID != "" {
			for _, item := range typed {
				if m, ok := item.(map[string]any); ok {
					if idVal, has := m["id"]; has && fmt.Sprint(idVal) == targetID {
						return extractObjectFromAny(m, targetID, depth+1)
					}
				}
			}
		}
		return extractObjectFromAny(typed[0], targetID, depth+1)
	case map[string]any:
		if targetID != "" {
			if idVal, has := typed["id"]; has && fmt.Sprint(idVal) == targetID {
				return typed, true
			}
		}
		for _, key := range objectWrapperKeys {
			if inner, exists := typed[key]; exists && inner != nil {
				switch inner.(type) {
				case map[string]any, []any:
					if obj, ok := extractObjectFromAny(inner, targetID, depth+1); ok {
						return obj, true
					}
				}
			}
		}
		for key, inner := range typed {
			if strings.HasSuffix(key, "_data") && inner != nil {
				if arr, ok := inner.([]any); ok && len(arr) > 0 {
					if obj, ok := extractObjectFromAny(arr, targetID, depth+1); ok {
						return obj, true
					}
				}
			}
		}
		return typed, true
	default:
		return nil, false
	}
}

// ExtractCollection normalizes arrays, {data, total} wrappers, and map-of-maps responses
// (such as /interfaces/physical's ether_info) into a slice of row maps + total count.
func ExtractCollection(raw json.RawMessage) ([]map[string]any, int) {
	v, err := DecodeAny(raw)
	if err != nil || v == nil {
		return []map[string]any{}, 0
	}
	return extractCollectionFromAny(v)
}

func extractCollectionFromAny(v any) ([]map[string]any, int) {
	switch typed := v.(type) {
	case []any:
		rows := toMapSlice(typed)
		return rows, len(rows)
	case map[string]any:
		if len(typed) == 0 {
			return []map[string]any{}, 0
		}
		total := Int(typed, "total", "count", "groups_total", "accounts_total", "dir_total", "proto_total", "time_total")

		// 1. Prefer explicit collection keys
		for _, key := range []string{
			"data", "rows", "list", "items", "records", "results",
			"ether_info", "proto3_day",
			"groups_data", "accounts_data", "dir_data", "proto_data", "time_data", "backup_info",
		} {
			if sub, exists := typed[key]; exists {
				if sub == nil {
					return []map[string]any{}, total
				}
				if arr, ok := sub.([]any); ok {
					rows := toMapSlice(arr)
					if total == 0 {
						total = len(rows)
					}
					return rows, total
				}
				if nestedMap, ok := sub.(map[string]any); ok {
					if len(nestedMap) == 0 {
						return []map[string]any{}, total
					}
					if rows, ok := unwrapMapOfMaps(nestedMap); ok {
						if total == 0 {
							total = len(rows)
						}
						return rows, total
					}
				}
			}
		}

		// 2. Single object-array key (e.g. iface_check, wans_stat_history)
		var arrayKeys []string
		var mapKeys []string
		for k, val := range typed {
			switch inner := val.(type) {
			case []any:
				if len(inner) == 0 || arrayHasObjects(inner) {
					arrayKeys = append(arrayKeys, k)
				}
			case map[string]any:
				if isMapOfMaps(inner) {
					mapKeys = append(mapKeys, k)
				}
			}
		}
		if len(arrayKeys) == 1 {
			arr := typed[arrayKeys[0]].([]any)
			rows := toMapSlice(arr)
			if total == 0 {
				total = len(rows)
			}
			return rows, total
		}
		if len(mapKeys) == 1 {
			if rows, ok := unwrapMapOfMaps(typed[mapKeys[0]].(map[string]any)); ok {
				if total == 0 {
					total = len(rows)
				}
				return rows, total
			}
		}

		// 3. Top-level map-of-maps
		if isMapOfMaps(typed) {
			if rows, ok := unwrapMapOfMaps(typed); ok {
				return rows, len(rows)
			}
		}

		return []map[string]any{typed}, 1
	default:
		return []map[string]any{}, 0
	}
}

func toMapSlice(arr []any) []map[string]any {
	rows := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			rows = append(rows, m)
		}
	}
	return rows
}

func arrayHasObjects(arr []any) bool {
	for _, item := range arr {
		if _, ok := item.(map[string]any); ok {
			return true
		}
	}
	return false
}

func isMapOfMaps(m map[string]any) bool {
	if len(m) == 0 {
		return false
	}
	for _, v := range m {
		inner, ok := v.(map[string]any)
		if !ok || len(inner) == 0 {
			return false
		}
	}
	return true
}

func unwrapMapOfMaps(m map[string]any) ([]map[string]any, bool) {
	if !isMapOfMaps(m) {
		return nil, false
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := make([]map[string]any, 0, len(keys))
	for _, k := range keys {
		entry := m[k].(map[string]any)
		copied := make(map[string]any, len(entry)+1)
		for ek, ev := range entry {
			copied[ek] = ev
		}
		if _, exists := copied["name"]; !exists {
			copied["name"] = k
		}
		rows = append(rows, copied)
	}
	return rows, true
}

// Lookup looks up the first matching key (exact first, then normalized case-insensitive).
func Lookup(m map[string]any, keys ...string) (any, bool) {
	if m == nil {
		return nil, false
	}
	for _, k := range keys {
		if val, ok := m[k]; ok && val != nil {
			return val, true
		}
	}
	for _, k := range keys {
		norm := normalizeKey(k)
		for mk, mv := range m {
			if mv != nil && normalizeKey(mk) == norm {
				return mv, true
			}
		}
	}
	return nil, false
}

func normalizeKey(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

// Str returns the string value for the first matching key.
func Str(m map[string]any, keys ...string) string {
	val, ok := Lookup(m, keys...)
	if !ok || val == nil {
		return ""
	}
	switch t := val.(type) {
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(t, 10)
	case int:
		return strconv.Itoa(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

// Int64 returns the int64 value for the first matching key.
func Int64(m map[string]any, keys ...string) int64 {
	val, ok := Lookup(m, keys...)
	if !ok || val == nil {
		return 0
	}
	return toInt64(val)
}

// Int returns the int value for the first matching key.
func Int(m map[string]any, keys ...string) int {
	return int(Int64(m, keys...))
}

// Float64 returns the float64 value for the first matching key (handles "0.50%" too).
func Float64(m map[string]any, keys ...string) float64 {
	val, ok := Lookup(m, keys...)
	if !ok || val == nil {
		return 0
	}
	return ToFloat64(val)
}

// Bool returns a boolean value handling bool, "yes"/"no", "1"/"0", "true"/"false".
func Bool(m map[string]any, keys ...string) bool {
	val, ok := Lookup(m, keys...)
	if !ok || val == nil {
		return false
	}
	switch t := val.(type) {
	case bool:
		return t
	case json.Number:
		n, _ := t.Int64()
		return n != 0
	case float64:
		return t != 0
	case int64:
		return t != 0
	case int:
		return t != 0
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "yes" || s == "true" || s == "1" || s == "on" || s == "enable" || s == "enabled"
	default:
		return false
	}
}

// ToFloat64 converts a JSON scalar (including percentage strings like "39%") into float64.
func ToFloat64(val any) float64 {
	switch t := val.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case int:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		s := strings.TrimSpace(t)
		s = strings.TrimSuffix(s, "%")
		if s == "" || s == "--" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func toInt64(val any) int64 {
	switch t := val.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case json.Number:
		if i, err := t.Int64(); err == nil {
			return i
		}
		f, _ := t.Float64()
		return int64(f)
	case string:
		s := strings.TrimSpace(t)
		if s == "" || s == "--" {
			return 0
		}
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return int64(f)
	default:
		return 0
	}
}
