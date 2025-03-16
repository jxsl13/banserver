package parser

import (
	"math"
	"strconv"
	"unsafe"
)

const (
	intBytes = int(unsafe.Sizeof(math.MaxInt) * 8)
)

func parseInt(s string) (int, error) {
	i64, err := strconv.ParseInt(s, 10, intBytes)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}

func mustParseInt(s string) int {
	i, err := parseInt(s)
	if err != nil {
		panic(err)
	}
	return i
}
