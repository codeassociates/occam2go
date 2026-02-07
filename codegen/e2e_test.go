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

// transpileCompileRun takes Occam source, transpiles to Go, compiles, runs,
// and returns the stdout output
func transpileCompileRun(t *testing.T, occamSource string) string {
	t.Helper()

	// Transpile
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

	// Create temp directory for this test
	tmpDir, err := os.MkdirTemp("", "occam2go-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write Go source
	goFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(goFile, []byte(goCode), 0644); err != nil {
		t.Fatalf("failed to write Go file: %v", err)
	}

	// Compile
	binFile := filepath.Join(tmpDir, "main")
	compileCmd := exec.Command("go", "build", "-o", binFile, goFile)
	compileOutput, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compilation failed: %v\nOutput: %s\nGo code:\n%s", err, compileOutput, goCode)
	}

	// Run
	runCmd := exec.Command(binFile)
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	return string(output)
}

func TestE2E_PrintInt(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 42
  print.int(x)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Addition(t *testing.T) {
	occam := `SEQ
  INT x, y:
  x := 10
  y := 20
  print.int(x + y)
`
	output := transpileCompileRun(t, occam)
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Multiplication(t *testing.T) {
	occam := `SEQ
  INT a, b, c:
  a := 3
  b := 4
  c := a * b
  print.int(c)
`
	output := transpileCompileRun(t, occam)
	expected := "12\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Expression(t *testing.T) {
	occam := `SEQ
  INT result:
  result := (2 + 3) * 4
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Procedure(t *testing.T) {
	occam := `PROC double(VAL INT x, INT result)
  SEQ
    result := x * 2

SEQ
  INT n, doubled:
  n := 21
  double(n, doubled)
  print.int(doubled)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_PAR(t *testing.T) {
	// Test that PAR executes both branches
	// We can't guarantee order, but both should run
	occam := `SEQ
  INT x, y:
  PAR
    x := 10
    y := 20
  print.int(x + y)
`
	output := transpileCompileRun(t, occam)
	expected := "30\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_MultiplePrints(t *testing.T) {
	occam := `SEQ
  print.int(1)
  print.int(2)
  print.int(3)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n2\n3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Boolean(t *testing.T) {
	occam := `SEQ
  BOOL flag:
  flag := TRUE
  print.bool(flag)
`
	output := transpileCompileRun(t, occam)
	expected := "true\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_Comparison(t *testing.T) {
	occam := `SEQ
  INT a, b:
  a := 5
  b := 3
  print.bool(a > b)
`
	output := transpileCompileRun(t, occam)
	expected := "true\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ComplexExpression(t *testing.T) {
	// Test: (10 + 5) * 2 - 6 / 3 = 15 * 2 - 2 = 30 - 2 = 28
	occam := `SEQ
  INT result:
  result := ((10 + 5) * 2) - (6 / 3)
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	// Note: Need to verify Go's integer division matches expectation
	output = strings.TrimSpace(output)
	if output != "28" {
		t.Errorf("expected 28, got %q", output)
	}
}

func TestE2E_Channel(t *testing.T) {
	// Test basic channel communication between parallel processes
	occam := `SEQ
  CHAN OF INT c:
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

func TestE2E_ChannelExpression(t *testing.T) {
	// Test sending an expression over a channel
	occam := `SEQ
  CHAN OF INT c:
  INT x, result:
  x := 10
  PAR
    c ! x * 2
    c ? result
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ChannelPingPong(t *testing.T) {
	// Test two-way communication: send a value, double it, send back
	occam := `SEQ
  CHAN OF INT request:
  CHAN OF INT response:
  INT result:
  PAR
    SEQ
      request ! 21
      response ? result
    SEQ
      INT x:
      request ? x
      response ! x * 2
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltBasic(t *testing.T) {
	// Test basic ALT: select from first ready channel
	occam := `SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  PAR
    c1 ! 42
    ALT
      c1 ? result
        print.int(result)
      c2 ? result
        print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltSecondChannel(t *testing.T) {
	// Test ALT selecting from second channel
	occam := `SEQ
  CHAN OF INT c1:
  CHAN OF INT c2:
  INT result:
  PAR
    c2 ! 99
    ALT
      c1 ? result
        print.int(result)
      c2 ? result
        print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "99\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_AltWithBody(t *testing.T) {
	// Test ALT with computation in body
	occam := `SEQ
  CHAN OF INT c:
  INT result:
  PAR
    c ! 10
    ALT
      c ? result
        SEQ
          result := result * 2
          print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IfBasic(t *testing.T) {
	// Test basic IF: first branch is true
	occam := `SEQ
  INT x, y:
  x := 5
  y := 0
  IF
    x > 0
      y := 1
    x = 0
      y := 2
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IfSecondBranch(t *testing.T) {
	// Test IF where second branch matches
	occam := `SEQ
  INT x, y:
  x := 0
  y := 0
  IF
    x > 0
      y := 1
    x = 0
      y := 2
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IfThreeBranches(t *testing.T) {
	// Test IF with three choices where the last matches
	occam := `SEQ
  INT x, y:
  x := 0
  y := 0
  IF
    x > 0
      y := 1
    x < 0
      y := 2
    x = 0
      y := 3
  print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_IfWithSeqBody(t *testing.T) {
	// Test IF with SEQ body in branches
	occam := `SEQ
  INT x, y:
  x := 10
  y := 0
  IF
    x > 5
      SEQ
        y := x * 2
        print.int(y)
    x <= 5
      SEQ
        y := x * 3
        print.int(y)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_WhileBasic(t *testing.T) {
	// Test basic WHILE loop
	occam := `SEQ
  INT x:
  x := 3
  WHILE x > 0
    SEQ
      print.int(x)
      x := x - 1
`
	output := transpileCompileRun(t, occam)
	expected := "3\n2\n1\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_WhileSum(t *testing.T) {
	// Test WHILE loop computing a sum
	occam := `SEQ
  INT i, sum:
  i := 1
  sum := 0
  WHILE i <= 5
    SEQ
      sum := sum + i
      i := i + 1
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "15\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_WhileNested(t *testing.T) {
	// Test nested WHILE loops (multiplication table style)
	occam := `SEQ
  INT i, j, product:
  i := 1
  WHILE i <= 2
    SEQ
      j := 1
      WHILE j <= 2
        SEQ
          product := i * j
          print.int(product)
          j := j + 1
      i := i + 1
`
	output := transpileCompileRun(t, occam)
	expected := "1\n2\n2\n4\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeq(t *testing.T) {
	// Test replicated SEQ: SEQ i = 0 FOR 5 prints 0, 1, 2, 3, 4
	occam := `SEQ i = 0 FOR 5
  print.int(i)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n1\n2\n3\n4\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeqWithExpression(t *testing.T) {
	// Test replicated SEQ with expression for count
	occam := `SEQ
  INT n:
  n := 3
  SEQ i = 0 FOR n
    print.int(i)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n1\n2\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeqWithStartOffset(t *testing.T) {
	// Test replicated SEQ with non-zero start
	occam := `SEQ i = 5 FOR 3
  print.int(i)
`
	output := transpileCompileRun(t, occam)
	expected := "5\n6\n7\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedSeqSum(t *testing.T) {
	// Test replicated SEQ computing sum: 1+2+3+4+5 = 15
	occam := `SEQ
  INT sum:
  sum := 0
  SEQ i = 1 FOR 5
    sum := sum + i
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "15\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ReplicatedPar(t *testing.T) {
	// Test replicated PAR: PAR i = 0 FOR n spawns n goroutines
	// Since PAR is concurrent, we use channels to verify all goroutines ran
	occam := `SEQ
  CHAN OF INT c:
  INT sum:
  sum := 0
  PAR
    PAR i = 0 FOR 5
      c ! i
    SEQ j = 0 FOR 5
      INT x:
      c ? x
      sum := sum + x
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	// sum should be 0+1+2+3+4 = 10
	expected := "10\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArrayBasic(t *testing.T) {
	// Test basic array: declare, store, load
	occam := `SEQ
  [5]INT arr:
  arr[0] := 42
  print.int(arr[0])
`
	output := transpileCompileRun(t, occam)
	expected := "42\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArrayWithLoop(t *testing.T) {
	// Test filling array with replicated SEQ and printing all elements
	occam := `SEQ
  [5]INT arr:
  SEQ i = 0 FOR 5
    arr[i] := i * 10
  SEQ i = 0 FOR 5
    print.int(arr[i])
`
	output := transpileCompileRun(t, occam)
	expected := "0\n10\n20\n30\n40\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArraySum(t *testing.T) {
	// Test computing sum of array elements
	occam := `SEQ
  [4]INT arr:
  arr[0] := 10
  arr[1] := 20
  arr[2] := 30
  arr[3] := 40
  INT sum:
  sum := 0
  SEQ i = 0 FOR 4
    sum := sum + arr[i]
  print.int(sum)
`
	output := transpileCompileRun(t, occam)
	expected := "100\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_ArrayExpressionIndex(t *testing.T) {
	// Test using variable and expression as array index
	occam := `SEQ
  [3]INT arr:
  INT idx:
  arr[0] := 100
  arr[1] := 200
  arr[2] := 300
  idx := 1
  print.int(arr[idx])
  print.int(arr[idx + 1])
`
	output := transpileCompileRun(t, occam)
	expected := "200\n300\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_FunctionIS(t *testing.T) {
	occam := `INT FUNCTION square(VAL INT x)
  IS x * x

SEQ
  INT n:
  n := square(7)
  print.int(n)
`
	output := transpileCompileRun(t, occam)
	expected := "49\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_FunctionValof(t *testing.T) {
	occam := `INT FUNCTION max(VAL INT a, VAL INT b)
  INT result:
  VALOF
    IF
      a > b
        result := a
      TRUE
        result := b
    RESULT result

SEQ
  print.int(max(10, 20))
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_FunctionInExpr(t *testing.T) {
	occam := `INT FUNCTION double(VAL INT x)
  IS x * 2

SEQ
  INT a:
  a := double(3) + double(4)
  print.int(a)
`
	output := transpileCompileRun(t, occam)
	expected := "14\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CaseBasic(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 2
  CASE x
    1
      print.int(10)
    2
      print.int(20)
    3
      print.int(30)
`
	output := transpileCompileRun(t, occam)
	expected := "20\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CaseElse(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 99
  CASE x
    1
      print.int(10)
    2
      print.int(20)
    ELSE
      print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "0\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_CaseExpression(t *testing.T) {
	occam := `SEQ
  INT x:
  x := 3
  CASE x + 1
    3
      print.int(30)
    4
      print.int(40)
    ELSE
      print.int(0)
`
	output := transpileCompileRun(t, occam)
	expected := "40\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestE2E_TimerRead(t *testing.T) {
	// Test reading a timer: value should be positive (microseconds since epoch)
	occam := `SEQ
  TIMER tim:
  INT t:
  tim ? t
  IF
    t > 0
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

func TestE2E_TimerAltTimeout(t *testing.T) {
	// Test ALT with timer timeout: no channel is ready, so timer fires
	occam := `SEQ
  TIMER tim:
  INT t:
  tim ? t
  CHAN OF INT c:
  INT result:
  result := 0
  ALT
    c ? result
      result := 1
    tim ? AFTER (t + 1000)
      result := 2
  print.int(result)
`
	output := transpileCompileRun(t, occam)
	expected := "2\n"
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
