package parse

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPackage_SimplePackage(t *testing.T) {
	tempDir := t.TempDir()

	const code = `
package simple

// Hello says hi.
func Hello() string {
	return "hi"
}
`
	filePath := filepath.Join(tempDir, "hello.go")
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	docPkg, err := LoadPackage(tempDir)
	if err != nil {
		t.Fatalf("LoadPackage failed: %v", err)
	}

	if docPkg.Name != "simple" {
		t.Errorf("expected package name 'simple', got %q", docPkg.Name)
	}

	if len(docPkg.Funcs) == 0 || docPkg.Funcs[0].Name != "Hello" {
		t.Errorf("expected to find function Hello in doc package")
	}
}
