package parser

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestNewGeneric(t *testing.T) {
	g := NewGeneric()
	if g == nil {
		t.Fatal("Returned value should not be nil")
	}
}

func TestExists(t *testing.T) {
	g := NewGeneric()
	if g.exists() {
		t.Fatal("Should not exist at this point")
	}

	g.getNames(&ast.Field{
		Names: []*ast.Ident{{Name: "test"}},
	}).getTypes(&ast.Field{
		Type: new(ast.Ident),
	})

	if !g.exists() {
		t.Fatal("Should exist at this point")
	}
}

func TestGetTypes(t *testing.T) {
	table := []struct {
		names         []*ast.Ident
		typ           func() ast.Expr
		expectedNames []string
		expectedTypes map[string]struct{}
	}{{
		names: []*ast.Ident{{Name: "any"}},
		typ: func() ast.Expr {
			e := new(ast.Ident)
			e.Name = "any"
			return e
		},
		expectedNames: []string{"any"},
		expectedTypes: map[string]struct{}{"any": {}},
	}, {
		names: []*ast.Ident{{Name: "int"}},
		typ: func() ast.Expr {
			e := new(ast.BinaryExpr)
			e.X = new(ast.Ident)
			e.X.(*ast.Ident).Name = "int"
			e.Y = new(ast.Ident)
			e.Y.(*ast.Ident).Name = "bool"
			return e
		},
		expectedNames: []string{"int"},
		expectedTypes: map[string]struct{}{"int": {}, "bool": {}},
	}, {
		names: []*ast.Ident{{Name: "interface"}},
		typ: func() ast.Expr {
			e := new(ast.InterfaceType)
			return e
		},
		expectedNames: []string{"interface"},
		expectedTypes: map[string]struct{}{"interface": {}},
	}}

	for _, entry := range table {
		g := NewGeneric()
		field := &ast.Field{Names: entry.names, Type: entry.typ()}
		g.getNames(field).getTypes(field)
		if !reflect.DeepEqual(g.Names, entry.expectedNames) {
			t.Errorf("Mismatched names: %v %v", g.Names, entry.expectedNames)
		}

		if !reflect.DeepEqual(g.Types, entry.expectedTypes) {
			t.Errorf("Mismatched types: %v %v", g.Types, entry.expectedTypes)
		}
	}
}
