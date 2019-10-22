/*
Package parser generates PlantUml http://plantuml.com/ Class diagrams for your golang projects
The main structure is the ClassParser which you can generate by calling the NewClassDiagram(dir)
function.

Pass the directory where the .go files are and the parser will analyze the code and build a structure
containing the information it needs to Render the class diagram.

call the Render() function and this will return a string with the class diagram.

See github.com/jfeliu007/goplantuml/cmd/goplantuml/main.go for a command that uses this functions and outputs the text to
the console.

*/
package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

//LineStringBuilder extends the strings.Builder and adds functionality to build a string with tabs and
//adding new lines
type LineStringBuilder struct {
	strings.Builder
}

const tab = "    "
const builtinPackageName = "__builtin__"

//WriteLineWithDepth will write the given text with added tabs at the beginning into the string builder.
func (lsb *LineStringBuilder) WriteLineWithDepth(depth int, str string) {
	lsb.WriteString(strings.Repeat(tab, depth))
	lsb.WriteString(str)
	lsb.WriteString("\n")
}

//RenderingOptions will allow the class parser to optionally enebale or disable the things to render.
type RenderingOptions struct {
	Aggregation bool
}

//ClassParser contains the structure of the parsed files. The structure is a map of package_names that contains
//a map of structure_names -> Structs
type ClassParser struct {
	renderingOptions   *RenderingOptions
	structure          map[string]map[string]*Struct
	currentPackageName string
	allInterfaces      map[string]struct{}
	allStructs         map[string]struct{}
	allImports         map[string]string
	allAliases         map[string]*Alias
}

//NewClassDiagram returns a new classParser with which can Render the class diagram of
// files int eh given directory
func NewClassDiagram(directoryPaths []string, ignoreDirectories []string, recursive bool) (*ClassParser, error) {
	classParser := &ClassParser{
		renderingOptions: &RenderingOptions{},
		structure:        make(map[string]map[string]*Struct),
		allInterfaces:    make(map[string]struct{}),
		allStructs:       make(map[string]struct{}),
		allImports:       make(map[string]string),
		allAliases:       make(map[string]*Alias),
	}
	ignoreDirectoryMap := map[string]struct{}{}
	for _, dir := range ignoreDirectories {
		ignoreDirectoryMap[dir] = struct{}{}
	}
	for _, directoryPath := range directoryPaths {
		if recursive {
			err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" {
						return filepath.SkipDir
					}
					if _, ok := ignoreDirectoryMap[path]; ok {
						return filepath.SkipDir
					}
					classParser.parseDirectory(path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			err := classParser.parseDirectory(directoryPath)
			if err != nil {
				return nil, err
			}
		}
	}

	for s := range classParser.allStructs {
		st := classParser.getStruct(s)
		if st != nil {
			for i := range classParser.allInterfaces {
				inter := classParser.getStruct(i)
				if st.ImplementsInterface(inter) {
					st.AddToExtends(i)
				}
			}
		}
	}
	return classParser, nil
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
			for _, d := range f.Imports {
				p.parseImports(d)
			}
			for _, d := range f.Decls {
				p.parseFileDeclarations(d)
			}
		}
	}
}

func (p *ClassParser) parseImports(impt *ast.ImportSpec) {
	if impt.Name != nil {
		splitPath := strings.Split(impt.Path.Value, "/")
		s := strings.TrimRight(splitPath[len(splitPath)-1], `"`)
		p.allImports[impt.Name.Name] = s
	}
}

func (p *ClassParser) parseDirectory(directoryPath string) error {
	fs := token.NewFileSet()
	result, err := parser.ParseDir(fs, directoryPath, nil, 0)
	if err != nil {
		return err
	}
	for _, v := range result {
		p.parsePackage(v)
	}
	return nil
}

