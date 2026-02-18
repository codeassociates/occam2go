package codegen

import (
	"strings"
	"testing"
)

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

func TestE2E_StopNotTaken(t *testing.T) {
	// Test STOP in an IF branch that is NOT taken — program completes normally
	occam := `SEQ
  INT x:
  x := 1
  IF
    x = 0
      STOP
    TRUE
      print.int(42)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_InitialDecl(t *testing.T) {
	occam := `SEQ
  INITIAL INT x IS 10:
  print.int(x)
  x := x + 5
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "10\n15\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
