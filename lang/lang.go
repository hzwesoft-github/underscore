package lang

import (
	"strconv"
	"strings"
)

func IsBlank(str string) bool {
	return strings.TrimSpace(str) == ""
}

func EqualsAny[T comparable](one T, values ...T) bool {
	for _, v := range values {
		if one == v {
			return true
		}
	}
	return false
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

func StrSplit2(str string, sep string) (string, string) {
	parts := strings.Split(str, sep)
	if len(parts) == 1 {
		return parts[0], ""
	}

	return parts[0], parts[1]
}

func StrSplit3(str string, sep string) (string, string, string) {
	parts := strings.Split(str, sep)
	if len(parts) == 1 {
		return parts[0], "", ""
	}

	if len(parts) == 2 {
		return parts[0], parts[1], ""
	}

	return parts[0], parts[1], parts[2]
}

func StrToSlice(strs ...string) []string {
	slice := make([]string, 0)
	for _, str := range strs {
		if str != "" {
			slice = append(slice, str)
		}
	}

	return slice
}

func SliceToStr2(slice []string) (string, string) {
	if len(slice) == 0 {
		return "", ""
	}

	if len(slice) == 1 {
		return slice[0], ""
	}

	return slice[0], slice[1]

}

func SliceToStr3(slice []string) (string, string, string) {
	if len(slice) == 0 {
		return "", "", ""
	}
	if len(slice) == 1 {
		return slice[0], "", ""
	}
	if len(slice) == 2 {
		return slice[0], slice[1], ""
	}

	return slice[0], slice[1], slice[2]
}

func SliceToStr4(slice []string) (string, string, string, string) {
	if len(slice) == 0 {
		return "", "", "", ""
	}
	if len(slice) == 1 {
		return slice[0], "", "", ""
	}
	if len(slice) == 2 {
		return slice[0], slice[1], "", ""
	}
	if len(slice) == 3 {
		return slice[0], slice[1], slice[2], ""
	}

	return slice[0], slice[1], slice[2], slice[3]
}
