package lang

import "strings"

func IsBlank(str string) bool {
	return strings.TrimSpace(str) == ""
}
