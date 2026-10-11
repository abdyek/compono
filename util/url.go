package util

import "strings"

func stripC0Spaces(s string) string {
	for len(s) > 0 && (s[0] <= 0x1F || s[0] == ' ') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] <= 0x1F || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}

// NormalizeURL returns url as the WHATWG URL parser reads it before it looks
// for a scheme: without leading and trailing C0 control characters (U+0000 to
// U+001F) and spaces, and without any ASCII tab or newline (U+0009, U+000A,
// U+000D) inside it.
func NormalizeURL(url string) string {
	url = stripC0Spaces(url)
	b := make([]byte, 0, len(url))
	for i := 0; i < len(url); i++ {
		c := url[i]
		if c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		b = append(b, c)
	}
	return string(b)
}

// DisallowedURLScheme returns the lowercased scheme of the normalized url and
// true when the url has a scheme other than http, https, mailto or tel. The
// url has a scheme when it starts with an ASCII letter followed only by ASCII
// letters, digits, '+', '-' and '.' up to its first ':'. Otherwise it is a
// relative url and is allowed.
func DisallowedURLScheme(url string) (string, bool) {
	url = NormalizeURL(url)
	if len(url) == 0 {
		return "", false
	}
	c := url[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
		return "", false
	}
	i := 1
	for i < len(url) {
		c := url[i]
		if c == ':' {
			break
		}
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '+' || c == '-' || c == '.') {
			return "", false
		}
		i++
	}
	if i >= len(url) || url[i] != ':' {
		return "", false
	}
	scheme := strings.ToLower(url[:i])
	switch scheme {
	case "http", "https", "mailto", "tel":
		return scheme, false
	default:
		return scheme, true
	}
}
