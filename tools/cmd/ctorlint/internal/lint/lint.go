package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/BooleanCat/go-functional/iter"
)

type PkgElements struct {
	files []*ast.File

	Name    string
	Structs map[string]Struct
	Ctors   []Ctor
}

func ParsePkg(pkg string, fset *token.FileSet, files []*ast.File) (PkgElements, error) {
	var (
		pkgStructs = make(map[string]Struct)
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

					pkgStructs[name] = Struct{
						Name:     name,
						Vis:      getVisibility(name),
						NoLint:   nolint,
						Type:     st,
						TypeSpec: spec,
						FileSet:  fset,
						File:     f,
					}
				}
			}
		}
	}

	return PkgElements{files, pkg, pkgStructs, pkgCtors}, nil
}

func (pkg PkgElements) StructsWithoutCtors() ([]Struct, error) {
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

type PkgGroup map[string]PkgElements

type StructInit struct {
	Struct *Struct
	Expr   ast.Expr
}

func (pg PkgGroup) InvalidStructInits() ([]StructInit, error) {
	var invalidInits []StructInit
	for _, pkg := range pg {
		for _, f := range pkg.files {
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
							// If we have a raw ident, it's in the current pkg
							case *ast.Ident:
								init, err := pg.initForStructInvalid(rhs, pkg.Name, typ.Name)
								if err != nil {
									return nil, err
								}

								if init != nil {
									invalidInits = append(invalidInits, *init)
								}

							// If we have a selector, then it's in another pkg.
							case *ast.SelectorExpr:
								init, err := pg.initForStructInvalid(rhs, typ.X.(*ast.Ident).Name, typ.Sel.Name)
								if err != nil {
									return nil, err
								}

								if init != nil {
									invalidInits = append(invalidInits, *init)
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

func (pg PkgGroup) initForStructInvalid(expr ast.Expr, pkgName, structName string) (*StructInit, error) {
	// If we don't know about this pkg, bail.
	pkgElems, ok := pg[pkgName]
	if !ok {
		return nil, nil
	}

	// If we know about this pkg, but we don't know about this _struct_, that's probably an error.
	strct, ok := pkgElems.Structs[structName]
	if !ok {
		return nil, fmt.Errorf("couldn't find struct '%s' in pkg '%s'; this is probably an error.", structName, pkgName)
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
