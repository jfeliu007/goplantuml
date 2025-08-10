package repository

import (
	"go/ast"
)

// Function holds the signature of a function with name, Parameters and Return values
type Function struct {
	Name                 string
	Parameters           []*Field
	ReturnValues         []string
	PackageName          string
	FullNameReturnValues []string
}

// SignturesAreEqual Returns true if the two functions have the same signature (parameter names are not checked)
func (f *Function) SignturesAreEqual(function *Function) bool {
	if f.Name != function.Name {
		return false
	}
	if len(f.Parameters) != len(function.Parameters) {
		return false
	}
	for i, parameter := range f.Parameters {
		if parameter.Type != function.Parameters[i].Type {
			return false
		}
	}
	if len(f.ReturnValues) != len(function.ReturnValues) {
		return false
	}
	for i, returnValue := range f.ReturnValues {
		if returnValue != function.ReturnValues[i] {
			return false
		}
	}
	return true
}

// generate and return a function object from the given Functype. The names must be passed to this
// function since the FuncType does not have this information
func getFunction(f *ast.FuncType, name string, aliases map[string]string, packageName string) *Function {
	function := &Function{
		Name:                 name,
		Parameters:           make([]*Field, 0),
		ReturnValues:         make([]string, 0),
		FullNameReturnValues: make([]string, 0),
		PackageName:          packageName,
	}
	params := f.Params
	if params != nil {
		for _, pa := range params.List {
			theType, _ := getFieldType(pa.Type, aliases)
			if pa.Names != nil {
				if pa.Names != nil {
					for _, fieldName := range pa.Names {
						function.Parameters = append(function.Parameters, &Field{
							Name:     fieldName.Name,
							Type:     replacePackageConstant(theType, ""),
							FullType: replacePackageConstant(theType, packageName),
						})
					}
				}
			} else {
				function.Parameters = append(function.Parameters, &Field{
					Name:     "",
					Type:     replacePackageConstant(theType, ""),
					FullType: replacePackageConstant(theType, packageName),
				})
			}
		}
	}

	results := f.Results
	if results != nil {
		for _, pa := range results.List {
			theType, _ := getFieldType(pa.Type, aliases)
			count := 1
			if pa.Names != nil {
				count = len(pa.Names)
			}
			for count > 0 {
				count--
				function.ReturnValues = append(function.ReturnValues, replacePackageConstant(theType, ""))
				function.FullNameReturnValues = append(function.FullNameReturnValues, replacePackageConstant(theType, packageName))
			}
		}
	}
	return function
}
