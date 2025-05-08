[![CI](https://github.com/thinktide/godocmd/actions/workflows/test.yml/badge.svg)](https://github.com/thinktide/godocmd/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/thinktide/godocmd/branch/main/graph/badge.svg)](https://codecov.io/gh/thinktide/godocmd)
[![Go Reference](https://pkg.go.dev/badge/github.com/thinktide/godocmd.svg)](https://pkg.go.dev/github.com/thinktide/godocmd)
![Go Version](https://img.shields.io/badge/go-1.21+-blue)
# godocmd

**godocmd** is a powerful Go documentation generator that outputs structured, customizable Markdown from your Go packages. It supports recursive package scanning, inclusion of private and undocumented symbols, and both CLI and programmatic usage.

---

## 🔧 Installation (CLI)

You can install the CLI tool using:

```bash
go install github.com/thinktide/godocmd/cmd@latest
```

> This will install a binary named `godocmd` in your `$GOBIN`.

---

## 🚀 CLI Usage

```bash
godocmd -d <directory> [flags]
```

### Common Flags:

| Flag                  | Alias | Description                                                         |
|-----------------------|-------|---------------------------------------------------------------------|
| `--dir`               | `-d`  | **Required.** The root directory to scan for Go packages.          |
| `--out`               | `-o`  | Output markdown file (defaults to stdout).                         |
| `--recursive`         | `-r`  | Recursively scan subdirectories.                                   |
| `--include-private`   | `-p`  | Include unexported (private) functions and types.                  |
| `--include-undocumented`       |       | Include symbols that lack GoDoc comments.                         |
| `--verbose`           |       | Output detailed logs for each step.                                |

### Example

```bash
# Generate docs recursively from ./models and write to docs.md
godocmd -d ./models -r -o docs.md --include-private --verbose
```

---

## 📦 Programmatic Usage

You can also use `godocmd` as a Go library.

### Install

```bash
go get github.com/thinktide/godocmd@latest
```

### Import and Use

```go
package main

import (
	"os"

	"github.com/thinktide/godocmd"
	"github.com/thinktide/godocmd/enums"
)

func main() {
	err := godocmd.GenerateMarkdown(
		"./myproject",   // root directory
		os.Stdout,       // output writer
		enums.Recursive, // enable recursion
		enums.Verbose,   // log verbosely
		enums.IncludePrivate, // include private symbols
	)
	if err != nil {
		panic(err)
	}
}
```

---

## 🧠 Features

- ✅ Recursive scanning with `enums.Recursive`
- ✅ Support for private types and functions via `enums.IncludePrivate`
- ✅ Strict documentation enforcement (exclude symbols with no GoDoc by default)
- ✅ Verbose logging support with `enums.Verbose`
- ✅ Outputs Markdown suitable for GitHub, wikis, or README sections

---

## 📁 Output Structure

- Each package starts with a comment header: `<!-- ./package/path -->`
- Structs include:
    - Go struct definition
    - JSON tags (if present)
    - DynamoDB tags (if present)
- Functions and methods show:
    - Full signature
    - GoDoc comments (if present)
    - Grouped under their receiver (for methods)

---

## 🧪 Contributing

We welcome issues and pull requests that improve Markdown output, formatting, or support for additional tag types (e.g., BSON, XML).

---

## 📄 License

This project is licensed under the [MIT License](./LICENSE).
