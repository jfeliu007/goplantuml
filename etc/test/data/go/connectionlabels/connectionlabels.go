package connectionlabels

type AbstractInterface interface {
	interfaceFunction() bool
}

type ImplementsAbstractInterface struct {
	AliasOfInt
	PublicUse AbstractInterface
}

func (iai *ImplementsAbstractInterface) interfaceFunction() bool {
	return true
}

type AliasOfInt int
