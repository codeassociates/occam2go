package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/parser"
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

func TestE2E_PrintInt(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 42
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Addition(t *testing.T) {
	occam := `SEQ
  INT x, y:
  x := 10
  y := 20
  print.int(x + y)
`
	output := transpileCompileRun(t, occam)
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Multiplication(t *testing.T) {
	occam := `SEQ
  INT a, b, c:
  a := 3
  b := 4
  c := a * b
  print.int(c)
`
	output := transpileCompileRun(t, occam)
	expected := "12\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Expression(t *testing.T) {
	occam := `SEQ
  INT result:
  result := (2 + 3) * 4
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Procedure(t *testing.T) {
	occam := `PROC double(VAL INT x, INT result)
  SEQ
    result := x * 2

SEQ
  INT n, doubled:
  n := 21
  double(n, doubled)
  print.int(doubled)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_PAR(t *testing.T) {
	// Test that PAR executes both branches
	// We can't guarantee order, but both should run
	occam := `SEQ
  INT x, y:
  PAR
    x := 10
    y := 20
  print.int(x + y)
`
	output := transpileCompileRun(t, occam)
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiplePrints(t *testing.T) {
	occam := `SEQ
  print.int(1)
  print.int(2)
  print.int(3)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n2\n3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Boolean(t *testing.T) {
	occam := `SEQ
  BOOL flag:
  flag := TRUE
  print.bool(flag)
`
	output := transpileCompileRun(t, occam)
	expected := "true\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Comparison(t *testing.T) {
	occam := `SEQ
  INT a, b:
  a := 5
  b := 3
  print.bool(a > b)
`
	output := transpileCompileRun(t, occam)
	expected := "true\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ComplexExpression(t *testing.T) {
	// Test: (10 + 5) * 2 - 6 / 3 = 15 * 2 - 2 = 30 - 2 = 28
	occam := `SEQ
  INT result:
  result := ((10 + 5) * 2) - (6 / 3)
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	// Note: Need to verify Go's integer division matches expectation
	output = strings.TrimSpace(output)
	if output != "28" {
		t.Errorf("expected 28, got %q", output)
	}
}

func TestE2E_Channel(t *testing.T) {
	// Test basic channel communication between parallel processes
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  PAR
    c ! 42
    c ? result
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChannelExpression(t *testing.T) {
	// Test sending an expression over a channel
	occam := `SEQ
  CHAN OF INT c:
  INT x, result:
  x := 10
  PAR
    c ! x * 2
    c ? result
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChannelPingPong(t *testing.T) {
	// Test two-way communication: send a value, double it, send back
	occam := `SEQ
  CHAN OF INT request:
  CHAN OF INT response:
  INT result:
  PAR
    SEQ
      request ! 21
      response ? result
    SEQ
      INT x:
      request ? x
      response ! x * 2
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltBasic(t *testing.T) {
	// Test basic ALT: select from first ready channel
	occam := `SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  PAR
    c1 ! 42
    ALT
      c1 ? result
        print.int(result)
      c2 ? result
        print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltSecondChannel(t *testing.T) {
	// Test ALT selecting from second channel
	occam := `SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  PAR
    c2 ! 99
    ALT
      c1 ? result
        print.int(result)
      c2 ? result
        print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltWithBody(t *testing.T) {
	// Test ALT with computation in body
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  PAR
    c ! 10
    ALT
      c ? result
        SEQ
          result := result * 2
          print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
