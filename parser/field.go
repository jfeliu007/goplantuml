package parser

import (
	"fmt"
	"strings"

	"go/ast"
)

const packageConstant = "{packageName}"

// Field can hold the name and type of any field
type Field struct {
	Name     string
	Type     string
	FullType string
}

// Returns a string representation of the given expression if it was recognized.
// Refer to the implementation to see the different string representations.
func getFieldType(exp ast.Expr, aliases map[string]string) (string, []string) {
	switch v := exp.(type) {
	case *ast.Ident:
		return getIdent(v, aliases)
	case *ast.ArrayType:
		return getArrayType(v, aliases)
	case *ast.SelectorExpr:
		return getSelectorExp(v, aliases)
	case *ast.MapType:
		return getMapType(v, aliases)
	case *ast.StarExpr:
		return getStarExp(v, aliases)
	case *ast.ChanType:
		return getChanType(v, aliases)
	case *ast.StructType:
		return getStructType(v, aliases)
	case *ast.InterfaceType:
		return getInterfaceType(v, aliases)
	case *ast.FuncType:
		return getFuncType(v, aliases)
	case *ast.Ellipsis:
		return getEllipsis(v, aliases)
	}
	return "", []string{}
}

func getIdent(v *ast.Ident, aliases map[string]string) (string, []string) {

	if isPrimitive(v) {
		return v.Name, []string{}
	}
	t := fmt.Sprintf("%s%s", packageConstant, v.Name)
	return t, []string{t}
}

func getArrayType(v *ast.ArrayType, aliases map[string]string) (string, []string) {
	t, fundamentalTypes := getFieldType(v.Elt, aliases)
	return fmt.Sprintf("[]%s", t), fundamentalTypes
}

func getSelectorExp(v *ast.SelectorExpr, aliases map[string]string) (string, []string) {

	packageName := v.X.(*ast.Ident).Name
	if realPackageName, ok := aliases[packageName]; ok {
		packageName = realPackageName
	}
	t := fmt.Sprintf("%s.%s", packageName, v.Sel.Name)
	return t, []string{t}
}

func getMapType(v *ast.MapType, aliases map[string]string) (string, []string) {

	t1, f1 := getFieldType(v.Key, aliases)
	t2, f2 := getFieldType(v.Value, aliases)
	return fmt.Sprintf("<font color=blue>map</font>[%s]%s", t1, t2), append(f1, f2...)
}

func getStarExp(v *ast.StarExpr, aliases map[string]string) (string, []string) {

	t, f := getFieldType(v.X, aliases)
	return fmt.Sprintf("*%s", t), f
}

func getChanType(v *ast.ChanType, aliases map[string]string) (string, []string) {

	t, f := getFieldType(v.Value, aliases)
	return fmt.Sprintf("<font color=blue>chan</font> %s", t), f
}

func getStructType(v *ast.StructType, aliases map[string]string) (string, []string) {

	fieldList := make([]string, 0)
	for _, field := range v.Fields.List {
		t, _ := getFieldType(field.Type, aliases)
		fieldList = append(fieldList, t)
	}
	return fmt.Sprintf("<font color=blue>struct</font>{%s}", strings.Join(fieldList, ", ")), []string{}
}

func getInterfaceType(v *ast.InterfaceType, aliases map[string]string) (string, []string) {

	methods := make([]string, 0)
	for _, field := range v.Methods.List {
		methodName := ""
		if field.Names != nil {
			methodName = field.Names[0].Name
		}
		t, _ := getFieldType(field.Type, aliases)
		methods = append(methods, methodName+" "+t)
	}
	return fmt.Sprintf("<font color=blue>interface</font>{%s}", strings.Join(methods, "; ")), []string{}
}

func getFuncType(v *ast.FuncType, aliases map[string]string) (string, []string) {

	function := getFunction(v, "", aliases, "")
	params := make([]string, 0)
	for _, pa := range function.Parameters {
		params = append(params, pa.Type)
	}
	returns := ""
	returnList := make([]string, 0)
	returnList = append(returnList, function.ReturnValues...)
	if len(returnList) > 1 {
		returns = fmt.Sprintf("(%s)", strings.Join(returnList, ", "))
	} else {
		returns = strings.Join(returnList, "")
	}
	return fmt.Sprintf("<font color=blue>func</font>(%s) %s", strings.Join(params, ", "), returns), []string{}
}

func getEllipsis(v *ast.Ellipsis, aliases map[string]string) (string, []string) {
	t, _ := getFieldType(v.Elt, aliases)
	return fmt.Sprintf("...%s", t), []string{}
}

var globalPrimitives = map[string]struct{}{
	"bool":        {},
	"string":      {},
	"int":         {},
	"int8":        {},
	"int16":       {},
	"int32":       {},
	"int64":       {},
	"uint":        {},
	"uint8":       {},
	"uint16":      {},
	"uint32":      {},
	"uint64":      {},
	"uintptr":     {},
	"byte":        {},
	"rune":        {},
	"float32":     {},
	"float64":     {},
	"complex64":   {},
	"complex128":  {},
	"error":       {},
	"*bool":       {},
	"*string":     {},
	"*int":        {},
	"*int8":       {},
	"*int16":      {},
	"*int32":      {},
	"*int64":      {},
	"*uint":       {},
	"*uint8":      {},
	"*uint16":     {},
	"*uint32":     {},
	"*uint64":     {},
	"*uintptr":    {},
	"*byte":       {},
	"*rune":       {},
	"*float32":    {},
	"*float64":    {},
	"*complex64":  {},
	"*complex128": {},
	"*error":      {},
}

func isPrimitive(ty *ast.Ident) bool {
	return isPrimitiveString(ty.Name)
}

func isPrimitiveString(t string) bool {
	_, ok := globalPrimitives[t]
	return ok
}

func replacePackageConstant(field, packageName string) string {
	if packageName != "" {
		packageName = fmt.Sprintf("%s.", packageName)
	}
	return strings.Replace(field, packageConstant, packageName, 1)
}