//parse the given declaration looking for classes, interfaces, or member functions
func (p *ClassParser) parseFileDeclarations(node ast.Decl) {
	switch decl := node.(type) {
	case *ast.GenDecl:
		spec := decl.Specs[0]
		var declarationType string
		var typeName string
		var alias *Alias
		switch v := spec.(type) {
		case *ast.TypeSpec:
			typeName = v.Name.Name
			switch c := v.Type.(type) {
			case *ast.StructType:
				declarationType = "class"
				for _, f := range c.Fields.List {
					p.getOrCreateStruct(typeName).AddField(f, p.allImports)
				}
			case *ast.InterfaceType:
				declarationType = "interface"
				for _, f := range c.Methods.List {
					switch t := f.Type.(type) {
					case *ast.FuncType:
						p.getOrCreateStruct(typeName).AddMethod(f, p.allImports)
						break
					case *ast.Ident:
						f, _ := getFieldType(t, p.allImports)
						st := p.getOrCreateStruct(typeName)
						f = replacePackageConstant(f, st.PackageName)
						st.AddToComposition(f)
						break
					}
				}
			default:
				declarationType = "alias"
				aliasType, _ := getFieldType(c, p.allImports)
				if !isPrimitiveString(typeName) {
					typeName = fmt.Sprintf("%s.%s", p.currentPackageName, typeName)
				}
				alias = getNewAlias(aliasType, p.currentPackageName, typeName)
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
		case "class":
			p.allStructs[fullName] = struct{}{}
		case "alias":
			p.allAliases[typeName] = alias
		}
	case *ast.FuncDecl:
		if decl.Recv != nil {
			// Only get in when the function is defined for a structure. Global functions are not needed for class diagram
			theType, _ := getFieldType(decl.Recv.List[0].Type, p.allImports)
			theType = replacePackageConstant(theType, "")
			if theType[0] == "*"[0] {
				theType = theType[1:]
			}
			structure := p.getOrCreateStruct(theType)
			if structure.Type == "" {
				structure.Type = "class"
			}

			fullName := fmt.Sprintf("%s.%s", p.currentPackageName, theType)
			p.allStructs[fullName] = struct{}{}
			structure.AddMethod(&ast.Field{
				Names:   []*ast.Ident{decl.Name},
				Doc:     decl.Doc,
				Type:    decl.Type,
				Tag:     nil,
				Comment: nil,
			}, p.allImports)
		}
	}
}

//Render returns a string of the class diagram that this parser has generated.
func (p *ClassParser) Render() string {
	str := &LineStringBuilder{}
	str.WriteLineWithDepth(0, "@startuml")
	for pack, structures := range p.structure {
		p.renderStructures(pack, structures, str)

	}
	for name, alias := range p.allAliases {
		renderAlias(name, alias, str)
	}
	str.WriteLineWithDepth(0, "@enduml")
	return str.String()
}

func (p *ClassParser) renderStructures(pack string, structures map[string]*Struct, str *LineStringBuilder) {
	if len(structures) > 0 {
		composition := &LineStringBuilder{}
		extends := &LineStringBuilder{}
		aggregations := &LineStringBuilder{}
		str.WriteLineWithDepth(0, fmt.Sprintf(`namespace %s {`, pack))
		for name, structure := range structures {
			p.renderStructure(structure, pack, name, str, composition, extends, aggregations)
		}
		str.WriteLineWithDepth(0, fmt.Sprintf(`}`))
		str.WriteLineWithDepth(0, composition.String())
		str.WriteLineWithDepth(0, extends.String())
		if p.renderingOptions.Aggregation {
			str.WriteLineWithDepth(0, aggregations.String())
		}
	}
}

func renderAlias(name string, alias *Alias, str *LineStringBuilder) {
	str.WriteLineWithDepth(0, fmt.Sprintf("%s #.. %s", alias.Name, alias.AliasOf))
}

func (p *ClassParser) renderStructure(structure *Struct, pack string, name string, str *LineStringBuilder, composition *LineStringBuilder, extends *LineStringBuilder, aggregations *LineStringBuilder) {

	privateFields := &LineStringBuilder{}
	publicFields := &LineStringBuilder{}
	privateMethods := &LineStringBuilder{}
	publicMethods := &LineStringBuilder{}
	sType := ""
	renderStructureType := structure.Type
	switch structure.Type {
	case "class":
		sType = "<< (S,Aquamarine) >>"
	case "alias":
		sType = "<< (T, #FF7700) >> "
		renderStructureType = "class"

	}
	str.WriteLineWithDepth(1, fmt.Sprintf(`%s %s %s {`, renderStructureType, name, sType))
	p.renderStructFields(structure, privateFields, publicFields)
	p.renderStructMethods(structure, privateMethods, publicMethods)
	p.renderCompositions(structure, name, composition)
	p.renderExtends(structure, name, extends)
	p.renderAggregations(structure, name, aggregations)
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

func (p *ClassParser) renderCompositions(structure *Struct, name string, composition *LineStringBuilder) {

	for c := range structure.Composition {
		if !strings.Contains(c, ".") {
			c = fmt.Sprintf("%s.%s", p.getPackageName(c, structure), c)
		}
		composition.WriteLineWithDepth(0, fmt.Sprintf(`%s *-- %s.%s`, c, structure.PackageName, name))
	}
}

func (p *ClassParser) renderAggregations(structure *Struct, name string, aggregations *LineStringBuilder) {

	for a := range structure.Aggregations {
		if !strings.Contains(a, ".") {
			a = fmt.Sprintf("%s.%s", p.getPackageName(a, structure), a)
		}
		if p.getPackageName(a, structure) != builtinPackageName {
			aggregations.WriteLineWithDepth(0, fmt.Sprintf(`%s.%s o-- %s`, structure.PackageName, name, a))
		}
	}
}

func (p *ClassParser) getPackageName(t string, st *Struct) string {

	packageName := st.PackageName
	if isPrimitiveString(t) {
		packageName = builtinPackageName
	}
	return packageName
}
func (p *ClassParser) renderExtends(structure *Struct, name string, extends *LineStringBuilder) {

	for c := range structure.Extends {
		if !strings.Contains(c, ".") {
			c = fmt.Sprintf("%s.%s", structure.PackageName, c)
		}
		extends.WriteLineWithDepth(0, fmt.Sprintf(`%s <|-- %s.%s`, c, structure.PackageName, name))
	}
}

func (p *ClassParser) renderStructMethods(structure *Struct, privateMethods *LineStringBuilder, publicMethods *LineStringBuilder) {

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
		if len(method.ReturnValues) > 0 {
			if len(method.ReturnValues) == 1 {
				returnValues = method.ReturnValues[0]
			} else {
				returnValues = fmt.Sprintf("(%s)", strings.Join(method.ReturnValues, ", "))
			}
		}
		if accessModifier == "-" {
			privateMethods.WriteLineWithDepth(2, fmt.Sprintf(`%s %s(%s) %s`, accessModifier, method.Name, strings.Join(parameterList, ", "), returnValues))
		} else {
			publicMethods.WriteLineWithDepth(2, fmt.Sprintf(`%s %s(%s) %s`, accessModifier, method.Name, strings.Join(parameterList, ", "), returnValues))
		}
	}
}

func (p *ClassParser) renderStructFields(structure *Struct, privateFields *LineStringBuilder, publicFields *LineStringBuilder) {
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
}

// Returns an initialized struct of the given name or returns the existing one if it was already created
func (p *ClassParser) getOrCreateStruct(name string) *Struct {
	result, ok := p.structure[p.currentPackageName][name]
	if !ok {
		result = &Struct{
			PackageName:  p.currentPackageName,
			Functions:    make([]*Function, 0),
			Fields:       make([]*Field, 0),
			Type:         "",
			Composition:  make(map[string]struct{}, 0),
			Extends:      make(map[string]struct{}, 0),
			Aggregations: make(map[string]struct{}, 0),
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

//SetRenderingOptions Sets the rendering options for the Render() Function
func (p *ClassParser) SetRenderingOptions(ro *RenderingOptions) {
	p.renderingOptions = ro
}
