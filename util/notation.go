package util

import (
	"regexp"
	"strings"
	"unicode"
)

func IsScreamingSnakeCase(val string) bool {
	re := regexp.MustCompile(`^[A-Z0-9]+(?:_[A-Z0-9]+)*$`)
	return re.MatchString(val)
}

func IsKebabCase(val string) bool {
	re := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	return re.MatchString(val)
}

func ToKebabCase(val string) string {
	if val == "" {
		return ""
	}

	var b strings.Builder
	runes := []rune(val)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					b.WriteByte('-')
				} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					b.WriteByte('-')
				}
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}

		if r == '_' || r == ' ' {
			if b.Len() > 0 {
				b.WriteByte('-')
			}
			continue
		}

		b.WriteRune(unicode.ToLower(r))
	}

	return strings.Trim(b.String(), "-")
}
