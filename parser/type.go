package parser

import (
	"go/ast"
)

type Generic struct {
	Names []string
	Types map[string]struct{} // Keep a set to cover the case when there are multiple type names using the same type.
}

func NewGeneric() *Generic {
	return &Generic{Types: make(map[string]struct{})}
}

func (g *Generic) exists() bool {
	return len(g.Names) != 0 && len(g.Types) != 0
}

func (g *Generic) getNames(field *ast.Field) *Generic {
	for _, name := range field.Names {
		g.Names = append(g.Names, name.String())
	}

	return g
}

func (g *Generic) getTypes(field *ast.Field) *Generic {
	switch f := field.Type.(type) {
	case *ast.Ident:
		g.Types[f.Name] = struct{}{}
	case *ast.BinaryExpr:
		switch x := f.X.(type) {
		case *ast.Ident:
			g.Types[x.Name] = struct{}{}
		}

		switch y := f.Y.(type) {
		case *ast.Ident:
			g.Types[y.Name] = struct{}{}
		}

		// The below, while ugly, handles scenarios where we have N binary expressions.
		// An example of this could be
		//   type foo[T string | bool | int | int16 | float64] struct{}
		switch f.X.(type) {
		case *ast.BinaryExpr:
			var x ast.Expr = f.X.(*ast.BinaryExpr)
			var process = true
			for process {
				switch xt := x.(type) {
				case *ast.Ident:
					g.Types[xt.Name] = struct{}{}
				}

				switch yt := x.(type) {
				case *ast.Ident:
					g.Types[yt.Name] = struct{}{}
				case *ast.BinaryExpr:
					switch ytt := yt.Y.(type) {
					case *ast.Ident:
						g.Types[ytt.Name] = struct{}{}
					}
				}

				newX, safe := x.(*ast.BinaryExpr)
				process = safe
				if safe {
					x = newX.X
				}
			}
		}
	case *ast.InterfaceType:
		g.Types["interface"] = struct{}{}
	}

	return g
}
