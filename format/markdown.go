package format

import (
	"fmt"
	"go/ast"
	"go/doc"
	"io"
	"reflect"
	"strings"
)

// StructFieldInfo represents metadata about a struct field, including its name, type,
// comments, and any struct tags like json or dynamodbav.
type StructFieldInfo struct {
	Name       string
	Type       string
	Comment    string
	JSONTag    string
	DynamoTag  string
	DynamoType string
}

// WriteMarkdownWithOptions writes documentation with filtering controls.
//
// Parameters:
//   - pkg: The parsed Go package to document
//   - out: The writer to output the markdown to
//   - includePrivate: Whether to include non-exported (unexported) symbols
//   - includeUndocumented: Whether to include symbols that lack GoDoc
//
// Returns:
//   - error: Any error encountered during rendering
func WriteMarkdownWithOptions(pkg *doc.Package, out io.Writer, includePrivate, includeUndocumented bool) error {
	fmt.Fprintf(out, "# Package %s\n\n", pkg.Name)

	if len(pkg.Funcs) == 0 && len(pkg.Types) == 0 {
		fmt.Fprintf(out, "_No exported symbols in package `%s`._\n", pkg.Name)
		return nil
	}

	for _, f := range pkg.Funcs {
		if !includePrivate && !isExported(f.Name) {
			continue
		}
		if !includeUndocumented && strings.TrimSpace(f.Doc) == "" {
			continue
		}
		printFunc(f, out)
	}

	for _, t := range pkg.Types {
		if !includePrivate && !isExported(t.Name) {
			continue
		}
		if !includeUndocumented && strings.TrimSpace(t.Doc) == "" {
			continue
		}

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
			if !includePrivate && !isExported(m.Name) {
				continue
			}
			if !includeUndocumented && strings.TrimSpace(m.Doc) == "" {
				continue
			}
			printFunc(m, out)
		}
	}

	return nil
}

// WriteMarkdown is a convenience alias that includes all symbols.
// Deprecated: use WriteMarkdownWithOptions with explicit flags.
func WriteMarkdown(pkg *doc.Package, out io.Writer) error {
	return WriteMarkdownWithOptions(pkg, out, true, true)
}

// printFunc writes a markdown section for a Go function or method.
//
// Parameters:
//   - f: The Go function or method to document
//   - out: The writer to output the markdown to
func printFunc(f *doc.Func, out io.Writer) {
	decl := formatFuncDecl(f.Decl)

	fmt.Fprintln(out, "\n---")
	if f.Recv != "" {
		recv := formatReceiverName(f.Decl)
		fmt.Fprintf(out, "## <small><em>%s.</em></small>%s\n\n", recv, f.Name)
	} else {
		fmt.Fprintf(out, "## %s\n\n", f.Name)
	}

	fmt.Fprintf(out, "```go\n%s\n```\n\n", decl)

	if f.Doc != "" {
		fmt.Fprintln(out, formatDocComment(f.Doc))
	}
}

// renderStructType formats a Go struct type as a Go code block and extracts field metadata.
//
// Parameters:
//   - spec: The AST TypeSpec for the struct
//   - structType: The AST StructType to render
//
// Returns:
//   - string: Formatted Go code block of the struct
//   - []StructFieldInfo: Metadata for each struct field
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
			raw := tag.Get("json")
			jsonTag = strings.Split(raw, ",")[0]
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

// renderJSONBlock returns a markdown code block for struct JSON tags.
//
// Parameters:
//   - fields: Slice of StructFieldInfo to extract json tags from
//
// Returns:
//   - string: A markdown-formatted code block with JSON field names
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
	b.WriteString("}\n```")
	return b.String()
}

// renderDynamoBlock returns a markdown code block for struct DynamoDB tags.
//
// Parameters:
//   - fields: Slice of StructFieldInfo to extract DynamoDB mappings from
//
// Returns:
//   - string: A markdown-formatted table of DynamoDB field mappings
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

// mapGoTypeToDynamoType converts a Go type to an approximate DynamoDB type.
//
// Parameters:
//   - goType: The Go type string to map
//
// Returns:
//   - string: The inferred DynamoDB-compatible type
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

// formatFuncDecl formats an AST FuncDecl into a Go-style function signature string.
//
// Parameters:
//   - decl: The AST function declaration
//
// Returns:
//   - string: A rendered Go-style function signature
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

// fieldListToString renders a list of field names and types.
//
// Parameters:
//   - f: The AST field node to render
//
// Returns:
//   - string: A comma-separated list of field declarations
func fieldListToString(f *ast.Field) string {
	var buf strings.Builder
	for i, n := range f.Names {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(n.Name + " ")
	}
	buf.WriteString(exprToString(f.Type))
	return buf.String()
}

// exprToString converts an AST expression into its Go code string form.
//
// Parameters:
//   - expr: The AST expression to convert
//
// Returns:
//   - string: A human-readable Go representation
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

// formatDocComment converts a GoDoc comment into markdown by trimming slashes and joining lines.
//
// Parameters:
//   - doc: The raw GoDoc comment string
//
// Returns:
//   - string: A cleaned-up markdown version of the comment
func formatDocComment(doc string) string {
	lines := strings.Split(doc, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(strings.TrimPrefix(line, "//"))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

// formatReceiverName extracts the receiver type name from a method declaration.
//
// Parameters:
//   - fn: The function declaration with a receiver
//
// Returns:
//   - string: The type name of the receiver
func formatReceiverName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	switch r := fn.Recv.List[0].Type.(type) {
	case *ast.StarExpr:
		if ident, ok := r.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return r.Name
	}
	return exprToString(fn.Recv.List[0].Type)
}

// isExported checks if a symbol name is exported (starts with an uppercase letter).
//
// Parameters:
//   - name: The identifier name to check
//
// Returns:
//   - bool: True if exported, false otherwise
func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := rune(name[0])
	return strings.ToUpper(string(r)) == string(r)
}
