package parser

import (
	"go/ast"
)

//Struct represent a struct in golang, it can be of Type "class" or "interface" and can be associated
//with other structs via Composition and Extends
type Struct struct {
	PackageName string
	Functions   []*Function
	Fields      []*Field
	Type        string
	Composition []string
	Extends     []string
}

// ImplementsInterface returns true if the struct st conforms ot the given interface
func (st *Struct) ImplementsInterface(inter *Struct) bool {
	if len(inter.Functions) == 0 {
		return false
	}
	for _, f1 := range inter.Functions {
		foundMatch := false
		for _, f2 := range st.Functions {
			if f1.SignturesAreEqual(f2) {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			return false
		}
	}
	return true
}

//AddToComposition adds the composition relation to the structure. We want to make sure that *ExampleStruct
//gets added as ExampleStruct so that we can properly build the relation later to the
//class identifier
func (st *Struct) AddToComposition(fType string) {
	if len(fType) == 0 {
		return
	}
	if fType[0] == "*"[0] {
		fType = fType[1:]
	}
	st.Composition = append(st.Composition, fType)
}

//AddToExtends Adds an extends relationship to this struct. We want to make sure that *ExampleStruct
//gets added as ExampleStruct so that we can properly build the relation later to the
//class identifier
func (st *Struct) AddToExtends(fType string) {
	if len(fType) == 0 {
		return
	}
	if fType[0] == "*"[0] {
		fType = fType[1:]
	}
	st.Extends = append(st.Extends, fType)
}

//AddField adds a field into this structure. It parses the ast.Field and extract all
//needed information
func (st *Struct) AddField(field *ast.Field) {
	if field.Names != nil {
		st.Fields = append(st.Fields, &Field{
			Name: field.Names[0].Name,
			Type: getFieldType(field.Type, ""),
		})
	} else if field.Type != nil {
		fType := getFieldType(field.Type, "")
		if fType[0] == "*"[0] {
			fType = fType[1:]
		}
		st.AddToComposition(fType)
	}
}

// Parse the Field and if it is an ast.FuncType, then add the methods into the structure
func (st *Struct) AddMethod(method *ast.Field) {
	f, ok := method.Type.(*ast.FuncType)
	if !ok {
		return
	}
	function := getFunction(f, method.Names[0].Name)
	st.Functions = append(st.Functions, function)
}
