package common

import (
	"fmt"
	"testing"
)

func TestMustGzipEncode(t *testing.T) {
	fmt.Println(MustGzipEncode([]byte("hello")))
}
