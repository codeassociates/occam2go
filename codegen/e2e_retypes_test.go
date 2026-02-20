package codegen

import "testing"

func TestE2E_RetypesFloat32ToInt(t *testing.T) {
	// VAL INT X RETYPES X : where X is a REAL32 parameter
	// Reinterpret float32(1.0) as int → IEEE 754: 0x3F800000 = 1065353216
	occam := `PROC show.bits(VAL REAL32 x)
  VAL INT bits RETYPES x :
  SEQ
    print.int(bits)
:

SEQ
  show.bits(REAL32 1)
`
	output := transpileCompileRun(t, occam)
	expected := "1065353216\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RetypesFloat32Zero(t *testing.T) {
	// float32(0.0) → bits = 0
	occam := `PROC show.bits(VAL REAL32 x)
  VAL INT bits RETYPES x :
  SEQ
    print.int(bits)
:

SEQ
  show.bits(REAL32 0)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RetypesFloat32NegOne(t *testing.T) {
	// float32(-1.0) → IEEE 754: 0xBF800000 = -1082130432 (as signed int32)
	occam := `PROC show.bits(VAL REAL32 x)
  VAL INT bits RETYPES x :
  SEQ
    print.int(bits)
:

SEQ
  REAL32 v:
  v := REAL32 1
  v := REAL32 0 - v
  show.bits(v)
`
	output := transpileCompileRun(t, occam)
	expected := "-1082130432\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RetypesSameNameShadow(t *testing.T) {
	// The classic pattern: VAL INT X RETYPES X : where param is also named X
	// Tests the RETYPES parameter rename mechanism (_rp_X)
	occam := `PROC bits.of(VAL REAL32 X)
  VAL INT X RETYPES X :
  SEQ
    print.int(X)
:

SEQ
  bits.of(REAL32 2)
`
	// float32(2.0) = 0x40000000 = 1073741824
	output := transpileCompileRun(t, occam)
	expected := "1073741824\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RetypesFloat64ToIntPair(t *testing.T) {
	// VAL [2]INT X RETYPES X : reinterpret float64 as two int32 words
	// float64(1.0) = 0x3FF0000000000000
	// lo = 0x00000000 = 0, hi = 0x3FF00000 = 1072693248
	occam := `PROC show.bits64(VAL REAL64 X)
  VAL [2]INT X RETYPES X :
  SEQ
    print.int(X[0])
    print.int(X[1])
:

SEQ
  show.bits64(REAL64 1)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n1072693248\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_RetypesFloat64Zero(t *testing.T) {
	// float64(0.0) → both words should be 0
	occam := `PROC show.bits64(VAL REAL64 X)
  VAL [2]INT X RETYPES X :
  SEQ
    print.int(X[0])
    print.int(X[1])
:

SEQ
  show.bits64(REAL64 0)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
