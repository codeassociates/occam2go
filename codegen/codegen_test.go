package codegen

import (
	"strings"
	"testing"

	"github.com/codeassociates/occam2go/lexer"
	"github.com/codeassociates/occam2go/parser"
)

func TestSimpleVarDecl(t *testing.T) {
	input := `INT x:
`
	output := transpile(t, input)

	if !strings.Contains(output, "var x int") {
		t.Errorf("expected 'var x int' in output, got:\n%s", output)
	}
}

func TestMultipleVarDecl(t *testing.T) {
	input := `INT x, y, z:
`
	output := transpile(t, input)

	if !strings.Contains(output, "var x, y, z int") {
		t.Errorf("expected 'var x, y, z int' in output, got:\n%s", output)
	}
}

func TestReal32VarDecl(t *testing.T) {
	input := `REAL32 x:
`
	output := transpile(t, input)

	if !strings.Contains(output, "var x float32") {
		t.Errorf("expected 'var x float32' in output, got:\n%s", output)
	}
}

func TestReal64VarDecl(t *testing.T) {
	input := `REAL64 x:
`
	output := transpile(t, input)

	if !strings.Contains(output, "var x float64") {
		t.Errorf("expected 'var x float64' in output, got:\n%s", output)
	}
}

func TestAssignment(t *testing.T) {
	input := `x := 42
`
	output := transpile(t, input)

	if !strings.Contains(output, "x = 42") {
		t.Errorf("expected 'x = 42' in output, got:\n%s", output)
	}
}

func TestBinaryExpression(t *testing.T) {
	input := `x := a + b
`
	output := transpile(t, input)

	if !strings.Contains(output, "x = (a + b)") {
		t.Errorf("expected 'x = (a + b)' in output, got:\n%s", output)
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := a = b\n", "x = (a == b)"},
		{"x := a <> b\n", "x = (a != b)"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("expected %q in output, got:\n%s", tt.expected, output)
		}
	}
}

func TestSeqBlock(t *testing.T) {
	input := `SEQ
  INT x:
  x := 10
`
	output := transpile(t, input)

	// SEQ becomes sequential Go code
	if !strings.Contains(output, "var x int") {
		t.Errorf("expected 'var x int' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "x = 10") {
		t.Errorf("expected 'x = 10' in output, got:\n%s", output)
	}
}

func TestParBlock(t *testing.T) {
	input := `PAR
  x := 1
  y := 2
`
	output := transpile(t, input)

	// PAR should use sync.WaitGroup
	if !strings.Contains(output, "sync.WaitGroup") {
		t.Errorf("expected sync.WaitGroup in output, got:\n%s", output)
	}
	if !strings.Contains(output, "wg.Add(2)") {
		t.Errorf("expected wg.Add(2) in output, got:\n%s", output)
	}
	if !strings.Contains(output, "go func()") {
		t.Errorf("expected 'go func()' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "wg.Wait()") {
		t.Errorf("expected wg.Wait() in output, got:\n%s", output)
	}
}

func TestProcDecl(t *testing.T) {
	input := `PROC foo(VAL INT x)
  y := x
`
	output := transpile(t, input)

	if !strings.Contains(output, "func foo(x int)") {
		t.Errorf("expected 'func foo(x int)' in output, got:\n%s", output)
	}
}

func TestProcDeclWithRefParam(t *testing.T) {
	input := `PROC bar(INT x)
  x := 10
`
	output := transpile(t, input)

	// Non-VAL parameter should be pointer
	if !strings.Contains(output, "func bar(x *int)") {
		t.Errorf("expected 'func bar(x *int)' in output, got:\n%s", output)
	}
}

func TestIfStatement(t *testing.T) {
	input := `IF
  x > 0
    y := 1
  x = 0
    y := 0
`
	output := transpile(t, input)

	if !strings.Contains(output, "if (x > 0)") {
		t.Errorf("expected 'if (x > 0)' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "} else if (x == 0)") {
		t.Errorf("expected '} else if (x == 0)' in output, got:\n%s", output)
	}
}

func TestReplicatedIf(t *testing.T) {
	input := `IF i = 0 FOR 5
  i = 3
    SKIP
`
	output := transpile(t, input)

	if !strings.Contains(output, "for i := 0; i < 0 + 5; i++") {
		t.Errorf("expected for loop in output, got:\n%s", output)
	}
	if !strings.Contains(output, "if (i == 3)") {
		t.Errorf("expected 'if (i == 3)' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "break") {
		t.Errorf("expected 'break' in output, got:\n%s", output)
	}
}

func TestArrayDecl(t *testing.T) {
	input := `[5]INT arr:
`
	output := transpile(t, input)

	if !strings.Contains(output, "arr := make([]int, 5)") {
		t.Errorf("expected 'arr := make([]int, 5)' in output, got:\n%s", output)
	}
}

func TestIndexedAssignment(t *testing.T) {
	input := `arr[2] := 10
`
	output := transpile(t, input)

	if !strings.Contains(output, "arr[2] = 10") {
		t.Errorf("expected 'arr[2] = 10' in output, got:\n%s", output)
	}
}

func transpile(t *testing.T, input string) string {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}

	gen := New()
	return gen.Generate(program)
}

func TestBitwiseOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := a /\\ b\n", "x = (a & b)"},
		{"x := a \\/ b\n", "x = (a | b)"},
		{"x := a >< b\n", "x = (a ^ b)"},
		{"x := a << 2\n", "x = (a << 2)"},
		{"x := a >> 2\n", "x = (a >> 2)"},
		{"x := ~ a\n", "x = ^a"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q: expected %q in output, got:\n%s", tt.input, tt.expected, output)
		}
	}
}

