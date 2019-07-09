package testingsupport

import f "fmt"

func (t *test) test() {
	f.Println("Hello Test")
}

type test struct {
	field int
}

type myInt int

var globalVariable int
