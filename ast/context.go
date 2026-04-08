package ast

import (
	"strconv"
	"strings"
)

func GetContextPathFromRaw(raw string) (string, []ValueAccessor) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "{{")
	raw = strings.TrimSuffix(raw, "}}")
	raw = strings.TrimSpace(raw)
	key, end, ok := scanContextPath(raw)
	if !ok {
		return "", nil
	}

	accessors := []ValueAccessor{}
	offset := end
	for {
		offset = skipAccessorSpaces(raw, offset)
		if offset >= len(raw) {
			break
		}

		switch raw[offset] {
		case '.':
			offset++
			offset = skipAccessorSpaces(raw, offset)
			keyEnd, ok := scanAccessorBase(raw, offset)
			if !ok {
				return key, accessors
			}
			accessors = append(accessors, ValueAccessor{
				Kind: "key",
				Key:  strings.TrimSpace(raw[offset:keyEnd]),
			})
			offset = keyEnd
		case '[':
			indexEnd := strings.Index(raw[offset:], "]")
			if indexEnd == -1 {
				return key, accessors
			}
			index, err := strconv.Atoi(strings.TrimSpace(raw[offset+1 : offset+indexEnd]))
			if err != nil {
				return key, accessors
			}
			accessors = append(accessors, ValueAccessor{
				Kind:  "index",
				Index: index,
			})
			offset = offset + indexEnd + 1
		default:
			return key, accessors
		}
	}

	return key, accessors
}

func GetContextKey(node Node) string {
	keyNode := FindNodeByRuleName(node.Children(), "context-key")
	if keyNode != nil {
		return strings.TrimSpace(string(keyNode.Raw()))
	}

	key, _ := GetContextPathFromRaw(strings.TrimSpace(string(node.Raw())))
	return key
}

func GetContextAccessors(node Node) []ValueAccessor {
	_, accessors := GetContextPathFromRaw(strings.TrimSpace(string(node.Raw())))
	return accessors
}

func ResolveContextValue(root Node, key string) ResolvedValue {
	contextWrapper := FindNodeByRuleName(root.Children(), "context-wrapper")
	if contextWrapper == nil {
		return ResolvedValue{MissingContextKey: key}
	}

	entry := FindNode(contextWrapper.Children(), func(node Node) bool {
		return IsRuleName(node, "context-entry") && GetContextKey(node) == key
	})
	if entry == nil {
		return ResolvedValue{MissingContextKey: key}
	}

	valueType := FindNodeByRuleName(entry.Children(), "context-value-type")
	if valueType == nil {
		return ResolvedValue{MissingContextKey: key}
	}

	return resolveValueNode(root, valueType, nil, nil)
}

func ResolveContextReferenceValue(root Node, node Node) ResolvedValue {
	key := GetContextKey(node)
	if key == "" {
		return ResolvedValue{}
	}

	resolved := ResolveContextValue(root, key)
	if resolved.MissingContextKey != "" {
		return resolved
	}

	return ApplyAccessors(resolved, GetContextAccessors(node))
}

func scanContextPath(raw string) (string, int, bool) {
	offset := skipAccessorSpaces(raw, 0)
	if !strings.HasPrefix(raw[offset:], "context") {
		return "", 0, false
	}
	offset += len("context")
	offset = skipAccessorSpaces(raw, offset)
	if offset >= len(raw) || raw[offset] != '(' {
		return "", 0, false
	}
	offset++
	offset = skipAccessorSpaces(raw, offset)

	start := offset
	for offset < len(raw) && raw[offset] != ')' {
		ch := raw[offset]
		if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
			offset++
			continue
		}
		if ch == '/' || ch == '-' || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			offset++
			continue
		}
		return "", 0, false
	}
	if offset >= len(raw) {
		return "", 0, false
	}

	key := strings.TrimSpace(raw[start:offset])
	if !isValidContextKey(key) {
		return "", 0, false
	}

	return key, offset + 1, true
}

func isValidContextKey(key string) bool {
	if key == "" || strings.HasPrefix(key, "/") || strings.HasSuffix(key, "/") || strings.Contains(key, "//") {
		return false
	}

	for _, segment := range strings.Split(key, "/") {
		if segment == "" {
			return false
		}
		for _, r := range segment {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				continue
			}
			return false
		}
	}

	return true
}
