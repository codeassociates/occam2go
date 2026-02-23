package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/parser"
)

func TestE2E_SkipStatement(t *testing.T) {
	// SKIP as a standalone statement — should be a no-op
	occam := `SEQ
  print.int(1)
  SKIP
  print.int(2)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_SkipInPar(t *testing.T) {
	// SKIP in a PAR branch — one branch does nothing
	occam := `SEQ
  INT x:
  PAR
    SKIP
    x := 42
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_StopReached(t *testing.T) {
	// STOP should print an error message to stderr and halt (deadlock via select{})
	// We verify the program exits with non-zero status and prints to stderr
	occamSource := `SEQ
  STOP
`
	l := lexer.New(occamSource)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	gen := New()
	goCode := gen.Generate(program)

	tmpDir, err := os.MkdirTemp("", "occam2go-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		t.Fatalf("failed to write Go file: %v", err)
	}

	binFile := filepath.Join(tmpDir, "main")
	compileCmd := exec.Command("go", "build", "-o", binFile, goFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compilation failed: %v\nOutput: %s\nGo code:\n%s", err, compileOutput, goCode)
	}

	// Run with a timeout — STOP causes a deadlock (select{})
	runCmd := exec.Command(binFile)
	err = runCmd.Start()
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// The program should deadlock, so we just verify it compiles and starts.
	// Kill it after a short delay.
	done := make(chan error, 1)
	go func() {
		done <- runCmd.Wait()
	}()

	select {
	case err := <-done:
		// If it exited, it should be non-zero (fatal error: all goroutines are asleep)
		if err == nil {
			t.Errorf("expected STOP to cause non-zero exit, but exited successfully")
		}
	case <-func() <-chan struct{} {
		ch := make(chan struct{})
		go func() {
			// Wait 2 seconds then signal
			exec.Command("sleep", "0.5").Run()
			close(ch)
		}()
		return ch
	}():
		// Expected: program is stuck in select{}, kill it
		runCmd.Process.Kill()
	}
}

func TestE2E_ModuloOperator(t *testing.T) {
	// \ is the modulo operator in occam, maps to % in Go
	occam := `SEQ
  INT x:
  x := 42 \ 5
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ModuloInExpression(t *testing.T) {
	// Modulo used in a larger expression
	occam := `SEQ
  INT x:
  x := (17 \ 5) + (10 \ 3)
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	// 17 % 5 = 2, 10 % 3 = 1, sum = 3
	expected := "3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltWithBooleanGuard(t *testing.T) {
	// ALT with boolean guard: FALSE guard disables a channel
	// Only send on c2 since c1's guard is FALSE and won't be selected
	occam := `SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  BOOL allow:
  allow := FALSE
  PAR
    c2 ! 42
    ALT
      allow & c1 ? result
        SKIP
      TRUE & c2 ? result
        SKIP
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltWithTrueGuard(t *testing.T) {
	// ALT where guard evaluates to TRUE for the first channel
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  PAR
    c ! 99
    ALT
      TRUE & c ? result
        print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltWithParenthesizedGuard(t *testing.T) {
	// Issue #78: parenthesized guard expression in ALT
	occam := `SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  INT mode:
  mode := 1
  PAR
    c2 ! 77
    ALT
      (mode <> 1) & c1 ? result
        SKIP
      (mode <> 0) & c2 ? result
        SKIP
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "77\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltGuardedSkip(t *testing.T) {
	// Issue #78: guard & SKIP in ALT (always-ready alternative)
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  BOOL ready:
  ready := TRUE
  result := 0
  PAR
    SEQ
      ALT
        ready & SKIP
          SKIP
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

func TestE2E_AltGuardedSkipTrue(t *testing.T) {
	// Issue #77: ALT with channel case and guarded SKIP where guard is TRUE
	// The SKIP fires immediately, then the channel send proceeds
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  BOOL ready:
  ready := TRUE
  result := 0
  PAR
    SEQ
      ALT
        ready & SKIP
          result := 99
        c ? result
          SKIP
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

func TestE2E_AltGuardedSkipFalse(t *testing.T) {
	// Issue #77: ALT with channel case and guarded SKIP where guard is FALSE
	// The SKIP guard is false, so the channel case fires
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  BOOL ready:
  ready := FALSE
  result := 0
  PAR
    SEQ
      ALT
        ready & SKIP
          result := 99
        c ? result
          SKIP
    c ! 77
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "77\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiLineAbbreviation(t *testing.T) {
	// Issue #79: IS at end of line as continuation
	occam := `SEQ
  VAL INT x IS
    42 :
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiLineAbbreviationExpr(t *testing.T) {
	// Issue #79: IS continuation with complex expression
	occam := `SEQ
  VAL INT a IS 10 :
  VAL INT b IS
    (a + 5) :
  print.int(b)
`
	output := transpileCompileRun(t, occam)
	expected := "15\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MostNegReal32(t *testing.T) {
	// MOSTNEG REAL32 → -math.MaxFloat32 (a very large negative number)
	occam := `SEQ
  REAL32 x:
  x := MOSTNEG REAL32
  IF
    x < (REAL32 0)
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

func TestE2E_MostPosReal32(t *testing.T) {
	// MOSTPOS REAL32 → math.MaxFloat32
	occam := `SEQ
  REAL32 x:
  x := MOSTPOS REAL32
  IF
    x > (REAL32 0)
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

func TestE2E_MostNegReal64(t *testing.T) {
	// MOSTNEG REAL64 → -math.MaxFloat64
	occam := `SEQ
  REAL64 x:
  x := MOSTNEG REAL64
  IF
    x < (REAL64 0)
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

func TestE2E_MostPosReal64(t *testing.T) {
	// MOSTPOS REAL64 → math.MaxFloat64
	occam := `SEQ
  REAL64 x:
  x := MOSTPOS REAL64
  IF
    x > (REAL64 0)
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

func TestE2E_ShorthandSliceFromZero(t *testing.T) {
	// [arr FOR m] — shorthand for [arr FROM 0 FOR m]
	occam := `SEQ
  [5]INT arr:
  SEQ i = 0 FOR 5
    arr[i] := i * 10
  INT sum:
  sum := 0
  VAL first3 IS [arr FOR 3]:
  SEQ i = 0 FOR 3
    sum := sum + first3[i]
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	// 0 + 10 + 20 = 30
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_StringToByteSliceWrapping(t *testing.T) {
	// When passing a string literal to a []BYTE param, it should wrap with []byte()
	occam := `PROC first.char(VAL []BYTE s, INT result)
  result := INT s[0]
:

SEQ
  INT ch:
  first.char("hello", ch)
  print.int(ch)
`
	output := transpileCompileRun(t, occam)
	// 'h' = 104
	expected := "104\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_GoReservedWordEscaping(t *testing.T) {
	// Test that occam identifiers matching Go reserved words are escaped
	// e.g., a variable named "string" should work
	occam := `SEQ
  INT len:
  len := 42
  print.int(len)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_GoReservedWordByte(t *testing.T) {
	// "byte" is a Go reserved word — should be escaped to _byte
	occam := `SEQ
  INT byte:
  byte := 99
  print.int(byte)
`
	output := transpileCompileRun(t, occam)
	expected := "99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiLineExpression(t *testing.T) {
	// Multi-line expression with continuation operator at end of line
	occam := `SEQ
  INT x:
  x := 10 +
    20 +
    12
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiLineParenExpression(t *testing.T) {
	// Expression spanning multiple lines inside parentheses
	occam := `SEQ
  INT x:
  x := (10
    + 20
    + 12)
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NegativeIntLiteral(t *testing.T) {
	// Negative integer literals (unary minus)
	occam := `SEQ
  INT x:
  x := -42
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "-42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NotOperator(t *testing.T) {
	// NOT boolean operator
	occam := `SEQ
  BOOL x:
  x := NOT TRUE
  print.bool(x)
`
	output := transpileCompileRun(t, occam)
	expected := "false\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_LogicalAndOr(t *testing.T) {
	// AND / OR operators
	occam := `SEQ
  BOOL a, b:
  a := TRUE AND FALSE
  b := TRUE OR FALSE
  print.bool(a)
  print.bool(b)
`
	output := transpileCompileRun(t, occam)
	expected := "false\ntrue\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_NestedIfInSeq(t *testing.T) {
	// Nested IF inside SEQ with variable declarations
	occam := `SEQ
  INT x:
  x := 5
  INT y:
  y := 0
  IF
    x > 3
      IF
        x < 10
          y := 1
        TRUE
          y := 2
    TRUE
      y := 3
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_WhileWithBreakCondition(t *testing.T) {
	// WHILE loop counting to a target
	occam := `SEQ
  INT sum, i:
  sum := 0
  i := 1
  WHILE i <= 10
    SEQ
      sum := sum + i
      i := i + 1
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	// 1+2+...+10 = 55
	expected := "55\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CaseWithMultipleArms(t *testing.T) {
	// CASE with several branches
	occam := `SEQ
  INT x, result:
  x := 3
  CASE x
    1
      result := 10
    2
      result := 20
    3
      result := 30
    ELSE
      result := 0
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_EqualNotEqual(t *testing.T) {
	// = and <> operators
	occam := `SEQ
  print.bool(5 = 5)
  print.bool(5 <> 3)
  print.bool(5 = 3)
  print.bool(5 <> 5)
`
	output := transpileCompileRun(t, occam)
	expected := "true\ntrue\nfalse\nfalse\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CompileOnly_StopInProc(t *testing.T) {
	// STOP inside a proc — just verify it compiles (don't run, it would deadlock)
	occamSource := `PROC fatal()
  STOP
:

SEQ
  print.int(42)
`
	l := lexer.New(occamSource)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	gen := New()
	goCode := gen.Generate(program)

	tmpDir, err := os.MkdirTemp("", "occam2go-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		t.Fatalf("failed to write Go file: %v", err)
	}

	// Just check it compiles
	compileCmd := exec.Command("go", "vet", goFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compilation failed: %v\nOutput: %s\nGo code:\n%s", err, compileOutput, goCode)
	}
}

func TestE2E_NestedReplicatedSeq(t *testing.T) {
	// Nested replicated SEQ — matrix-like access
	occam := `SEQ
  INT sum:
  sum := 0
  SEQ i = 0 FOR 3
    SEQ j = 0 FOR 3
      sum := sum + ((i * 3) + j)
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	// 0+1+2+3+4+5+6+7+8 = 36
	expected := "36\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArraySliceAssignment(t *testing.T) {
	// [arr FROM n FOR m] := src — copy slice
	occam := `SEQ
  [5]INT dst:
  [3]INT src:
  SEQ i = 0 FOR 5
    dst[i] := 0
  src[0] := 10
  src[1] := 20
  src[2] := 30
  [dst FROM 1 FOR 3] := src
  SEQ i = 0 FOR 5
    print.int(dst[i])
`
	output := transpileCompileRun(t, occam)
	expected := "0\n10\n20\n30\n0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_FunctionCallInCondition(t *testing.T) {
	// Function call used as condition in IF
	occam := `BOOL FUNCTION is.positive(VAL INT x)
  IS x > 0

SEQ
  IF
    is.positive(42)
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

func TestE2E_RecursiveFunction(t *testing.T) {
	// Recursive function (factorial)
	occam := `INT FUNCTION factorial(VAL INT n)
  INT result:
  VALOF
    IF
      n <= 1
        result := 1
      TRUE
        result := n * factorial(n - 1)
    RESULT result

SEQ
  print.int(factorial(5))
`
	output := transpileCompileRun(t, occam)
	expected := "120\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiLineProcParams(t *testing.T) {
	// Procedure with parameters spanning multiple lines (paren suppression)
	occam := `PROC add(
  VAL INT a,
  VAL INT b,
  INT result)
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

func TestE2E_VetOutputClean(t *testing.T) {
	// Verify go vet passes on generated code for a non-trivial program
	occamSource := `PROC compute(VAL INT n)
  INT x:
  PROC helper()
    x := n * 2
  :
  SEQ
    helper()
    print.int(x)
:

SEQ
  compute(21)
`
	l := lexer.New(occamSource)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	gen := New()
	goCode := gen.Generate(program)

	tmpDir, err := os.MkdirTemp("", "occam2go-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		t.Fatalf("failed to write Go file: %v", err)
	}

	vetCmd := exec.Command("go", "vet", goFile)
	vetOutput, err := vetCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go vet failed: %v\nOutput: %s\nGo code:\n%s", err, vetOutput, goCode)
	}

	// Also verify it runs correctly
	output := transpileCompileRun(t, occamSource)
	if strings.TrimSpace(output) != "42" {
		t.Errorf("expected 42, got %q", output)
	}
}
