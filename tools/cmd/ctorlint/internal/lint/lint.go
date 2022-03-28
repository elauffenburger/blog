package lint

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/BooleanCat/go-functional/iter"
)

func LintPkg(pkg string, files []*ast.File) ([]Strct, error) {
	var (
		pkgStructs []Strct
		pkgCtors   []Ctor
	)

	for _, f := range files {
		for _, decl := range f.Decls {
			switch t := decl.(type) {
			case *ast.FuncDecl:
				name := t.Name.Name

				// A ctor must:
				//   - not be a method on a type
				//   - return something
				//   - have a name that starts with "New"
				if !(t.Recv == nil && t.Type.Results != nil && strings.HasPrefix(name, "New")) {
					break
				}

				pkgCtors = append(pkgCtors, Ctor{
					Name:       name,
					Constructs: strings.TrimPrefix(name, "New"),
					Vis:        getVisibility(name),
					Decl:       t,
				})

			case *ast.GenDecl:
				// Make sure this is a type decl.
				if t.Tok != token.TYPE {
					break
				}

				for _, s := range t.Specs {
					spec := s.(*ast.TypeSpec)

					// Make sure this is a struct.
					st, isStruct := spec.Type.(*ast.StructType)
					if !isStruct {
						break
					}

					name := spec.Name.Name

					nolint := false
					if t.Doc != nil {
						for _, c := range t.Doc.List {
							if strings.HasPrefix(c.Text, "//nolint:ctors") {
								nolint = true
								break
							}
						}
					}

					pkgStructs = append(pkgStructs, Strct{name, getVisibility(name), nolint, st, spec})
				}
			}
		}
	}

	unmatchedStrcts := iter.Collect[Strct](
		iter.Filter[Strct](iter.Lift(pkgStructs), func(s Strct) bool {
			// If this struct is private, skip it.
			if s.Vis == Unexported {
				return false
			}

			// If we're not supposed to lint this struct, skip it.
			if s.NoLint {
				return false
			}

			for _, c := range pkgCtors {
				// If we found a matching ctor, return mark as matched.
				if c.MatchesStruct(s) {
					return false
				}
			}

			return true
		}),
	)

	return unmatchedStrcts, nil
}
