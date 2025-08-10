package namedimports

import customTime "time"

// MyTypeWithNamedImport demonstrates named import usage in struct fields
type MyTypeWithNamedImport struct {
	*customTime.Duration
	Timestamp customTime.Time
}

// ProcessTimeWithNamedImport demonstrates method with named import types
func (m *MyTypeWithNamedImport) ProcessTimeWithNamedImport(d customTime.Duration) customTime.Time {
	return customTime.Now().Add(d)
}
