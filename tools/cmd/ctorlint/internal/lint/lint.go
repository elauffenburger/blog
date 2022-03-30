package lint

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/BooleanCat/go-functional/iter"
	"golang.org/x/tools/go/packages"
)

type Pkg struct {
	Pkg *packages.Package

	Structs map[string]Struct
	Ctors   []Ctor
}

func ParsePkg(pkg *packages.Package) (Pkg, error) {
	var (
		pkgStructs = make(map[string]Struct)
		pkgCtors   []Ctor
	)

	for _, f := range pkg.Syntax {
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

					pkgStructs[name] = Struct{
						Name:     name,
						Vis:      getVisibility(name),
						NoLint:   nolint,
						Type:     st,
						TypeSpec: spec,
						Pkg:      pkg,
						File:     f,
					}
				}
			}
		}
	}

	return Pkg{pkg, pkgStructs, pkgCtors}, nil
}

func (pkg Pkg) StructsWithoutCtors() ([]Struct, error) {
	var pkgStructs []Struct
	for _, s := range pkg.Structs {
		pkgStructs = append(pkgStructs, s)
	}

	unmatched := iter.Collect[Struct](
		iter.Filter[Struct](iter.Lift(pkgStructs), func(s Struct) bool {
			// If this struct is private, skip it.
			if s.Vis == Unexported {
				return false
			}

			// If we're not supposed to lint this struct, skip it.
			if s.NoLint {
				return false
			}

			for _, c := range pkg.Ctors {
				// If we found a matching ctor, return mark as matched.
				if c.MatchesStruct(s) {
					return false
				}
			}

			return true
		}),
	)

	return unmatched, nil
}

type PkgGroup map[string]Pkg

type StructInit struct {
	Struct *Struct
	Expr   ast.Expr
}

func (pg PkgGroup) InvalidStructInits() ([]StructInit, error) {
	var invalidInits []StructInit
	for _, pkg := range pg {
		for _, f := range pkg.Pkg.Syntax {
			// Go through all decls for this file.
			for _, decl := range f.Decls {
				// If this isn't a fn decl, skip it.
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}

				// Go through each stmt in the fn body.
				for _, stmt := range fn.Body.List {
					// If this isn't an assignment stmt, skip it.
					asgn, ok := stmt.(*ast.AssignStmt)
					if !ok {
						continue
					}

					// Check out the rhs elements of the assignment.
					for _, rhs := range asgn.Rhs {
						switch rhsT := rhs.(type) {
						// If we're creating a composit literal:
						case *ast.CompositeLit:
							switch typ := rhsT.Type.(type) {
							// If we have a raw ident, it's in the current pkg.
							case *ast.Ident:
								init, err := pg.initForStructInvalid(rhs, pkg, typ.Name)
								if err != nil {
									return nil, err
								}

								if init != nil {
									invalidInits = append(invalidInits, *init)
								}

							// If we have a selector, then it's in another pkg.
							case *ast.SelectorExpr:
								localPkgName := typ.X.(*ast.Ident).Name

								var pkgPath string
								for _, i := range f.Imports {
									if i.Name != nil {
										pkg := i.Name.Name
										if pkg == localPkgName {
											pkgPath = i.Path.Value
											break
										}
									} else {
										impPath := strings.Trim(i.Path.Value, `"`)
										impPathParts := strings.Split(impPath, "/")
										impPkgName := impPathParts[len(impPathParts)-1]

										if impPkgName == localPkgName {
											pkgPath = impPath
											break
										}
									}
								}

								// If we don't know about this pkg, bail.
								pkg, ok := pg[pkgPath]
								if !ok {
									continue
								}

								invInit, err := pg.initForStructInvalid(rhs, pkg, typ.Sel.Name)
								if err != nil {
									return nil, err
								}

								if invInit != nil {
									invalidInits = append(invalidInits, *invInit)
								}
							}
						}
					}
				}
			}
		}
	}

	return invalidInits, nil
}

func (pg PkgGroup) initForStructInvalid(expr ast.Expr, pkg Pkg, typeName string) (*StructInit, error) {
	// If this isn't a struct, bail.
	strct, ok := pkg.Structs[typeName]
	if !ok {
		return nil, nil
	}

	// If we're not supposed to lint this struct, bail.
	if strct.NoLint {
		return nil, nil
	}

	// Otherwise, this is indeed an invalid struct init!
	return &StructInit{&strct, expr}, nil
}

func (pg PkgGroup) StructsWithoutCtors() ([]Struct, error) {
	var structs []Struct
	for _, pkg := range pg {
		pkgStructs, err := pkg.StructsWithoutCtors()
		if err != nil {
			return nil, err
		}

		structs = append(structs, pkgStructs...)
	}

	return structs, nil
}
