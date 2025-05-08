package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/thinktide/godocmd/format"
	"github.com/thinktide/godocmd/parse"

	"github.com/urfave/cli/v2"
)

// main is the CLI entrypoint for the godocmd tool.
// It parses CLI flags, loads one or more Go packages, and generates Markdown documentation.
//
// Parameters:
//   - none (standard CLI entrypoint)
//
// Returns:
//   - none (exits the program on error)
func main() {
	app := &cli.App{
		Name:  "godocmd",
		Usage: "Generate Markdown documentation from Go packages",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dir",
				Aliases:  []string{"d"},
				Usage:    "Directory to scan for Go packages",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "out",
				Aliases: []string{"o"},
				Usage:   "Output markdown file (default is stdout)",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Usage:   "Recursively find and document all Go packages under the given directory",
			},
			&cli.BoolFlag{
				Name:    "include-private",
				Aliases: []string{"p"},
				Usage:   "Include unexported (non-exported) functions and types in the output",
			},
		},
		Action: func(c *cli.Context) error {
			dir := c.String("dir")
			outPath := c.String("out")
			recursive := c.Bool("recursive")
			includePrivate := c.Bool("include-private")
			return runRecursive(dir, outPath, recursive, includePrivate)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// runRecursive handles recursive or non-recursive documentation generation.
//
// Parameters:
//   - rootDir: The base directory to scan
//   - outPath: File path for output, or empty string for stdout
//   - recursive: Whether to search directories recursively
//   - includePrivate: Whether to include non-exported symbols in the output
//
// Returns:
//   - error: An error if scanning or writing fails
func runRecursive(rootDir, outPath string, recursive, includePrivate bool) error {
	var out *os.File
	var err error

	if outPath != "" {
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
		out, err = os.Create(outPath)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	var dirs []string
	if recursive {
		dirs, err = collectGoPackageDirs(rootDir)
		if err != nil {
			return fmt.Errorf("walking directory tree: %w", err)
		}
		log.Printf("📂 found %d package dirs", len(dirs))
	} else {
		dirs = []string{rootDir}
	}

	for i, dir := range dirs {
		log.Printf("📦 processing %s", dir)

		docPkg, err := parse.LoadPackage(dir)
		if err != nil {
			log.Printf("⚠️ skipping %s: %v", dir, err)
			continue
		}

		if len(docPkg.Types) == 0 && len(docPkg.Funcs) == 0 {
			log.Printf("⚠️ skipping %s: no exported types or functions", dir)
			continue
		}

		fmt.Fprintf(out, "\n<!-- %s -->\n\n", dir)
		if err := format.WriteMarkdown(docPkg, out, includePrivate); err != nil {
			log.Printf("⚠️ failed to write markdown for %s: %v", dir, err)
			continue
		}

		if recursive && i < len(dirs)-1 {
			fmt.Fprintln(out, "\n---\n")
		}
	}

	return nil
}

// collectGoPackageDirs walks the file tree from the given root and returns all directories
// that contain one or more non-test .go source files.
//
// Parameters:
//   - root: The root directory to start scanning from
//
// Returns:
//   - []string: A slice of valid Go package directories
//   - error: Any error encountered during scanning
func collectGoPackageDirs(root string) ([]string, error) {
	var dirs []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}

		base := filepath.Base(path)
		if base == "vendor" || base == "testdata" || (base != "." && strings.HasPrefix(base, ".")) {
			log.Printf("🚫 skipping special dir: %s", path)
			return filepath.SkipDir
		}

		log.Printf("🔍 checking: %s", path)

		entries, err := os.ReadDir(path)
		if err != nil {
			log.Printf("❌ cannot read dir %s: %v", path, err)
			return nil
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
				log.Printf("✅ valid Go package: %s", path)
				dirs = append(dirs, path)
				break
			}
		}
		return nil
	})

	return dirs, err
}
