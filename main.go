package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"godocmd/format"
	"godocmd/parse"
)

func main() {
	app := &cli.App{
		Name:  "godoc-md",
		Usage: "Generate markdown documentation from GoDoc comments",
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
		},
		Action: func(c *cli.Context) error {
			dir := c.String("dir")
			outPath := c.String("out")

			docPkg, err := parse.LoadPackage(dir)
			if err != nil {
				return err
			}

			var out *os.File
			if outPath != "" {
				// Ensure parent directories exist
				if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
					return fmt.Errorf("creating output directory: %w", err)
				}

				out, err = os.Create(outPath)
				if err != nil {
					return fmt.Errorf("creating output file: %w", err)
				}
				defer func(out *os.File) {
					err := out.Close()
					if err != nil {
						log.Printf("error closing output file: %v", err)
					}
				}(out)

			} else {
				out = os.Stdout
			}

			return format.WriteMarkdown(docPkg, out)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
