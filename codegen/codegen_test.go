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
