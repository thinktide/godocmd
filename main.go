package main

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "godoc-md",
		Usage: "Generate markdown from GoDoc comments",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dir",
				Aliases:  []string{"d"},
				Usage:    "Directory to start scanning for Go packages",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			dir := c.String("dir")
			return run(dir)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, absDir, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing directory: %w", err)
	}

	for _, pkg := range pkgs {
		docPkg := doc.New(pkg, absDir, 0)
		generateMarkdown(docPkg)
	}
	return nil
}

func generateMarkdown(pkg *doc.Package) {
	fmt.Printf("# Package %s\n\n", pkg.Name)

	for _, f := range pkg.Funcs {
		printFunc(f)
	}
	for _, t := range pkg.Types {
		fmt.Printf("\n## type %s\n\n", t.Name)
		if t.Doc != "" {
			fmt.Println(t.Doc)
			fmt.Println()
		}

		for _, m := range t.Methods {
			printFunc(m)
		}
		for _, f := range t.Funcs {
			printFunc(f)
		}
	}
}

func printFunc(f *doc.Func) {
	decl := formatFuncDecl(f.Decl)

	fmt.Printf("\n## func %s\n\n", f.Name)
	fmt.Printf("```go\n%s\n```\n\n", decl)

	if f.Doc != "" {
		fmt.Println(formatDocComment(f.Doc))
	}
}

func formatFuncDecl(decl *ast.FuncDecl) string {
	var buf strings.Builder

	buf.WriteString("func ")
	if decl.Recv != nil {
		buf.WriteString("(")
		for _, f := range decl.Recv.List {
			buf.WriteString(fieldListToString(f))
		}
		buf.WriteString(") ")
	}
	buf.WriteString(decl.Name.Name)
	buf.WriteString("(")
	if decl.Type.Params != nil {
		for i, p := range decl.Type.Params.List {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fieldListToString(p))
		}
	}
	buf.WriteString(")")
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		buf.WriteString(" ")
		if len(decl.Type.Results.List) > 1 {
			buf.WriteString("(")
		}
		for i, r := range decl.Type.Results.List {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fieldListToString(r))
		}
		if len(decl.Type.Results.List) > 1 {
			buf.WriteString(")")
		}
	}
	return buf.String()
}

func fieldListToString(f *ast.Field) string {
	var buf strings.Builder
	for i, n := range f.Names {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(n.Name)
		buf.WriteString(" ")
	}
	buf.WriteString(exprToString(f.Type))
	return buf.String()
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func formatDocComment(doc string) string {
	lines := strings.Split(doc, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(strings.TrimPrefix(line, "//"))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
