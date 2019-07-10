package parser

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestGetFunction(t *testing.T) {

	tt := []struct {
		Name           string
		Func           *ast.FuncType
		ExpectedResult *Function
		FunctionName   string
	}{
		{
			Name: "Function with to typed parameters",
			Func: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "param1",
								},
							},
							Type: &ast.Ident{
								Name: "int",
							},
						},
					},
				},
			},
			ExpectedResult: &Function{
				Name: "TestFunction",
				Parameters: []*Field{
					{
						Name:     "param1",
						Type:     "int",
						FullType: "int",
					},
				},
			},
			FunctionName: "TestFunction",
		},
		{
			Name:           "Function with to parameters only one typed",
			Func:           &ast.FuncType{},
			ExpectedResult: &Function{},
			FunctionName:   "TestFunction",
		},
		{
			Name: "Function with to typed parameters",
			Func: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "param1",
								},
								{
									Name: "param2i",
								},
							},
							Type: &ast.Ident{
								Name: "int",
							},
						},
					},
				},
			},
			ExpectedResult: &Function{
				Name: "TestFunction",
				Parameters: []*Field{
					{
						Name:     "param1",
						Type:     "int",
						FullType: "int",
					},
					{
						Name:     "param2",
						Type:     "int",
						FullType: "int",
					},
				},
			},
			FunctionName: "TestFunction",
		},
		{
			Name:           "Function with to parameters only one typed",
			Func:           &ast.FuncType{},
			ExpectedResult: &Function{},
			FunctionName:   "TestFunction",
		},
	}

	for _, tc := range tt {
		function := getFunction(tc.Func, tc.FunctionName, map[string]string{
			"main": "main",
		}, "main")

		if !reflect.DeepEqual(function, tc.ExpectedResult) {
			t.Errorf("Expected function to be %v, got %v", tc.ExpectedResult, function)
		}
	}
}
