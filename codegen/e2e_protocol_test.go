package codegen

import "testing"

func TestE2E_SimpleProtocol(t *testing.T) {
	// Simple protocol: just a named type alias
	occam := `PROTOCOL SIGNAL IS INT

SEQ
  CHAN OF SIGNAL c:
  INT result:
  PAR
    c ! 42
    c ? result
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SequentialProtocol(t *testing.T) {
	// Sequential protocol: send/receive multiple values
	occam := `PROTOCOL PAIR IS INT ; INT

SEQ
  CHAN OF PAIR c:
  INT x, y:
  PAR
    c ! 10 ; 20
    c ? x ; y
  print.int(x)
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "10\n20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_VariantProtocol(t *testing.T) {
	// Variant protocol: tagged union with CASE receive
	occam := `PROTOCOL MSG
  CASE
    data; INT
    quit

SEQ
  CHAN OF MSG c:
  INT result:
  result := 0
  PAR
    c ! data ; 42
    c ? CASE
      data ; result
        print.int(result)
      quit
        print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_VariantProtocolNoPayload(t *testing.T) {
	// Variant protocol with no-payload tag
	occam := `PROTOCOL MSG
  CASE
    data; INT
    quit

SEQ
  CHAN OF MSG c:
  INT result:
  result := 0
  PAR
    c ! quit
    c ? CASE
      data ; result
        print.int(result)
      quit
        print.int(99)
`
	output := transpileCompileRun(t, occam)
	expected := "99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_VariantProtocolDottedTags(t *testing.T) {
	// Variant protocol with dotted tag names (e.g., bar.data)
	occam := `PROTOCOL BAR.PROTO
  CASE
    bar.data; INT
    bar.terminate
    bar.blank; INT

SEQ
  CHAN OF BAR.PROTO c:
  INT result:
  result := 0
  PAR
    SEQ
      c ! bar.data ; 42
      c ! bar.terminate
    SEQ
      c ? CASE
        bar.data ; result
          print.int(result)
        bar.terminate
          print.int(0)
        bar.blank ; result
          print.int(result)
      c ? CASE
        bar.data ; result
          print.int(result)
        bar.terminate
          print.int(99)
        bar.blank ; result
          print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ProtocolWithProc(t *testing.T) {
	// Protocol channel passed as PROC parameter
	occam := `PROTOCOL PAIR IS INT ; INT

PROC sender(CHAN OF PAIR out)
  out ! 3 ; 7

SEQ
  CHAN OF PAIR c:
  INT a, b:
  PAR
    sender(c)
    c ? a ; b
  print.int(a + b)
`
	output := transpileCompileRun(t, occam)
	expected := "10\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
