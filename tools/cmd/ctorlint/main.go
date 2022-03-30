package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/elauffenburger/blog/tools/cmd/ctorlint/internal/lint"
	"github.com/elauffenburger/blog/tools/cmd/ctorlint/internal/utils"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
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
	var cpuprofile string

	cmd := cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			dirsToLint := args

			if cpuprofile != "" {
				f, err := os.Create(cpuprofile)
				if err != nil {
					log.Fatal(err)
				}

				pprof.StartCPUProfile(f)
				defer pprof.StopCPUProfile()
			}

			pkgs := make(map[string]*packages.Package)

			var pkgErrs error
			for _, dir := range dirsToLint {
				cfg := &packages.Config{
					Mode: packages.NeedSyntax | packages.NeedFiles | packages.NeedTypes | packages.NeedDeps,
					Dir:  dir,
				}

				var pkgsToLint []string
				filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
					// If this isn't a dir, ignore it.
					if !d.IsDir() {
						return nil
					}

					files, err := os.ReadDir(path)
					if err != nil {
						return err
					}

					// If this dir contains any go files, it's considered a package.
					for _, f := range files {
						if filepath.Ext(f.Name()) == ".go" {
							pkgsToLint = append(pkgsToLint, path)
							break
						}
					}

					return nil
				})

				ps, err := packages.Load(cfg, pkgsToLint...)
				if err != nil {
					return err
				}

			pkgsLoop:
				for _, pkg := range ps {
					for _, err := range pkg.Errors {
						if strings.HasPrefix(err.Msg, "build constraints exclude all") {
							continue pkgsLoop
						}

						pkgErrs = multierror.Append(pkgErrs, err)
					}

					pkgs[pkg.ID] = pkg
				}
			}

			if pkgErrs != nil {
				return pkgErrs
			}

			// Parse each pkg into a `lint.PkgElements` that's lookupable by pkg name
			// and add to the package group.
			pg := make(lint.PkgGroup)
			for _, pkg := range pkgs {
				parsedPkg, err := lint.ParsePkg(pkg)
				if err != nil {
					return err
				}

				pg[pkg.ID] = parsedPkg
			}

			// Find and report invalid stuff!

			structsWithoutCtors, err := pg.StructsWithoutCtors()
			if err != nil {
				return err
			}

			if len(structsWithoutCtors) > 0 {
				for _, s := range structsWithoutCtors {
					fmt.Printf("type without ctor: %s: %s\n", s.Pkg.Fset.Position(s.Type.Pos()), s.Name)
				}
			}

			invalidStructInits, err := pg.InvalidStructInits()
			if err != nil {
				return err
			}

			for _, init := range invalidStructInits {
				s := init.Struct

				fmt.Printf("type init without ctor: %s: %s\n", s.Pkg.Fset.Position(init.Expr.Pos()), s.Name)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&cpuprofile, "cpuprofile", "", "where to write CPU profiling report to")

	utils.NoError(cmd.Execute())
}
