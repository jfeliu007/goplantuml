package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"unicode"
)

//LineStringBuilder extends the strings.Builder and adds functionality to build a string with tabs and
//adding new lines
type LineStringBuilder struct {
	strings.Builder
}

const tab = "    "

//WriteLineWithDepth will write the given text with added tabs at the begining into the string builder.
func (lsb *LineStringBuilder) WriteLineWithDepth(depth int, str string) {
	lsb.WriteString(strings.Repeat(tab, depth))
	lsb.WriteString(str)
	lsb.WriteString("\n")
}

//ClassParser contains the structure of the parsed files. The structure is a map of package_names that contains
//a map of structure_names -> Structs
type ClassParser struct {
	structure          map[string]map[string]*Struct
	currentPackageName string
	allInterfaces      map[string]struct{}
	allStructs         map[string]struct{}
}

//Parameter can hold the name and type of any field
type Parameter struct {
	Name string
	Type string
}

//Function holds the signature of a function with name, Parameters and Return values
type Function struct {
	Name         string
	Parameters   []*Parameter
	ReturnValues []string
}

//Struct represent a struct in golang, it can be of Type "class" or "interface" and can be associated
//with other structs via Composition and Extends
type Struct struct {
	Functions   []*Function
	Fields      []*Parameter
	Type        string
	Composition []string
	Extends     []string
}

