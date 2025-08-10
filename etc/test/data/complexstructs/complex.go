package complexstructs

import (
	"fmt"
	"strings"
)

// ComplexStruct demonstrates complex field types and relationships
type ComplexStruct struct {
	SimpleField    string
	SliceField     []int
	MapField       map[string]int
	ChannelField   chan bool
	FunctionField  func(string) error
	AliasField     CustomAlias
	ComposedStruct EmbeddedStruct
}

// EmbeddedStruct demonstrates struct embedding
type EmbeddedStruct struct {
	EmbeddedField string
}

// CustomAlias demonstrates type alias with complex underlying type
type CustomAlias func(strings.Builder) bool

// ProcessorInterface demonstrates interface with complex methods
type ProcessorInterface interface {
	Process(data map[string]interface{}) ([]byte, error)
	Transform(input <-chan string) <-chan string
}

// ComplexProcessor implements ProcessorInterface
type ComplexProcessor struct {
	config map[string]string
}

func (cp *ComplexProcessor) Process(data map[string]interface{}) ([]byte, error) {
	return []byte(fmt.Sprintf("%v", data)), nil
}

func (cp *ComplexProcessor) Transform(input <-chan string) <-chan string {
	output := make(chan string)
	go func() {
		for s := range input {
			output <- strings.ToUpper(s)
		}
		close(output)
	}()
	return output
}
