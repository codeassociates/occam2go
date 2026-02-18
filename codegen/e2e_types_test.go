package codegen

import "testing"

func TestE2E_TypeConversionIntFromByte(t *testing.T) {
	occam := `SEQ
  BYTE b:
  b := 65
  INT x:
  x := INT b
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "65\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_TypeConversionByteFromInt(t *testing.T) {
	occam := `SEQ
  INT n:
  n := 72
  BYTE b:
  b := BYTE n
  print.int(INT b)
`
	output := transpileCompileRun(t, occam)
	expected := "72\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_TypeConversionInExpression(t *testing.T) {
	occam := `SEQ
  BYTE b:
  b := 10
  INT x:
  x := INT b + 1
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "11\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Real32VarDecl(t *testing.T) {
	occam := `SEQ
  REAL32 x:
  x := REAL32 5
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "5\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Real64VarDecl(t *testing.T) {
	occam := `SEQ
  REAL64 x:
  x := REAL64 10
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "10\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Real32ToReal64Conversion(t *testing.T) {
	occam := `SEQ
  REAL32 a:
  a := REAL32 7
  REAL64 b:
  b := REAL64 a
  print.int(INT b)
`
	output := transpileCompileRun(t, occam)
	expected := "7\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Real32Array(t *testing.T) {
	occam := `SEQ
  [3]REAL32 arr:
  arr[0] := REAL32 10
  arr[1] := REAL32 20
  arr[2] := REAL32 30
  INT sum:
  sum := (INT arr[0]) + (INT arr[1]) + (INT arr[2])
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "60\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_HexLiteral(t *testing.T) {
	occam := `SEQ
  INT x:
  x := #FF
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "255\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_HexLiteralInExpression(t *testing.T) {
	occam := `SEQ
  INT x:
  x := #0A + #14
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_BitwiseAnd(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 12 /\ 10
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "8\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_BitwiseOr(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 12 \/ 10
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "14\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_BitwiseXor(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 12 >< 10
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "6\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_BitwiseNot(t *testing.T) {
	occam := `SEQ
  INT x:
  x := ~ 0
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "-1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LeftShift(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 1 << 4
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "16\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RightShift(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 16 >> 2
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "4\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AfterExpression(t *testing.T) {
	// Test AFTER as a boolean expression in IF
	occam := `SEQ
  INT t1, t2:
  t1 := 100
  t2 := 200
  IF
    t2 AFTER t1
      print.int(1)
    TRUE
      print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ByteLiteral(t *testing.T) {
	occam := `SEQ
  BYTE x:
  x := 'A'
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "65\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ByteLiteralEscape(t *testing.T) {
	occam := `SEQ
  BYTE x:
  x := '*n'
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "10\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
