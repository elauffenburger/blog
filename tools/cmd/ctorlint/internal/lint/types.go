package lint

import (
	"go/ast"
)

type Visibility string

const (
	Exported   Visibility = "pub"
	Unexported            = ""
)

type Struct struct {
	Name     string
	Vis      Visibility
	NoLint   bool
	Type     *ast.StructType
	TypeSpec *ast.TypeSpec
	File     *ast.File
}

type Ctor struct {
	Name       string
	Constructs string
	Vis        Visibility
	Decl       *ast.FuncDecl
}

func (c Ctor) MatchesStruct(s Struct) bool {
	for _, r := range c.Decl.Type.Results.List {
		// Flatten out *T to T
		expr := r.Type
		for {
			stexpr, isStar := expr.(*ast.StarExpr)
			if !isStar {
				break
			}

			expr = stexpr.X
		}

		switch rt := expr.(type) {
		case *ast.Ident:
			if rt.Name == s.Name {
				return true
			}
		}

	}

	return false
}