//NewClassDiagram returns a new classParser with which can Render the class diagram of
// files int eh given directory
func NewClassDiagram(directoryPath string) (*ClassParser, error) {
	classParser := &ClassParser{
		structure:     make(map[string]map[string]*Struct),
		allInterfaces: make(map[string]struct{}),
		allStructs:    make(map[string]struct{}),
	}
	fs := token.NewFileSet()
	result, err := parser.ParseDir(fs, directoryPath, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, v := range result {
		classParser.parsePackage(v)
	}
	for s := range classParser.allStructs {
		st := classParser.getStruct(s)
		if st != nil {
			for i := range classParser.allInterfaces {
				inter := classParser.getStruct(i)
				if classParser.structImplementsInterface(st, inter) {
					classParser.addToExtends(st, i)
				}
			}
		}
	}
	return classParser, nil
}

// returns true if the struct st conforms ot the given interface
func (p *ClassParser) structImplementsInterface(st *Struct, inter *Struct) bool {
	if len(inter.Functions) == 0 {
		return false
	}
	for _, f1 := range inter.Functions {
		foundMatch := false
		for _, f2 := range st.Functions {
			if signturesAreEqual(f1, f2) {
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

//Returns true if the two functions have the same signature (parameter names are not checked)
func signturesAreEqual(f1, f2 *Function) bool {
	result := true
	result = result && (f2.Name == f1.Name)
	result = result && reflect.DeepEqual(f1.ReturnValues, f2.ReturnValues)
	result = result && (len(f1.Parameters) == len(f2.Parameters))
	if result {
		for i, p := range f1.Parameters {
			if p.Type != f2.Parameters[i].Type {
				return false
			}
		}
	}
	return result
}

//parse the given ast.Package into the ClassParser structure
func (p *ClassParser) parsePackage(node ast.Node) {
	pack := node.(*ast.Package)
	p.currentPackageName = pack.Name
	_, ok := p.structure[p.currentPackageName]
	if !ok {
		p.structure[p.currentPackageName] = make(map[string]*Struct)
	}
	for fileName, f := range pack.Files {
		if !strings.HasSuffix(fileName, "_test.go") {
			for _, d := range f.Decls {
				p.parseFileDeclarations(d)
			}
		}
	}
}

//parse the given declaration looking for classes, interfaces, or member functions
func (p *ClassParser) parseFileDeclarations(node ast.Decl) {
	switch decl := node.(type) {
	case *ast.GenDecl:
		spec := decl.Specs[0]
		var declarationType string
		var typeName string
		switch v := spec.(type) {
		case *ast.TypeSpec:
			typeName = v.Name.Name
			switch c := v.Type.(type) {
			case *ast.StructType:
				declarationType = "class"
				for _, f := range c.Fields.List {
					p.addFieldToStruct(p.getOrCreateStruct(typeName), f)
				}
				break
			case *ast.InterfaceType:
				declarationType = "interface"
				for _, f := range c.Methods.List {
					p.addMethodToStruct(p.getOrCreateStruct(typeName), f)
				}
				break
			default:
				// Not needed for class diagrams (Imports, global variables, regular functions, etc)
				return
			}
		default:
			// Not needed for class diagrams (Imports, global variables, regular functions, etc)
			return
		}
		p.getOrCreateStruct(typeName).Type = declarationType
		fullName := fmt.Sprintf("%s.%s", p.currentPackageName, typeName)
		switch declarationType {
		case "interface":
			p.allInterfaces[fullName] = struct{}{}
			break
		case "class":
			p.allStructs[fullName] = struct{}{}
			break
		}
		break
	case *ast.FuncDecl:
		if decl.Recv != nil {
			// Only get in when the function is defined for a structure. Global functions are not needed for class diagram
			theType := p.getFieldType(decl.Recv.List[0].Type, false)
			if theType[0] == "*"[0] {
				theType = theType[1:]
			}
			structure := p.getOrCreateStruct(theType)
			if structure.Type == "" {
				structure.Type = "class"
			}
			p.addMethodToStruct(structure, &ast.Field{
				Names:   []*ast.Ident{decl.Name},
				Doc:     decl.Doc,
				Type:    decl.Type,
				Tag:     nil,
				Comment: nil,
			})
		}
		break
	}
}

// Parse the Field and if it is an ast.FuncType, then add the methods into the structure
func (p *ClassParser) addMethodToStruct(s *Struct, method *ast.Field) {
	f, ok := method.Type.(*ast.FuncType)
	if !ok {
		return
	}
	function := p.getFunction(f, method.Names[0].Name)
	s.Functions = append(s.Functions, function)
}

// generate and return a function object from the given Functype. The names must be passed to this
// function since the FuncType does not have this information
func (p *ClassParser) getFunction(f *ast.FuncType, name string) *Function {
	function := &Function{
		Name:         name,
		Parameters:   make([]*Parameter, 0),
		ReturnValues: make([]string, 0),
	}
	params := f.Params
	if params != nil {
		for _, pa := range params.List {
			fieldName := ""
			if pa.Names != nil {
				fieldName = pa.Names[0].Name
			}
			function.Parameters = append(function.Parameters, &Parameter{
				Name: fieldName,
				Type: p.getFieldType(pa.Type, false),
			})
		}
	}

	results := f.Results
	if results != nil {
		for _, pa := range results.List {
			function.ReturnValues = append(function.ReturnValues, p.getFieldType(pa.Type, false))
		}
	}
	return function
}

// analize the field and if the field is a composition field, then adds a composition
// entry into the structure, otherwhise, add the corresponding field information into the structure
func (p *ClassParser) addFieldToStruct(s *Struct, field *ast.Field) {
	if field.Names != nil {
		s.Fields = append(s.Fields, &Parameter{
			Name: field.Names[0].Name,
			Type: p.getFieldType(field.Type, false),
		})
	} else if field.Type != nil {
		fType := p.getFieldType(field.Type, true)
		if fType[0] == "*"[0] {
			fType = fType[1:]
		}
		p.addToComposition(s, fType)
	}
}

//add the composition relation to the structure. We want to make sure that *ExampleStruct
//gets added as ExampleStruct so that we can properly build the relation later to the
//class identifier
func (p *ClassParser) addToComposition(s *Struct, fType string) {
	if len(fType) == 0 {
		return
	}
	if fType[0] == "*"[0] {
		fType = fType[1:]
	}
	s.Composition = append(s.Composition, fType)
}

//Adds an extends relationship to this struct. We want to make sure that *ExampleStruct
//gets added as ExampleStruct so that we can properly build the relation later to the
//class identifier
func (p *ClassParser) addToExtends(s *Struct, fType string) {
	if len(fType) == 0 {
		return
	}
	if fType[0] == "*"[0] {
		fType = fType[1:]
	}
	s.Extends = append(s.Extends, fType)
}

//Render returns a string of the class diagram that this parser has generated.
func (p *ClassParser) Render() string {
	str := &LineStringBuilder{}
	str.WriteLineWithDepth(0, "@startuml")
	for pack, structures := range p.structure {
		composition := &LineStringBuilder{}
		extends := &LineStringBuilder{}
		if len(structures) > 0 {
			str.WriteLineWithDepth(0, fmt.Sprintf(`namespace %s {`, pack))
			for name, structure := range structures {
				privateFields := &LineStringBuilder{}
				publicFields := &LineStringBuilder{}
				privateMethods := &LineStringBuilder{}
				publicMethods := &LineStringBuilder{}
				str.WriteLineWithDepth(1, fmt.Sprintf(`%s %s {`, structure.Type, name))
				for _, field := range structure.Fields {
					accessModifier := "+"
					if unicode.IsLower(rune(field.Name[0])) {
						accessModifier = "-"
					}
					if accessModifier == "-" {
						privateFields.WriteLineWithDepth(2, fmt.Sprintf(`%s %s %s`, accessModifier, field.Name, field.Type))
					} else {
						publicFields.WriteLineWithDepth(2, fmt.Sprintf(`%s %s %s`, accessModifier, field.Name, field.Type))
					}
				}
				for _, c := range structure.Composition {
					composition.WriteLineWithDepth(0, fmt.Sprintf(`%s *-- %s.%s`, c, pack, name))
				}
				for _, c := range structure.Extends {
					extends.WriteLineWithDepth(0, fmt.Sprintf(`%s <|-- %s.%s`, c, pack, name))
				}
				for _, method := range structure.Functions {
					accessModifier := "+"
					if unicode.IsLower(rune(method.Name[0])) {
						accessModifier = "-"
					}
					parameterList := make([]string, 0)
					for _, p := range method.Parameters {
						parameterList = append(parameterList, fmt.Sprintf("%s %s", p.Name, p.Type))
					}
					returnValues := ""
					if len(method.ReturnValues) > 1 {
						returnValues = fmt.Sprintf("(%s)", strings.Join(method.ReturnValues, ", "))
					}
					if accessModifier == "-" {
						privateMethods.WriteLineWithDepth(2, fmt.Sprintf(`%s %s(%s) %s`, accessModifier, method.Name, strings.Join(parameterList, ", "), returnValues))
					} else {
						publicMethods.WriteLineWithDepth(2, fmt.Sprintf(`%s %s(%s) %s`, accessModifier, method.Name, strings.Join(parameterList, ", "), returnValues))
					}
				}
				if privateFields.Len() > 0 {
					str.WriteLineWithDepth(0, privateFields.String())
				}
				if publicFields.Len() > 0 {
					str.WriteLineWithDepth(0, publicFields.String())
				}
				if privateMethods.Len() > 0 {
					str.WriteLineWithDepth(0, privateMethods.String())
				}
				if publicMethods.Len() > 0 {
					str.WriteLineWithDepth(0, publicMethods.String())
				}
				str.WriteLineWithDepth(1, fmt.Sprintf(`}`))
			}
			str.WriteLineWithDepth(0, fmt.Sprintf(`}`))
			str.WriteLineWithDepth(0, composition.String())
			str.WriteLineWithDepth(0, extends.String())
		}

	}
	str.WriteString("@enduml")
	return str.String()
}

// Returns an initialized struct of the given name or returns the existing one if it was already created
func (p *ClassParser) getOrCreateStruct(name string) *Struct {
	result, ok := p.structure[p.currentPackageName][name]
	if !ok {
		result = &Struct{
			Functions:   make([]*Function, 0),
			Fields:      make([]*Parameter, 0),
			Type:        "",
			Composition: make([]string, 0),
			Extends:     make([]string, 0),
		}
		p.structure[p.currentPackageName][name] = result
	}
	return result
}

// Returns an existing struct only if it was created. nil otherwhise
func (p *ClassParser) getStruct(structName string) *Struct {
	split := strings.SplitN(structName, ".", 2)
	pack, ok := p.structure[split[0]]
	if !ok {
		return nil
	}
	return pack[split[1]]
}

//Returns a string representation of the given expression if it was recognized.
//Refer to the implementation to see the different string representations.
func (p *ClassParser) getFieldType(exp ast.Expr, includePackageName bool) string {
	packageName := ""
	if includePackageName {
		packageName = fmt.Sprintf("%s.", p.currentPackageName)
	}
	switch v := exp.(type) {
	case *ast.Ident:
		return fmt.Sprintf("%s%s", packageName, v.Name)
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", p.getFieldType(v.Elt, false))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", v.X.(*ast.Ident).Name, p.getFieldType(v.Sel, false))
	case *ast.MapType:
		return fmt.Sprintf("<font color=blue>map</font>[%s]%s", p.getFieldType(v.Key, false), p.getFieldType(v.Value, false))
	case *ast.StarExpr:
		return fmt.Sprintf("*%s%s", packageName, p.getFieldType(v.X, false))
	case *ast.ChanType:
		return fmt.Sprintf("<font color=blue>chan</font> %s", p.getFieldType(v.Value, false))
	case *ast.StructType:
		fieldList := make([]string, 0)
		for _, field := range v.Fields.List {
			fieldList = append(fieldList, p.getFieldType(field.Type, false))
		}
		return fmt.Sprintf("<font color=blue>struct</font>{%s}", strings.Join(fieldList, ", "))
	case *ast.InterfaceType:
		methods := make([]string, 0)
		for _, field := range v.Methods.List {
			methodName := ""
			if field.Names != nil {
				methodName = field.Names[0].Name
			}
			methods = append(methods, methodName+" "+p.getFieldType(field.Type, false))
		}
		return fmt.Sprintf("<font color=blue>interface</font>{%s}", strings.Join(methods, "; "))
	case *ast.FuncType:
		function := p.getFunction(v, "")
		params := make([]string, 0)
		for _, pa := range function.Parameters {
			params = append(params, pa.Type)
		}
		returns := ""
		returnList := make([]string, 0)
		for _, re := range function.ReturnValues {
			returnList = append(returnList, re)
		}
		if len(returnList) > 1 {
			returns = fmt.Sprintf("(%s)", strings.Join(returnList, ", "))
		} else {
			returns = strings.Join(returnList, "")
		}
		return fmt.Sprintf("<font color=blue>func</font>(%s) %s", strings.Join(params, ", "), returns)
	case *ast.Ellipsis:
		return fmt.Sprintf("<...%s", p.getFieldType(v.Elt, false))
	}
	return ""
}
