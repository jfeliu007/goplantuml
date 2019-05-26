package parser

import (
	"go/ast"
	"reflect"
)

//Function holds the signature of a function with name, Parameters and Return values
type Function struct {
	Name         string
	Parameters   []*Field
	ReturnValues []string
}

//Returns true if the two functions have the same signature (parameter names are not checked)
func (f *Function) SignturesAreEqual(function *Function) bool {
	result := true
	result = result && (function.Name == f.Name)
	result = result && reflect.DeepEqual(f.ReturnValues, function.ReturnValues)
	result = result && (len(f.Parameters) == len(function.Parameters))
	if result {
		for i, p := range f.Parameters {
			if p.Type != function.Parameters[i].Type {
				return false
			}
		}
	}
	return result
}

// generate and return a function object from the given Functype. The names must be passed to this
// function since the FuncType does not have this information
func getFunction(f *ast.FuncType, name string) *Function {
	function := &Function{
		Name:         name,
		Parameters:   make([]*Field, 0),
		ReturnValues: make([]string, 0),
	}
	params := f.Params
	if params != nil {
		for _, pa := range params.List {
			fieldName := ""
			if pa.Names != nil {
				fieldName = pa.Names[0].Name
			}
			function.Parameters = append(function.Parameters, &Field{
				Name: fieldName,
				Type: getFieldType(pa.Type, ""),
			})
		}
	}

	results := f.Results
	if results != nil {
		for _, pa := range results.List {
			function.ReturnValues = append(function.ReturnValues, getFieldType(pa.Type, ""))
		}
	}
	return function
}
