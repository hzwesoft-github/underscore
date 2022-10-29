package lang

import (
	"strconv"
	"strings"
)

func IsBlank(str string) bool {
	return strings.TrimSpace(str) == ""
}

func TernaryOperator[T any](cond bool, value1 T, value2 T) T {
	if cond {
		return value1
	}

	return value2
}

func StrToInt[T int | int8 | int16 | int32 | int64](str string, base, bitSize int, defaultVal T) T {
	ret, err := strconv.ParseInt(str, base, bitSize)
	if err != nil {
		return defaultVal
	}

	return T(ret)
}

func StrToUint[T uint | uint8 | uint16 | uint32 | uint64](str string, base, bitSize int, defaultVal T) T {
	ret, err := strconv.ParseUint(str, base, bitSize)
	if err != nil {
		return defaultVal
	}

	return T(ret)
}

func StrToBool(str string, defaultVal bool) bool {
	ret, err := strconv.ParseBool(str)
	if err != nil {
		return defaultVal
	}

	return ret
}
