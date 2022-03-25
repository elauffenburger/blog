package main

import (
	"path/filepath"

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

			filesByPkg := make(map[string][]string)
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
					filesByPkg[pkg] = append(filesByPkg[pkg], f)
				}
			}

			for pkg, files := range filesByPkg {
				if err := lintPkg(pkg, files); err != nil {
					return err
				}
			}

			return nil
		},
	}

	noerror(cmd.Execute())
}

func NewStrPtrWithErr() (*string, error) {
	return nil, nil
}

//nolint:ctors
type foo struct {
}

// some other comment!
type Foo struct {
}

//nolint:ctors
type Bar struct {
}

func NewFooPtrWithErr() (Foo, error) {
	return Foo{}, nil
}
