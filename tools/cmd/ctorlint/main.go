package main

import (
	"fmt"

	"github.com/elauffenburger/blog/tools/cmd/ctorlint/internal/lint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

/*
 * Rules:
 * 	- If a struct `X` is exported and flagged for linting, it must:
 *    - Have a corresponding ctor
 *    - Not be created without invoking a ctor
 *      - Meaning, the zero value or a struct literal would be invalid
 *
 *  - For a ctor to match a struct `X`, it must:
 *    - Start with the name "New"
 *    - Have a return value of `X|*X` or `(X|*X, error)`
 *
 *  - A struct can be excluded from linting with a `nolint:ctors` comment.
 */

var analyzer = &analysis.Analyzer{
	Name: "checkctors",
	Doc:  `does the things`,
	Run: func(p *analysis.Pass) (interface{}, error) {
		pkg, err := lint.ParsePkg(p.Pkg, p.Files)
		if err != nil {
			return nil, err
		}

		structs, err := pkg.StructsWithoutCtors()
		if err != nil {
			return nil, err
		}

		for _, s := range structs {
			p.Report(analysis.Diagnostic{
				Pos:     s.Type.Pos(),
				End:     s.Type.End(),
				Message: fmt.Sprintf("struct %s is missing a valid constructor", s.Name),
			})
		}

		invInits, err := pkg.InvalidStructInits()
		if err != nil {
			return nil, err
		}

		for _, i := range invInits {
			p.Report(analysis.Diagnostic{
				Pos:     i.Expr.Pos(),
				End:     i.Expr.End(),
				Message: fmt.Sprintf("must use ctor to construct %s", i.Struct.Name),
			})
		}

		return nil, nil
	},
}

func main() {
	singlechecker.Main(analyzer)
}
