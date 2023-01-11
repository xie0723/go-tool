package functions

import (
	"strings"
	"unicode"
)

func PathSplit(path string) []string {
	l := strings.FieldsFunc(path, func(v rune) bool {
		if unicode.IsSpace(v) || v == '\\' || v == '/' {
			return true
		}
		return false
	})
	return l
}
