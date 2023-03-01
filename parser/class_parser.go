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
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/afero"
)

// LineStringBuilder extends the strings.Builder and adds functionality to build a string with tabs and
// adding new lines
type LineStringBuilder struct {
	strings.Builder
}

const tab = "    "
const builtinPackageName = "__builtin__"
const implements = `"implements"`
const extends = `"extends"`
const aggregates = `"uses"`
const aliasOf = `"alias of"`

// WriteLineWithDepth will write the given text with added tabs at the beginning into the string builder.
func (lsb *LineStringBuilder) WriteLineWithDepth(depth int, str string) {
	lsb.WriteString(strings.Repeat(tab, depth))
	lsb.WriteString(str)
	lsb.WriteString("\n")
}

// ClassDiagramOptions will provide a way for callers of the NewClassDiagramFs() function to pass all the necessary arguments.
type ClassDiagramOptions struct {
	FileSystem         afero.Fs
	Directories        []string
	IgnoredDirectories []string
	RenderingOptions   map[RenderingOption]interface{}
	Recursive          bool
}

// RenderingOptions will allow the class parser to optionally enebale or disable the things to render.
type RenderingOptions struct {
	Title                   string
	Notes                   string
	Aggregations            bool
	Fields                  bool
	Methods                 bool
	Compositions            bool
	Implementations         bool
	Aliases                 bool
	ConnectionLabels        bool
	AggregatePrivateMembers bool
	PrivateMembers          bool
}

const aliasComplexNameComment = "'This class was created so that we can correctly have an alias pointing to this name. Since it contains dots that can break namespaces"

const (
	// RenderAggregations is to be used in the SetRenderingOptions argument as the key to the map, when value is true, it will set the parser to render aggregations
	RenderAggregations RenderingOption = iota

	// RenderCompositions is to be used in the SetRenderingOptions argument as the key to the map, when value is true, it will set the parser to render compositions
	RenderCompositions

	// RenderImplementations is to be used in the SetRenderingOptions argument as the key to the map, when value is true, it will set the parser to render implementations
	RenderImplementations

	// RenderAliases is to be used in the SetRenderingOptions argument as the key to the map, when value is true, it will set the parser to render aliases
	RenderAliases

	// RenderFields is to be used in the SetRenderingOptions argument as the key to the map, when value is true, it will set the parser to render fields
	RenderFields

	// RenderMethods is to be used in the SetRenderingOptions argument as the key to the map, when value is true, it will set the parser to render methods
	RenderMethods

	// RenderConnectionLabels is to be used in the SetRenderingOptions argument as the key to the map, when value is true, it will set the parser to render the connection labels
	RenderConnectionLabels

	// RenderTitle is the options for the Title of the diagram. The value of this will be rendered as a title unless empty
	RenderTitle

	// RenderNotes contains a list of notes to be rendered in the class diagram
	RenderNotes

	// AggregatePrivateMembers is to be used in the SetRenderingOptions argument as the key to the map, when value is true, it will connect aggregations with private members
	AggregatePrivateMembers

	// RenderPrivateMembers is used if private members (fields, methods) should be rendered
	RenderPrivateMembers
)

// RenderingOption is an alias for an it so it is easier to use it as options in a map (see SetRenderingOptions(map[RenderingOption]bool) error)
type RenderingOption int

// ClassParser contains the structure of the parsed files. The structure is a map of package_names that contains
// a map of structure_names -> Structs
type ClassParser struct {
	renderingOptions   *RenderingOptions
	structure          map[string]map[string]*Struct
	currentPackageName string
	allInterfaces      map[string]struct{}
	allStructs         map[string]struct{}
	allImports         map[string]string
	allAliases         map[string]*Alias
	allRenamedStructs  map[string]map[string]string
}

