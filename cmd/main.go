package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/thinktide/godocmd"
	"github.com/thinktide/godocmd/enums"
	"github.com/urfave/cli/v2"
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
				Usage:   "Recursively find and document all Go packages under the given directory",
			},
			&cli.BoolFlag{
				Name:    "include-private",
				Aliases: []string{"p"},
				Usage:   "Include unexported (non-exported) functions and types in the output",
			},
			&cli.BoolFlag{
				Name:  "include-undocumented",
				Usage: "Include functions and types that lack GoDoc comments",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose log output",
			},
		},
		Action: func(c *cli.Context) error {
			dir := c.String("dir")
			outPath := c.String("out")

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

			var flags []enums.MarkdownFlag
			if c.Bool("recursive") {
				flags = append(flags, enums.Recursive)
			}
			if c.Bool("include-private") {
				flags = append(flags, enums.IncludePrivate)
			}
			if c.Bool("include-undocumented") {
				flags = append(flags, enums.IncludeUndocumented)
			}
			if c.Bool("verbose") {
				flags = append(flags, enums.Verbose)
			}

			return godocmd.GenerateMarkdown(dir, out, flags...)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
