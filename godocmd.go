package godocmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/thinktide/godocmd/enums"
	"github.com/thinktide/godocmd/format"
	"github.com/thinktide/godocmd/parse"
)

// GenerateMarkdown recursively walks the provided directory and writes
// markdown documentation for each Go package to the given writer.
//
// By default, it excludes non-exported (private) symbols and undocumented items.
// Behavior can be modified by passing any combination of flags from the enums package.
//
// Parameters:
//   - rootDir: The base directory to scan
//   - out: The writer to output markdown to (e.g., os.Stdout or a file)
//   - flags: One or more enums.MarkdownFlag values to alter generation behavior
//
// Returns:
//   - error: Any error encountered during parsing or output
func GenerateMarkdown(rootDir string, out io.Writer, flags ...enums.MarkdownFlag) error {
	cfg := struct {
		recursive           bool
		includePrivate      bool
		includeUndocumented bool
		verbose             bool
	}{}

	for _, flag := range flags {
		switch flag {
		case enums.Recursive:
			cfg.recursive = true
		case enums.IncludePrivate:
			cfg.includePrivate = true
		case enums.IncludeUndocumented:
			cfg.includeUndocumented = true
		case enums.Verbose:
			cfg.verbose = true
		}
	}

	dirs := []string{rootDir}
	if cfg.recursive {
		var err error
		dirs, err = collectGoPackageDirs(rootDir)
		if err != nil {
			return fmt.Errorf("collecting package dirs: %w", err)
		}
		if cfg.verbose {
			fmt.Fprintf(os.Stderr, "🔍 Found %d Go package directories\n", len(dirs))
		}
	}

	for i, dir := range dirs {
		if cfg.verbose {
			fmt.Fprintf(os.Stderr, "📦 Parsing package: %s\n", dir)
		}

		docPkg, err := parse.LoadPackage(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Skipping %s: %v\n", dir, err)
			continue
		}

		if len(docPkg.Types)+len(docPkg.Funcs) == 0 {
			if cfg.verbose {
				fmt.Fprintf(os.Stderr, "⚠️  Skipping %s: no exported symbols\n", dir)
			}
			continue
		}

		fmt.Fprintf(out, "<!-- %s -->\n\n", dir)
		if err := format.WriteMarkdownWithOptions(docPkg, out, cfg.includePrivate, cfg.includeUndocumented); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to write markdown for %s: %v\n", dir, err)
			continue
		}

		if i < len(dirs)-1 {
			fmt.Fprint(out, "\n\n---\n")
		}
	}

	return nil
}

// collectGoPackageDirs walks the file tree and collects valid Go package directories.
//
// Parameters:
//   - root: The root directory to start scanning from
//
// Returns:
//   - []string: List of package paths that contain buildable Go source files
//   - error: Any filesystem error encountered
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
