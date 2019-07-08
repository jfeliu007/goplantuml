package parser

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestStructImplementsInterface(t *testing.T) {
	tt := []struct {
		name           string
		structure      *Struct
		inter          *Struct
		expectedResult bool
	}{
		{
			name: "Correct implementation",
			structure: &Struct{
				Functions: []*Function{
					{
						Name: "foo",
						Parameters: []*Field{
							{
								Name: "a",
								Type: "int",
							},
							{
								Name: "b",
								Type: "string",
							},
						},
						FullNameReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			inter: &Struct{
				Functions: []*Function{
					{
						Name: "foo",
						Parameters: []*Field{
							{
								Type: "int",
							},
							{
								Type: "string",
							},
						},
						FullNameReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			expectedResult: true,
		}, {
			name: "Parameters not in order",
			structure: &Struct{
				Functions: []*Function{
					{
						Name: "foo",
						Parameters: []*Field{
							{
								Name: "a",
								Type: "int",
							},
							{
								Name: "b",
								Type: "string",
							},
						},
						FullNameReturnValues: []string{"int", "error"},
					},
				},
				Type: "interface",
			},
			inter: &Struct{
				Functions: []*Function{
					{
						Name: "foo",
						Parameters: []*Field{
							{
								Name: "b",
								Type: "string",
							},
							{
								Name: "a",
								Type: "int",
							},
						},
						FullNameReturnValues: []string{"int", "error"},
					},
				},
				Type: "interface",
			},
			expectedResult: false,
		}, {
			name: "Empty Interface",
			structure: &Struct{
				Functions: []*Function{
					{
						Name: "foo",
						Parameters: []*Field{
							{
								Name: "a",
								Type: "int",
							},
							{
								Name: "b",
								Type: "string",
							},
						},
						FullNameReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			inter: &Struct{
				Functions: []*Function{},
				Type:      "interface",
			},
			expectedResult: false,
		}, {
			name: "Different Function Names",
			structure: &Struct{
				Functions: []*Function{
					{
						Name: "foo",
						Parameters: []*Field{
							{
								Name: "a",
								Type: "int",
							},
							{
								Name: "b",
								Type: "string",
							},
						},
						FullNameReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			inter: &Struct{
				Functions: []*Function{
					{
						Name: "bar",
						Parameters: []*Field{
							{
								Name: "a",
								Type: "int",
							},
							{
								Name: "b",
								Type: "string",
							},
						},
						FullNameReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			expectedResult: false,
		}, {
			name: "Return value different",
			structure: &Struct{
				Functions: []*Function{
					{
						Name: "foo",
						Parameters: []*Field{
							{
								Name: "a",
								Type: "int",
							},
							{
								Name: "b",
								Type: "string",
							},
						},
						FullNameReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			inter: &Struct{
				Functions: []*Function{
					{
						Name: "foo",
						Parameters: []*Field{
							{
								Type: "int",
							},
							{
								Type: "string",
							},
						},
						FullNameReturnValues: []string{"error", "int"},
					},
				},
				Type: "class",
			},
			expectedResult: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.structure.ImplementsInterface(tc.inter)
			if result != tc.expectedResult {
				t.Errorf("Expected result to be %t, got %t", tc.expectedResult, result)
			}
		})

	}
}

func TestAddToComposition(t *testing.T) {
	st := &Struct{
		Functions: []*Function{
			{
				Name: "foo",
				Parameters: []*Field{
					{
						Type: "int",
					},
					{
						Type: "string",
					},
				},
				ReturnValues:         []string{"error", "int"},
				FullNameReturnValues: []string{"error", "int"},
			},
		},
		Type:        "class",
		PackageName: "test",
		Fields:      make([]*Field, 0),
		Composition: make(map[string]struct{}),
		Extends:     make(map[string]struct{}),
	}
	st.AddToComposition("Foo")

	if !arrayContains(st.Composition, "Foo") {
		t.Errorf("TestAddToComposition: Expected CompositionArray to have %s, but it contains %v", "Foo", st.Composition)
	}

	st.AddToComposition("")

	if arrayContains(st.Composition, "") {
		t.Errorf(`TestAddToComposition: Expected CompositionArray to not have "", but it contains %v`, st.Composition)
	}
	testArray := map[string]struct{}{
		"Foo": struct{}{},
	}
	if !reflect.DeepEqual(st.Composition, testArray) {

		t.Errorf("TestAddToComposition: Expected CompositionArray to be %v, but it contains %v", testArray, st.Composition)
	}

	st.AddToComposition("*Foo2")

	if !arrayContains(st.Composition, "Foo2") {
		t.Errorf("TestAddToComposition: Expected CompositionArray to have %s, but it contains %v", "Foo2", st.Composition)
	}
}
func TestAddToExtension(t *testing.T) {
	st := &Struct{
		Functions: []*Function{
			{
				Name: "foo",
				Parameters: []*Field{
					{
						Type: "int",
					},
					{
						Type: "string",
					},
				},
				ReturnValues:         []string{"error", "int"},
				FullNameReturnValues: []string{"error", "int"},
			},
		},
		Type:        "class",
		PackageName: "test",
		Fields:      make([]*Field, 0),
		Composition: make(map[string]struct{}),
		Extends:     make(map[string]struct{}),
	}
	st.AddToExtends("Foo")

	if !arrayContains(st.Extends, "Foo") {
		t.Errorf("TestAddToComposition: Expected Extends Array to have %s, but it contains %v", "Foo", st.Composition)
	}

	st.AddToExtends("")

	if arrayContains(st.Extends, "") {
		t.Errorf(`TestAddToComposition: Expected Extends Array to not have "", but it contains %v`, st.Composition)
	}
	testArray := map[string]struct{}{
		"Foo": struct{}{},
	}
	if !reflect.DeepEqual(st.Extends, testArray) {
		t.Errorf("TestAddToComposition: Expected Extends Array to be %v, but it contains %v", testArray, st.Composition)
	}

	st.AddToExtends("*Foo2")

	if !arrayContains(st.Extends, "Foo2") {
		t.Errorf("TestAddToComposition: Expected Extends Array to have %s, but it contains %v", "Foo2", st.Composition)
	}
}

func arrayContains(a map[string]struct{}, text string) bool {

	found := false
	for v := range a {
		if v == text {
			found = true
			break
		}
	}
	return found
}

func TestAddField(t *testing.T) {
	st := &Struct{
		PackageName: "main",
		Functions: []*Function{
			{
				Name:                 "foo",
				Parameters:           []*Field{},
				ReturnValues:         []string{"error", "int"},
				FullNameReturnValues: []string{"error", "int"},
			},
		},
		Type:        "class",
		Fields:      make([]*Field, 0),
		Composition: make(map[string]struct{}),
		Extends:     make(map[string]struct{}),
	}
	st.AddField(&ast.Field{
		Names: []*ast.Ident{
			{
				Name: "foo",
			},
		},
		Type: &ast.Ident{
			Name: "int",
		},
	}, make(map[string]string))
	if len(st.Fields) != 1 {
		t.Errorf("TestAddField: Expected st.Fields to have exactly one element but it has %d elements", len(st.Fields))
	}
	testField := &Field{
		Name: "foo",
		Type: "int",
	}
	if !reflect.DeepEqual(st.Fields[0], testField) {
		t.Errorf("TestAddField: Expected st.Fields[0] to have %v, got %v", testField, st.Fields[0])
	}
	st.AddField(&ast.Field{
		Names: nil,
		Type: &ast.StarExpr{
			X: &ast.Ident{
				Name: "FooComposed",
			},
		},
	}, make(map[string]string))
	if !arrayContains(st.Composition, "FooComposed") {
		t.Errorf("TestAddField: Expecting FooComposed to be part of the compositions ,but the array had %v", st.Composition)
	}
}

func TestAddMethod(t *testing.T) {
	st := &Struct{
		PackageName: "main",
		Functions:   []*Function{},
		Type:        "class",
	}
	st.AddMethod(&ast.Field{
		Names: []*ast.Ident{
			{
				Name: "foo",
			},
		},
		Type: &ast.Ident{},
	}, make(map[string]string))
	if len(st.Functions) != 0 {
		t.Errorf("TestAddMethod: Expected Functions array to be empty but it contains %v", st.Functions)
	}
	st.AddMethod(&ast.Field{
		Names: []*ast.Ident{
			{
				Name: "foo",
			},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: "var1",
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: "FooComposed",
							},
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: nil,
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: "FooComposed",
							},
						},
					},
				},
			},
		},
	}, make(map[string]string))
	if len(st.Functions) != 1 {
		t.Errorf("TestAddMethod: Expected st.Functions to have exactly one element but it has %d elements", len(st.Functions))
	}
	testFunction := &Function{
		PackageName: "main",
		Name:        "foo",
		Parameters: []*Field{
			{
				Name: "var1",
				Type: "*FooComposed",
			},
		},
		ReturnValues:         []string{"*FooComposed"},
		FullNameReturnValues: []string{"*main.FooComposed"},
	}
	if !st.Functions[0].SignturesAreEqual(testFunction) {
		t.Errorf("TestAddMethod: Expected st.Function[0] to have %v, got %v", testFunction, st.Functions[0])
	}
}
