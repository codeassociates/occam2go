package codegen

import "testing"

func TestE2E_IfBasic(t *testing.T) {
	// Test basic IF: first branch is true
	occam := `SEQ
  INT x, y:
  x := 5
  y := 0
  IF
    x > 0
      y := 1
    x = 0
      y := 2
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IfSecondBranch(t *testing.T) {
	// Test IF where second branch matches
	occam := `SEQ
  INT x, y:
  x := 0
  y := 0
  IF
    x > 0
      y := 1
    x = 0
      y := 2
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IfThreeBranches(t *testing.T) {
	// Test IF with three choices where the last matches
	occam := `SEQ
  INT x, y:
  x := 0
  y := 0
  IF
    x > 0
      y := 1
    x < 0
      y := 2
    x = 0
      y := 3
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IfWithSeqBody(t *testing.T) {
	// Test IF with SEQ body in branches
	occam := `SEQ
  INT x, y:
  x := 10
  y := 0
  IF
    x > 5
      SEQ
        y := x * 2
        print.int(y)
    x <= 5
      SEQ
        y := x * 3
        print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_WhileBasic(t *testing.T) {
	// Test basic WHILE loop
	occam := `SEQ
  INT x:
  x := 3
  WHILE x > 0
    SEQ
      print.int(x)
      x := x - 1
`
	output := transpileCompileRun(t, occam)
	expected := "3\n2\n1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_WhileSum(t *testing.T) {
	// Test WHILE loop computing a sum
	occam := `SEQ
  INT i, sum:
  i := 1
  sum := 0
  WHILE i <= 5
    SEQ
      sum := sum + i
      i := i + 1
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "15\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_WhileNested(t *testing.T) {
	// Test nested WHILE loops (multiplication table style)
	occam := `SEQ
  INT i, j, product:
  i := 1
  WHILE i <= 2
    SEQ
      j := 1
      WHILE j <= 2
        SEQ
          product := i * j
          print.int(product)
          j := j + 1
      i := i + 1
`
	output := transpileCompileRun(t, occam)
	expected := "1\n2\n2\n4\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CaseBasic(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 2
  CASE x
    1
      print.int(10)
    2
      print.int(20)
    3
      print.int(30)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CaseElse(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 99
  CASE x
    1
      print.int(10)
    2
      print.int(20)
    ELSE
      print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CaseExpression(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 3
  CASE x + 1
    3
      print.int(30)
    4
      print.int(40)
    ELSE
      print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "40\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiStatementIfBody(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 5
  IF
    x > 0
      INT y:
      y := x + 10
      print.int(y)
    TRUE
      SKIP
`
	output := transpileCompileRun(t, occam)
	expected := "15\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiStatementCaseBody(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 2
  CASE x
    1
      print.int(10)
    2
      INT y:
      y := x * 100
      print.int(y)
    ELSE
      print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "200\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiStatementWhileBody(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 0
  WHILE x < 3
    INT step:
    step := 1
    x := x + step
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedReplicatedIfWithDefault(t *testing.T) {
	// Replicated IF as a choice within outer IF, with TRUE default
	occam := `SEQ
  [5]INT arr:
  INT result:
  SEQ i = 0 FOR 5
    arr[i] := i * 10
  IF
    IF i = 0 FOR 5
      arr[i] > 25
        result := arr[i]
    TRUE
      result := -1
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedReplicatedIfNoMatch(t *testing.T) {
	// Replicated IF where no choice matches, falls through to TRUE
	occam := `SEQ
  [3]INT arr:
  INT result:
  SEQ i = 0 FOR 3
    arr[i] := i
  IF
    IF i = 0 FOR 3
      arr[i] > 100
        result := arr[i]
    TRUE
      result := -1
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "-1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedReplicatedIfWithPrecedingChoice(t *testing.T) {
	// Normal choice before replicated IF, then default
	occam := `SEQ
  [3]INT arr:
  INT result:
  SEQ i = 0 FOR 3
    arr[i] := i
  INT x:
  x := 99
  IF
    x > 100
      result := x
    IF i = 0 FOR 3
      arr[i] = 2
        result := arr[i]
    TRUE
      result := -1
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedNonReplicatedIf(t *testing.T) {
	// Non-replicated nested IF (choices inlined into parent)
	occam := `SEQ
  INT x:
  INT result:
  x := 5
  IF
    IF
      x > 10
        result := 1
      x > 3
        result := 2
    TRUE
      result := 0
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChannelDirAtCallSite(t *testing.T) {
	occam := `PROC worker(CHAN OF INT in?, CHAN OF INT out!)
  INT x:
  in ? x
  out ! x + 1
:
SEQ
  CHAN OF INT a:
  CHAN OF INT b:
  PAR
    worker(a?, b!)
    SEQ
      a ! 10
      INT result:
      b ? result
      print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "11\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
