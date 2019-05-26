package test1

type Foo struct {
	i int
	m map[string]float32
	a []*struct {
		integer int
		boolean bool
	}
	*Bar
}

type Bar struct {
}
