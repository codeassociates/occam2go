package preproc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefineAndIfDefined(t *testing.T) {
	pp := New()
	src := `#DEFINE FOO
#IF DEFINED (FOO)
hello
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out, "\n")
	if strings.TrimSpace(lines[2]) != "hello" {
		t.Errorf("expected 'hello' on line 3, got %q", lines[2])
	}
}

func TestIfFalseExcludes(t *testing.T) {
	pp := New()
	src := `#IF FALSE
visible
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "visible") {
		t.Error("expected #IF FALSE to exclude content")
	}
}

func TestIfTrue(t *testing.T) {
	pp := New()
	src := `#IF TRUE
visible
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "visible") {
		t.Error("expected #IF TRUE to include content")
	}
}

func TestElse(t *testing.T) {
	pp := New()
	src := `#IF FALSE
wrong
#ELSE
right
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "wrong") {
		t.Error("should not contain 'wrong'")
	}
	if !strings.Contains(out, "right") {
		t.Error("should contain 'right'")
	}
}

func TestElseNotTakenWhenIfTrue(t *testing.T) {
	pp := New()
	src := `#IF TRUE
right
#ELSE
wrong
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "right") {
		t.Error("should contain 'right'")
	}
	if strings.Contains(out, "wrong") {
		t.Error("should not contain 'wrong'")
	}
}

func TestNestedIf(t *testing.T) {
	pp := New()
	src := `#DEFINE A
#IF DEFINED (A)
outer
#IF FALSE
inner-hidden
#ENDIF
outer2
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "outer") {
		t.Error("should contain 'outer'")
	}
	if strings.Contains(out, "inner-hidden") {
		t.Error("should not contain 'inner-hidden'")
	}
	if !strings.Contains(out, "outer2") {
		t.Error("should contain 'outer2'")
	}
}

func TestNotDefined(t *testing.T) {
	pp := New()
	src := `#IF NOT DEFINED (MISSING)
visible
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "visible") {
		t.Error("NOT DEFINED of missing symbol should be true")
	}
}

func TestLineCountPreservation(t *testing.T) {
	pp := New()
	src := `line1
#IF FALSE
excluded
#ENDIF
line5
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out, "\n")
	// Original has 6 lines (including trailing empty from final \n)
	srcLines := strings.Split(src, "\n")
	if len(lines) != len(srcLines) {
		t.Errorf("line count mismatch: got %d, want %d", len(lines), len(srcLines))
	}
	if lines[0] != "line1" {
		t.Errorf("line 1: got %q, want %q", lines[0], "line1")
	}
	if lines[4] != "line5" {
		t.Errorf("line 5: got %q, want %q", lines[4], "line5")
	}
}

func TestCommentPragmaUseIgnored(t *testing.T) {
	pp := New()
	src := `#COMMENT "this is a comment"
#PRAGMA SHARED
#USE "somelib"
hello
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "hello") {
		t.Error("should contain 'hello'")
	}
	if strings.Contains(out, "COMMENT") || strings.Contains(out, "PRAGMA") || strings.Contains(out, "USE") {
		t.Error("directives should be replaced with blank lines")
	}
}

func TestEqualityExpression(t *testing.T) {
	pp := New()
	// TARGET.BITS.PER.WORD is predefined as "64"
	src := `#IF (TARGET.BITS.PER.WORD = 64)
is64
#ENDIF
#IF (TARGET.BITS.PER.WORD = 32)
is32
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "is64") {
		t.Error("should match 64-bit")
	}
	if strings.Contains(out, "is32") {
		t.Error("should not match 32-bit")
	}
}

