package repository

import (
	"fmt"
	"go/ast"
	"strings"
)

const packageConstant = "{packageName}"

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
	default:
		return "", nil
	}
}

func getIdent(v *ast.Ident, aliases map[string]string) (string, []string) {
	if alias, ok := aliases[v.Name]; ok {
		return alias, nil
	}
	return v.Name, nil
}

func getArrayType(v *ast.ArrayType, aliases map[string]string) (string, []string) {
	field, subFields := getFieldType(v.Elt, aliases)
	return fmt.Sprintf("[]%s", field), subFields
}

func getSelectorExp(v *ast.SelectorExpr, aliases map[string]string) (string, []string) {
	theType, _ := getFieldType(v.X, aliases)
	return fmt.Sprintf("%s.%s", theType, v.Sel.Name), nil
}

// Updated to remove HTML font tag (PlantUML member type syntax error対策)
func getMapType(v *ast.MapType, aliases map[string]string) (string, []string) {
	key, keyDependencies := getFieldType(v.Key, aliases)
	value, valueDependencies := getFieldType(v.Value, aliases)
	return fmt.Sprintf("map[%s]%s", key, value), append(keyDependencies, valueDependencies...)
}

func getStarExp(v *ast.StarExpr, aliases map[string]string) (string, []string) {
	field, subFields := getFieldType(v.X, aliases)
	return fmt.Sprintf("*%s", field), subFields
}

func getChanType(v *ast.ChanType, aliases map[string]string) (string, []string) {
	switch v.Dir {
	case ast.SEND:
		chType, deps := getFieldType(v.Value, aliases)
		return fmt.Sprintf("chan<- %s", chType), deps
	case ast.RECV:
		chType, deps := getFieldType(v.Value, aliases)
		return fmt.Sprintf("<-chan %s", chType), deps
	default:
		chType, deps := getFieldType(v.Value, aliases)
		return fmt.Sprintf("chan %s", chType), deps
	}
}

// Updated to remove HTML font tag
func getStructType(v *ast.StructType, aliases map[string]string) (string, []string) {
	var fields []string
	var allDeps []string
	for _, field := range v.Fields.List {
		fieldType, deps := getFieldType(field.Type, aliases)
		fields = append(fields, fieldType)
		allDeps = append(allDeps, deps...)
	}
	return fmt.Sprintf("struct{%s}", strings.Join(fields, ", ")), allDeps
}

// Updated to remove HTML font tag
func getInterfaceType(v *ast.InterfaceType, aliases map[string]string) (string, []string) {
	return "interface{}", nil
}

func getFuncType(v *ast.FuncType, aliases map[string]string) (string, []string) {
	params := []string{}
	var allDeps []string
	if v.Params != nil {
		for _, param := range v.Params.List {
			paramType, deps := getFieldType(param.Type, aliases)
			params = append(params, paramType)
			allDeps = append(allDeps, deps...)
		}
	}
	results := []string{}
	if v.Results != nil {
		for _, result := range v.Results.List {
			resultType, deps := getFieldType(result.Type, aliases)
			results = append(results, resultType)
			allDeps = append(allDeps, deps...)
		}
	}
	return fmt.Sprintf("func(%s) (%s)", strings.Join(params, ", "), strings.Join(results, ", ")), allDeps
}

func getEllipsis(v *ast.Ellipsis, aliases map[string]string) (string, []string) {
	field, subFields := getFieldType(v.Elt, aliases)
	return fmt.Sprintf("...%s", field), subFields
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

func isPrimitiveString(t string) bool {
	_, ok := globalPrimitives[t]
	return ok
}

func replacePackageConstant(field, packageName string) string {
	if strings.Contains(field, packageConstant) {
		field = strings.ReplaceAll(field, packageConstant, packageName)
	}
	return field
}

// generate and return a function object from the given Functype. The names must be passed to this
// function since the FuncType does not have this information
