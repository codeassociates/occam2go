package codegen

import "testing"

func TestE2EEntryHarnessEcho(t *testing.T) {
	// An echoing program that reads characters until 'Z' and echoes each one.
	// Uses the standard occam entry-point PROC signature.
	input := `PROC echo(CHAN OF BYTE keyboard?, screen!, error!)
  BYTE ch:
  SEQ
    keyboard ? ch
    WHILE ch <> 'Z'
      SEQ
        screen ! ch
        keyboard ? ch
:
`
	// Pipe "hello Z" — the program should echo "hello " (everything before Z)
	output := transpileCompileRunWithInput(t, input, "hello Z")

	expected := "hello "
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}
