package model

import (
	"strings"

	"github.com/shopspring/decimal"
)

// we can reuse these utility functions in both model_catalog and provider_model
func extractDefaultParameters(value any) map[string]*decimal.Decimal {
	result := map[string]*decimal.Decimal{}
	params, ok := value.(map[string]any)
	if !ok {
		return result
	}
	for key, raw := range params {
		if raw == nil {
			result[key] = nil
			continue
		}
		switch v := raw.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				result[key] = nil
				continue
			}
			if d, err := decimal.NewFromString(v); err == nil {
				val := d
				result[key] = &val
			}
		case float64:
			d := decimal.NewFromFloat(v)
			result[key] = &d
		case float32:
			d := decimal.NewFromFloat32(v)
			result[key] = &d
		default:
			// ignore unsupported types
		}
	}
	return result
}

func extractStringSlice(value any) []string {
	list := []string{}
	switch arr := value.(type) {
	case []any:
		for _, item := range arr {
			if str, ok := item.(string); ok {
				list = append(list, strings.TrimSpace(str))
			}
		}
	case []string:
		for _, item := range arr {
			list = append(list, strings.TrimSpace(item))
		}
	}
	return list
}

func extractStringSliceFromMap(raw map[string]any, path ...string) []string {
	current := any(raw)
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = m[key]
	}
	return extractStringSlice(current)
}

func getString(raw map[string]any, key string) (string, bool) {
	if raw == nil {
		return "", false
	}
	if value, ok := raw[key]; ok {
		if str, ok := value.(string); ok {
			return strings.TrimSpace(str), true
		}
	}
	return "", false
}

func copyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	dest := make(map[string]any, len(source))
	for k, v := range source {
		dest[k] = v
	}
	return dest
}

func floatFromAny(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if strings.TrimSpace(v) == "" {
			return 0, false
		}
		if parsed, err := decimal.NewFromString(v); err == nil {
			return parsed.InexactFloat64(), true
		}
	}
	return 0, false
}

func containsString(list []string, target string) bool {
	target = strings.ToLower(target)
	for _, item := range list {
		if strings.ToLower(item) == target {
			return true
		}
	}
	return false
}

func normalizeURL(baseURL string) string {
	normalized := strings.TrimSpace(baseURL)
	normalized = strings.TrimRight(normalized, "/")
	return normalized
}
