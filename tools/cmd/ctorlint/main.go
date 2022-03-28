package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"github.com/elauffenburger/blog/tools/cmd/ctorlint/internal/lint"
	"github.com/elauffenburger/blog/tools/cmd/ctorlint/internal/utils"
	"github.com/mattn/go-zglob"
	"github.com/spf13/cobra"
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

func main() {
	cmd := cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			dirsToLint := args

			fset := token.NewFileSet()
			astFilesByPkg := make(map[string][]*ast.File)
			for _, dir := range dirsToLint {
				absDir, err := filepath.Abs(dir)
				if err != nil {
					return err
				}

				files, err := zglob.Glob(filepath.Join(absDir, "/**/*.go"))
				if err != nil {
					return err
				}

				for _, f := range files {
					pkg := filepath.Dir(f)

					astFile, err := parser.ParseFile(fset, f, nil, parser.ParseComments)
					if err != nil {
						return err
					}

					astFilesByPkg[pkg] = append(astFilesByPkg[pkg], astFile)
				}
			}

			for pkg, files := range astFilesByPkg {
				unmatched, err := lint.LintPkg(pkg, files)
				if err != nil {
					return err
				}

				fmt.Printf("unmatched structs: %#v\n\n", unmatched)
			}

			return nil
		},
	}

	utils.NoError(cmd.Execute())
}
