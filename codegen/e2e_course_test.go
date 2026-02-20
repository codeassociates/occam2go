package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/parser"
	"github.com/codeassociates/occam2go/preproc"
)

// transpileCompileRunWithDefines is like transpileCompileRunFromFile but
// accepts preprocessor defines (e.g., TARGET.BITS.PER.WORD=32).
func transpileCompileRunWithDefines(t *testing.T, mainFile string, includePaths []string, defines map[string]string) string {
	t.Helper()

	pp := preproc.New(preproc.WithIncludePaths(includePaths), preproc.WithDefines(defines))
	expanded, err := pp.ProcessFile(mainFile)
	if err != nil {
		t.Fatalf("preprocessor error: %v", err)
	}
	if len(pp.Errors()) > 0 {
		for _, e := range pp.Errors() {
			t.Errorf("preprocessor warning: %s", e)
		}
	}

	// Transpile
	l := lexer.New(expanded)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	gen := New()
	goCode := gen.Generate(program)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "occam2go-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write Go source
	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		t.Fatalf("failed to write Go file: %v", err)
	}

	// Compile
	binFile := filepath.Join(tmpDir, "main")
	compileCmd := exec.Command("go", "build", "-o", binFile, goFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compilation failed: %v\nOutput: %s\nGo code:\n%s", err, compileOutput, goCode)
	}

	// Run
	runCmd := exec.Command(binFile)
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	return string(output)
}

func TestE2E_HelloWorldCourseModule(t *testing.T) {
	// Find the kroc directory relative to this test file
	krocDir := filepath.Join("..", "kroc", "modules", "course")
	mainFile := filepath.Join(krocDir, "examples", "hello_world.occ")
	includeDir := filepath.Join(krocDir, "libsrc")

	// Check that the files exist
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		t.Skip("kroc course module not found, skipping")
	}

	defines := map[string]string{
		"TARGET.BITS.PER.WORD": "32",
	}

	output := transpileCompileRunWithDefines(t, mainFile, []string{includeDir}, defines)
	expected := "Hello World\r\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
