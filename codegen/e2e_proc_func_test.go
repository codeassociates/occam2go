package codegen

import "testing"

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

func TestE2E_FunctionIS(t *testing.T) {
	occam := `INT FUNCTION square(VAL INT x)
  IS x * x

SEQ
  INT n:
  n := square(7)
  print.int(n)
`
	output := transpileCompileRun(t, occam)
	expected := "49\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_FunctionValof(t *testing.T) {
	occam := `INT FUNCTION max(VAL INT a, VAL INT b)
  INT result:
  VALOF
    IF
      a > b
        result := a
      TRUE
        result := b
    RESULT result

SEQ
  print.int(max(10, 20))
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_FunctionInExpr(t *testing.T) {
	occam := `INT FUNCTION double(VAL INT x)
  IS x * 2

SEQ
  INT a:
  a := double(3) + double(4)
  print.int(a)
`
	output := transpileCompileRun(t, occam)
	expected := "14\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ValAbbreviation(t *testing.T) {
	occam := `SEQ
  VAL INT x IS 42:
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AbbreviationWithExpression(t *testing.T) {
	occam := `SEQ
  INT a:
  a := 10
  VAL INT b IS (a + 5):
  print.int(b)
`
	output := transpileCompileRun(t, occam)
	expected := "15\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ProcColonTerminator(t *testing.T) {
	occam := `PROC double(VAL INT x, INT result)
  SEQ
    result := x * 2
:

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

func TestE2E_FunctionISColonTerminator(t *testing.T) {
	occam := `INT FUNCTION square(VAL INT x)
  IS x * x
:

SEQ
  INT n:
  n := square(7)
  print.int(n)
`
	output := transpileCompileRun(t, occam)
	expected := "49\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_FunctionValofColonTerminator(t *testing.T) {
	occam := `INT FUNCTION max(VAL INT a, VAL INT b)
  INT result:
  VALOF
    IF
      a > b
        result := a
      TRUE
        result := b
    RESULT result
:

SEQ
  print.int(max(10, 20))
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultipleProcColonTerminator(t *testing.T) {
	occam := `PROC add(VAL INT a, VAL INT b, INT result)
  result := a + b
:

PROC double(VAL INT x, INT result)
  SEQ
    result := x * 2
:

SEQ
  INT sum, doubled:
  add(10, 11, sum)
  double(sum, doubled)
  print.int(doubled)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiResultFunction(t *testing.T) {
	occam := `INT, INT FUNCTION swap(VAL INT a, VAL INT b)
  INT x, y:
  VALOF
    SEQ
      x := b
      y := a
    RESULT x, y

SEQ
  INT p, q:
  p, q := swap(10, 20)
  print.int(p)
  print.int(q)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n10\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiResultFunctionThreeValues(t *testing.T) {
	occam := `INT, INT, INT FUNCTION rotate(VAL INT a, VAL INT b, VAL INT c)
  INT x, y, z:
  VALOF
    SEQ
      x := b
      y := c
      z := a
    RESULT x, y, z

SEQ
  INT p, q, r:
  p, q, r := rotate(1, 2, 3)
  print.int(p)
  print.int(q)
  print.int(r)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n3\n1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NonValAbbreviation(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 7
  INT y IS x:
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "7\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedProcClosure(t *testing.T) {
	occam := `PROC outer(VAL INT n)
  INT x:
  PROC inner()
    x := n * 2
  :
  SEQ
    inner()
    print.int(x)
:

SEQ
  outer(21)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedProcWithParams(t *testing.T) {
	occam := `PROC outer()
  INT result:
  PROC add(VAL INT a, VAL INT b)
    result := a + b
  :
  SEQ
    add(10, 32)
    print.int(result)
:

SEQ
  outer()
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedFunctionIS(t *testing.T) {
	occam := `PROC compute()
  INT FUNCTION double(VAL INT x)
    IS x * 2
  SEQ
    print.int(double(21))
:

SEQ
  compute()
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedFunctionValof(t *testing.T) {
	occam := `PROC compute()
  INT FUNCTION max(VAL INT a, VAL INT b)
    INT result:
    VALOF
      IF
        a > b
          result := a
        TRUE
          result := b
      RESULT result
  SEQ
    print.int(max(42, 10))
:

SEQ
  compute()
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultipleNestedProcs(t *testing.T) {
	occam := `PROC compute()
  INT x:
  PROC setten()
    x := 10
  :
  PROC setthirtytwo()
    x := 32
  :
  SEQ
    setten()
    INT a:
    a := x
    setthirtytwo()
    print.int(a + x)
:

SEQ
  compute()
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ProcLocalVarDecls(t *testing.T) {
	occam := `PROC foo(VAL INT n)
  INT x:
  INT y:
  SEQ
    x := n
    y := n * 2
    print.int(x + y)
:

SEQ
  foo(14)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
