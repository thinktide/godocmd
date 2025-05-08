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