// NewClassDiagramWithOptions returns a new classParser with which can Render the class diagram of
// files in the given directory passed in the ClassDiargamOptions. This will also alow for different types of FileSystems
// Passed since it is part of the ClassDiagramOptions as well.
func NewClassDiagramWithOptions(options *ClassDiagramOptions) (*ClassParser, error) {
	classParser := &ClassParser{
		renderingOptions: &RenderingOptions{
			Aggregations:     false,
			Fields:           true,
			Methods:          true,
			Compositions:     true,
			Implementations:  true,
			Aliases:          true,
			ConnectionLabels: false,
			Title:            "",
			Notes:            "",
		},
		structure:         make(map[string]map[string]*Struct),
		allInterfaces:     make(map[string]struct{}),
		allStructs:        make(map[string]struct{}),
		allImports:        make(map[string]string),
		allAliases:        make(map[string]*Alias),
		allRenamedStructs: make(map[string]map[string]string),
	}
	ignoreDirectoryMap := map[string]struct{}{}
	for _, dir := range options.IgnoredDirectories {
		ignoreDirectoryMap[dir] = struct{}{}
	}
	for _, directoryPath := range options.Directories {
		if options.Recursive {
			err := afero.Walk(options.FileSystem, directoryPath, func(path string, info os.FileInfo, err error) error {
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
	classParser.SetRenderingOptions(options.RenderingOptions)
	return classParser, nil
}

// NewClassDiagram returns a new classParser with which can Render the class diagram of
// files in the given directory
func NewClassDiagram(directoryPaths []string, ignoreDirectories []string, recursive bool) (*ClassParser, error) {
	options := &ClassDiagramOptions{
		Directories:        directoryPaths,
		IgnoredDirectories: ignoreDirectories,
		Recursive:          recursive,
		RenderingOptions:   map[RenderingOption]interface{}{},
		FileSystem:         afero.NewOsFs(),
	}
	return NewClassDiagramWithOptions(options)
}

// parse the given ast.Package into the ClassParser structure
func (p *ClassParser) parsePackage(node ast.Node) {
	pack := node.(*ast.Package)
	p.currentPackageName = pack.Name
	_, ok := p.structure[p.currentPackageName]
	if !ok {
		p.structure[p.currentPackageName] = make(map[string]*Struct)
	}
	var sortedFiles []string
	for fileName := range pack.Files {
		sortedFiles = append(sortedFiles, fileName)
	}

	sort.Strings(sortedFiles)
	for _, fileName := range sortedFiles {
		if strings.HasSuffix(fileName, "_test.go") {
			continue
		}

		f := pack.Files[fileName]
		for _, d := range f.Imports {
			p.parseImports(d)
		}
		for _, d := range f.Decls {
			p.parseFileDeclarations(d)
		}
	}
}

func (p *ClassParser) parseImports(impt *ast.ImportSpec) {
	if impt.Name != nil {
		splitPath := strings.Split(impt.Path.Value, "/")
		s := strings.Trim(splitPath[len(splitPath)-1], `"`)
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

// parse the given declaration looking for classes, interfaces, or member functions
func (p *ClassParser) parseFileDeclarations(node ast.Decl) {
	switch decl := node.(type) {
	case *ast.GenDecl:
		p.handleGenDecl(decl)
	case *ast.FuncDecl:
		p.handleFuncDecl(decl)
	}
}

func (p *ClassParser) handleFuncDecl(decl *ast.FuncDecl) {
	if decl.Recv != nil {
		if decl.Recv.List == nil {
			return
		}

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

func handleGenDecStructType(p *ClassParser, typeName string, c *ast.StructType, typeParams *ast.FieldList) {
	for _, f := range c.Fields.List {
		p.getOrCreateStruct(typeName).AddField(f, p.allImports)
	}

	if typeParams == nil {
		return
	}

	for _, tp := range typeParams.List {
		p.getOrCreateStruct(typeName).AddTypeParam(tp)
	}
}

func handleGenDecInterfaceType(p *ClassParser, typeName string, c *ast.InterfaceType) {
	for _, f := range c.Methods.List {
		switch t := f.Type.(type) {
		case *ast.FuncType:
			p.getOrCreateStruct(typeName).AddMethod(f, p.allImports)
		case *ast.Ident:
			f, _ := getFieldType(t, p.allImports)
			st := p.getOrCreateStruct(typeName)
			f = replacePackageConstant(f, st.PackageName)
			st.AddToComposition(f)
		}
	}
}

func (p *ClassParser) handleGenDecl(decl *ast.GenDecl) {
	if decl.Specs == nil || len(decl.Specs) < 1 {
		// This might be a type of General Declaration we do not know how to handle.
		return
	}
	for _, spec := range decl.Specs {
		p.processSpec(spec)
	}
}

func (p *ClassParser) processSpec(spec ast.Spec) {
	var typeName string
	var alias *Alias
	declarationType := "alias"
	switch v := spec.(type) {
	case *ast.TypeSpec:
		typeName = v.Name.Name
		switch c := v.Type.(type) {
		case *ast.StructType:
			declarationType = "class"
			handleGenDecStructType(p, typeName, c, v.TypeParams)
		case *ast.InterfaceType:
			declarationType = "interface"
			handleGenDecInterfaceType(p, typeName, c)
		default:
			basicType, _ := getFieldType(getBasicType(c), p.allImports)

			aliasType, _ := getFieldType(c, p.allImports)
			aliasType = replacePackageConstant(aliasType, "")
			if !isPrimitiveString(typeName) {
				typeName = fmt.Sprintf("%s.%s", p.currentPackageName, typeName)
			}
			packageName := p.currentPackageName
			if isPrimitiveString(basicType) {
				packageName = builtinPackageName
			}
			alias = getNewAlias(fmt.Sprintf("%s.%s", packageName, aliasType), p.currentPackageName, typeName)

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
		if strings.Count(alias.Name, ".") > 1 {
			pack := strings.SplitN(alias.Name, ".", 2)
			if _, ok := p.allRenamedStructs[pack[0]]; !ok {
				p.allRenamedStructs[pack[0]] = map[string]string{}
			}
			renamedClass := generateRenamedStructName(pack[1])
			p.allRenamedStructs[pack[0]][renamedClass] = pack[1]
		}
	}
}

// If this element is an array or a pointer, this function will return the type that is closer to these
// two definitions. For example []***map[int] string will return map[int]string
func getBasicType(theType ast.Expr) ast.Expr {
	switch t := theType.(type) {
	case *ast.ArrayType:
		return getBasicType(t.Elt)
	case *ast.StarExpr:
		return getBasicType(t.X)
	case *ast.MapType:
		return getBasicType(t.Value)
	case *ast.ChanType:
		return getBasicType(t.Value)
	case *ast.Ellipsis:
		return getBasicType(t.Elt)
	}
	return theType
}

// Render returns a string of the class diagram that this parser has generated.
func (p *ClassParser) Render() string {
	str := &LineStringBuilder{}
	str.WriteLineWithDepth(0, "@startuml")
	if p.renderingOptions.Title != "" {
		str.WriteLineWithDepth(0, fmt.Sprintf(`title %s`, p.renderingOptions.Title))
	}
	if note := strings.TrimSpace(p.renderingOptions.Notes); note != "" {
		str.WriteLineWithDepth(0, "legend")
		str.WriteLineWithDepth(0, note)
		str.WriteLineWithDepth(0, "end legend")
	}

	var packages []string
	for pack := range p.structure {
		packages = append(packages, pack)
	}
	sort.Strings(packages)
	for _, pack := range packages {
		structures := p.structure[pack]
		p.renderStructures(pack, structures, str)

	}
	if p.renderingOptions.Aliases {
		p.renderAliases(str)
	}
	if !p.renderingOptions.Fields {
		str.WriteLineWithDepth(0, "hide fields")
	}
	if !p.renderingOptions.Methods {
		str.WriteLineWithDepth(0, "hide methods")
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

		names := []string{}
		for name := range structures {
			names = append(names, name)
		}

		sort.Strings(names)

		for _, name := range names {
			structure := structures[name]
			p.renderStructure(structure, pack, name, str, composition, extends, aggregations)
		}
		var orderedRenamedStructs []string
		for tempName := range p.allRenamedStructs[pack] {
			orderedRenamedStructs = append(orderedRenamedStructs, tempName)
		}
		sort.Strings(orderedRenamedStructs)
		for _, tempName := range orderedRenamedStructs {
			name := p.allRenamedStructs[pack][tempName]
			str.WriteLineWithDepth(1, fmt.Sprintf(`class "%s" as %s {`, name, tempName))
			str.WriteLineWithDepth(2, aliasComplexNameComment)
			str.WriteLineWithDepth(1, "}")
		}
		str.WriteLineWithDepth(0, `}`)
		if p.renderingOptions.Compositions {
			str.WriteLineWithDepth(0, composition.String())
		}
		if p.renderingOptions.Implementations {
			str.WriteLineWithDepth(0, extends.String())
		}
		if p.renderingOptions.Aggregations {
			str.WriteLineWithDepth(0, aggregations.String())
		}
	}
}

func (p *ClassParser) renderAliases(str *LineStringBuilder) {
	aliasString := ""
	if p.renderingOptions.ConnectionLabels {
		aliasString = aliasOf
	}
	orderedAliases := AliasSlice{}
	for _, alias := range p.allAliases {
		orderedAliases = append(orderedAliases, *alias)
	}
	sort.Sort(orderedAliases)
	for _, alias := range orderedAliases {
		aliasName := alias.Name
		if strings.Count(alias.Name, ".") > 1 {
			split := strings.SplitN(alias.Name, ".", 2)
			if aliasRename, ok := p.allRenamedStructs[split[0]]; ok {
				renamed := generateRenamedStructName(split[1])
				if _, ok := aliasRename[renamed]; ok {
					aliasName = fmt.Sprintf("%s.%s", split[0], renamed)
				}
			}
		}
		str.WriteLineWithDepth(0, fmt.Sprintf(`"%s" #.. %s"%s"`, aliasName, aliasString, alias.AliasOf))
	}
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

	types := ""
	if structure.Generics.exists() {
		types = "<"
		for t := range structure.Generics.Types {
			types += fmt.Sprintf("%s, ", t)
		}
		types = strings.TrimSuffix(types, ", ")
		types += " constrains "
		for _, n := range structure.Generics.Names {
			types += fmt.Sprintf("%s, ", n)
		}
		types = strings.TrimSuffix(types, ", ")
		types += ">"
	}

	str.WriteLineWithDepth(1, fmt.Sprintf(`%s %s%s %s {`, renderStructureType, name, types, sType))
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
	str.WriteLineWithDepth(1, `}`)
}

func (p *ClassParser) renderCompositions(structure *Struct, name string, composition *LineStringBuilder) {
	orderedCompositions := []string{}

	for c := range structure.Composition {
		if !strings.Contains(c, ".") {
			c = fmt.Sprintf("%s.%s", p.getPackageName(c, structure), c)
		}
		composedString := ""
		if p.renderingOptions.ConnectionLabels {
			composedString = extends
		}
		c = fmt.Sprintf(`"%s" *-- %s"%s.%s"`, c, composedString, structure.PackageName, name)
		orderedCompositions = append(orderedCompositions, c)
	}
	sort.Strings(orderedCompositions)
	for _, c := range orderedCompositions {
		composition.WriteLineWithDepth(0, c)
	}
}

func (p *ClassParser) renderAggregations(structure *Struct, name string, aggregations *LineStringBuilder) {
	aggregationMap := structure.Aggregations
	if p.renderingOptions.AggregatePrivateMembers {
		p.updatePrivateAggregations(structure, aggregationMap)
	}
	p.renderAggregationMap(aggregationMap, structure, aggregations, name)
}

func (p *ClassParser) updatePrivateAggregations(structure *Struct, aggregationsMap map[string]struct{}) {
	for agg := range structure.PrivateAggregations {
		aggregationsMap[agg] = struct{}{}
	}
}

func (p *ClassParser) renderAggregationMap(aggregationMap map[string]struct{}, structure *Struct, aggregations *LineStringBuilder, name string) {
	var orderedAggregations []string
	for a := range aggregationMap {
		orderedAggregations = append(orderedAggregations, a)
	}

	sort.Strings(orderedAggregations)

	for _, a := range orderedAggregations {
		if !strings.Contains(a, ".") {
			a = fmt.Sprintf("%s.%s", p.getPackageName(a, structure), a)
		}
		aggregationString := ""
		if p.renderingOptions.ConnectionLabels {
			aggregationString = aggregates
		}
		if p.getPackageName(a, structure) != builtinPackageName {
			aggregations.WriteLineWithDepth(0, fmt.Sprintf(`"%s.%s"%s o-- "%s"`, structure.PackageName, name, aggregationString, a))
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

	orderedExtends := []string{}
	for c := range structure.Extends {
		if !strings.Contains(c, ".") {
			c = fmt.Sprintf("%s.%s", structure.PackageName, c)
		}
		implementString := ""
		if p.renderingOptions.ConnectionLabels {
			implementString = implements
		}
		c = fmt.Sprintf(`"%s" <|-- %s"%s.%s"`, c, implementString, structure.PackageName, name)
		orderedExtends = append(orderedExtends, c)
	}
	sort.Strings(orderedExtends)
	for _, c := range orderedExtends {
		extends.WriteLineWithDepth(0, c)
	}
}

func (p *ClassParser) renderStructMethods(structure *Struct, privateMethods *LineStringBuilder, publicMethods *LineStringBuilder) {
	for _, method := range structure.Functions {
		accessModifier := "+"
		if unicode.IsLower(rune(method.Name[0])) {
			if !p.renderingOptions.PrivateMembers {
				continue
			}

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
			if !p.renderingOptions.PrivateMembers {
				continue
			}

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
			PackageName:         p.currentPackageName,
			Functions:           make([]*Function, 0),
			Fields:              make([]*Field, 0),
			Type:                "",
			Generics:            NewGeneric(),
			Composition:         make(map[string]struct{}, 0),
			Extends:             make(map[string]struct{}, 0),
			Aggregations:        make(map[string]struct{}, 0),
			PrivateAggregations: make(map[string]struct{}, 0),
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

// SetRenderingOptions Sets the rendering options for the Render() Function
func (p *ClassParser) SetRenderingOptions(ro map[RenderingOption]interface{}) error {
	for option, val := range ro {
		switch option {
		case RenderAggregations:
			p.renderingOptions.Aggregations = val.(bool)
		case RenderAliases:
			p.renderingOptions.Aliases = val.(bool)
		case RenderCompositions:
			p.renderingOptions.Compositions = val.(bool)
		case RenderFields:
			p.renderingOptions.Fields = val.(bool)
		case RenderImplementations:
			p.renderingOptions.Implementations = val.(bool)
		case RenderMethods:
			p.renderingOptions.Methods = val.(bool)
		case RenderConnectionLabels:
			p.renderingOptions.ConnectionLabels = val.(bool)
		case RenderTitle:
			p.renderingOptions.Title = val.(string)
		case RenderNotes:
			p.renderingOptions.Notes = val.(string)
		case AggregatePrivateMembers:
			p.renderingOptions.AggregatePrivateMembers = val.(bool)
		case RenderPrivateMembers:
			p.renderingOptions.PrivateMembers = val.(bool)
		default:
			return fmt.Errorf("Invalid Rendering option %v", option)
		}

	}
	return nil
}
func generateRenamedStructName(currentName string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(currentName, "")
}
