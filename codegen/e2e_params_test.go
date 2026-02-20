package codegen

import "testing"

func TestE2E_ResultQualifier(t *testing.T) {
	// RESULT INT x is semantically the same as non-VAL (pointer param)
	occam := `PROC compute(VAL INT a, VAL INT b, RESULT INT sum)
  sum := a + b
:

SEQ
  INT s:
  compute(10, 32, s)
  print.int(s)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ResultQualifierMultiple(t *testing.T) {
	// Multiple RESULT params
	occam := `PROC divmod(VAL INT a, VAL INT b, RESULT INT quot, RESULT INT rem)
  SEQ
    quot := a / b
    rem := a \ b
:

SEQ
  INT q, r:
  divmod(42, 5, q, r)
  print.int(q)
  print.int(r)
`
	output := transpileCompileRun(t, occam)
	expected := "8\n2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_FixedSizeArrayParam(t *testing.T) {
	// [2]INT param → pointer to fixed-size array
	occam := `PROC swap([2]INT arr)
  INT tmp:
  SEQ
    tmp := arr[0]
    arr[0] := arr[1]
    arr[1] := tmp
:

SEQ
  [2]INT pair:
  pair[0] := 10
  pair[1] := 20
  swap(pair)
  print.int(pair[0])
  print.int(pair[1])
`
	output := transpileCompileRun(t, occam)
	expected := "20\n10\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SharedTypeChanParams(t *testing.T) {
	// PROC f(CHAN OF INT a?, b?) — type applies to both a and b
	occam := `PROC relay(CHAN OF INT input?, output!)
  INT x:
  SEQ
    input ? x
    output ! x
:

SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  PAR
    c1 ! 42
    relay(c1, c2)
    SEQ
      c2 ? result
      print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SharedTypeIntParams(t *testing.T) {
	// PROC f(VAL INT a, b) — type applies to both a and b
	occam := `PROC add(VAL INT a, b, INT result)
  result := a + b
:

SEQ
  INT r:
  add(10, 32, r)
  print.int(r)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ValOpenArrayByteParam(t *testing.T) {
	// VAL []BYTE param with string literal → wraps with []byte()
	occam := `PROC show.length(VAL []BYTE s)
  print.int(SIZE s)
:

SEQ
  show.length("hello")
`
	output := transpileCompileRun(t, occam)
	expected := "5\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