func TestStringLiteral(t *testing.T) {
	input := `x := "hello world"
`
	output := transpile(t, input)

	if !strings.Contains(output, `x = "hello world"`) {
		t.Errorf("expected 'x = \"hello world\"' in output, got:\n%s", output)
	}
}

func TestByteLiteral(t *testing.T) {
	input := "x := 'A'\n"
	output := transpile(t, input)

	if !strings.Contains(output, "x = byte(65)") {
		t.Errorf("expected 'x = byte(65)' in output, got:\n%s", output)
	}
}

func TestByteLiteralEscape(t *testing.T) {
	input := "x := '*n'\n"
	output := transpile(t, input)

	if !strings.Contains(output, "x = byte(10)") {
		t.Errorf("expected 'x = byte(10)' in output, got:\n%s", output)
	}
}

func TestStop(t *testing.T) {
	input := "STOP\n"
	output := transpile(t, input)

	if !strings.Contains(output, `fmt.Fprintln(os.Stderr, "STOP encountered")`) {
		t.Errorf("expected fmt.Fprintln(os.Stderr, ...) in output, got:\n%s", output)
	}
	if !strings.Contains(output, "select {}") {
		t.Errorf("expected 'select {}' in output, got:\n%s", output)
	}
	if !strings.Contains(output, `"os"`) {
		t.Errorf("expected os import in output, got:\n%s", output)
	}
}

func TestTypeConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := INT y\n", "x = int(y)"},
		{"x := BYTE n\n", "x = byte(n)"},
		{"x := REAL count\n", "x = float64(count)"},
		{"x := BOOL flag\n", "x = bool(flag)"},
		{"x := REAL32 y\n", "x = float32(y)"},
		{"x := REAL64 y\n", "x = float64(y)"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q: expected %q in output, got:\n%s", tt.input, tt.expected, output)
		}
	}
}

func TestMostNegMostPos(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := MOSTNEG INT\n", "x = math.MinInt"},
		{"x := MOSTPOS INT\n", "x = math.MaxInt"},
		{"x := MOSTNEG BYTE\n", "x = 0"},
		{"x := MOSTPOS BYTE\n", "x = 255"},
		{"x := MOSTNEG REAL32\n", "x = -math.MaxFloat32"},
		{"x := MOSTPOS REAL32\n", "x = math.MaxFloat32"},
		{"x := MOSTNEG REAL64\n", "x = -math.MaxFloat64"},
		{"x := MOSTPOS REAL64\n", "x = math.MaxFloat64"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q: expected %q in output, got:\n%s", tt.input, tt.expected, output)
		}
	}
}

