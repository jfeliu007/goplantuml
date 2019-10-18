package parser

import (
	"io/ioutil"
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

func TestRenderStructFields(t *testing.T) {
	parser := getEmptyParser("main")

	st := &Struct{
		Fields: []*Field{
			{
				Name: "privateField",
				Type: "int",
			},
			{
				Name: "PublicField",
				Type: "string",
			},
		},
	}
	privateFields := &LineStringBuilder{}
	publicFields := &LineStringBuilder{}
	parser.renderStructFields(st, privateFields, publicFields)
	if privateFields.String() != "        - privateField int\n" {
		t.Errorf("TestRenderStructFields: expected privateFields to be [        - privateField int\\n] got [%v]", privateFields.String())
	}
	if publicFields.String() != "        + PublicField string\n" {
		t.Errorf("TestRenderStructFields: expected publicFields to be [        + PublicField int\\n] got [%v]", publicFields.String())
	}
}

func TestRenderStructutes(t *testing.T) {

	structMap := map[string]*Struct{
		"MainClass": getTestStruct(),
	}
	lineB := &LineStringBuilder{}
	parser := getEmptyParser("main")
	parser.renderStructures("main", structMap, lineB)
	expectedResult := "namespace main {\n    class MainClass << (S,Aquamarine) >> {\n        - privateField int\n\n        + PublicField error\n\n        - foo( int,  string) (error, int)\n\n        + Boo( string,  int) int\n\n    }\n}\nfoopack.AnotherClass *-- main.MainClass\n\nmain.NewClass <|-- main.MainClass\n\n"
	if lineB.String() != expectedResult {
		t.Errorf("TestRenderStructures: expected %s, got %s", expectedResult, lineB.String())
	}
}

func TestRenderStructure(t *testing.T) {
	parser := getEmptyParser("main")
	st := getTestStruct()
	lineBuilder := &LineStringBuilder{}
	compositionBuilder := &LineStringBuilder{}
	extendBuilder := &LineStringBuilder{}
	parser.renderStructure(st, "main", "TestClass", lineBuilder, compositionBuilder, extendBuilder)
	expectedLineBuilder := "    class TestClass << (S,Aquamarine) >> {\n        - privateField int\n\n        + PublicField error\n\n        - foo( int,  string) (error, int)\n\n        + Boo( string,  int) int\n\n    }\n"
	if lineBuilder.String() != expectedLineBuilder {
		t.Errorf("TestRenderStructure: Expected lineBuilder [%s] got [%s]", expectedLineBuilder, lineBuilder.String())
	}
	expectedComposition := "foopack.AnotherClass *-- main.TestClass\n"
	if compositionBuilder.String() != expectedComposition {
		t.Errorf("TestRenderStructure: Expected compositionBuilder %s got %s", expectedComposition, compositionBuilder.String())
	}
	expectedExtends := "main.NewClass <|-- main.TestClass\n"
	if extendBuilder.String() != expectedExtends {
		t.Errorf("TestRenderStructure: Expected extendBuilder %s got %s", expectedExtends, extendBuilder.String())
	}
}

func getTestStruct() *Struct {
	return &Struct{
		Type:        "class",
		PackageName: "main",
		Composition: map[string]struct{}{
			"foopack.AnotherClass": {},
		},
		Extends: map[string]struct{}{
			"NewClass": {},
		},
		Fields: []*Field{
			{
				Name: "privateField",
				Type: "int",
			},
			{
				Name: "PublicField",
				Type: "error",
			},
		},
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
			{
				Name: "Boo",
				Parameters: []*Field{
					{
						Type: "string",
					},
					{
						Type: "int",
					},
				},
				ReturnValues: []string{"int"},
			},
		},
	}
}

