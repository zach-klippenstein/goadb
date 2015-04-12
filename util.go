package goadb

import "strings"
import (
	"regexp"
)

var (
	whitespaceRegex = regexp.MustCompile(`^\s*$`)
)

func containsWhitespace(str string) bool {
	return strings.ContainsAny(str, " \t\v")
}

func isBlank(str string) bool {
	return whitespaceRegex.MatchString(str)
}
