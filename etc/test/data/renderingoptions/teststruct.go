package renderingoptions

// TestStruct demonstrates different rendering options
type TestStruct struct {
	PublicField    string
	privateField   int
	ExportedField2 bool
}

// PublicMethod is an exported method
func (t *TestStruct) PublicMethod() string {
	return t.PublicField
}

// privateMethod is an unexported method
func (t *TestStruct) privateMethod() int {
	return t.privateField
}

// ExportedMethodWithParams demonstrates method with parameters
func (t *TestStruct) ExportedMethodWithParams(param1 string, param2 int) (string, error) {
	return param1, nil
}