func TestRenderCompositions(t *testing.T) {
	parser := getEmptyParser("main")
	st := &Struct{
		PackageName: "main",
		Composition: map[string]struct{}{
			"foopack.AnotherClass": {},
		},
		Extends: map[string]struct{}{
			"foopack.YetAnotherClass": {},
		},
	}
	extendsBuilder := &LineStringBuilder{}
	parser.renderCompositions(st, "TestClass", extendsBuilder)
	expectedResult := "foopack.AnotherClass *-- main.TestClass\n"
	if extendsBuilder.String() != expectedResult {
		t.Errorf("TestRenderCompositions: Expected %s got %s", expectedResult, extendsBuilder.String())
	}

	st = &Struct{
		PackageName: "main",
		Composition: map[string]struct{}{
			"AnotherClass": {},
		},
	}
	extendsBuilder = &LineStringBuilder{}
	parser.renderCompositions(st, "TestClass", extendsBuilder)
	expectedResult = "main.AnotherClass *-- main.TestClass\n"
	if extendsBuilder.String() != expectedResult {
		t.Errorf("TestRenderCompositions: Expected %s got %s", expectedResult, extendsBuilder.String())
	}

	st = &Struct{
		PackageName: "main",
		Composition: map[string]struct{}{
			"int": {},
		},
	}
	extendsBuilder = &LineStringBuilder{}
	parser.renderCompositions(st, "TestClass", extendsBuilder)
	expectedResult = builtinPackageName + ".int *-- main.TestClass\n"
	if extendsBuilder.String() != expectedResult {
		t.Errorf("TestRenderCompositions: Expected %s got %s", expectedResult, extendsBuilder.String())
	}
}
func TestRenderExtends(t *testing.T) {
	parser := getEmptyParser("main")
	st := &Struct{
		PackageName: "main",
		Extends: map[string]struct{}{
			"foopack.AnotherClass": {},
		},
	}
	extendsBuilder := &LineStringBuilder{}
	parser.renderExtends(st, "TestClass", extendsBuilder)
	expectedResult := "foopack.AnotherClass <|-- main.TestClass\n"
	if extendsBuilder.String() != expectedResult {
		t.Errorf("TestRenderExtends: Expected %s got %s", expectedResult, extendsBuilder.String())
	}

	st = &Struct{
		PackageName: "main",
		Extends: map[string]struct{}{
			"AnotherClass": {},
		},
	}
	extendsBuilder = &LineStringBuilder{}
	parser.renderExtends(st, "TestClass", extendsBuilder)
	expectedResult = "main.AnotherClass <|-- main.TestClass\n"
	if extendsBuilder.String() != expectedResult {
		t.Errorf("TestRenderExtends: Expected %s got %s", expectedResult, extendsBuilder.String())
	}
}
func TestRenderStructMethods(t *testing.T) {

	parser := getEmptyParser("main")

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
			{
				Name: "Bar",
				Parameters: []*Field{
					{
						Type: "int",
					},
					{
						Type: "string",
					},
				},
				ReturnValues: []string{"int"},
			},
		},
	}
	privateFunctions := &LineStringBuilder{}
	publicFunctions := &LineStringBuilder{}
	parser.renderStructMethods(st, privateFunctions, publicFunctions)
	if privateFunctions.String() != "        - foo( int,  string) (error, int)\n" {
		t.Errorf("TestRenderStructMethods: expected privateFields to be [        - foo( int,  string) (error, int)\\n] got [%v]", privateFunctions.String())
	}
	if publicFunctions.String() != "        + Bar( int,  string) int\n" {
		t.Errorf("TestRenderStructMethods: expected publicFields to be [        + Bar( int,  string) int\\n] got [%v]", publicFunctions.String())
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

func TestWriteWithLineDepth(t *testing.T) {
	b := &LineStringBuilder{}
	b.WriteLineWithDepth(1, "Hello Test")
	expected := "    Hello Test\n"
	if b.String() != expected {
		t.Errorf("TestWriteWithLineTest: expected %s, got %s", expected, b.String())
	}
}

func TestNewClassDiagram(t *testing.T) {
	tt := []struct {
		Name            string
		Path            string
		ExpectedError   string
		Recursive       bool
		ExpectedStructs []struct {
			Name   string
			Type   string
			Exists bool
		}
	}{
		{
			Name:          "Directory Missing not recursive",
			ExpectedError: "open ./no_path: no such file or directory",
			Path:          "./no_path",
			Recursive:     false,
		},
		{
			Name:          "Directory Missing recursive",
			ExpectedError: "lstat ./no_path: no such file or directory",
			Path:          "./no_path",
			Recursive:     true,
		},
		{
			Name:          "Recursive",
			ExpectedError: "",
			Path:          "../testingsupport",
			Recursive:     true,
			ExpectedStructs: []struct {
				Name   string
				Type   string
				Exists bool
			}{
				{
					Name:   "testingsupport.test",
					Type:   "class",
					Exists: true,
				},
				{
					Name:   "subfolder.test2",
					Type:   "interface",
					Exists: true,
				},
			},
		},
		{
			Name:          "Not Recursive",
			ExpectedError: "",
			Path:          "../testingsupport",
			Recursive:     false,
			ExpectedStructs: []struct {
				Name   string
				Type   string
				Exists bool
			}{
				{
					Name:   "testingsupport.test",
					Type:   "class",
					Exists: true,
				},
				{
					Name:   "subfolder.test2",
					Type:   "interface",
					Exists: false,
				},
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			parser, err := NewClassDiagram([]string{tc.Path}, tc.Recursive)

			if tc.ExpectedError != "" {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if err.Error() != tc.ExpectedError {
					t.Errorf("Expected error to be %s, got %s", tc.ExpectedError, err.Error())
					return
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %s", err.Error())
					return
				}
				for _, st := range tc.ExpectedStructs {
					stt := parser.getStruct(st.Name)
					if st.Exists {
						if stt == nil || stt.Type != st.Type {
							t.Errorf("Expected structure %v to exist with the correct type, but got %v", st, stt)
						}
					} else {
						if stt != nil {
							t.Errorf("Expected %s to not exists but got %v", st.Name, stt)
						}
					}
				}
			}
		})
	}
}

func TestRender(t *testing.T) {

	parser, err := NewClassDiagram([]string{"../testingsupport"}, false)
	if err != nil {
		t.Errorf("TestRender: expected no errors, got %s", err.Error())
		return
	}

	resultRender := parser.Render()
	result, err := ioutil.ReadFile("../testingsupport/testingsupport.puml")
	if err != nil {
		t.Errorf("TestRender: expected no errors reading testing file, got %s", err.Error())
	}
	resultAlt, err := ioutil.ReadFile("../testingsupport/testingsupport alt.puml")
	if string(result) != resultRender && string(resultAlt) != resultRender {
		t.Errorf("TestRender: Expected renders to be the same as %s or %s, but got %s", result, resultAlt, resultRender)
	}
}

func TestGetPackageName(t *testing.T) {
	p := getEmptyParser("main")
	s := &Struct{
		PackageName: "main",
	}
	ty := p.getPackageName("int", s)
	if ty != builtinPackageName {
		t.Errorf("TestGetPackageName: expecting [%s], got [%s]", builtinPackageName, ty)
	}
}

func TestMultipleFolders(t *testing.T) {
	parser, err := NewClassDiagram([]string{"../testingsupport/subfolder", "../testingsupport/subfolder2"}, false)

	if err != nil {
		t.Errorf("TestMultipleFolders: expected no errors, got %s", err.Error())
		return
	}

	resultRender := parser.Render()
	result, err := ioutil.ReadFile("../testingsupport/subfolder1-2.puml")
	if err != nil {
		t.Errorf("TestMultipleFolders: expected no errors reading testing file, got %s", err.Error())
	}
	resultAlt, err := ioutil.ReadFile("../testingsupport/subfolder1-2alt.puml")
	if string(result) != resultRender && string(resultAlt) != resultRender {
		t.Errorf("TestMultipleFolders: Expected renders to be the same as %s or %s, but got %s", result, resultAlt, resultRender)
	}
}
