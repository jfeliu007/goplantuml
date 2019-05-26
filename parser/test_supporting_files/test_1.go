package test1

import "strings"

type Foo struct {
	i  int
	i2 *strings.Builder
	m  map[string]float32
	a  []*struct {
		integer int
		boolean bool
	}
	*Bar
	*strings.Builder
	t *struct{}
}

type Bar struct {
}
