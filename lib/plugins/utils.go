package plugins

import (
	"regexp"
)

var escapeReg = regexp.MustCompile(`[.*+?^${}()|[\]\\]`)

func escapeRegExp(str string) string {
	return escapeReg.ReplaceAllString(str, "\\$&")
}
