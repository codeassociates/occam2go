package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/parser"
)

func TestE2E_UntypedValAbbreviation(t *testing.T) {
	occam := `SEQ
  VAL x IS 42 :
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArrayLiteralIndexing(t *testing.T) {
	occam := `SEQ
  VAL arr IS [10, 20, 30] :
  print.int(arr[1])
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiLineBooleanIF(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 1
  IF
    (x > 0) AND
      (x < 10)
      print.int(x)
    TRUE
      print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CAUSEERROR(t *testing.T) {
	occamSource := `PROC main()
  CAUSEERROR()
:
`
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

	// Compile
	binFile := filepath.Join(tmpDir, "main")
	compileCmd := exec.Command("go", "build", "-o", binFile, goFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compilation failed: %v\nOutput: %s\nGo code:\n%s", err, compileOutput, goCode)
	}

	// Run — expect non-zero exit code (panic)
	runCmd := exec.Command(binFile)
	err = runCmd.Run()
	if err == nil {
		t.Fatalf("expected CAUSEERROR to cause a non-zero exit, but program exited successfully")
	}
}
