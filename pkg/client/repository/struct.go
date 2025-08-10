package repository

import (
	"go/ast"
	"strings"
)

// Field can hold the name and type of any field
type Field struct {
	Name     string
	Type     string
	FullType string
}

// Struct contains all the information needed to generate a class diagram structure
type Struct struct {
	Type                 string
	PackageName          string
	Functions            []*Function
	Fields               []*Field
	Composition          map[string]struct{}
	Extends              map[string]struct{}
	Aggregations         map[string]struct{}
	PrivateAggregations  map[string]struct{}
	PrivateCompositions  map[string]struct{}
	RecursiveAggregation map[string]struct{}
}

// normalizeTypeName removes leading pointer stars for relationship targets
func normalizeTypeName(name string) string {
	return strings.TrimLeft(name, "*")
}

// AddField Parse the Field and if it's a pointer field, then it will create composition to the pointed reference
// and if it is an embedded type, then it will create an extends relationship
func (st *Struct) AddField(field *ast.Field, aliases map[string]string) {
	for _, name := range field.Names {
		isPublic := name.IsExported()
		if st.Fields == nil {
			st.Fields = []*Field{}
		}
		fieldType, subFields := getFieldType(field.Type, aliases)
		fieldType = replacePackageConstant(fieldType, st.PackageName)
		f := &Field{
			Name: name.Name,
			Type: fieldType,
		}
		st.Fields = append(st.Fields, f)
		for _, subField := range subFields {
			subField = replacePackageConstant(subField, st.PackageName)
			if !isPrimitiveString(subField) {
				if isPublic {
					st.AddToAggregation(subField)
				} else {
					st.AddToPrivateAggregation(subField)
				}
			}
		}
		if !isPrimitiveString(fieldType) {
			if isPublic {
				st.AddToAggregation(fieldType)
			} else {
				st.AddToPrivateAggregation(fieldType)
			}
		}
	}
	// If field.Names is nil, it means that it is an embedded field and should be treated as an extends relationship
	if field.Names == nil {
		fieldType, _ := getFieldType(field.Type, aliases)
		fieldType = replacePackageConstant(fieldType, st.PackageName)
		if fieldType[0] == "*"[0] {
			fieldType = fieldType[1:]
			st.AddToComposition(fieldType)
		} else {
			st.AddToExtends(fieldType)
		}
	}
}

// AddMethod Parse the Field and if it is an ast.FuncType, then add the methods into the structure
func (st *Struct) AddMethod(method *ast.Field, aliases map[string]string) {
	f, ok := method.Type.(*ast.FuncType)
	if !ok {
		return
	}
	function := getFunction(f, method.Names[0].Name, aliases, st.PackageName)
	st.Functions = append(st.Functions, function)
}

// AddToComposition adds a struct name to the composition relation of the struct
func (st *Struct) AddToComposition(structName string) {
	structName = normalizeTypeName(structName)
	if st.Composition == nil {
		st.Composition = make(map[string]struct{})
	}
	st.Composition[structName] = struct{}{}
}

// AddToExtends adds a struct name to the extends relation of the struct
func (st *Struct) AddToExtends(structName string) {
	structName = normalizeTypeName(structName)
	if st.Extends == nil {
		st.Extends = make(map[string]struct{})
	}
	st.Extends[structName] = struct{}{}
}

// AddToAggregation adds a struct name to the aggregation relation of the struct
func (st *Struct) AddToAggregation(structName string) {
	structName = normalizeTypeName(structName)
	if st.Aggregations == nil {
		st.Aggregations = make(map[string]struct{})
	}
	st.Aggregations[structName] = struct{}{}
}

// AddToPrivateAggregation adds a struct name to the private aggregation relation of the struct
func (st *Struct) AddToPrivateAggregation(structName string) {
	structName = normalizeTypeName(structName)
	if st.PrivateAggregations == nil {
		st.PrivateAggregations = make(map[string]struct{})
	}
	st.PrivateAggregations[structName] = struct{}{}
}

// ImplementsInterface returns true if the struct has the methods to implement the given interface struct
func (st *Struct) ImplementsInterface(inter *Struct) bool {
	if inter.Type != "interface" {
		return false
	}
	if len(inter.Functions) == 0 {
		return false
	}
	for _, interfaceFunction := range inter.Functions {
		found := false
		for _, structFunction := range st.Functions {
			if structFunction.SignturesAreEqual(interfaceFunction) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
