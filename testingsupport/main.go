package testingsupport

import (
	f "fmt"
	"strings"
)

func (t *test) test() {
	f.Println("Hello Test")
}

type test struct {
	field int
}

type myInt int

var globalVariable int

//TestComplicatedAlias for testing purposes only
type TestComplicatedAlias func(strings.Builder) bool
