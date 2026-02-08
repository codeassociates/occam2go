package codegen

import (
	"strings"
	"testing"

	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/parser"
)

func TestSimpleVarDecl(t *testing.T) {
	input := `INT x:
`
	output := transpile(t, input)

	if !strings.Contains(output, "var x int") {
		t.Errorf("expected 'var x int' in output, got:\n%s", output)
	}
}

func TestMultipleVarDecl(t *testing.T) {
	input := `INT x, y, z:
`
	output := transpile(t, input)

	if !strings.Contains(output, "var x, y, z int") {
		t.Errorf("expected 'var x, y, z int' in output, got:\n%s", output)
	}
}

func TestAssignment(t *testing.T) {
	input := `x := 42
`
	output := transpile(t, input)

	if !strings.Contains(output, "x = 42") {
		t.Errorf("expected 'x = 42' in output, got:\n%s", output)
	}
}

func TestBinaryExpression(t *testing.T) {
	input := `x := a + b
`
	output := transpile(t, input)

	if !strings.Contains(output, "x = (a + b)") {
		t.Errorf("expected 'x = (a + b)' in output, got:\n%s", output)
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := a = b\n", "x = (a == b)"},
		{"x := a <> b\n", "x = (a != b)"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("expected %q in output, got:\n%s", tt.expected, output)
		}
	}
}

func TestSeqBlock(t *testing.T) {
	input := `SEQ
  INT x:
  x := 10
`
	output := transpile(t, input)

	// SEQ becomes sequential Go code
	if !strings.Contains(output, "var x int") {
		t.Errorf("expected 'var x int' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "x = 10") {
		t.Errorf("expected 'x = 10' in output, got:\n%s", output)
	}
}

func TestParBlock(t *testing.T) {
	input := `PAR
  x := 1
  y := 2
`
	output := transpile(t, input)

	// PAR should use sync.WaitGroup
	if !strings.Contains(output, "sync.WaitGroup") {
		t.Errorf("expected sync.WaitGroup in output, got:\n%s", output)
	}
	if !strings.Contains(output, "wg.Add(2)") {
		t.Errorf("expected wg.Add(2) in output, got:\n%s", output)
	}
	if !strings.Contains(output, "go func()") {
		t.Errorf("expected 'go func()' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "wg.Wait()") {
		t.Errorf("expected wg.Wait() in output, got:\n%s", output)
	}
}

func TestProcDecl(t *testing.T) {
	input := `PROC foo(VAL INT x)
  y := x
`
	output := transpile(t, input)

	if !strings.Contains(output, "func foo(x int)") {
		t.Errorf("expected 'func foo(x int)' in output, got:\n%s", output)
	}
}

func TestProcDeclWithRefParam(t *testing.T) {
	input := `PROC bar(INT x)
  x := 10
`
	output := transpile(t, input)

	// Non-VAL parameter should be pointer
	if !strings.Contains(output, "func bar(x *int)") {
		t.Errorf("expected 'func bar(x *int)' in output, got:\n%s", output)
	}
}

func TestIfStatement(t *testing.T) {
	input := `IF
  x > 0
    y := 1
  x = 0
    y := 0
`
	output := transpile(t, input)

	if !strings.Contains(output, "if (x > 0)") {
		t.Errorf("expected 'if (x > 0)' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "} else if (x == 0)") {
		t.Errorf("expected '} else if (x == 0)' in output, got:\n%s", output)
	}
}

func TestArrayDecl(t *testing.T) {
	input := `[5]INT arr:
`
	output := transpile(t, input)

	if !strings.Contains(output, "arr := make([]int, 5)") {
		t.Errorf("expected 'arr := make([]int, 5)' in output, got:\n%s", output)
	}
}

func TestIndexedAssignment(t *testing.T) {
	input := `arr[2] := 10
`
	output := transpile(t, input)

	if !strings.Contains(output, "arr[2] = 10") {
		t.Errorf("expected 'arr[2] = 10' in output, got:\n%s", output)
	}
}

func transpile(t *testing.T, input string) string {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	gen := New()
	return gen.Generate(program)
}

func TestBitwiseOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := a /\\ b\n", "x = (a & b)"},
		{"x := a \\/ b\n", "x = (a | b)"},
		{"x := a >< b\n", "x = (a ^ b)"},
		{"x := a << 2\n", "x = (a << 2)"},
		{"x := a >> 2\n", "x = (a >> 2)"},
		{"x := ~ a\n", "x = ^a"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q: expected %q in output, got:\n%s", tt.input, tt.expected, output)
		}
	}
}

func TestStringLiteral(t *testing.T) {
	input := `x := "hello world"
`
	output := transpile(t, input)

	if !strings.Contains(output, `x = "hello world"`) {
		t.Errorf("expected 'x = \"hello world\"' in output, got:\n%s", output)
	}
}

func TestStop(t *testing.T) {
	input := "STOP\n"
	output := transpile(t, input)

	if !strings.Contains(output, `fmt.Fprintln(os.Stderr, "STOP encountered")`) {
		t.Errorf("expected fmt.Fprintln(os.Stderr, ...) in output, got:\n%s", output)
	}
	if !strings.Contains(output, "select {}") {
		t.Errorf("expected 'select {}' in output, got:\n%s", output)
	}
	if !strings.Contains(output, `"os"`) {
		t.Errorf("expected os import in output, got:\n%s", output)
	}
}

func TestTypeConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := INT y\n", "x = int(y)"},
		{"x := BYTE n\n", "x = byte(n)"},
		{"x := REAL count\n", "x = float64(count)"},
		{"x := BOOL flag\n", "x = bool(flag)"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q: expected %q in output, got:\n%s", tt.input, tt.expected, output)
		}
	}
}

func TestStringLiteralInProcCall(t *testing.T) {
	input := `print.string("hello")
`
	output := transpile(t, input)

	if !strings.Contains(output, `fmt.Println("hello")`) {
		t.Errorf("expected 'fmt.Println(\"hello\")' in output, got:\n%s", output)
	}
}
