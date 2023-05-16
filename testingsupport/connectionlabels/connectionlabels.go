package connectionlabels

// AbstractInterface for testing purposes
type AbstractInterface interface {
	interfaceFunction() bool
}

// ImplementsAbstractInterface for testing purposes
type ImplementsAbstractInterface struct {
	AliasOfInt
	PublicUse AbstractInterface
}

func (iai *ImplementsAbstractInterface) interfaceFunction() bool {
	return true
}

// AliasOfInt for testing purposes
type AliasOfInt int
