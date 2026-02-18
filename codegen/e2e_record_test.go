package codegen

import "testing"

func TestE2E_RecordBasic(t *testing.T) {
	occam := `RECORD POINT
  INT x:
  INT y:

SEQ
  POINT p:
  p[x] := 10
  p[y] := 20
  print.int(p[x])
  print.int(p[y])
`
	output := transpileCompileRun(t, occam)
	expected := "10\n20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RecordWithProc(t *testing.T) {
	occam := `RECORD POINT
  INT x:
  INT y:

PROC setPoint(POINT p, VAL INT a, VAL INT b)
  SEQ
    p[x] := a
    p[y] := b

SEQ
  POINT p:
  setPoint(p, 3, 7)
  print.int(p[x])
  print.int(p[y])
`
	output := transpileCompileRun(t, occam)
	expected := "3\n7\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RecordWithValProc(t *testing.T) {
	occam := `RECORD POINT
  INT x:
  INT y:

PROC printPoint(VAL POINT p)
  SEQ
    print.int(p[x])
    print.int(p[y])

SEQ
  POINT p:
  p[x] := 42
  p[y] := 99
  printPoint(p)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RecordInExpression(t *testing.T) {
	occam := `RECORD POINT
  INT x:
  INT y:

SEQ
  POINT p:
  p[x] := 3
  p[y] := 4
  INT sum:
  sum := p[x] + p[y]
  print.int(sum)
  print.int(p[x] * p[y])
`
	output := transpileCompileRun(t, occam)
	expected := "7\n12\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