func TestMostNegImportsMath(t *testing.T) {
	output := transpile(t, "x := MOSTNEG INT\n")
	if !strings.Contains(output, `"math"`) {
		t.Errorf("expected math import in output, got:\n%s", output)
	}
}

func TestMostNegByteNoMathImport(t *testing.T) {
	output := transpile(t, "x := MOSTNEG BYTE\n")
	if strings.Contains(output, `"math"`) {
		t.Errorf("expected no math import for MOSTNEG BYTE, got:\n%s", output)
	}
}

func TestStringLiteralInProcCall(t *testing.T) {
	input := `print.string("hello")
`
	output := transpile(t, input)

	if !strings.Contains(output, `fmt.Println("hello")`) {
		t.Errorf("expected 'fmt.Println(\"hello\")' in output, got:\n%s", output)
	}
}

func TestCheckedArithmeticCodegen(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := a PLUS b\n", "x = (a + b)"},
		{"x := a MINUS b\n", "x = (a - b)"},
		{"x := a TIMES b\n", "x = (a * b)"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("expected %q in output for %q, got:\n%s", tt.expected, tt.input, output)
		}
	}
}

func TestSimpleProtocolType(t *testing.T) {
	input := `PROTOCOL SIGNAL IS INT
`
	output := transpile(t, input)

	if !strings.Contains(output, "type _proto_SIGNAL = int") {
		t.Errorf("expected 'type _proto_SIGNAL = int' in output, got:\n%s", output)
	}
}

func TestSequentialProtocolType(t *testing.T) {
	input := `PROTOCOL PAIR IS INT ; BYTE
`
	output := transpile(t, input)

	if !strings.Contains(output, "type _proto_PAIR struct {") {
		t.Errorf("expected struct declaration in output, got:\n%s", output)
	}
	if !strings.Contains(output, "_0 int") {
		t.Errorf("expected '_0 int' field in output, got:\n%s", output)
	}
	if !strings.Contains(output, "_1 byte") {
		t.Errorf("expected '_1 byte' field in output, got:\n%s", output)
	}
}

func TestVariantProtocolType(t *testing.T) {
	input := `PROTOCOL MSG
  CASE
    text; INT
    quit
`
	output := transpile(t, input)

	if !strings.Contains(output, "type _proto_MSG interface {") {
		t.Errorf("expected interface declaration in output, got:\n%s", output)
	}
	if !strings.Contains(output, "_is_MSG()") {
		t.Errorf("expected marker method in output, got:\n%s", output)
	}
	if !strings.Contains(output, "type _proto_MSG_text struct {") {
		t.Errorf("expected text struct in output, got:\n%s", output)
	}
	if !strings.Contains(output, "type _proto_MSG_quit struct{}") {
		t.Errorf("expected quit struct in output, got:\n%s", output)
	}
}

