package lint_test

import (
	_ "embed"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/BooleanCat/go-functional/iter"
	"github.com/elauffenburger/blog/tools/cmd/ctorlint/internal/lint"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

//go:embed .testing/testsrc.go
var testsrc string

func TestStructsWithoutCtors(t *testing.T) {
	pkg := parseTestSrc(t)

	unmatchedStructs, err := pkg.StructsWithoutCtors()
	require.NoError(t, err)

	unmatchedByName := iter.Fold[lint.Struct](
		iter.Lift(unmatchedStructs),
		make(map[string]lint.Struct),
		func(acc map[string]lint.Struct, s lint.Struct) map[string]lint.Struct {
			acc[s.Name] = s

			return acc
		},
	)

	t.Run("are reported when they", func(t *testing.T) {
		t.Run("have no ctor", func(t *testing.T) {
			require.Contains(t, unmatchedByName, "InvalidNoCtor")
		})

		t.Run("have an invalid ctor", func(t *testing.T) {
			require.Contains(t, unmatchedByName, "InvalidCtor")
		})
	})

	t.Run("are not reported when they", func(t *testing.T) {
		t.Run("are missing a ctor but are unexported", func(t *testing.T) {
			require.NotContains(t, unmatchedByName, "valid")
		})

		t.Run("are missing a ctor but are unexported and nolint", func(t *testing.T) {
			require.NotContains(t, unmatchedByName, "validNoLint")
		})

		t.Run("are missing a ctor but are nolint", func(t *testing.T) {
			require.NotContains(t, unmatchedByName, "ValidNoLint")
		})

		t.Run("have a T ctor", func(t *testing.T) {
			require.NotContains(t, unmatchedByName, "ValidT")
		})

		t.Run("have a *T ctor", func(t *testing.T) {
			require.NotContains(t, unmatchedByName, "ValidTPtr")
		})

		t.Run("have a valid (T, error) ctor", func(t *testing.T) {
			require.NotContains(t, unmatchedByName, "ValidTErr")
		})

		t.Run("have a valid (*T, error) ctor", func(t *testing.T) {
			require.NotContains(t, unmatchedByName, "ValidTPtrErr")
		})
	})
}

func TestInvalidStructInits(t *testing.T) {
	pkg := parseTestSrc(t)
	pkgGroup := lint.PkgGroup{pkg.Pkg.Name: pkg}

	invalidInits, err := pkgGroup.InvalidStructInits()
	require.NoError(t, err)

	require.Len(t, invalidInits, 1)
}

func parseTestSrc(t *testing.T) lint.Pkg {
	fileset := token.NewFileSet()

	f, err := parser.ParseFile(fileset, "testsrc.go", testsrc, parser.ParseComments)
	require.NoError(t, err)

	astPkg := packages.Package{
		Name:   "testsrc",
		Syntax: []*ast.File{f},
	}

	pkg, err := lint.ParsePkg(&astPkg)
	require.NoError(t, err)

	return pkg
}
