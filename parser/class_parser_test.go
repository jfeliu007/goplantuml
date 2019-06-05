package parser

import (
	"reflect"
	"testing"
)

func TestLineBuilder(t *testing.T) {
	s := &LineStringBuilder{}
	s.WriteLineWithDepth(1, "text")
	result := "    text\n"
	if s.String() != result {
		t.Errorf("TestLineBuilder: Expected text to be %s got %s", result, s.String())
	}

}

func TestGetOrCreateStruct(t *testing.T) {
	tt := []struct {
		name          string
		packageName   string
		nameToLookFor string
		structureName string
		structure     *Struct
		expectedEmpty bool
	}{
		{
			name:          "Struct is present",
			packageName:   "main",
			nameToLookFor: "Foo",
			structureName: "Foo",
			structure: &Struct{
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
						ReturnValues: []string{"error", "int"},
					},
				},
				Type: "class",
			},
			expectedEmpty: false,
		}, {
			name:          "Struct is not present",
			packageName:   "main",
			nameToLookFor: "Wrong",
			structureName: "Foo",
			structure: &Struct{
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
						ReturnValues: []string{"error", "int"},
					},
				},
				Type: "class",
			},
			expectedEmpty: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			parser := &ClassParser{
				currentPackageName: tc.packageName,
				structure:          make(map[string]map[string]*Struct),
				allInterfaces:      make(map[string]struct{}),
				allStructs:         make(map[string]struct{}),
			}
			parser.structure[tc.packageName] = make(map[string]*Struct)
			if tc.structure != nil {
				parser.structure[tc.packageName][tc.structureName] = tc.structure
			}

			st := parser.getOrCreateStruct(tc.nameToLookFor)
			if tc.expectedEmpty {
				if !reflect.DeepEqual(st, &Struct{
					PackageName: parser.currentPackageName,
					Functions:   make([]*Function, 0),
					Fields:      make([]*Field, 0),
					Type:        "",
					Composition: make(map[string]struct{}, 0),
					Extends:     make(map[string]struct{}, 0),
				}) {
					t.Errorf("Expected resulting structure to be equal to %v, got %v", tc.structure, st)
				}
			} else {

				if st == nil {
					t.Error("Expected a Struct, nil received")
				}
				if !reflect.DeepEqual(st, tc.structure) {
					t.Errorf("Expected resulting structure to be equal to %v, got %v", tc.structure, st)
				}
			}

		})
	}
}

func TestGetStruct(t *testing.T) {
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
				ReturnValues: []string{"error", "int"},
			},
		},
		Type: "class",
	}
	parser := getEmptyParser("main")
	parser.structure["main"] = make(map[string]*Struct)
	parser.structure["main"]["foo"] = st
	stt := parser.getStruct("main.foo")

	if stt == nil {
		t.Errorf("TestGetStruct: Extected %T, got nil", st)
	}
	if !reflect.DeepEqual(st, stt) {
		t.Errorf("TestGetStruct: Expected both structures to be equal, got %v %v", st, stt)
	}
	stt = parser.getStruct("main.wrong")
	if stt != nil {
		t.Errorf("TestGetStruct: Extected nil, got %T", st)
	}
	stt = parser.getStruct("wrong")
	if stt != nil {
		t.Errorf("TestGetStruct: Extected nil, got %T", st)
	}
}

func getEmptyParser(packageName string) *ClassParser {
	result := &ClassParser{
		currentPackageName: packageName,
		structure:          make(map[string]map[string]*Struct),
		allInterfaces:      make(map[string]struct{}),
		allStructs:         make(map[string]struct{}),
	}
	result.structure[packageName] = make(map[string]*Struct)
	return result
}
