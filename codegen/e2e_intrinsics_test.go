package codegen

import "testing"

func TestE2E_LONGPROD(t *testing.T) {
	// LONGPROD(a, b, c) = a*b+c as 64-bit, returns (hi, lo)
	// 100000 * 100000 + 0 = 10000000000
	// 10000000000 = 2 * 2^32 + 1410065408
	// hi = 2, lo = 1410065408
	occam := `PROC main()
  INT hi, lo:
  SEQ
    hi, lo := LONGPROD(100000, 100000, 0)
    print.int(hi)
    print.int(lo)
:
`
	output := transpileCompileRun(t, occam)
	expected := "2\n1410065408\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LONGPRODWithCarry(t *testing.T) {
	// LONGPROD(a, b, carry) = a*b+carry
	// 3 * 4 + 5 = 17, fits in lo word
	occam := `PROC main()
  INT hi, lo:
  SEQ
    hi, lo := LONGPROD(3, 4, 5)
    print.int(hi)
    print.int(lo)
:
`
	output := transpileCompileRun(t, occam)
	expected := "0\n17\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LONGDIV(t *testing.T) {
	// LONGDIV(hi, lo, divisor) divides (hi:lo) by divisor → (quotient, remainder)
	// (0:42) / 5 = quotient 8, remainder 2
	occam := `PROC main()
  INT quot, rem:
  SEQ
    quot, rem := LONGDIV(0, 42, 5)
    print.int(quot)
    print.int(rem)
:
`
	output := transpileCompileRun(t, occam)
	expected := "8\n2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LONGDIVLargeValue(t *testing.T) {
	// (2:1409286144) / 100000 = 10000000000 / 100000 = 100000
	// Use the result from LONGPROD to roundtrip
	occam := `PROC main()
  INT hi, lo, quot, rem:
  SEQ
    hi, lo := LONGPROD(100000, 100000, 0)
    quot, rem := LONGDIV(hi, lo, 100000)
    print.int(quot)
    print.int(rem)
:
`
	output := transpileCompileRun(t, occam)
	expected := "100000\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LONGSUM(t *testing.T) {
	// LONGSUM(a, b, carry) = a+b+carry as 64-bit → (carry_out, sum)
	// 10 + 20 + 0 = 30, no carry
	occam := `PROC main()
  INT carry, sum:
  SEQ
    carry, sum := LONGSUM(10, 20, 0)
    print.int(carry)
    print.int(sum)
:
`
	output := transpileCompileRun(t, occam)
	expected := "0\n30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LONGSUMOverflow(t *testing.T) {
	// LONGSUM with overflow using smaller values that fit cleanly in uint32
	// LONGSUM(0xFFFFFFFF, 1, 0): uint32 max + 1 = 0x1_0000_0000
	// hi (carry) = 1, lo = 0
	occam := `PROC main()
  INT carry, sum:
  SEQ
    carry, sum := LONGSUM(-1, 1, 0)
    print.int(carry)
    print.int(sum)
:
`
	output := transpileCompileRun(t, occam)
	// uint32(-1) = 0xFFFFFFFF, uint32(1) = 1, sum = 0x100000000
	// carry = 1, sum = 0
	expected := "1\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LONGDIFF(t *testing.T) {
	// LONGDIFF(a, b, borrow) = a-b-borrow → (borrow_out, diff)
	// 30 - 10 - 0 = 20, no borrow
	occam := `PROC main()
  INT borrow, diff:
  SEQ
    borrow, diff := LONGDIFF(30, 10, 0)
    print.int(borrow)
    print.int(diff)
:
`
	output := transpileCompileRun(t, occam)
	expected := "0\n20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LONGDIFFBorrow(t *testing.T) {
	// LONGDIFF(10, 30, 0): 10-30 = underflow
	// uint32(10) - uint32(30) = wraps → borrow=1
	// uint32 result: 0xFFFFFFEC → int32 = -20
	occam := `PROC main()
  INT borrow, diff:
  SEQ
    borrow, diff := LONGDIFF(10, 30, 0)
    print.int(borrow)
    print.int(diff)
:
`
	output := transpileCompileRun(t, occam)
	expected := "1\n-20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NORMALISE(t *testing.T) {
	// NORMALISE(hi, lo) shifts left until MSB is set
	// NORMALISE(0, 1) — value is 1, needs 63 left shifts to set bit 63
	occam := `PROC main()
  INT places, nhi, nlo:
  SEQ
    places, nhi, nlo := NORMALISE(0, 1)
    print.int(places)
:
`
	output := transpileCompileRun(t, occam)
	expected := "63\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NORMALISEZero(t *testing.T) {
	// NORMALISE(0, 0) — zero value returns 64 shifts, (0, 0)
	occam := `PROC main()
  INT places, nhi, nlo:
  SEQ
    places, nhi, nlo := NORMALISE(0, 0)
    print.int(places)
    print.int(nhi)
    print.int(nlo)
:
`
	output := transpileCompileRun(t, occam)
	expected := "64\n0\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SHIFTRIGHT(t *testing.T) {
	// SHIFTRIGHT(hi, lo, n) — shift 64-bit (hi:lo) right by n
	// SHIFTRIGHT(0, 16, 2) = shift 16 right by 2 = (0, 4)
	occam := `PROC main()
  INT rhi, rlo:
  SEQ
    rhi, rlo := SHIFTRIGHT(0, 16, 2)
    print.int(rhi)
    print.int(rlo)
:
`
	output := transpileCompileRun(t, occam)
	expected := "0\n4\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SHIFTLEFT(t *testing.T) {
	// SHIFTLEFT(hi, lo, n) — shift 64-bit (hi:lo) left by n
	// SHIFTLEFT(0, 1, 4) = shift 1 left by 4 = (0, 16)
	occam := `PROC main()
  INT rhi, rlo:
  SEQ
    rhi, rlo := SHIFTLEFT(0, 1, 4)
    print.int(rhi)
    print.int(rlo)
:
`
	output := transpileCompileRun(t, occam)
	expected := "0\n16\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SHIFTLEFTCrossWord(t *testing.T) {
	// Shift a value from lo into hi word
	// SHIFTLEFT(0, 1, 32) = (1, 0) — bit moves from lo to hi
	occam := `PROC main()
  INT rhi, rlo:
  SEQ
    rhi, rlo := SHIFTLEFT(0, 1, 32)
    print.int(rhi)
    print.int(rlo)
:
`
	output := transpileCompileRun(t, occam)
	expected := "1\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
