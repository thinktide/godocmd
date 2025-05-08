package parse

import (
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// LoadPackage loads the Go package from the specified directory and returns its documentation.
//
// Parameters:
//   - dir: The path to the package directory to load
//
// Returns:
//   - *doc.Package: The parsed documentation package
//   - error: Any error encountered while parsing the directory
func LoadPackage(dir string) (*doc.Package, error) {
	fileSet := token.NewFileSet()

	files, err := parser.ParseDir(fileSet, dir, func(fi os.FileInfo) bool {
		// Skip test files
		return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, pkg := range files {
		return doc.New(pkg, dir, doc.AllDecls), nil
	}

	return nil, nil
}
