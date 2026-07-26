// Package expression evaluates ${...} expressions against a context map.
// If a string is a single expression, the resolved value preserves its Go type.
// Mixed literal+expression strings are interpolated to a string result.
package expression

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var exprPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// Evaluate resolves all ${...} expressions in s against ctx.
// A string that is exactly one expression (e.g. "${count}") returns the
// native Go value. Mixed strings return a string with interpolated values.
func Evaluate(s string, ctx map[string]any) (any, error) {
	matches := exprPattern.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return s, nil
	}

	// Single expression spanning the entire string → preserve type.
	if len(matches) == 1 && matches[0][0] == 0 && matches[0][1] == len(s) {
		path := s[matches[0][2]:matches[0][3]]
		return resolvePath(strings.TrimSpace(path), ctx)
	}

	// Mixed: interpolate everything to string.
	var b strings.Builder
	prev := 0
	for _, loc := range matches {
		b.WriteString(s[prev:loc[0]])
		path := s[loc[2]:loc[3]]
		val, err := resolvePath(strings.TrimSpace(path), ctx)
		if err != nil {
			return nil, err
		}
		b.WriteString(fmt.Sprintf("%v", val))
		prev = loc[1]
	}
	b.WriteString(s[prev:])
	return b.String(), nil
}

// ContainsExpression reports whether s includes any ${...} pattern.
func ContainsExpression(s string) bool {
	return exprPattern.MatchString(s)
}

// ExtractPaths returns every dot-path referenced in ${...} expressions.
func ExtractPaths(s string) []string {
	matches := exprPattern.FindAllStringSubmatch(s, -1)
	paths := make([]string, 0, len(matches))
	for _, m := range matches {
		paths = append(paths, strings.TrimSpace(m[1]))
	}
	return paths
}

// resolvePath navigates a dot-separated path through nested maps/slices.
func resolvePath(path string, root map[string]any) (any, error) {
	parts := strings.Split(path, ".")
	var current any = root

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("expression: path %q not found (missing key %q at segment %d)", path, part, i)
			}
			current = val
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("expression: path %q: segment %q is not a valid index for slice at segment %d", path, part, i)
			}
			if idx < 0 || idx >= len(v) {
				return nil, fmt.Errorf("expression: path %q: index %d out of range (len %d) at segment %d", path, idx, len(v), i)
			}
			current = v[idx]
		default:
			// Try reflection for other slice/map types.
			rv := reflect.ValueOf(current)
			if rv.Kind() == reflect.Map {
				key := reflect.ValueOf(part)
				val := rv.MapIndex(key)
				if !val.IsValid() {
					return nil, fmt.Errorf("expression: path %q not found (missing key %q at segment %d)", path, part, i)
				}
				current = val.Interface()
			} else if rv.Kind() == reflect.Slice {
				idx, err := strconv.Atoi(part)
				if err != nil {
					return nil, fmt.Errorf("expression: path %q: segment %q is not a valid index at segment %d", path, part, i)
				}
				if idx < 0 || idx >= rv.Len() {
					return nil, fmt.Errorf("expression: path %q: index %d out of range (len %d) at segment %d", path, idx, rv.Len(), i)
				}
				current = rv.Index(idx).Interface()
			} else {
				return nil, fmt.Errorf("expression: path %q: cannot traverse %T at segment %d (%q)", path, current, i, part)
			}
		}
	}
	return current, nil
}
