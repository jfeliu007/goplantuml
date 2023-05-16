package parser

import "fmt"

// Alias defines a type that is an alias for some other type
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

// AliasSlice implement the sort.Interface interface to allow for proper sorting of an alias slice
type AliasSlice []Alias

// Len is the number of elements in the collection.
func (as AliasSlice) Len() int {
	return len(as)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (as AliasSlice) Less(i, j int) bool {
	return fmt.Sprintf("%s %s %s", as[i].Name, as[i].PackageName, as[i].AliasOf) < fmt.Sprintf("%s %s %s", as[j].Name, as[j].PackageName, as[j].AliasOf)
}

// Swap swaps the elements with indexes i and j.
func (as AliasSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}
