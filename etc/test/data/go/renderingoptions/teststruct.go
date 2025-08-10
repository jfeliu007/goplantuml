package renderingoptions

type TestStruct struct {
	PublicField  string
	privateField int
}

func (t *TestStruct) PublicMethod() string {
	return t.PublicField
}

func (t *TestStruct) privateMethod() int {
	return t.privateField
}
