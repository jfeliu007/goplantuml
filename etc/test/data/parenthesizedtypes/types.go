package parenthesizedtypes

type (
	// Foo is a test interface for testing parenthesized type declarations
	Foo interface {
		FooMethod()
	}

	// Bar is another test interface in the same parenthesized block
	Bar interface {
		BarMethod()
	}

	// BazStruct demonstrates struct in parenthesized declaration
	BazStruct struct {
		Field1 string
		Field2 int
	}

	// CustomString demonstrates type alias in parenthesized declaration
	CustomString string
)
