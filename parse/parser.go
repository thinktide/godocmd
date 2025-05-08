package parse

import (
	"fmt"
	"go/doc"
	"go/parser"
	"go/token"
	"path/filepath"
)

func LoadPackage(dir string) (*doc.Package, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("invalid directory: %w", err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, absDir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing directory: %w", err)
	}

	// pick the first package
	for _, pkg := range pkgs {
		return doc.New(pkg, absDir, 0), nil
	}

	return nil, fmt.Errorf("no Go package found in directory: %s", dir)
}
