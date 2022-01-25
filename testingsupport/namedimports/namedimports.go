package namedimports

import time "time"

// MyType for testing purposes when a named import is used as an anonymous
// field.
type MyType struct {
	*time.Duration
}
