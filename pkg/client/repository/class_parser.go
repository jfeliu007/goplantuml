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
package repository

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
)

type LineStringBuilder struct { // restored struct definition
	strings.Builder
}

const tab = "    "

// builtinPackageName used by alias.go via getNewAlias
const builtinPackageName = "__builtin__" //nolint:unused // referenced in alias.go
var _ = builtinPackageName               // reference to avoid unused lint; used also in alias.go

// WriteLineWithDepth will write the given text with added tabs at the beginning into the string builder.
func (lsb *LineStringBuilder) WriteLineWithDepth(depth int, str string) { // keep method with struct now defined
	for i := 0; i < depth; i++ {
		lsb.WriteString(tab)
	}
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
	// --- added ---
	CustomResources []string
	CustomKeywords  map[string][]string
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

// ClassParser contains the structure of the parsed files. The structure is a map of package_paths that contains
// a map of structure_names -> Structs
type ClassParser struct {
	renderingOptions   *RenderingOptions
	structure          map[string]map[string]*Struct
	currentPackageName string
	currentPackagePath string // Full package path for hierarchy support
	allInterfaces      map[string]struct{}
	allStructs         map[string]struct{}
	allImports         map[string]string
	allAliases         map[string]*Alias
	allRenamedStructs  map[string]map[string]string
	packagePaths       map[string]string   // Maps package name to full path
	customResources    []string            // Custom resource patterns for function categorization
	customKeywords     map[string][]string // Custom keywords for function categorization (resource_name -> keywords)
	arrowLines         map[string]struct{} // 重複排除用に既出矢印を保持
	arrowStats         map[string]int      // 矢印種類別統計
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
		CustomResources:    []string{},
		CustomKeywords:     map[string][]string{},
	}
	return NewClassDiagramWithOptions(options)
}

