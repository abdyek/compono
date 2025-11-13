package util

import "regexp"

func IsScreamingSnakeCase(val string) bool {
	re := regexp.MustCompile(`^[A-Z0-9]+(?:_[A-Z0-9]+)*$`)
	return re.MatchString(val)
}
