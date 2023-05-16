package subfolder2

// Subfolder2 structure for testing purpose only
type Subfolder2 struct {
}

// SubfolderFunction is for testing purposes
func (s *Subfolder2) SubfolderFunction(b bool, i int) bool {
	return true
}

func (s *Subfolder2) SubfolderFunctionWithReturnListParametrized() (a, b, c []byte, err error) {
	return
}
