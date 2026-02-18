// Package preproc implements a textual preprocessor for occam source files.
// It handles #IF/#ELSE/#ENDIF conditional compilation, #DEFINE symbols,
// #INCLUDE file inclusion, and ignores #COMMENT/#PRAGMA/#USE directives.
// The output is a single expanded string suitable for feeding into the lexer.
package preproc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Option configures a Preprocessor.
type Option func(*Preprocessor)

// WithIncludePaths sets the search paths for #INCLUDE resolution.
func WithIncludePaths(paths []string) Option {
	return func(pp *Preprocessor) {
		pp.includePaths = paths
	}
}

// WithDefines sets predefined symbols.
func WithDefines(defs map[string]string) Option {
	return func(pp *Preprocessor) {
		for k, v := range defs {
			pp.defines[k] = v
		}
	}
}

// Preprocessor performs textual preprocessing of occam source.
type Preprocessor struct {
	defines      map[string]string
	includePaths []string
	errors       []string
	processing   map[string]bool // absolute paths currently being processed (circular include detection)
}

// New creates a new Preprocessor with the given options.
func New(opts ...Option) *Preprocessor {
	pp := &Preprocessor{
		defines:    map[string]string{},
		processing: map[string]bool{},
	}
	// Predefined symbols
	pp.defines["TARGET.BITS.PER.WORD"] = "64"

	for _, opt := range opts {
		opt(pp)
	}
	return pp
}

// Errors returns any errors accumulated during processing.
func (pp *Preprocessor) Errors() []string {
	return pp.errors
}

// ProcessFile reads and processes a file, resolving #INCLUDE directives.
func (pp *Preprocessor) ProcessFile(filename string) (string, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path %q: %w", filename, err)
	}

	if pp.processing[absPath] {
		return "", fmt.Errorf("circular include detected: %s", filename)
	}
	pp.processing[absPath] = true
	defer delete(pp.processing, absPath)

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot read %q: %w", filename, err)
	}

	return pp.processSource(string(data), filepath.Dir(absPath))
}

// ProcessSource processes occam source text with no file context.
// #INCLUDE directives will only resolve against includePaths.
func (pp *Preprocessor) ProcessSource(source string) (string, error) {
	return pp.processSource(source, "")
}

// processSource performs line-by-line preprocessing.
// baseDir is the directory of the current file (for relative #INCLUDE resolution).
func (pp *Preprocessor) processSource(source string, baseDir string) (string, error) {
	lines := strings.Split(source, "\n")
	var out strings.Builder
	var condStack []condState

	for i, line := range lines {
		if i > 0 {
			out.WriteByte('\n')
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") {
			directive, rest := parseDirectiveLine(trimmed)

			switch directive {
			case "DEFINE":
				if isActive(condStack) {
					sym := strings.TrimSpace(rest)
					if sym != "" {
						pp.defines[sym] = ""
					}
				}
				out.WriteString("") // blank line preserves line numbers

			case "IF":
				val := pp.evalExpr(rest)
				condStack = append(condStack, condState{active: val, seenTrue: val})
				out.WriteString("")

			case "ELSE":
				if len(condStack) == 0 {
					pp.errors = append(pp.errors, fmt.Sprintf("line %d: #ELSE without matching #IF", i+1))
				} else {
					top := &condStack[len(condStack)-1]
					if top.seenTrue {
						top.active = false
					} else {
						top.active = true
						top.seenTrue = true
					}
				}
				out.WriteString("")

			case "ENDIF":
				if len(condStack) == 0 {
					pp.errors = append(pp.errors, fmt.Sprintf("line %d: #ENDIF without matching #IF", i+1))
				} else {
					condStack = condStack[:len(condStack)-1]
				}
				out.WriteString("")

			case "INCLUDE":
				if isActive(condStack) {
					included, err := pp.resolveAndInclude(rest, baseDir)
					if err != nil {
						return "", fmt.Errorf("line %d: %w", i+1, err)
					}
					out.WriteString(included)
				} else {
					out.WriteString("")
				}

			case "COMMENT", "PRAGMA", "USE":
				out.WriteString("") // no-op, blank line

			default:
				// Unknown directive — pass through if active
				if isActive(condStack) {
					out.WriteString(line)
				} else {
					out.WriteString("")
				}
			}
		} else {
			if isActive(condStack) {
				out.WriteString(line)
			} else {
				out.WriteString("") // blank line preserves line numbers
			}
		}
	}

	if len(condStack) > 0 {
		pp.errors = append(pp.errors, fmt.Sprintf("unterminated #IF (missing %d #ENDIF)", len(condStack)))
	}

	return out.String(), nil
}

