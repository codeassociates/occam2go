package codegen

import "testing"

func TestE2E_ValByteArrayAbbreviation(t *testing.T) {
	// VAL []BYTE s IS "hello": — open array byte abbreviation
	occam := `SEQ
  VAL []BYTE s IS "hello":
  print.int(SIZE s)
`
	output := transpileCompileRun(t, occam)
	expected := "5\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_PrintString(t *testing.T) {
	// print.string should output the string content
	occam := `SEQ
  print.string("hello world")
`
	output := transpileCompileRun(t, occam)
	expected := "hello world\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_PrintNewline(t *testing.T) {
	// print.newline should output a blank line
	occam := `SEQ
  print.int(1)
  print.newline()
  print.int(2)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n\n2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_PrintStringAndNewline(t *testing.T) {
	// Combined usage of print.string and print.newline
	occam := `SEQ
  print.string("first")
  print.string("second")
`
	output := transpileCompileRun(t, occam)
	expected := "first\nsecond\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_StringWithEscapes(t *testing.T) {
	// Occam escape sequences in string: *n = newline, *t = tab
	occam := `SEQ
  print.string("a*tb")
`
	output := transpileCompileRun(t, occam)
	expected := "a\tb\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
