package testingsupport

import (
	f "fmt"
	"strings"
)

func (t *test) test() {
	f.Println("Hello Test")
}

type test struct {
	field  int
	field2 TestComplicatedAlias
}

type myInt int

var globalVariable int

// TestComplicatedAlias for testing purposes only
type TestComplicatedAlias func(strings.Builder) bool
