package enums

// MarkdownFlag defines available flags that alter the behavior of the godocmd.GenerateMarkdown function.
// These can be passed as variadic arguments to enable recursive scanning, inclusion of private or undocumented
// symbols, and verbose output during generation.
type MarkdownFlag int

const (
	// Recursive enables recursive scanning of all subdirectories starting from the provided root.
	Recursive MarkdownFlag = iota

	// IncludePrivate includes non-exported (unexported) functions and types in the markdown output.
	IncludePrivate

	// IncludeUndocumented includes functions, methods, and types that lack GoDoc comments.
	IncludeUndocumented

	// Verbose enables detailed logging of package parsing, filtering, and markdown rendering steps.
	Verbose
)
