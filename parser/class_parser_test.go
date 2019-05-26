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
		t.Errorf("Expected text to be %s got %s", result, s.String())
	}

}

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
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Name: "a",
								Type: "int",
							},
							&Parameter{
								Name: "b",
								Type: "string",
							},
						},
						ReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			inter: &Struct{
				Functions: []*Function{
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Type: "int",
							},
							&Parameter{
								Type: "string",
							},
						},
						ReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			expectedResult: true,
		}, {
			name: "Parameters not in order",
			structure: &Struct{
				Functions: []*Function{
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Name: "a",
								Type: "int",
							},
							&Parameter{
								Name: "b",
								Type: "string",
							},
						},
						ReturnValues: []string{"int", "error"},
					},
				},
				Type: "interface",
			},
			inter: &Struct{
				Functions: []*Function{
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Name: "b",
								Type: "string",
							},
							&Parameter{
								Name: "a",
								Type: "int",
							},
						},
						ReturnValues: []string{"int", "error"},
					},
				},
				Type: "interface",
			},
			expectedResult: false,
		}, {
			name: "Empty Interface",
			structure: &Struct{
				Functions: []*Function{
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Name: "a",
								Type: "int",
							},
							&Parameter{
								Name: "b",
								Type: "string",
							},
						},
						ReturnValues: []string{"int", "error"},
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
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Name: "a",
								Type: "int",
							},
							&Parameter{
								Name: "b",
								Type: "string",
							},
						},
						ReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			inter: &Struct{
				Functions: []*Function{
					&Function{
						Name: "bar",
						Parameters: []*Parameter{
							&Parameter{
								Name: "a",
								Type: "int",
							},
							&Parameter{
								Name: "b",
								Type: "string",
							},
						},
						ReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			expectedResult: false,
		}, {
			name: "Return value different",
			structure: &Struct{
				Functions: []*Function{
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Name: "a",
								Type: "int",
							},
							&Parameter{
								Name: "b",
								Type: "string",
							},
						},
						ReturnValues: []string{"int", "error"},
					},
				},
				Type: "class",
			},
			inter: &Struct{
				Functions: []*Function{
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Type: "int",
							},
							&Parameter{
								Type: "string",
							},
						},
						ReturnValues: []string{"error", "int"},
					},
				},
				Type: "class",
			},
			expectedResult: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			parser := &ClassParser{}
			result := parser.structImplementsInterface(tc.structure, tc.inter)
			if result != tc.expectedResult {
				t.Errorf("Expected result to be %t, got %t", tc.expectedResult, result)
			}
		})

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
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Type: "int",
							},
							&Parameter{
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
					&Function{
						Name: "foo",
						Parameters: []*Parameter{
							&Parameter{
								Type: "int",
							},
							&Parameter{
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
					Functions:   make([]*Function, 0),
					Fields:      make([]*Parameter, 0),
					Type:        "",
					Composition: make([]string, 0),
					Extends:     make([]string, 0),
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
			&Function{
				Name: "foo",
				Parameters: []*Parameter{
					&Parameter{
						Type: "int",
					},
					&Parameter{
						Type: "string",
					},
				},
				ReturnValues: []string{"error", "int"},
			},
		},
		Type: "class",
	}
	parser := &ClassParser{
		currentPackageName: "main",
		structure:          make(map[string]map[string]*Struct),
		allInterfaces:      make(map[string]struct{}),
		allStructs:         make(map[string]struct{}),
	}
	parser.structure["main"] = make(map[string]*Struct)
	parser.structure["main"]["foo"] = st
	stt := parser.getStruct("main.foo")

	if stt == nil {
		t.Errorf("Extected %T, got nil", st)
	}
	if !reflect.DeepEqual(st, stt) {
		t.Errorf("Expected both structures to be equal, got %v %v", st, stt)
	}
	stt = parser.getStruct("main.wrong")
	if stt != nil {
		t.Errorf("Extected nil, got %T", st)
	}
	stt = parser.getStruct("wrong")
	if stt != nil {
		t.Errorf("Extected nil, got %T", st)
	}
}