func TestIncludeGuardPattern(t *testing.T) {
	pp := New()
	src := `#IF NOT (DEFINED (MY.MODULE))
#DEFINE MY.MODULE
content
#ENDIF
#IF NOT (DEFINED (MY.MODULE))
#DEFINE MY.MODULE
duplicate
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "content") {
		t.Error("first include should have content")
	}
	if strings.Contains(out, "duplicate") {
		t.Error("second include should be guarded")
	}
}

func TestWithDefinesOption(t *testing.T) {
	pp := New(WithDefines(map[string]string{"MY.FLAG": ""}))
	src := `#IF DEFINED (MY.FLAG)
flagged
#ENDIF
`
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "flagged") {
		t.Error("pre-defined flag should be recognized")
	}
}

func TestPredefinedTargetBits(t *testing.T) {
	pp := New()
	if _, ok := pp.defines["TARGET.BITS.PER.WORD"]; !ok {
		t.Error("TARGET.BITS.PER.WORD should be predefined")
	}
}

// --- File-based tests for #INCLUDE ---

func TestIncludeFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create included file
	incContent := "INT x:\nx := 42\n"
	os.WriteFile(filepath.Join(tmpDir, "lib.inc"), []byte(incContent), 0644)

	// Create main file
	mainContent := `#INCLUDE "lib.inc"
print.int(x)
`
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	pp := New()
	out, err := pp.ProcessFile(mainFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "INT x:") {
		t.Error("should include content from lib.inc")
	}
	if !strings.Contains(out, "print.int(x)") {
		t.Error("should contain main file content")
	}
}

func TestIncludeWithSearchPath(t *testing.T) {
	tmpDir := t.TempDir()
	libDir := filepath.Join(tmpDir, "libs")
	os.Mkdir(libDir, 0755)

	// Create included file in lib directory
	os.WriteFile(filepath.Join(libDir, "helper.inc"), []byte("INT helper:\n"), 0644)

	// Create main file that includes from a different directory
	mainContent := `#INCLUDE "helper.inc"
done
`
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	pp := New(WithIncludePaths([]string{libDir}))
	out, err := pp.ProcessFile(mainFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "INT helper:") {
		t.Error("should find file via include path")
	}
}

func TestIncludeGuardWithFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create guarded module file
	modContent := `#IF NOT (DEFINED (MY.MODULE))
#DEFINE MY.MODULE
INT shared:
#ENDIF
`
	os.WriteFile(filepath.Join(tmpDir, "my.module"), []byte(modContent), 0644)

	// Create main file that includes twice
	mainContent := `#INCLUDE "my.module"
#INCLUDE "my.module"
done
`
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	pp := New()
	out, err := pp.ProcessFile(mainFile)
	if err != nil {
		t.Fatal(err)
	}
	// "INT shared:" should appear only once
	count := strings.Count(out, "INT shared:")
	if count != 1 {
		t.Errorf("expected 'INT shared:' once, found %d times", count)
	}
}

func TestNestedIncludes(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "inner.inc"), []byte("inner-content\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "outer.inc"), []byte("#INCLUDE \"inner.inc\"\nouter-content\n"), 0644)

	mainContent := `#INCLUDE "outer.inc"
main-content
`
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	pp := New()
	out, err := pp.ProcessFile(mainFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "inner-content") {
		t.Error("should contain nested include content")
	}
	if !strings.Contains(out, "outer-content") {
		t.Error("should contain outer include content")
	}
	if !strings.Contains(out, "main-content") {
		t.Error("should contain main content")
	}
}

func TestCircularIncludeError(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "a.inc"), []byte("#INCLUDE \"b.inc\"\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.inc"), []byte("#INCLUDE \"a.inc\"\n"), 0644)

	mainFile := filepath.Join(tmpDir, "a.inc")
	pp := New()
	_, err := pp.ProcessFile(mainFile)
	if err == nil {
		t.Error("expected circular include error")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected 'circular' in error, got: %s", err)
	}
}

func TestIncludeFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	mainContent := `#INCLUDE "nonexistent.inc"
`
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	pp := New()
	_, err := pp.ProcessFile(mainFile)
	if err == nil {
		t.Error("expected file not found error")
	}
	if !strings.Contains(err.Error(), "cannot find") {
		t.Errorf("expected 'cannot find' in error, got: %s", err)
	}
}

func TestUnterminatedIf(t *testing.T) {
	pp := New()
	src := `#IF TRUE
hello
`
	_, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(pp.Errors()) == 0 {
		t.Error("expected unterminated #IF error")
	}
}

func TestElseWithoutIf(t *testing.T) {
	pp := New()
	src := `#ELSE
hello
`
	_, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(pp.Errors()) == 0 {
		t.Error("expected #ELSE without #IF error")
	}
}

func TestEndifWithoutIf(t *testing.T) {
	pp := New()
	src := `#ENDIF
`
	_, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(pp.Errors()) == 0 {
		t.Error("expected #ENDIF without #IF error")
	}
}

// --- Source map tests ---

func TestSourceMapSimple(t *testing.T) {
	pp := New()
	src := "line1\nline2\nline3"
	out, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(out, "\n")
	sm := pp.SourceMap()
	if len(sm) != len(lines) {
		t.Fatalf("source map length %d != output lines %d", len(sm), len(lines))
	}
	for i, loc := range sm {
		if loc.File != "<input>" {
			t.Errorf("entry %d: file = %q, want %q", i, loc.File, "<input>")
		}
		if loc.Line != i+1 {
			t.Errorf("entry %d: line = %d, want %d", i, loc.Line, i+1)
		}
	}
}

func TestSourceMapWithDirectives(t *testing.T) {
	pp := New()
	src := "#DEFINE FOO\n#IF TRUE\nhello\n#ENDIF\nworld"
	_, err := pp.ProcessSource(src)
	if err != nil {
		t.Fatal(err)
	}
	sm := pp.SourceMap()
	// 5 source lines → 5 source map entries
	if len(sm) != 5 {
		t.Fatalf("source map length = %d, want 5", len(sm))
	}
	// Each entry maps to its original line in <input>
	for i, loc := range sm {
		if loc.File != "<input>" {
			t.Errorf("entry %d: file = %q, want %q", i, loc.File, "<input>")
		}
		if loc.Line != i+1 {
			t.Errorf("entry %d: line = %d, want %d", i, loc.Line, i+1)
		}
	}
}

func TestSourceMapWithInclude(t *testing.T) {
	tmpDir := t.TempDir()

	// Create included file (2 lines + trailing newline = 3 entries after split)
	os.WriteFile(filepath.Join(tmpDir, "inc.occ"), []byte("incA\nincB\n"), 0644)

	// Create main file: line1, #INCLUDE, line3
	mainContent := "line1\n#INCLUDE \"inc.occ\"\nline3\n"
	mainFile := filepath.Join(tmpDir, "main.occ")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	pp := New()
	out, err := pp.ProcessFile(mainFile)
	if err != nil {
		t.Fatal(err)
	}

	sm := pp.SourceMap()
	outLines := strings.Split(out, "\n")
	if len(sm) != len(outLines) {
		t.Fatalf("source map length %d != output lines %d", len(sm), len(outLines))
	}

	// Entry 0: main.occ line 1
	if sm[0].Line != 1 || !strings.HasSuffix(sm[0].File, "main.occ") {
		t.Errorf("entry 0: got {%s, %d}, want {main.occ, 1}", sm[0].File, sm[0].Line)
	}

	// Entries 1-3: inc.occ lines 1-3
	incFile := filepath.Join(tmpDir, "inc.occ")
	for i := 1; i <= 3; i++ {
		if sm[i].File != incFile {
			t.Errorf("entry %d: file = %q, want %q", i, sm[i].File, incFile)
		}
		if sm[i].Line != i {
			t.Errorf("entry %d: line = %d, want %d", i, sm[i].Line, i)
		}
	}

	// Entry 4: main.occ line 3
	if sm[4].Line != 3 || !strings.HasSuffix(sm[4].File, "main.occ") {
		t.Errorf("entry 4: got {%s, %d}, want {main.occ, 3}", sm[4].File, sm[4].Line)
	}
}
