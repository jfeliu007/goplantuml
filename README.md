[![godoc reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jfeliu007/goplantuml/parser) [![Go Report Card](https://goreportcard.com/badge/github.com/jfeliu007/goplantuml)](https://goreportcard.com/report/github.com/jfeliu007/goplantuml) [![codecov](https://codecov.io/gh/jfeliu007/goplantuml/branch/master/graph/badge.svg)](https://codecov.io/gh/jfeliu007/goplantuml) [![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![GitHub release](https://img.shields.io/github/release/jfeliu007/goplantuml.svg)](https://github.com/jfeliu007/goplantuml/releases/)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) 
[![DUMELS Diagram](https://server.dumels.com/api/v1/badge/23ff0222-e93b-4e9f-a4ef-4d5d9b7a5c7d)](https://www.dumels.com/shdv/p/0497cfda-9623-4c02-8aea-ac040817c128) 
# GoPlantUML V2

GoPlantUML is an open-source tool developed to streamline the process of generating PlantUML diagrams from Go source code. With GoPlantUML, developers can effortlessly visualize the structure and relationships within their Go projects, aiding in code comprehension and documentation. By parsing Go source code and producing PlantUML diagrams, GoPlantUML empowers developers to create clear and concise visual representations of their codebase architecture, package dependencies, and function interactions. This tool simplifies the documentation process and enhances collaboration among team members by providing a visual overview of complex Go projects. GoPlantUML is actively maintained and welcomes contributions from the Go community.

## Want to try it on your code? 
Take a look at [www.dumels.com](https://www.dumels.com). We have created dumels using this library. 

## Code of Conduct
Please, review the code of conduct [here](https://github.com/jfeliu007/goplantuml/blob/master/CODE_OF_CONDUCT.md "here").

### Prerequisites
golang 1.17 or above

### Installing

```
go get github.com/jfeliu007/goplantuml/parser
go install github.com/jfeliu007/goplantuml/cmd/goplantuml@latest
```

This will install the command goplantuml in your GOPATH bin folder.

### Usage

```
goplantuml [-recursive] path/to/gofiles path/to/gofiles2
```
```
goplantuml [-recursive] path/to/gofiles path/to/gofiles2 > diagram_file_name.puml
```
```
Usage of goplantuml:
  -aggregate-private-members
        Show aggregations for private members. Ignored if -show-aggregations is not used.
  -hide-connections
        hides all connections in the diagram
  -hide-fields
        hides fields
  -hide-methods
        hides methods
  -ignore string
        comma separated list of folders to ignore
  -notes string
        Comma separated list of notes to be added to the diagram
  -output string
        output file path. If omitted, then this will default to standard output
  -recursive
        walk all directories recursively
  -show-aggregations
        renders public aggregations even when -hide-connections is used (do not render by default)
  -show-aliases
        Shows aliases even when -hide-connections is used
  -show-compositions
        Shows compositions even when -hide-connections is used
  -show-connection-labels
        Shows labels in the connections to identify the connections types (e.g. extends, implements, aggregates, alias of
  -show-implementations
        Shows implementations even when -hide-connections is used
  -show-options-as-note
        Show a note in the diagram with the none evident options ran with this CLI
  -title string
        Title of the generated diagram
  -hide-private-members
        Hides all private members (fields and methods)
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
See it on kroki.io

![Diagram](https://kroki.io/plantuml/svg/eNqtVNtqGzEQfd-v0OPW9eUH0lDqJBBIXJOY5iHkQdkd26JaSUiz4BL87x1d1tZmbTClelhbmpkz1zOzGYNqq8EVxXeH3GLbyELxBpzhFTDDrQPLPgpGp5LcOfaMtq0wPfnzld21qkKhlWOvb6PukssFyDoIl9wSNoLNpKs_BphDK9Qme53rxmgnPBIZDsS3OwQVMJMoyPZZnA9CwXOQ_WiFrA9ZRPsXKxC8yovA7Q0Y3Ja1_zKhcOyjSRF9YUPkuf8u88r4M_EWVJnWArtaa4Ws0lLbb--yheuGm6uZf7x-jbhvF6iMYqkzD1VrLShc8uo338CCSvm5cBPGpbxXVOE19c_9UyQxj6T1se9jx5j-D_CgdveNkdBQhu6QQumQpUKMfXOo5ulKrTnah0FNdSmVroFxh9MF_Rmq3QkJN0C9tDxM7VHfv_b0eV0_Am51vdLRaemO0TRBwkbeMox4z3QD2FGhXCcluvtpHzN1bN1nfwHplLu1F5zzRoYrnXGmZ5gxbGiUmHSRAeX0084tcIQU37lESLPLIPwszuvFhMljCTsT2nC7M9Y3u5JtDfmwv2stO0ZGHj9R7GDLUzQ97Joe8U-Rpr-DBkBdGy_AOfg8u-2egDaE-sWJD8PttS_i3U27nTWapKG108FCK2hnU_q0sf8CImDLsA==)

[Edit this diagram](https://niolesk.top/#https://kroki.io/plantuml/svg/eNqtVNtqGzEQfd-v0OPW9eUH0lDqJBBIXJOY5iHkQdkd26JaSUiz4BL87x1d1tZmbTClelhbmpkz1zOzGYNqq8EVxXeH3GLbyELxBpzhFTDDrQPLPgpGp5LcOfaMtq0wPfnzld21qkKhlWOvb6PukssFyDoIl9wSNoLNpKs_BphDK9Qme53rxmgnPBIZDsS3OwQVMJMoyPZZnA9CwXOQ_WiFrA9ZRPsXKxC8yovA7Q0Y3Ja1_zKhcOyjSRF9YUPkuf8u88r4M_EWVJnWArtaa4Ws0lLbb--yheuGm6uZf7x-jbhvF6iMYqkzD1VrLShc8uo338CCSvm5cBPGpbxXVOE19c_9UyQxj6T1se9jx5j-D_CgdveNkdBQhu6QQumQpUKMfXOo5ulKrTnah0FNdSmVroFxh9MF_Rmq3QkJN0C9tDxM7VHfv_b0eV0_Am51vdLRaemO0TRBwkbeMox4z3QD2FGhXCcluvtpHzN1bN1nfwHplLu1F5zzRoYrnXGmZ5gxbGiUmHSRAeX0084tcIQU37lESLPLIPwszuvFhMljCTsT2nC7M9Y3u5JtDfmwv2stO0ZGHj9R7GDLUzQ97Joe8U-Rpr-DBkBdGy_AOfg8u-2egDaE-sWJD8PttS_i3U27nTWapKG108FCK2hnU_q0sf8CImDLsA==)


```
goplantuml $GOPATH/src/github.com/jfeliu007/goplantuml/parser > ClassDiagram.puml
// Generates a file ClassDiagram.puml with the previous specifications
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

//MyStruct2 will be directly composed of MyStruct1 so it will have a composition relationship with it
type MyStruct2 struct {
	MyStruct1
}

//MyStruct3 will have a foo() function but the return value is not a bool, so it will not have any relationship with MyInterface
type MyStruct3 struct {
    Foo MyStruct1
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

        + Foo MyStruct1

    }
}
testingsupport.MyStruct1 *-- testingsupport.MyStruct2

testingsupport.MyInterface <|-- testingsupport.MyStruct1

testingsupport.MyStruct3 o-- testingsupport.MyStruct1

@enduml
```

![alt text](https://raw.githubusercontent.com/jfeliu007/goplantuml/master/example/example.png)

### Diagram using www.dumels.com
[UML Diagram](https://www.dumels.com/diagram/23ff0222-e93b-4e9f-a4ef-4d5d9b7a5c7d)

### Diagram rendered in plantuml online server
[UML Diagram](https://www.plantuml.com/plantuml/uml/SoWkIImgAStDuSfBp4qjBaXCJbKeIIqkoSnBBoujACWlAb6evb80WioyajIYD92qRwKdd0sIX09TXRJyV0rDXQJy_1mki6Wjc4pEIImk1ceABYagJIunLB2nKT08rd4iB4tCJIpAp4lLLB2p8zaO8np6uDHWJAozN70HRGMt_7o4ms6EgUL23HyzXDUqT7KLS4WQSM5gGmIZJGrkdOOOEX5-oiUhpI4rBmKOim00)

For instructions on how to render these diagrams locally using plantuml please visit [https://plantuml.com](https://plantuml.com)
