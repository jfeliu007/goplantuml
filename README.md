[![godoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jfeliu007/goplantuml/parser) [![Go Report Card](https://goreportcard.com/badge/github.com/jfeliu007/goplantuml)](https://goreportcard.com/report/github.com/jfeliu007/goplantuml) 
# GoPlantUML

PlantUML Class Diagram Generator for golang projects. Generates class diagram text compatible with plantuml with the information of all structures and interfaces as well as the relationship among them.

### Prerequisites
golang

### Installing

```
go get github.com/jfeliu007/goplantuml/parser
go get github.com/jfeliu007/goplantuml/cmd/goplantuml
cd $GOPATH/src/github.com/jfeliu007/goplantuml
go install ./...
```

This will install the command goplantuml in your GOPATH bin folder.

### Usage

```
goplantuml [-recursive] path/to/gofiles
```
```
goplantuml [-recursive] path/to/gofiles > diagram_file_name.puml
```

#### Example
```
goplantuml $GOPATH/src/github.com/jfeliu007/goplantuml/parser
```
```
// echoes

@startuml
namespace parser {
    class Struct {
        + Functions []*Function
        + Fields []*Parameter
        + Type string
        + Composition []string
        + Extends []string

    }
    class LineStringBuilder {
        + WriteLineWithDepth(depth int, str string) 

    }
    class ClassParser {
        - structure <font color=blue>map</font>[string]<font color=blue>map</font>[string]*Struct
        - currentPackageName string
        - allInterfaces <font color=blue>map</font>[string]<font color=blue>struct</font>{}
        - allStructs <font color=blue>map</font>[string]<font color=blue>struct</font>{}

        - structImplementsInterface(st *Struct, inter *Struct) 
        - parsePackage(node ast.Node) 
        - parseFileDeclarations(node ast.Decl) 
        - addMethodToStruct(s *Struct, method *ast.Field) 
        - getFunction(f *ast.FuncType, name string) 
        - addFieldToStruct(s *Struct, field *ast.Field) 
        - addToComposition(s *Struct, fType string) 
        - addToExtends(s *Struct, fType string) 
        - getOrCreateStruct(name string) 
        - getStruct(structName string) 
        - getFieldType(exp ast.Expr, includePackageName bool) 

        + Render() 

    }
    class Parameter {
        + Name string
        + Type string

    }
    class Function {
        + Name string
        + Parameters []*Parameter
        + ReturnValues []string

    }
}
strings.Builder *-- parser.LineStringBuilder


@enduml
```
```
goplantuml $GOPATH/src/github.com/jfeliu007/goplantuml/parser > ClassDiagram.puml
// Generates a file ClassDiagram.plum with the previous specifications
```

### The following diagram is generated based on the file in https://raw.githubusercontent.com/jfeliu007/goplantuml/master/ClassDiagram.puml
![Alt text](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/jfeliu007/goplantuml/master/ClassDiagram.puml?raw=true "Title")
