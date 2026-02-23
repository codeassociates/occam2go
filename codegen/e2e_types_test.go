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

func TestE2E_MostNegInt(t *testing.T) {
	occam := `SEQ
  INT x:
  x := MOSTNEG INT
  BOOL neg:
  IF
    x < 0
      neg := TRUE
    TRUE
      neg := FALSE
  print.bool(neg)
`
	output := transpileCompileRun(t, occam)
	expected := "true\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MostPosInt(t *testing.T) {
	occam := `SEQ
  INT x:
  x := MOSTPOS INT
  BOOL pos:
  IF
    x > 0
      pos := TRUE
    TRUE
      pos := FALSE
  print.bool(pos)
`
	output := transpileCompileRun(t, occam)
	expected := "true\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MostNegByte(t *testing.T) {
	occam := `SEQ
  BYTE x:
  x := MOSTNEG BYTE
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MostPosByte(t *testing.T) {
	occam := `SEQ
  BYTE x:
  x := MOSTPOS BYTE
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "255\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MostNegInExpression(t *testing.T) {
	// Test MOSTNEG INT used in comparison (like utils.occ does)
	occam := `SEQ
  INT n:
  n := MOSTNEG INT
  IF
    n = (MOSTNEG INT)
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

func TestE2E_CheckedArithmeticPLUS(t *testing.T) {
	occam := `SEQ
  INT x:
  SEQ
    x := 3 PLUS 4
    print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "7\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CheckedArithmeticMINUS(t *testing.T) {
	occam := `SEQ
  INT x:
  SEQ
    x := 10 MINUS 3
    print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "7\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CheckedArithmeticTIMES(t *testing.T) {
	occam := `SEQ
  INT x:
  SEQ
    x := 6 TIMES 7
    print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CheckedArithmeticWrapping(t *testing.T) {
	// MOSTPOS INT PLUS 1 should wrap to MOSTNEG INT (modular arithmetic)
	// Use a variable so Go doesn't detect constant overflow at compile time
	occam := `SEQ
  INT x:
  SEQ
    x := MOSTPOS INT
    x := x PLUS 1
    BOOL neg:
    IF
      x = (MOSTNEG INT)
        neg := TRUE
      TRUE
        neg := FALSE
    print.bool(neg)
`
	output := transpileCompileRun(t, occam)
	expected := "true\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Int16VarDeclAndConversion(t *testing.T) {
	occam := `SEQ
  INT16 x:
  x := 1000
  INT y:
  y := INT x
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "1000\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Int32VarDeclAndConversion(t *testing.T) {
	occam := `SEQ
  INT32 x:
  x := 100000
  INT y:
  y := INT x
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "100000\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Int64VarDeclAndConversion(t *testing.T) {
	occam := `SEQ
  INT64 x:
  x := 123456789
  INT y:
  y := INT x
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "123456789\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MostNegMostPosInt16(t *testing.T) {
	occam := `SEQ
  INT16 x:
  x := MOSTNEG INT16
  print.int(INT x)
  x := MOSTPOS INT16
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "-32768\n32767\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MostNegMostPosInt32(t *testing.T) {
	occam := `SEQ
  INT32 x:
  x := MOSTNEG INT32
  print.int(INT x)
  x := MOSTPOS INT32
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "-2147483648\n2147483647\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Int16TypeConversionFromInt(t *testing.T) {
	occam := `SEQ
  INT n:
  n := 42
  INT16 x:
  x := INT16 n
  print.int(INT x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_BoolToInt(t *testing.T) {
	occam := `SEQ
  BOOL a:
  a := TRUE
  INT x:
  x := INT a
  print.int(x)
  a := FALSE
  x := INT a
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_BoolToByte(t *testing.T) {
	occam := `SEQ
  BOOL a:
  a := TRUE
  BYTE b:
  b := BYTE a
  print.int(INT b)
  a := FALSE
  b := BYTE a
  print.int(INT b)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IntToBool(t *testing.T) {
	occam := `SEQ
  INT n:
  n := 42
  BOOL a:
  a := BOOL n
  print.bool(a)
  n := 0
  a := BOOL n
  print.bool(a)
`
	output := transpileCompileRun(t, occam)
	expected := "true\nfalse\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ByteToBool(t *testing.T) {
	occam := `SEQ
  BYTE b:
  b := 1
  BOOL a:
  a := BOOL b
  print.bool(a)
  b := 0
  a := BOOL b
  print.bool(a)
`
	output := transpileCompileRun(t, occam)
	expected := "true\nfalse\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ComparisonToInt(t *testing.T) {
	occam := `SEQ
  INT a, b:
  a := 5
  b := 3
  INT result:
  result := INT (a > b)
  print.int(result)
  result := INT (a < b)
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
