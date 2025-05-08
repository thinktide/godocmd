package parse

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// parseGoDocPackage is a test helper that builds a *doc.Package from raw Go source code.
//
// Parameters:
//   - name: The package name
//   - code: Go source string to parse
//
// Returns:
//   - *doc.Package: Parsed documentation package
func parseGoDocPackage(name, code string) *doc.Package {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, name+".go", code, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	pkg, err := doc.NewFromFiles(fset, []*ast.File{file}, "./"+name, doc.AllDecls)
	if err != nil {
		panic(err)
	}

	return pkg
}

func TestParseGoDocPackage_ExtractsFuncs(t *testing.T) {
	src := `
		package testpkg

		// Add returns the sum of two integers.
		func Add(a, b int) int {
			return a + b
		}
	`
	pkg := parseGoDocPackage("testpkg", src)
	if len(pkg.Funcs) != 1 {
		t.Fatalf("expected 1 func, got %d", len(pkg.Funcs))
	}
	if pkg.Funcs[0].Name != "Add" {
		t.Errorf("expected func name 'Add', got %s", pkg.Funcs[0].Name)
	}
	if !strings.Contains(pkg.Funcs[0].Doc, "Add returns") {
		t.Errorf("expected doc comment to mention 'Add returns', got: %s", pkg.Funcs[0].Doc)
	}
}

func TestParseGoDocPackage_HandlesNoFuncs(t *testing.T) {
	src := `
		package testpkg

		type Foo struct{}
	`
	pkg := parseGoDocPackage("testpkg", src)
	if len(pkg.Funcs) != 0 {
		t.Fatalf("expected 0 funcs, got %d", len(pkg.Funcs))
	}
	if len(pkg.Types) != 1 || pkg.Types[0].Name != "Foo" {
		t.Errorf("expected type Foo to be parsed, got %+v", pkg.Types)
	}
}
