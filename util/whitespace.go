package util

import "strings"

// Whitespace is the whitespace of Compono: U+0020 (space) and U+0009 (tab).
const Whitespace = " \t"

// TrimWhitespace returns s without the Compono whitespace at its start and end.
func TrimWhitespace(s string) string { return strings.Trim(s, Whitespace) }
