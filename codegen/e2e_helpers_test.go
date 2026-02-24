package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/parser"
	"github.com/codeassociates/occam2go/preproc"
)

// transpileCompileRun takes Occam source, transpiles to Go, compiles, runs,
// and returns the stdout output
func transpileCompileRun(t *testing.T, occamSource string) string {
	t.Helper()

	// Transpile
	l := lexer.New(occamSource)
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

	// Create temp directory for this test
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

// transpileCompileRunFromFile takes an occam file path, preprocesses it,
// then transpiles, compiles, and runs.
func transpileCompileRunFromFile(t *testing.T, mainFile string, includePaths []string) string {
	t.Helper()

	pp := preproc.New(preproc.WithIncludePaths(includePaths))
	expanded, err := pp.ProcessFile(mainFile)
	if err != nil {
		t.Fatalf("preprocessor error: %v", err)
	}
	if len(pp.Errors()) > 0 {
		for _, e := range pp.Errors() {
			t.Errorf("preprocessor warning: %s", e)
		}
	}

	return transpileCompileRun(t, expanded)
}

// transpileCompileRunWithInput takes Occam source that uses the entry-point
// PROC pattern (CHAN OF BYTE keyboard?, screen!, error!), transpiles to Go,
// initialises a Go module (needed for golang.org/x/term), compiles, pipes
// the given input to stdin, and returns the stdout output.
func transpileCompileRunWithInput(t *testing.T, occamSource, stdin string) string {
	t.Helper()

	// Transpile
	l := lexer.New(occamSource)
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

	// Initialise Go module (needed for golang.org/x/term dependency)
	modInit := exec.Command("go", "mod", "init", "test")
	modInit.Dir = tmpDir
	if out, err := modInit.CombinedOutput(); err != nil {
		t.Fatalf("go mod init failed: %v\n%s", err, out)
	}
	modTidy := exec.Command("go", "mod", "tidy")
	modTidy.Dir = tmpDir
	if out, err := modTidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s\nGo code:\n%s", err, out, goCode)
	}

	// Compile
	binFile := filepath.Join(tmpDir, "main")
	compileCmd := exec.Command("go", "build", "-o", binFile, ".")
	compileCmd.Dir = tmpDir
	if out, err := compileCmd.CombinedOutput(); err != nil {
		t.Fatalf("compilation failed: %v\nOutput: %s\nGo code:\n%s", err, out, goCode)
	}

	// Run with piped stdin
	runCmd := exec.Command(binFile)
	runCmd.Stdin = strings.NewReader(stdin)
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	return string(output)
}
