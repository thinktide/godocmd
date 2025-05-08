package format

import (
	"bytes"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestWriteMarkdown_StructWithTags(t *testing.T) {
	const input = `
package testpkg

// User represents a system user.
type User struct {
	Name string  ` + "`json:\"name\" dynamodbav:\"username\"`" + `
	Age  int     ` + "`json:\"age,omitempty\" dynamodbav:\"user_age\"`" + `
}
`

	docPkg := parseGoDocPackage("testpkg", input)
	var buf bytes.Buffer

	err := WriteMarkdown(docPkg, &buf)
	if err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}

	out := buf.String()

	assertContains(t, out, "## User", "missing struct name header")
	assertContains(t, out, `"name",`, "missing JSON tag 'name'")
	assertNotContains(t, out, "omitempty", "should have stripped omitempty from JSON tag")
	assertContains(t, out, "username", "missing DynamoDB tag 'username'")
	assertContains(t, out, "user_age", "missing DynamoDB tag 'user_age'")
}

func parseGoDocPackage(name string, code string) *doc.Package {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name+".go", code, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	pkg := map[string]*ast.File{name + ".go": file}
	return doc.New(&ast.Package{
		Name:  name,
		Files: pkg,
	}, "./", 0)
}

func assertContains(t *testing.T, haystack, needle, message string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q: %s", needle, message)
	}
}

func assertNotContains(t *testing.T, haystack, needle, message string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Errorf("expected output NOT to contain %q: %s", needle, message)
	}
}

func TestWriteMarkdown_WithUndocumentedFiltering(t *testing.T) {
	const input = `
package testpkg

// Visible function.
func Exported() {}

func internal() {}

// Visible is a test struct that should be included.
type Visible struct{}
type invisible struct{}
`

	docPkg := parseGoDocPackage("testpkg", input)

	var buf bytes.Buffer
	err := WriteMarkdownWithOptions(docPkg, &buf, false, false)
	if err != nil {
		t.Fatalf("WriteMarkdownWithOptions failed: %v", err)
	}

	out := buf.String()

	assertContains(t, out, "## Exported", "expected exported function to be included")
	assertNotContains(t, out, "internal", "expected unexported function to be excluded")
	assertContains(t, out, "## Visible", "expected exported type to be included")
	assertNotContains(t, out, "invisible", "expected unexported type to be excluded")
}

func TestWriteMarkdown_IncludePrivateAndUndocumented(t *testing.T) {
	const input = `
package testpkg

func Alpha() {}

type Beta struct{}
`

	docPkg := parseGoDocPackage("testpkg", input)

	var buf bytes.Buffer
	err := WriteMarkdownWithOptions(docPkg, &buf, true, true)
	if err != nil {
		t.Fatalf("WriteMarkdownWithOptions failed: %v", err)
	}

	out := buf.String()
	assertContains(t, out, "## Alpha", "expected private function to be included")
	assertContains(t, out, "## Beta", "expected private type to be included")
}
