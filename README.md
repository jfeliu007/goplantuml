[![godoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jfeliu007/goplantuml/parser) [![Go Report Card](https://goreportcard.com/badge/github.com/jfeliu007/goplantuml)](https://goreportcard.com/report/github.com/jfeliu007/goplantuml) [![codecov](https://codecov.io/gh/jfeliu007/goplantuml/branch/master/graph/badge.svg)](https://codecov.io/gh/jfeliu007/goplantuml) [![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/jfeliu007/goplantuml.svg)](https://github.com/jfeliu007/goplantuml/releases/)
[![Build Status](https://travis-ci.org/jfeliu007/goplantuml.svg?branch=master)](https://travis-ci.org/jfeliu007/goplantuml)
# GoPlantUML

PlantUML Class Diagram Generator for golang projects. Generates class diagram text compatible with plantuml with the information of all structures and interfaces as well as the relationship among them.

## Code of Conduct
Please, review the code of conduct [here](https://github.com/jfeliu007/goplantuml/blob/master/CODE_OF_CONDUCT.md "here").

### Prerequisites
golang 1.10 or above

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

There are two different relationships considered in goplantuml:
- Interface implementation
- Type Composition

The following example contains interface implementations and composition. Notice how the signature of the functions
```golang
package testingsupport

//MyInterface only has one method, notice the signature return value
type MyInterface interface {
	foo() bool
}

//MyStruct1 will implement the foo() bool function so it will have an "extends" association with MyInterface
type MyStruct1 struct {
}

func (s1 *MyStruct1) foo() bool {
	return true
}

//MyStruct2 will be direclty composed of MyStruct1 so it will have a composition relationship with it
type MyStruct2 struct {
	MyStruct1
}

//MyStruct3 will have a foo() function but the return value is not a bool, so it will not have any relationship with MyInterface
type MyStruct3 struct {
}

func (s3 *MyStruct3) foo() {

}
```
This will be generated from the previous code
```
@startuml
namespace testingsupport {
    interface MyInterface  {
        - foo() bool

    }
    class MyStruct1 << (S,Aquamarine) >> {
        - foo() bool

    }
    class MyStruct2 << (S,Aquamarine) >> {
    }
    class MyStruct3 << (S,Aquamarine) >> {
        - foo() 

    }
}
testingsupport.MyStruct1 *-- testingsupport.MyStruct2

testingsupport.MyInterface <|-- testingsupport.MyStruct1

@enduml
```

![alt text](https://raw.githubusercontent.com/jfeliu007/goplantuml/master/example/example.png)

### The following diagram is generated based on the file in https://raw.githubusercontent.com/jfeliu007/goplantuml/master/ClassDiagram.puml
![Alt text](http://www.plantuml.com/plantuml/proxy?cache=no&src=https://raw.githubusercontent.com/jfeliu007/goplantuml/master/ClassDiagram.puml?raw=true "Title")
