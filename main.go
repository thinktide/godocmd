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
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"
)

type StructFieldInfo struct {
	Name       string
	Type       string
	Comment    string
	JSONTag    string
	DynamoTag  string
	DynamoType string
}

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
			return run(dir, outPath)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(dir, outPath string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, absDir, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing directory: %w", err)
	}

	var out *os.File
	if outPath != "" {
		out, err = os.Create(outPath)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	for _, pkg := range pkgs {
		docPkg := doc.New(pkg, absDir, 0)
		generateMarkdown(docPkg, out)
	}
	return nil
}

func generateMarkdown(pkg *doc.Package, out *os.File) {
	fmt.Fprintf(out, "# Package %s\n\n", pkg.Name)

	for _, f := range pkg.Funcs {
		printFunc(f, out)
	}

	for _, t := range pkg.Types {
		fmt.Fprintln(out, "\n---")
		fmt.Fprintf(out, "## %s\n\n", t.Name)

		if t.Doc != "" {
			fmt.Fprintln(out, formatDocComment(t.Doc))
			fmt.Fprintln(out)
		}

		for _, spec := range t.Decl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					structText, fields := renderStructType(typeSpec, structType)
					fmt.Fprintf(out, "```go\n%s\n```\n\n", structText)

					if jsonOut := renderJSONBlock(fields); jsonOut != "" {
						fmt.Fprintln(out, jsonOut)
					}
					if dynamoOut := renderDynamoBlock(fields); dynamoOut != "" {
						fmt.Fprintln(out, dynamoOut)
					}
				}
			}
		}

		for _, m := range t.Methods {
			printFunc(m, out)
		}
	}
}

func printFunc(f *doc.Func, out *os.File) {
	decl := formatFuncDecl(f.Decl)
	fmt.Fprintln(out, "\n---")
	fmt.Fprintf(out, "## %s\n\n", f.Name)
	fmt.Fprintf(out, "```go\n%s\n```\n\n", decl)
	if f.Doc != "" {
		fmt.Fprintln(out, formatDocComment(f.Doc))
	}
}

func renderStructType(spec *ast.TypeSpec, structType *ast.StructType) (string, []StructFieldInfo) {
	var b strings.Builder
	var fields []StructFieldInfo

	maxFieldLen := 0
	maxTypeLen := 0
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		name := field.Names[0].Name
		typ := exprToString(field.Type)
		if len(name) > maxFieldLen {
			maxFieldLen = len(name)
		}
		if len(typ) > maxTypeLen {
			maxTypeLen = len(typ)
		}
	}

	b.WriteString(fmt.Sprintf("type %s struct {\n", spec.Name.Name))
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		name := field.Names[0].Name
		typ := exprToString(field.Type)
		comment := ""
		if field.Comment != nil && len(field.Comment.List) > 0 {
			comment = strings.TrimPrefix(field.Comment.List[0].Text, "//")
			comment = strings.TrimSpace(comment)
		}

		jsonTag, dynamoTag := "", ""
		if field.Tag != nil {
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			jsonTag = tag.Get("json")
			dynamoTag = tag.Get("dynamodbav")
		}

		line := fmt.Sprintf("    %-*s %-*s", maxFieldLen, name, maxTypeLen, typ)
		if comment != "" {
			line += " // " + comment
		}
		b.WriteString(line + "\n")

		fields = append(fields, StructFieldInfo{
			Name:       name,
			Type:       typ,
			Comment:    comment,
			JSONTag:    jsonTag,
			DynamoTag:  dynamoTag,
			DynamoType: mapGoTypeToDynamoType(typ),
		})
	}
	b.WriteString("}")
	return b.String(), fields
}

func renderJSONBlock(fields []StructFieldInfo) string {
	var b strings.Builder
	var tags []string
	for _, f := range fields {
		if f.JSONTag != "" {
			tags = append(tags, fmt.Sprintf(`  "%s",`, f.JSONTag))
		}
	}
	if len(tags) == 0 {
		return ""
	}
	b.WriteString("#### JSON\n\n```json\n{\n")
	for _, tag := range tags {
		b.WriteString(tag + "\n")
	}
	b.WriteString("}\n```\n")
	return b.String()
}

func renderDynamoBlock(fields []StructFieldInfo) string {
	var b strings.Builder
	var tags []string
	for _, f := range fields {
		if f.DynamoTag != "" {
			tags = append(tags, fmt.Sprintf("%-25s %s", f.DynamoTag, f.DynamoType))
		}
	}
	if len(tags) == 0 {
		return ""
	}
	b.WriteString("#### DynamoDB\n\n```sql\n")
	for _, line := range tags {
		b.WriteString(line + "\n")
	}
	b.WriteString("```\n")
	return b.String()
}

func mapGoTypeToDynamoType(goType string) string {
	switch strings.TrimPrefix(goType, "*") {
	case "string":
		return "String"
	case "bool":
		return "Boolean"
	case "int", "int64", "int32", "int16", "int8":
		return "Number"
	case "float32", "float64":
		return "Number"
	default:
		return "String"
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
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	case *ast.InterfaceType:
		return "interface{}"
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
