package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/thinktide/godocmd"
)

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
				Usage:   "Recursively find and document all Go packages",
			},
			&cli.BoolFlag{
				Name:    "include-private",
				Aliases: []string{"p"},
				Usage:   "Include unexported (non-exported) functions and types",
			},
		},
		Action: func(c *cli.Context) error {
			dir := c.String("dir")
			outPath := c.String("out")
			recursive := c.Bool("recursive")
			includePrivate := c.Bool("include-private")

			var out *os.File
			var err error

			if outPath != "" {
				if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
					return err
				}
				out, err = os.Create(outPath)
				if err != nil {
					return err
				}
				defer out.Close()
			} else {
				out = os.Stdout
			}

			return godocmd.GenerateMarkdown(dir, out, recursive, includePrivate)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