// condState tracks one level of #IF/#ELSE nesting.
type condState struct {
	active   bool // currently emitting lines?
	seenTrue bool // has any branch been true?
}

// isActive returns true if all condition stack levels are active.
func isActive(stack []condState) bool {
	for _, s := range stack {
		if !s.active {
			return false
		}
	}
	return true
}

// parseDirectiveLine splits "#DIRECTIVE rest" into (directive, rest).
func parseDirectiveLine(trimmed string) (string, string) {
	// trimmed starts with "#"
	s := trimmed[1:] // skip '#'
	s = strings.TrimSpace(s)

	idx := strings.IndexAny(s, " \t")
	if idx == -1 {
		return strings.ToUpper(s), ""
	}
	return strings.ToUpper(s[:idx]), strings.TrimSpace(s[idx+1:])
}

// resolveAndInclude resolves an #INCLUDE filename and processes the included file.
func (pp *Preprocessor) resolveAndInclude(rest string, baseDir string) (string, error) {
	filename := stripQuotes(rest)
	if filename == "" {
		return "", fmt.Errorf("#INCLUDE with empty filename")
	}

	// Try to find the file
	resolved := pp.resolveIncludePath(filename, baseDir)
	if resolved == "" {
		return "", fmt.Errorf("cannot find included file %q", filename)
	}

	return pp.ProcessFile(resolved)
}

// resolveIncludePath searches for a file: first relative to baseDir, then in includePaths.
func (pp *Preprocessor) resolveIncludePath(filename string, baseDir string) string {
	// First: relative to current file's directory
	if baseDir != "" {
		candidate := filepath.Join(baseDir, filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Then: each include path
	for _, dir := range pp.includePaths {
		candidate := filepath.Join(dir, filename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// stripQuotes removes surrounding double quotes from a string.
func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// evalExpr evaluates a preprocessor conditional expression.
// Supports: TRUE, FALSE, DEFINED (SYMBOL), NOT (expr), (SYMBOL = value)
func (pp *Preprocessor) evalExpr(expr string) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false
	}

	// TRUE / FALSE
	if expr == "TRUE" {
		return true
	}
	if expr == "FALSE" {
		return false
	}

	// NOT (expr) or NOT DEFINED (...)
	if strings.HasPrefix(expr, "NOT ") || strings.HasPrefix(expr, "NOT(") {
		inner := strings.TrimPrefix(expr, "NOT")
		inner = strings.TrimSpace(inner)
		return !pp.evalExpr(inner)
	}

	// DEFINED (SYMBOL)
	if strings.HasPrefix(expr, "DEFINED") {
		inner := strings.TrimPrefix(expr, "DEFINED")
		inner = strings.TrimSpace(inner)
		sym := stripParens(inner)
		_, ok := pp.defines[sym]
		return ok
	}

	// Parenthesized expression
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		inner := expr[1 : len(expr)-1]
		inner = strings.TrimSpace(inner)

		// Check for equality: SYMBOL = value
		if eqIdx := strings.Index(inner, "="); eqIdx >= 0 {
			lhs := strings.TrimSpace(inner[:eqIdx])
			rhs := strings.TrimSpace(inner[eqIdx+1:])
			lhsVal, ok := pp.defines[lhs]
			if !ok {
				return false
			}
			return lhsVal == rhs
		}

		// Otherwise recurse
		return pp.evalExpr(inner)
	}

	// Bare symbol — treat as DEFINED
	_, ok := pp.defines[expr]
	return ok
}

// stripParens removes surrounding parentheses and whitespace.
func stripParens(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '(' && s[len(s)-1] == ')' {
		return strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}
