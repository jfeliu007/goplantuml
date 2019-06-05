package parser

import (
	"fmt"
	"strings"

	"go/ast"
)

const packageConstant = "{packageName}"

//Field can hold the name and type of any field
type Field struct {
	Name     string
	Type     string
	FullType string
}

//Returns a string representation of the given expression if it was recognized.
//Refer to the implementation to see the different string representations.
func getFieldType(exp ast.Expr, aliases map[string]string) string {
	switch v := exp.(type) {
	case *ast.Ident:
		if isPrimitive(v) {
			return v.Name
		}
		return fmt.Sprintf("%s%s", packageConstant, v.Name)
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", getFieldType(v.Elt, aliases))
	case *ast.SelectorExpr:
		packageName := v.X.(*ast.Ident).Name
		if alias, ok := aliases[packageName]; ok {
			packageName = alias
		}
		return fmt.Sprintf("%s.%s", packageName, v.Sel.Name)
	case *ast.MapType:
		return fmt.Sprintf("<font color=blue>map</font>[%s]%s", getFieldType(v.Key, aliases), getFieldType(v.Value, aliases))
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", getFieldType(v.X, aliases))
	case *ast.ChanType:
		return fmt.Sprintf("<font color=blue>chan</font> %s", getFieldType(v.Value, aliases))
	case *ast.StructType:
		fieldList := make([]string, 0)
		for _, field := range v.Fields.List {
			fieldList = append(fieldList, getFieldType(field.Type, aliases))
		}
		return fmt.Sprintf("<font color=blue>struct</font>{%s}", strings.Join(fieldList, ", "))
	case *ast.InterfaceType:
		methods := make([]string, 0)
		for _, field := range v.Methods.List {
			methodName := ""
			if field.Names != nil {
				methodName = field.Names[0].Name
			}
			methods = append(methods, methodName+" "+getFieldType(field.Type, aliases))
		}
		return fmt.Sprintf("<font color=blue>interface</font>{%s}", strings.Join(methods, "; "))
	case *ast.FuncType:
		function := getFunction(v, "", aliases, "")
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
		return fmt.Sprintf("...%s", getFieldType(v.Elt, aliases))
	}
	return ""
}

var globalPrimitives = map[string]struct{}{
	"bool":       {},
	"string":     {},
	"int":        {},
	"int8":       {},
	"int16":      {},
	"int32":      {},
	"int64":      {},
	"uint":       {},
	"uint8":      {},
	"uint16":     {},
	"uint32":     {},
	"uint64":     {},
	"uintptr":    {},
	"byte":       {},
	"rune":       {},
	"float32":    {},
	"float64":    {},
	"complex64":  {},
	"complex128": {},
}

func isPrimitive(ty *ast.Ident) bool {
	_, ok := globalPrimitives[ty.Name]
	return ok
}

func replacePackageConstant(field, packageName string) string {
	if packageName != "" {
		packageName = fmt.Sprintf("%s.", packageName)
	}
	return strings.Replace(field, packageConstant, packageName, 1)
}