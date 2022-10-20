package lang

import (
	"fmt"
	"testing"
)

func TestAddToSlice(t *testing.T) {
	var strs []string
	AddToSlice(&strs, "1", "2")

	fmt.Println(strs)
}
