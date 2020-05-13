package parser

import (
	"go/ast"
	"unicode"
)

//Struct represent a struct in golang, it can be of Type "class" or "interface" and can be associated
//with other structs via Composition and Extends
type Struct struct {
	PackageName         string
	Functions           []*Function
	Fields              []*Field
	Type                string
	Composition         map[string]struct{}
	Extends             map[string]struct{}
	Aggregations        map[string]struct{}
	PrivateAggregations map[string]struct{}
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
	st.Composition[fType] = struct{}{}
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
	st.Extends[fType] = struct{}{}
}

//AddToAggregation adds an aggregation type to the list of aggregations
func (st *Struct) AddToAggregation(fType string) {
	st.Aggregations[fType] = struct{}{}
}

//addToPrivateAggregation adds an aggregation type to the list of aggregations for private members
func (st *Struct) addToPrivateAggregation(fType string) {
	st.PrivateAggregations[fType] = struct{}{}
}

//AddField adds a field into this structure. It parses the ast.Field and extract all
//needed information
func (st *Struct) AddField(field *ast.Field, aliases map[string]string) {
	theType, fundamentalTypes := getFieldType(field.Type, aliases)
	theType = replacePackageConstant(theType, "")
	if field.Names != nil {
		theType = replacePackageConstant(theType, "")
		newField := &Field{
			Name: field.Names[0].Name,
			Type: theType,
		}
		st.Fields = append(st.Fields, newField)
		if unicode.IsUpper(rune(newField.Name[0])) {
			for _, t := range fundamentalTypes {
				st.AddToAggregation(replacePackageConstant(t, st.PackageName))
			}
		} else {
			for _, t := range fundamentalTypes {
				st.addToPrivateAggregation(replacePackageConstant(t, st.PackageName))
			}
		}
	} else if field.Type != nil {
		if theType[0] == "*"[0] {
			theType = theType[1:]
		}
		st.AddToComposition(theType)
	}
}

//AddMethod Parse the Field and if it is an ast.FuncType, then add the methods into the structure
func (st *Struct) AddMethod(method *ast.Field, aliases map[string]string) {
	f, ok := method.Type.(*ast.FuncType)
	if !ok {
		return
	}
	function := getFunction(f, method.Names[0].Name, aliases, st.PackageName)
	st.Functions = append(st.Functions, function)
}