func TestRecordType(t *testing.T) {
	input := `RECORD POINT
  INT x:
  INT y:
`
	output := transpile(t, input)

	if !strings.Contains(output, "type POINT struct {") {
		t.Errorf("expected 'type POINT struct {' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "x int") {
		t.Errorf("expected 'x int' field in output, got:\n%s", output)
	}
	if !strings.Contains(output, "y int") {
		t.Errorf("expected 'y int' field in output, got:\n%s", output)
	}
}

func TestRecordFieldAssignmentCodegen(t *testing.T) {
	input := `RECORD POINT
  INT x:
  INT y:
SEQ
  POINT p:
  p[x] := 5
`
	output := transpile(t, input)

	if !strings.Contains(output, "p.x = 5") {
		t.Errorf("expected 'p.x = 5' in output, got:\n%s", output)
	}
}

func TestChanArrayDeclGen(t *testing.T) {
	input := `[5]CHAN OF INT cs:
`
	output := transpile(t, input)

	if !strings.Contains(output, "cs := make([]chan int, 5)") {
		t.Errorf("expected 'cs := make([]chan int, 5)' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "for _i := range cs { cs[_i] = make(chan int) }") {
		t.Errorf("expected init loop in output, got:\n%s", output)
	}
}

func TestIndexedSendGen(t *testing.T) {
	input := `cs[0] ! 42
`
	output := transpile(t, input)

	if !strings.Contains(output, "cs[0] <- 42") {
		t.Errorf("expected 'cs[0] <- 42' in output, got:\n%s", output)
	}
}

func TestIndexedReceiveGen(t *testing.T) {
	input := `cs[0] ? x
`
	output := transpile(t, input)

	if !strings.Contains(output, "x = <-cs[0]") {
		t.Errorf("expected 'x = <-cs[0]' in output, got:\n%s", output)
	}
}

func TestChanArrayParamGen(t *testing.T) {
	input := `PROC worker([]CHAN OF INT cs)
  SKIP
`
	output := transpile(t, input)

	if !strings.Contains(output, "func worker(cs []chan int)") {
		t.Errorf("expected 'func worker(cs []chan int)' in output, got:\n%s", output)
	}
}

func TestChanDirParamGen(t *testing.T) {
	input := `PROC worker(CHAN OF INT input?, CHAN OF INT output!)
  SEQ
    INT x:
    input ? x
    output ! x
`
	output := transpile(t, input)

	if !strings.Contains(output, "func worker(input <-chan int, output chan<- int)") {
		t.Errorf("expected directed channel types in output, got:\n%s", output)
	}
}

func TestChanArrayDirParamGen(t *testing.T) {
	input := `PROC worker([]CHAN OF INT cs?, []CHAN OF INT out!)
  SKIP
`
	output := transpile(t, input)

	if !strings.Contains(output, "cs []<-chan int") {
		t.Errorf("expected '[]<-chan int' for input chan array, got:\n%s", output)
	}
	if !strings.Contains(output, "out []chan<- int") {
		t.Errorf("expected '[]chan<- int' for output chan array, got:\n%s", output)
	}
}

func TestRecordFieldAccessCodegen(t *testing.T) {
	input := `RECORD POINT
  INT x:
  INT y:
SEQ
  POINT p:
  INT v:
  v := p[x]
`
	output := transpile(t, input)

	if !strings.Contains(output, "v = p.x") {
		t.Errorf("expected 'v = p.x' in output, got:\n%s", output)
	}
}

func TestSizeOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x := SIZE arr\n", "x = len(arr)"},
		{"x := SIZE arr + 1\n", "x = (len(arr) + 1)"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q: expected %q in output, got:\n%s", tt.input, tt.expected, output)
		}
	}
}

func TestOpenArrayParamGen(t *testing.T) {
	input := `PROC worker(VAL []INT arr, []BYTE data)
  SKIP
`
	output := transpile(t, input)

	if !strings.Contains(output, "func worker(arr []int, data []byte)") {
		t.Errorf("expected 'func worker(arr []int, data []byte)' in output, got:\n%s", output)
	}
}

func TestAbbreviation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"VAL INT x IS 42:\n", "x := 42"},
		{"VAL BOOL flag IS TRUE:\n", "flag := true"},
		{"INT y IS z:\n", "y := z"},
		{"INITIAL INT x IS 42:\n", "x := 42"},
		{"INITIAL BOOL done IS FALSE:\n", "done := false"},
	}

	for _, tt := range tests {
		output := transpile(t, tt.input)
		if !strings.Contains(output, tt.expected) {
			t.Errorf("for input %q: expected %q in output, got:\n%s", tt.input, tt.expected, output)
		}
	}
}

func TestMultiAssignmentSimple(t *testing.T) {
	input := `a, b := 1, 2
`
	output := transpile(t, input)
	if !strings.Contains(output, "a, b = 1, 2") {
		t.Errorf("expected 'a, b = 1, 2' in output, got:\n%s", output)
	}
}

func TestMultiAssignmentIndexed(t *testing.T) {
	input := `x[0], x[1] := x[1], x[0]
`
	output := transpile(t, input)
	if !strings.Contains(output, "x[0], x[1] = x[1], x[0]") {
		t.Errorf("expected 'x[0], x[1] = x[1], x[0]' in output, got:\n%s", output)
	}
}

func TestMultiAssignmentMixed(t *testing.T) {
	input := `a, x[0] := 1, 2
`
	output := transpile(t, input)
	if !strings.Contains(output, "a, x[0] = 1, 2") {
		t.Errorf("expected 'a, x[0] = 1, 2' in output, got:\n%s", output)
	}
}
