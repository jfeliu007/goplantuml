package connectionlabels

// AbstractInterface for testing connection labels
type AbstractInterface interface {
	interfaceFunction() bool
}

// ImplementsAbstractInterface demonstrates interface implementation and composition
type ImplementsAbstractInterface struct {
	AliasOfInt
	PublicUse AbstractInterface
}

func (iai *ImplementsAbstractInterface) interfaceFunction() bool {
	return true
}

// AliasOfInt demonstrates type alias connections
type AliasOfInt int
