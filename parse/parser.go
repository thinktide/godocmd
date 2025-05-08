package parse

import (
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// LoadPackage parses a Go package from the specified directory and returns its documentation.
// It handles both relative and absolute paths, and ensures the directory contains valid Go code.
//
// Parameters:
//   - dir: The directory path containing the Go package to parse
//
// Returns:
//   - *doc.Package: The parsed package documentation
//   - error: Any error that occurred during parsing, including invalid directory or no Go package found
func LoadPackage(dir string) (*doc.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("invalid directory: %w", err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing directory: %w", err)
	}

	// pick the first package
	for _, pkg := range pkgs {
		return doc.New(pkg, absDir, 0), nil
	}

	return nil, fmt.Errorf("no Go package found in directory: %s", dir)
}
