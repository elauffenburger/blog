package main

import (
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

			// Glob all `.go` files from the provided dirs and parse them into `ast.File`s grouped by pkg.
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

			// Parse each pkg into a `lint.PkgElements` that's lookupable by pkg name.
			var pkgs lint.PkgGroup
			for pkg, files := range astFilesByPkg {
				pkgElems, err := lint.ParsePkg(pkg, files)
				if err != nil {
					return err
				}

				pkgs[pkg] = pkgElems
			}

			return nil
		},
	}

	utils.NoError(cmd.Execute())
}
