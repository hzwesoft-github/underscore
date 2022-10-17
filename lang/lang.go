package lang

import "strings"

func IsBlank(str string) bool {
	return strings.TrimSpace(str) == ""
}

func TernaryOperator[T any](cond bool, value1 T, value2 T) T {
	if cond {
		return value1
	}

	return value2
}
