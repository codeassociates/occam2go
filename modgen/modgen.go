// Package modgen generates .module files from KRoC SConscript build files.
// It uses regex-based pattern matching to extract Split('''...''') variable
// assignments and OccamLibrary() calls, then produces an occam module file
// with include guards.
//
// Note: This does not execute the Python code in SConscript files, so it only
// works with simple, declarative SConscript files. Files that use Python
// control flow (loops, conditionals, etc.) are not supported.
package modgen

import (
	"fmt"
	"regexp"
	"strings"
)

// Library represents an OccamLibrary extracted from SConscript.
type Library struct {
	Name     string   // e.g. "course.lib"
	Sources  []string // source files
	Includes []string // --include files from OCCBUILDFLAGS
	Needs    []string // --need dependencies from OCCBUILDFLAGS
}

// ParseSConscript parses a SConscript file's content and extracts library definitions.
func ParseSConscript(content string) []Library {
	vars := extractSplitVars(content)
	return extractLibraries(content, vars)
}

// GenerateModule creates a .module file from a Library.
// moduleName is the guard symbol (e.g. "COURSE.MODULE").
func GenerateModule(lib Library, moduleName string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "#IF NOT (DEFINED (%s))\n", moduleName)
	fmt.Fprintf(&b, "#DEFINE %s\n", moduleName)

	// Include files from OCCBUILDFLAGS first
	for _, inc := range lib.Includes {
		fmt.Fprintf(&b, "#INCLUDE \"%s\"\n", inc)
	}

	// Then source files
	for _, src := range lib.Sources {
		fmt.Fprintf(&b, "#INCLUDE \"%s\"\n", src)
	}

	b.WriteString("#ENDIF\n")
	return b.String()
}

// splitVarRe matches: varname = Split('''...''') or Split("""...""")
var splitVarRe = regexp.MustCompile(`(\w+)\s*=\s*Split\(\s*(?:'''|""")([^'"]*)(?:'''|""")\s*\)`)

// extractSplitVars finds all variable = Split('''...''') assignments.
func extractSplitVars(content string) map[string][]string {
	vars := map[string][]string{}
	for _, m := range splitVarRe.FindAllStringSubmatch(content, -1) {
		name := m[1]
		files := splitWhitespace(m[2])
		vars[name] = files
	}
	return vars
}

// libCallRe matches OccamLibrary calls (both direct and via return).
// Captures: library name, source variable, and optional OCCBUILDFLAGS.
var libCallRe = regexp.MustCompile(
	`(?:local\.)?OccamLibrary\(\s*` +
		`['"]([^'"]+)['"]\s*,\s*` + // library name
		`(\w+)\s*` + // source variable
		`(?:,[^)]*?)?\)`, // optional extra args
)

// flagsRe extracts OCCBUILDFLAGS value.
var flagsRe = regexp.MustCompile(`OCCBUILDFLAGS\s*=\s*['"]([^'"]+)['"]`)

func extractLibraries(content string, vars map[string][]string) []Library {
	var libs []Library

	for _, m := range libCallRe.FindAllStringSubmatch(content, -1) {
		lib := Library{
			Name: m[1],
		}

		srcVar := m[2]
		if files, ok := vars[srcVar]; ok {
			lib.Sources = files
		}

		// Look for OCCBUILDFLAGS in the full match
		fullMatch := m[0]
		if fm := flagsRe.FindStringSubmatch(fullMatch); fm != nil {
			parseFlags(fm[1], &lib)
		}

		libs = append(libs, lib)
	}

	return libs
}

// parseFlags extracts --include and --need values from OCCBUILDFLAGS.
func parseFlags(flags string, lib *Library) {
	parts := strings.Fields(flags)
	for i := 0; i < len(parts); i++ {
		switch parts[i] {
		case "--include":
			if i+1 < len(parts) {
				lib.Includes = append(lib.Includes, parts[i+1])
				i++
			}
		case "--need":
			if i+1 < len(parts) {
				lib.Needs = append(lib.Needs, parts[i+1])
				i++
			}
		}
	}
}

func splitWhitespace(s string) []string {
	var result []string
	for _, f := range strings.Fields(s) {
		if f != "" {
			result = append(result, f)
		}
	}
	return result
}
