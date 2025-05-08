package godocmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/thinktide/godocmd/format"
	"github.com/thinktide/godocmd/parse"
)

// GenerateMarkdown recursively walks a directory and generates markdown documentation
// for each Go package it finds.
//
// Parameters:
//   - rootDir: The directory to scan for packages
//   - out: An io.Writer to write the generated markdown to
//   - recursive: Whether to scan subdirectories
//   - includePrivate: Whether to include unexported symbols
//
// Returns:
//   - error: Any error encountered during parsing or writing
func GenerateMarkdown(rootDir string, out io.Writer, recursive, includePrivate bool) error {
	dirs, err := collectGoPackageDirs(rootDir)
	if err != nil {
		return err
	}

	for i, dir := range dirs {
		docPkg, err := parse.LoadPackage(dir)
		if err != nil {
			continue
		}
		if len(docPkg.Funcs)+len(docPkg.Types) == 0 {
			continue
		}

		fmt.Fprintf(out, "<!-- %s -->\n\n", dir)
		if err := format.WriteMarkdown(docPkg, out, includePrivate); err != nil {
			continue
		}
		if i < len(dirs)-1 {
			fmt.Fprintln(out, "\n---\n")
		}
	}
	return nil
}

func collectGoPackageDirs(root string) ([]string, error) {
	var dirs []string
	seen := map[string]bool{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		dir := filepath.Dir(path)
		if seen[dir] {
			return nil
		}
		trimmed := strings.TrimPrefix(dir, "./")
		if trimmed == "vendor" || trimmed == "docs" || strings.HasPrefix(trimmed, ".") ||
			strings.HasPrefix(trimmed, "terraform") || strings.Contains(trimmed, "/.github") {
			return nil
		}
		seen[dir] = true
		dirs = append(dirs, dir)
		return nil
	})

	return dirs, err
}