// NewClassDiagramWithOptions returns a new classParser with which can Render the class diagram of
// files in the given directory passed in the ClassDiargamOptions. This will also alow for different types of FileSystems
// Passed since it is part of the ClassDiagramOptions as well.
func NewClassDiagramWithOptions(options *ClassDiagramOptions) (*ClassParser, error) {
	classParser := &ClassParser{
		renderingOptions: &RenderingOptions{
			Aggregations:     true, // Enable dependency arrows by default
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
		packagePaths:      make(map[string]string),
		customResources:   make([]string, 0),
		customKeywords:    make(map[string][]string),
		arrowLines:        make(map[string]struct{}),
		arrowStats:        make(map[string]int),
	}
	// apply pre-parse customization so classification uses them
	if len(options.CustomResources) > 0 {
		classParser.customResources = options.CustomResources
	}
	if len(options.CustomKeywords) > 0 {
		classParser.customKeywords = options.CustomKeywords
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

// parseDirectory processes a directory for Go source files
func (p *ClassParser) parseDirectory(directoryPath string) error {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, directoryPath, nil, 0)
	if err != nil {
		return err
	}
	for pkgName, pkg := range pkgs {
		// collect files for type checking
		var files []*ast.File
		for fileName, f := range pkg.Files {
			if !strings.HasSuffix(fileName, "_test.go") {
				files = append(files, f)
			}
		}
		if len(files) == 0 {
			continue
		}
		conf := types.Config{Importer: nil, FakeImportC: true, IgnoreFuncBodies: true}
		_, _ = conf.Check(pkgName, fs, files, nil) // ignore type errors for now
		p.consumeParsedPackage(pkgName, directoryPath, pkg.Files)
	}
	return nil
}

// consumeParsedPackage replaces parsePackage without using deprecated ast.Package
func (p *ClassParser) consumeParsedPackage(pkgName, directoryPath string, fileMap map[string]*ast.File) {
	p.currentPackageName = pkgName
	p.currentPackagePath = p.extractPackagePath(directoryPath)
	packageKey := p.currentPackagePath
	if packageKey == "" {
		packageKey = p.currentPackageName
	}
	p.packagePaths[p.currentPackageName] = packageKey
	if _, ok := p.structure[packageKey]; !ok {
		p.structure[packageKey] = make(map[string]*Struct)
	}
	var sortedFiles []string
	for fileName := range fileMap {
		sortedFiles = append(sortedFiles, fileName)
	}
	sort.Strings(sortedFiles)
	for _, fileName := range sortedFiles {
		if strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		f := fileMap[fileName]
		for _, d := range f.Imports {
			p.parseImports(d)
		}
		for _, d := range f.Decls {
			p.parseFileDeclarations(d)
		}
	}
}

// extractPackagePath extracts a package path from directory path
func (p *ClassParser) extractPackagePath(directoryPath string) string {
	// Normalize the path and extract meaningful package path
	cleanPath := filepath.Clean(directoryPath)

	// Remove common prefixes and create a meaningful package path
	parts := strings.Split(cleanPath, string(filepath.Separator))
	var relevantParts []string

	// Skip common root parts like ".", "/", "Users", etc.
	startIndex := 0
	for i, part := range parts {
		if part == "pkg" || part == "src" || part == "internal" {
			startIndex = i
			break
		}
	}

	if startIndex < len(parts) {
		relevantParts = parts[startIndex:]
	} else {
		// Fallback: use last few meaningful parts
		if len(parts) >= 2 {
			relevantParts = parts[len(parts)-2:]
		} else {
			relevantParts = parts
		}
	}

	// Create package path by joining relevant parts
	packagePath := strings.Join(relevantParts, ".")

	// Clean up the package path
	packagePath = strings.ReplaceAll(packagePath, "-", "_")
	packagePath = strings.ReplaceAll(packagePath, " ", "_")

	return packagePath
}

// parseImports processes import statements
func (p *ClassParser) parseImports(impt *ast.ImportSpec) {
	if impt.Name != nil {
		splitPath := strings.Split(impt.Path.Value, "/")
		s := strings.Trim(splitPath[len(splitPath)-1], `"`)
		p.allImports[impt.Name.Name] = s
	}
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

// handleFuncDecl processes function declarations
func (p *ClassParser) handleFuncDecl(decl *ast.FuncDecl) {
	if decl.Recv != nil {
		if decl.Recv.List == nil {
			return
		}

		// Process method functions - functions defined for a structure
		theType, _ := getFieldType(decl.Recv.List[0].Type, p.allImports)
		theType = replacePackageConstant(theType, "")
		if theType[0] == "*"[0] {
			theType = theType[1:]
		}
		structure := p.getOrCreateStruct(theType)
		if structure.Type == "" {
			structure.Type = "class"
		}

		packageKey := p.currentPackagePath
		if packageKey == "" {
			packageKey = p.currentPackageName
		}
		fullName := fmt.Sprintf("%s.%s", packageKey, theType)
		p.allStructs[fullName] = struct{}{}
		structure.AddMethod(&ast.Field{
			Names:   []*ast.Ident{decl.Name},
			Doc:     decl.Doc,
			Type:    decl.Type,
			Tag:     nil,
			Comment: nil,
		}, p.allImports)
	} else {
		// Process package-level functions
		// Categorize functions by resource type based on function names
		resourceType := p.extractResourceTypeFromFunction(decl.Name.Name)
		functionsClassName := fmt.Sprintf("%sFunctions", resourceType)

		structure := p.getOrCreateStruct(functionsClassName)
		if structure.Type == "" {
			structure.Type = "class"
		}

		packageKey := p.currentPackagePath
		if packageKey == "" {
			packageKey = p.currentPackageName
		}
		fullName := fmt.Sprintf("%s.%s", packageKey, functionsClassName)
		p.allStructs[fullName] = struct{}{}

		// Add the function as a method to the resource-specific class
		methodField := &ast.Field{
			Names:   []*ast.Ident{decl.Name},
			Doc:     decl.Doc,
			Type:    decl.Type,
			Tag:     nil,
			Comment: nil,
		}
		structure.AddMethod(methodField, p.allImports)

		// Extract dependencies from function parameters and return values
		p.extractFunctionDependencies(functionsClassName, decl.Type, packageKey)
	}
}

// extractResourceTypeFromFunction extracts resource type from function name
func (p *ClassParser) extractResourceTypeFromFunction(funcName string) string {
	// Check custom resource patterns first if any
	for _, pattern := range p.customResources {
		if strings.Contains(funcName, pattern) {
			return pattern
		}
	}

	// Check custom keywords for each resource type
	for resourceType, keywords := range p.customKeywords {
		for _, keyword := range keywords {
			if strings.Contains(funcName, keyword) {
				return resourceType
			}
		}
	}

	// Default category for unmatched functions
	return "General"
}

// extractFunctionDependencies extracts dependencies from function parameters and return values
func (p *ClassParser) extractFunctionDependencies(fromClass string, funcType *ast.FuncType, packageKey string) {
	if funcType == nil {
		return
	}

	fromStruct := p.getOrCreateStruct(fromClass)

	// Extract dependencies from parameters
	if funcType.Params != nil {
		for _, field := range funcType.Params.List {
			deps := p.extractTypeDependencies(field.Type)
			for _, dep := range deps {
				// Add dependency only if it's not a primitive type and not the same class
				if !isPrimitiveString(dep) && dep != fromClass {
					fromStruct.AddToAggregation(dep)
				}
			}
		}
	}

	// Extract dependencies from return values
	if funcType.Results != nil {
		for _, field := range funcType.Results.List {
			deps := p.extractTypeDependencies(field.Type)
			for _, dep := range deps {
				// Add dependency only if it's not a primitive type and not the same class
				if !isPrimitiveString(dep) && dep != fromClass {
					fromStruct.AddToAggregation(dep)
				}
			}
		}
	}
}

// extractTypeDependencies recursively extracts type dependencies from AST expressions
func (p *ClassParser) extractTypeDependencies(expr ast.Expr) []string {
	var deps []string

	switch e := expr.(type) {
	case *ast.Ident:
		// Simple identifier (e.g., "User", "string")
		if !isPrimitiveString(e.Name) {
			deps = append(deps, e.Name)
		}
	case *ast.SelectorExpr:
		// Package.Type (e.g., "request.UserRequest")
		if x, ok := e.X.(*ast.Ident); ok {
			typeName := fmt.Sprintf("%s.%s", x.Name, e.Sel.Name)
			deps = append(deps, typeName)
		}
	case *ast.StarExpr:
		// Pointer type (e.g., "*User")
		subDeps := p.extractTypeDependencies(e.X)
		deps = append(deps, subDeps...)
	case *ast.ArrayType:
		// Array type (e.g., "[]User")
		subDeps := p.extractTypeDependencies(e.Elt)
		deps = append(deps, subDeps...)
	case *ast.MapType:
		// Map type (e.g., "map[string]User")
		keyDeps := p.extractTypeDependencies(e.Key)
		valueDeps := p.extractTypeDependencies(e.Value)
		deps = append(deps, keyDeps...)
		deps = append(deps, valueDeps...)
	}

	return deps
}

// handleGenDecl processes general declarations
func (p *ClassParser) handleGenDecl(decl *ast.GenDecl) {
	if len(decl.Specs) == 0 { // simplified nil/len check
		// This might be a type of General Declaration we do not know how to handle.
		return
	}
	for _, spec := range decl.Specs {
		p.processSpec(spec)
	}
}

// processSpec processes type specifications
func (p *ClassParser) processSpec(spec ast.Spec) {
	var typeName string
	declarationType := "alias"
	switch v := spec.(type) {
	case *ast.TypeSpec:
		typeName = v.Name.Name
		switch c := v.Type.(type) {
		case *ast.StructType:
			declarationType = "class"
			p.handleGenDecStructType(typeName, c)
		case *ast.InterfaceType:
			declarationType = "interface"
			p.handleGenDecInterfaceType(typeName, c)
		default:
			// Handle other types as aliases
		}
	default:
		// Not needed for class diagrams (Imports, global variables, regular functions, etc)
		return
	}
	p.getOrCreateStruct(typeName).Type = declarationType

	packageKey := p.currentPackagePath
	if packageKey == "" {
		packageKey = p.currentPackageName
	}
	fullName := fmt.Sprintf("%s.%s", packageKey, typeName)
	switch declarationType {
	case "interface":
		p.allInterfaces[fullName] = struct{}{}
	case "class":
		p.allStructs[fullName] = struct{}{}
	}
}

// handleGenDecStructType processes struct type declarations
func (p *ClassParser) handleGenDecStructType(typeName string, c *ast.StructType) {
	for _, f := range c.Fields.List {
		p.getOrCreateStruct(typeName).AddField(f, p.allImports)
	}
}

// handleGenDecInterfaceType processes interface type declarations
func (p *ClassParser) handleGenDecInterfaceType(typeName string, c *ast.InterfaceType) {
	for _, f := range c.Methods.List {
		switch t := f.Type.(type) {
		case *ast.FuncType:
			p.getOrCreateStruct(typeName).AddMethod(f, p.allImports)
		case *ast.Ident:
			fieldType, _ := getFieldType(t, p.allImports)
			st := p.getOrCreateStruct(typeName)
			fieldType = replacePackageConstant(fieldType, st.PackageName)
			st.AddToComposition(fieldType)
		}
	}
}

// getOrCreateStruct gets or creates a struct in the parser
func (p *ClassParser) getOrCreateStruct(structName string) *Struct {
	packageKey := p.currentPackagePath
	if packageKey == "" {
		packageKey = p.currentPackageName
	}

	if _, ok := p.structure[packageKey]; !ok {
		p.structure[packageKey] = make(map[string]*Struct)
	}
	if st, ok := p.structure[packageKey][structName]; ok {
		return st
	}
	p.structure[packageKey][structName] = &Struct{
		PackageName: packageKey,
	}
	return p.structure[packageKey][structName]
}

// getStruct gets a struct by its full name
func (p *ClassParser) getStruct(structName string) *Struct {
	split := strings.Split(structName, ".")
	if len(split) == 2 {
		packageName := split[0]
		typeName := split[1]
		if pack, ok := p.structure[packageName]; ok {
			if st, ok := pack[typeName]; ok {
				return st
			}
		}
	}
	return nil
}

// SetRenderingOptions sets the rendering options for the parser
func (p *ClassParser) SetRenderingOptions(options map[RenderingOption]interface{}) {
	if options == nil {
		return
	}

	for option, value := range options {
		switch option {
		case RenderAggregations:
			if v, ok := value.(bool); ok {
				p.renderingOptions.Aggregations = v
			}
		case RenderFields:
			if v, ok := value.(bool); ok {
				p.renderingOptions.Fields = v
			}
		case RenderMethods:
			if v, ok := value.(bool); ok {
				p.renderingOptions.Methods = v
			}
		case RenderCompositions:
			if v, ok := value.(bool); ok {
				p.renderingOptions.Compositions = v
			}
		case RenderImplementations:
			if v, ok := value.(bool); ok {
				p.renderingOptions.Implementations = v
			}
		case RenderAliases:
			if v, ok := value.(bool); ok {
				p.renderingOptions.Aliases = v
			}
		case RenderConnectionLabels:
			if v, ok := value.(bool); ok {
				p.renderingOptions.ConnectionLabels = v
			}
		case RenderTitle:
			if v, ok := value.(string); ok {
				p.renderingOptions.Title = v
			}
		case RenderNotes:
			if v, ok := value.(string); ok {
				p.renderingOptions.Notes = v
			}
		case AggregatePrivateMembers:
			if v, ok := value.(bool); ok {
				p.renderingOptions.AggregatePrivateMembers = v
			}
		case RenderPrivateMembers:
			if v, ok := value.(bool); ok {
				p.renderingOptions.PrivateMembers = v
			}
		}
	}
}

// SetCustomResources sets custom resource patterns for function categorization
func (p *ClassParser) SetCustomResources(resources []string) {
	if resources != nil {
		p.customResources = resources
	}
}

// SetCustomKeywords sets custom keywords for different categories
func (p *ClassParser) SetCustomKeywords(keywords map[string][]string) {
	if keywords != nil {
		p.customKeywords = keywords
	}
}

// emitArrow は矢印行の重複を避けて出力し種類統計を記録する
func (p *ClassParser) emitArrow(str *LineStringBuilder, depth int, line, kind string) {
	if _, ok := p.arrowLines[line]; ok {
		return
	}
	p.arrowLines[line] = struct{}{}
	if kind != "" {
		p.arrowStats[kind]++
	}
	str.WriteLineWithDepth(depth, line)
}

func (p *ClassParser) renderInterfaceImplementationArrows(pack string, structures map[string]*Struct, str *LineStringBuilder) {
	for name, st := range structures {
		if st.Type == "class" && len(name) > 0 && name[0] >= 'a' && name[0] <= 'z' {
			cand := strings.ToUpper(string(name[0])) + name[1:]
			if inter, ok := structures[cand]; ok && inter.Type == "interface" {
				from := fmt.Sprintf("%s.%s", pack, cand)
				to := fmt.Sprintf("%s.%s", pack, name)
				p.emitArrow(str, 0, fmt.Sprintf(`%s <|.. %s : implements`, from, to), "implements")
			}
		}
	}
}

// extractPackageFromType extracts package name from a type name (e.g., "pkg.Type" -> "pkg")
func (p *ClassParser) extractPackageFromType(typeName string) string {
	if typeName == "" {
		return ""
	}
	
	// Remove pointer prefix
	cleanType := strings.TrimPrefix(typeName, "*")
	
	// Check if it contains a dot (package.Type format)
	if strings.Contains(cleanType, ".") {
		parts := strings.Split(cleanType, ".")
		if len(parts) >= 2 {
			// Return all parts except the last one (which is the type name)
			return strings.Join(parts[:len(parts)-1], ".")
		}
	}
	
	// Check if this type exists in any of our parsed packages
	for packageName, structures := range p.structure {
		for structName := range structures {
			if structName == cleanType {
				return packageName
			}
		}
	}
	
	return ""
}

// findPackagesByPattern find packages by pattern matching
func (p *ClassParser) findPackagesByPattern(pattern string) []string {
	var matchedPackages []string
	for pkg := range p.structure {
		if strings.Contains(pkg, pattern) {
			matchedPackages = append(matchedPackages, pkg)
		}
	}
	return matchedPackages
}

// renderLayerDependencies はレイヤー間の依存関係を矢印で表現する
func (p *ClassParser) renderLayerDependencies(str *LineStringBuilder, fromLayer, toLayer, relationship string) {
    fromPackages := p.findPackagesByPattern(fromLayer)
    toPackages := p.findPackagesByPattern(toLayer)
    for _, fromPkg := range fromPackages {
        for fromStruct := range p.structure[fromPkg] {
            for _, toPkg := range toPackages {
                for toStruct := range p.structure[toPkg] {
                    fromFull := fmt.Sprintf("%s.%s", fromPkg, fromStruct)
                    toFull := fmt.Sprintf("%s.%s", toPkg, toStruct)
                    p.emitArrow(str, 0, fmt.Sprintf(`%s ..> %s : %s`, fromFull, toFull, relationship), "layer")
                }
            }
        }
    }
}

// renderResourceDependencies renders architectural dependencies between layers
func (p *ClassParser) renderResourceDependencies(str *LineStringBuilder) {
    str.WriteLineWithDepth(0, "")
    str.WriteLineWithDepth(0, "'=== Architectural Dependencies ===")
    p.renderLayerDependencies(str, "controller", "usecase", "calls")
    p.renderLayerDependencies(str, "handler", "service", "calls")
    p.renderLayerDependencies(str, "usecase", "repository", "uses")
    p.renderLayerDependencies(str, "service", "repository", "uses")
    p.renderLayerDependencies(str, "repository", "model", "manages")
    p.renderLayerDependencies(str, "repository", "entity", "manages")
}

// renderPackageDependencies renders dependencies between packages only
func (p *ClassParser) renderPackageDependencies(str *LineStringBuilder) {
	packageDeps := make(map[string]map[string]struct{}) // source -> targets
	
	// Collect all package dependencies
	for pack, structures := range p.structure {
		if packageDeps[pack] == nil {
			packageDeps[pack] = make(map[string]struct{})
		}
		
		for _, st := range structures {
			// Check all aggregations
			for aggr := range st.Aggregations {
				if p.isDependencyValid(aggr, structures) {
					targetPackage := p.extractPackageFromType(aggr)
					if targetPackage != "" && targetPackage != pack {
						packageDeps[pack][targetPackage] = struct{}{}
					}
				}
			}
			
			// Check private aggregations if enabled
			if p.renderingOptions.AggregatePrivateMembers {
				for aggr := range st.PrivateAggregations {
					if p.isDependencyValid(aggr, structures) {
						targetPackage := p.extractPackageFromType(aggr)
						if targetPackage != "" && targetPackage != pack {
							packageDeps[pack][targetPackage] = struct{}{}
						}
					}
				}
			}
			
			// Check compositions
			if p.renderingOptions.Compositions {
				for comp := range st.Composition {
					if p.isDependencyValid(comp, structures) {
						targetPackage := p.extractPackageFromType(comp)
						if targetPackage != "" && targetPackage != pack {
							packageDeps[pack][targetPackage] = struct{}{}
						}
					}
				}
			}
			
			// Check extends
			for base := range st.Extends {
				if p.isDependencyValid(base, structures) {
					targetPackage := p.extractPackageFromType(base)
					if targetPackage != "" && targetPackage != pack {
						packageDeps[pack][targetPackage] = struct{}{}
					}
				}
			}
		}
	}
	
	// Output package-level dependencies
	str.WriteLineWithDepth(0, "")
	str.WriteLineWithDepth(0, "'=== Package Dependencies ===")
	for sourcePackage, targets := range packageDeps {
		for targetPackage := range targets {
			p.emitArrow(str, 0, fmt.Sprintf(`%s ..> %s : uses`, sourcePackage, targetPackage), "package_uses")
		}
	}
}

// Render returns a string of the class diagram that this parser has generated.
func (p *ClassParser) Render() string {
    str := &LineStringBuilder{}
    str.WriteLineWithDepth(0, "@startuml")
    // Set default direction to top-down for vertical layout (縦方向に長く)
    str.WriteLineWithDepth(0, "!define DIRECTION top to bottom direction")
    str.WriteLineWithDepth(0, "top to bottom direction")
    str.WriteLineWithDepth(0, "skinparam linetype ortho")

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

    // removed category.* virtual namespaces (vertical grouping now inline)

    // Render package-level dependencies instead of individual class dependencies
    p.renderPackageDependencies(str)

    // Disable detailed architectural dependencies to focus on package-level view
    // if p.renderingOptions.Aggregations {
    //     p.renderResourceDependencies(str)
    // }
    // 矢印統計出力
    p.renderArrowStats(str)
    if !p.renderingOptions.Fields {
        str.WriteLineWithDepth(0, "hide fields")
    }
    if !p.renderingOptions.Methods {
        str.WriteLineWithDepth(0, "hide methods")
    }
    str.WriteLineWithDepth(0, "@enduml")
    return str.String()
}

// renderStructures renders all structures in a package
func (p *ClassParser) renderStructures(pack string, structures map[string]*Struct, str *LineStringBuilder) {
	if len(structures) > 0 {
		str.WriteLineWithDepth(0, fmt.Sprintf(`namespace %s {`, pack))
		
		// collect and sort names by category precedence then name
		var names []string
		for n := range structures {
			names = append(names, n)
		}
		precedence := []string{"Common", "User", "Group", "Member", "Role"}
		precMap := map[string]int{}
		for i, c := range precedence {
			precMap[c] = i
		}
		sort.Slice(names, func(i, j int) bool {
			ci := p.matchCategory(names[i])
			cj := p.matchCategory(names[j])
			ri, okI := precMap[ci]
			if !okI {
				ri = 1000
			}
			rj, okJ := precMap[cj]
			if !okJ {
				rj = 1000
			}
			if ri != rj {
				return ri < rj
			}
			// same rank -> lexical
			if names[i] != names[j] {
				return names[i] < names[j]
			}
			return false
		})
		
		// Group by category and render with frames
		categories := make(map[string][]string)
		for _, name := range names {
			cat := p.matchCategory(name)
			if cat == "" {
				cat = "General"
			}
			categories[cat] = append(categories[cat], name)
		}
		
		// Render categories in precedence order
		for _, cat := range precedence {
			if items, ok := categories[cat]; ok {
				str.WriteLineWithDepth(1, fmt.Sprintf(`frame "%s" {`, cat))
				for _, name := range items {
					p.renderStructure(structures[name], name, str)
				}
				str.WriteLineWithDepth(1, "}")
			}
		}
		
		// Render uncategorized items
		if items, ok := categories["General"]; ok {
			for _, name := range items {
				p.renderStructure(structures[name], name, str)
			}
		}
		
		// Only render interface implementation arrows within the same package
		p.renderInterfaceImplementationArrows(pack, structures, str)
		// Don't render individual dependencies here - will be handled at package level
		str.WriteLineWithDepth(0, "}")
	}
}

// matchCategory returns first matching custom keyword category for a given name (case-insensitive)
func (p *ClassParser) matchCategory(name string) string {
	ln := strings.ToLower(name)
	for cat, kws := range p.customKeywords {
		for _, kw := range kws {
			if kw == "" {
				continue
			}
			if strings.Contains(ln, strings.ToLower(kw)) {
				return cat
			}
		}
	}
	return ""
}

// renderStructure renders a single structure
func (p *ClassParser) renderStructure(structure *Struct, name string, str *LineStringBuilder) {
	sType := ""
	renderStructureType := structure.Type
	switch structure.Type {
	case "class":
		sType = "<< (S,Aquamarine) >>"
	case "alias":
		sType = "<< (T, #FF7700) >> "
		renderStructureType = "class"
	}
	// Remove category stereotyping since we're now using frames
	str.WriteLineWithDepth(2, fmt.Sprintf(`%s %s %s {`, renderStructureType, name, sType))
	if p.renderingOptions.Fields {
		p.renderStructFields(structure, str)
	}
	if p.renderingOptions.Methods {
		p.renderStructMethods(structure, str)
	}
	str.WriteLineWithDepth(2, "}")
}

// renderStructFields renders struct fields
func (p *ClassParser) renderStructFields(structure *Struct, str *LineStringBuilder) {
	for _, field := range structure.Fields {
		accessModifier := "+"
		if !isExported(field.Name) {
			accessModifier = "-"
		}
		// PlantUML requires ':' between name and type
		str.WriteLineWithDepth(3, fmt.Sprintf(`%s %s : %s`, accessModifier, field.Name, field.Type))
	}
}

// renderStructMethods renders struct methods
func (p *ClassParser) renderStructMethods(structure *Struct, str *LineStringBuilder) {
	for _, method := range structure.Functions {
		accessModifier := "+"
		if !isExported(method.Name) {
			accessModifier = "-"
		}

		parameterList := []string{}
		for _, param := range method.Parameters {
			parameterList = append(parameterList, param.Type)
		}

		ret := ""
		if len(method.ReturnValues) > 0 {
			ret = strings.Join(method.ReturnValues, ", ")
			if len(method.ReturnValues) > 1 {
				ret = fmt.Sprintf("(%s)", ret)
			}
		}

		if ret == "" {
			str.WriteLineWithDepth(3, fmt.Sprintf(`%s %s(%s)`, accessModifier, method.Name, strings.Join(parameterList, ", ")))
		} else {
			str.WriteLineWithDepth(3, fmt.Sprintf(`%s %s(%s) : %s`, accessModifier, method.Name, strings.Join(parameterList, ", "), ret))
		}
	}
}

// isExported checks if a name is exported (starts with uppercase)
func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// renderArrowStats renders arrow usage statistics
func (p *ClassParser) renderArrowStats(str *LineStringBuilder) {
    if len(p.arrowStats) == 0 {
        return
    }
    str.WriteLineWithDepth(0, "")
    str.WriteLineWithDepth(0, "'=== Arrow Stats ===")
    keys := make([]string, 0, len(p.arrowStats))
    for k := range p.arrowStats {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    for _, k := range keys {
        str.WriteLineWithDepth(0, fmt.Sprintf("' %s: %d", k, p.arrowStats[k]))
    }
}

func (p *ClassParser) isDependencyValid(dep string, structures map[string]*Struct) bool {
    if dep == "" || isPrimitiveString(dep) {
        return false
    }
    // 配列/スライス型（[]type）は PlantUML 図では正しい参照として扱えないので除外
    if strings.HasPrefix(dep, "[]") {
        return false
    }
    // マップ型（map[key]value）も除外
    if strings.HasPrefix(dep, "map[") {
        return false
    }
    // interface{} は PlantUML で問題を起こすので除外
    if dep == "interface{}" {
        return false
    }
    return true
}