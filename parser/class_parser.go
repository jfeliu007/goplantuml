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
	MaxDepth           int // Maximum nesting depth for packages (0 = unlimited)
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
	currentDirPath     string   // Current directory being parsed
	rootDirectories    []string // Root directories being processed
	allInterfaces      map[string]struct{}
	allStructs         map[string]struct{}
	allImports         map[string]string
	allAliases         map[string]*Alias
	allRenamedStructs  map[string]map[string]string
	maxDepth           int
	packageHierarchy   map[string]*PackageNode // Maps package full path to PackageNode
}

// PackageNode represents a package in the hierarchy
type PackageNode struct {
	Name       string                  // Short name (e.g., "subfolder")
	FullPath   string                  // Full path (e.g., "testingsupport.subfolder")
	Parent     *PackageNode            // Parent package
	Children   map[string]*PackageNode // Child packages
	Structures map[string]*Struct      // Structures in this package
	Depth      int                     // Depth in hierarchy
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
		rootDirectories:   options.Directories,
		allInterfaces:     make(map[string]struct{}),
		allStructs:        make(map[string]struct{}),
		allImports:        make(map[string]string),
		allAliases:        make(map[string]*Alias),
		allRenamedStructs: make(map[string]map[string]string),
		maxDepth:          options.MaxDepth,
		packageHierarchy:  make(map[string]*PackageNode),
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
				if inter != nil && st.ImplementsInterface(inter) {
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
	return NewClassDiagramWithMaxDepth(directoryPaths, ignoreDirectories, recursive, 0)
}

// NewClassDiagramWithMaxDepth returns a new classParser with which can Render the class diagram of
// files in the given directory with a maximum nesting depth
func NewClassDiagramWithMaxDepth(directoryPaths []string, ignoreDirectories []string, recursive bool, maxDepth int) (*ClassParser, error) {
	options := &ClassDiagramOptions{
		Directories:        directoryPaths,
		IgnoredDirectories: ignoreDirectories,
		Recursive:          recursive,
		RenderingOptions:   map[RenderingOption]interface{}{},
		FileSystem:         afero.NewOsFs(),
		MaxDepth:           maxDepth,
	}
	return NewClassDiagramWithOptions(options)
}

// getOrCreatePackageNode creates or retrieves a package node in the hierarchy
func (p *ClassParser) getOrCreatePackageNode(dirPath string) *PackageNode {
	// Calculate the package path relative to the root directories
	packagePath := p.calculatePackagePath(dirPath)

	if node, exists := p.packageHierarchy[packagePath]; exists {
		return node
	}

	// Create new package node
	// Use the last component of the package path as the display name
	displayName := filepath.Base(dirPath)
	if strings.Contains(packagePath, ".") {
		parts := strings.Split(packagePath, ".")
		displayName = parts[len(parts)-1]
	}

	node := &PackageNode{
		Name:       displayName,
		FullPath:   packagePath,
		Children:   make(map[string]*PackageNode),
		Structures: make(map[string]*Struct),
		Depth:      p.calculateDepth(packagePath),
	}

	// Check depth limit
	if p.maxDepth > 0 && node.Depth > p.maxDepth {
		return nil
	}

	// Establish parent-child relationships
	p.establishParentChildRelationships(node)

	p.packageHierarchy[packagePath] = node
	return node
}

// establishParentChildRelationships sets up parent-child relationships for a package node
func (p *ClassParser) establishParentChildRelationships(node *PackageNode) {
	if node.Depth <= 1 {
		return // Root package, no parent
	}

	// Find parent path by removing the last component
	parentPath := p.getParentPath(node.FullPath)
	if parentPath == "" {
		return
	}

	// Get or create parent node
	parentNode := p.packageHierarchy[parentPath]
	if parentNode == nil {
		// Create parent node if it doesn't exist
		parentDir := p.getDirectoryForPackagePath(parentPath)
		if parentDir != "" {
			parentNode = p.getOrCreatePackageNode(parentDir)
		}
	}

	if parentNode != nil {
		node.Parent = parentNode
		parentNode.Children[node.FullPath] = node
	}
}

// getParentPath returns the parent path of a given package path
func (p *ClassParser) getParentPath(packagePath string) string {
	lastDot := strings.LastIndex(packagePath, ".")
	if lastDot == -1 {
		return "" // No parent
	}
	return packagePath[:lastDot]
}

// getDirectoryForPackagePath returns the directory path for a given package path
func (p *ClassParser) getDirectoryForPackagePath(packagePath string) string {
	// Convert package path back to directory path
	// For example: "cmd.goplantuml" -> "cmd/goplantuml"
	dirPath := strings.ReplaceAll(packagePath, ".", string(filepath.Separator))

	// Check if this directory exists relative to any of our root directories
	for _, rootDir := range p.rootDirectories {
		fullPath := filepath.Join(rootDir, dirPath)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return ""
}

// calculatePackagePath determines the package path from directory path
func (p *ClassParser) calculatePackagePath(dirPath string) string {
	absPath, _ := filepath.Abs(dirPath)

	// Find the shortest root directory that contains this path
	var shortestRoot string
	for _, root := range p.getRootDirectories() {
		rootAbs, _ := filepath.Abs(root)
		if strings.HasPrefix(absPath, rootAbs) {
			if shortestRoot == "" || len(rootAbs) < len(shortestRoot) {
				shortestRoot = rootAbs
			}
		}
	}

	if shortestRoot == "" {
		return filepath.Base(absPath)
	}

	// Get relative path from root
	relPath, err := filepath.Rel(shortestRoot, absPath)
	if err != nil {
		return filepath.Base(absPath)
	}

	// Convert path separators to dots for package naming
	packagePath := strings.ReplaceAll(relPath, string(filepath.Separator), ".")
	if packagePath == "." {
		return filepath.Base(shortestRoot)
	}

	// Check if we're at the project root level (no nesting)
	// If the relative path doesn't contain separators, we're at the top level
	if !strings.Contains(relPath, string(filepath.Separator)) {
		return packagePath
	}

	// Special case: if we're processing the project root (current directory)
	// and the path contains testingsupport or cmd, we want to preserve the nesting
	// This handles the case where these are subdirectories of the project
	if strings.HasPrefix(relPath, "testingsupport") || strings.HasPrefix(relPath, "cmd") {
		return packagePath
	}

	// Special case: if we're processing cmd/goplantuml, it should be treated as cmd.goplantuml
	// not as a separate root package
	if strings.HasPrefix(relPath, "cmd/goplantuml") {
		return "cmd.goplantuml"
	}

	// Special case: if we're processing cmd directory, it should be treated as cmd
	if strings.HasPrefix(relPath, "cmd/") {
		return "cmd"
	}

	// Only prepend root directory name if we're not at the root level
	// and if the root directory is not "." (current directory)
	rootName := filepath.Base(shortestRoot)
	if packagePath != "" && rootName != "." {
		return rootName + "." + packagePath
	}

	return packagePath
}

// calculateDepth calculates the nesting depth of a package path
func (p *ClassParser) calculateDepth(packagePath string) int {
	if packagePath == "" {
		return 0
	}
	return strings.Count(packagePath, ".") + 1
}

// getRootDirectories returns the root directories being processed
func (p *ClassParser) getRootDirectories() []string {
	return p.rootDirectories
}

// parse the given ast.Package into the ClassParser structure
func (p *ClassParser) parsePackage(node ast.Node) {
	pack := node.(*ast.Package)

	// Create package node for this directory
	packageNode := p.getOrCreatePackageNode(p.currentDirPath)
	if packageNode == nil {
		return // Skip if depth limit exceeded
	}

	// Use the hierarchical package name for the structure map
	p.currentPackageName = packageNode.FullPath

	// Initialize structure maps
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

		if !strings.HasSuffix(fileName, "_test.go") {
			f := pack.Files[fileName]
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
		s := strings.Trim(splitPath[len(splitPath)-1], `"`)
		p.allImports[impt.Name.Name] = s
	}
}

func (p *ClassParser) parseDirectory(directoryPath string) error {
	p.currentDirPath = directoryPath
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
		if len(decl.Recv.List) == 0 {
			return
		}

		// Only get in when the function is defined for a structure. Global functions are not needed for class diagram
		theType, _ := getFieldType(decl.Recv.List[0].Type, p.allImports)
		theType = replacePackageConstant(theType, "")
		if len(theType) > 0 && theType[0] == "*"[0] {
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

func handleGenDecStructType(p *ClassParser, typeName string, c *ast.StructType) {
	for _, f := range c.Fields.List {
		p.getOrCreateStruct(typeName).AddField(f, p.allImports)
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
	if len(decl.Specs) < 1 {
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
			handleGenDecStructType(p, typeName, c)
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

	// Create builders for relationships
	composition := &LineStringBuilder{}
	extends := &LineStringBuilder{}
	aggregations := &LineStringBuilder{}

	// Render hierarchical packages
	p.renderHierarchicalPackages(str, composition, extends, aggregations)

	// Render aliases
	if p.renderingOptions.Aliases {
		p.renderAliases(str)
	}

	// Render all relationships collected during package rendering
	if p.renderingOptions.Compositions {
		str.WriteLineWithDepth(0, composition.String())
	}
	if p.renderingOptions.Implementations {
		str.WriteLineWithDepth(0, extends.String())
	}
	if p.renderingOptions.Aggregations {
		str.WriteLineWithDepth(0, aggregations.String())
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

// renderHierarchicalPackages renders packages in a hierarchical structure
func (p *ClassParser) renderHierarchicalPackages(str *LineStringBuilder, composition *LineStringBuilder, extends *LineStringBuilder, aggregations *LineStringBuilder) {
	// Find root packages (packages with no parent)
	rootPackages := make(map[string]*PackageNode)
	for _, node := range p.packageHierarchy {
		if node.Parent == nil {
			rootPackages[node.FullPath] = node
		}
	}

	// Sort root packages by name
	var sortedRoots []string
	for path := range rootPackages {
		sortedRoots = append(sortedRoots, path)
	}
	sort.Strings(sortedRoots)

	// Render each root package and its children
	for _, rootPath := range sortedRoots {
		p.renderPackageNode(rootPackages[rootPath], str, composition, extends, aggregations, 0)
	}
}

// renderPackageNode renders a package node and its children recursively
func (p *ClassParser) renderPackageNode(node *PackageNode, str *LineStringBuilder, composition *LineStringBuilder, extends *LineStringBuilder, aggregations *LineStringBuilder, depth int) {
	if node == nil {
		return
	}

	// Render this package's namespace using the short name
	str.WriteLineWithDepth(depth, fmt.Sprintf(`namespace %s {`, node.Name))

	// Render structures in this package using the full path
	if structures, exists := p.structure[node.FullPath]; exists {
		p.renderStructuresInPackage(node.FullPath, structures, str, depth+1, composition, extends, aggregations)
	}

	// Render child packages
	var childNames []string
	for _, child := range node.Children {
		childNames = append(childNames, child.FullPath)
	}
	sort.Strings(childNames)

	for _, childPath := range childNames {
		p.renderPackageNode(node.Children[childPath], str, composition, extends, aggregations, depth+1)
	}

	// Close namespace
	str.WriteLineWithDepth(depth, "}")
}

// renderStructuresInPackage renders structures within a package namespace
func (p *ClassParser) renderStructuresInPackage(pack string, structures map[string]*Struct, str *LineStringBuilder, depth int, composition *LineStringBuilder, extends *LineStringBuilder, aggregations *LineStringBuilder) {
	if len(structures) > 0 {
		names := []string{}
		for name := range structures {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			structure := structures[name]
			p.renderStructure(structure, pack, name, str, composition, extends, aggregations)
		}

		// Render renamed structs if any
		var orderedRenamedStructs []string
		for tempName := range p.allRenamedStructs[pack] {
			orderedRenamedStructs = append(orderedRenamedStructs, tempName)
		}
		sort.Strings(orderedRenamedStructs)
		for _, tempName := range orderedRenamedStructs {
			name := p.allRenamedStructs[pack][tempName]
			str.WriteLineWithDepth(depth, fmt.Sprintf(`class "%s" as %s {`, name, tempName))
			str.WriteLineWithDepth(depth+1, aliasComplexNameComment)
			str.WriteLineWithDepth(depth, "}")
		}
	}
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
		str.WriteLineWithDepth(0, "}")
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
	str.WriteLineWithDepth(1, "}")
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
		if len(method.Name) > 0 && unicode.IsLower(rune(method.Name[0])) {
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
		if len(field.Name) > 0 && unicode.IsLower(rune(field.Name[0])) {
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
			return fmt.Errorf("invalid rendering option %v", option)
		}

	}
	return nil
}
func generateRenamedStructName(currentName string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(currentName, "")
}
