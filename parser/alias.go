package parser

import "fmt"

//Alias defines a type that is an alias for some other type
type Alias struct {
	Name        string
	PackageName string
	AliasOf     string
}

func getNewAlias(name, packageName, aliasOf string) *Alias {
	if isPrimitiveString(name) {
		name = fmt.Sprintf("%s.%s", builtinPackageName, name)
	}
	return &Alias{
		Name:        name,
		PackageName: packageName,
		AliasOf:     aliasOf,
	}
}
