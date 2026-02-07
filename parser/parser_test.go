package parser

import (
	"testing"

	"github.com/codeassociates/occam2go/ast"
	"github.com/codeassociates/occam2go/lexer"
)

func TestVarDecl(t *testing.T) {
	input := `INT x:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	decl, ok := program.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected VarDecl, got %T", program.Statements[0])
	}

	if decl.Type != "INT" {
		t.Errorf("expected type INT, got %s", decl.Type)
	}

	if len(decl.Names) != 1 || decl.Names[0] != "x" {
		t.Errorf("expected name 'x', got %v", decl.Names)
	}
}

func TestMultipleVarDecl(t *testing.T) {
	input := `INT x, y, z:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	decl, ok := program.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected VarDecl, got %T", program.Statements[0])
	}

	expected := []string{"x", "y", "z"}
	if len(decl.Names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(decl.Names))
	}
	for i, name := range expected {
		if decl.Names[i] != name {
			t.Errorf("expected name %s at position %d, got %s", name, i, decl.Names[i])
		}
	}
}

func TestAssignment(t *testing.T) {
	input := `x := 5
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	if assign.Name != "x" {
		t.Errorf("expected name 'x', got %s", assign.Name)
	}

	intLit, ok := assign.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral, got %T", assign.Value)
	}

	if intLit.Value != 5 {
		t.Errorf("expected value 5, got %d", intLit.Value)
	}
}

func TestBinaryExpression(t *testing.T) {
	input := `x := a + b * c
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	// Should be: a + (b * c) due to precedence
	binExpr, ok := assign.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}

	if binExpr.Operator != "+" {
		t.Errorf("expected +, got %s", binExpr.Operator)
	}

	// Right side should be b * c
	rightBin, ok := binExpr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right to be BinaryExpr, got %T", binExpr.Right)
	}

	if rightBin.Operator != "*" {
		t.Errorf("expected *, got %s", rightBin.Operator)
	}
}

func TestSeqBlock(t *testing.T) {
	input := `SEQ
  INT x:
  x := 10
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	seq, ok := program.Statements[0].(*ast.SeqBlock)
	if !ok {
		t.Fatalf("expected SeqBlock, got %T", program.Statements[0])
	}

	if len(seq.Statements) != 2 {
		t.Fatalf("expected 2 statements in SEQ, got %d", len(seq.Statements))
	}

	_, ok = seq.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Errorf("expected first statement to be VarDecl, got %T", seq.Statements[0])
	}

	_, ok = seq.Statements[1].(*ast.Assignment)
	if !ok {
		t.Errorf("expected second statement to be Assignment, got %T", seq.Statements[1])
	}
}

func TestParBlock(t *testing.T) {
	input := `PAR
  x := 1
  y := 2
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	par, ok := program.Statements[0].(*ast.ParBlock)
	if !ok {
		t.Fatalf("expected ParBlock, got %T", program.Statements[0])
	}

	if len(par.Statements) != 2 {
		t.Fatalf("expected 2 statements in PAR, got %d", len(par.Statements))
	}
}

func TestChanDecl(t *testing.T) {
	input := `CHAN OF INT c:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	decl, ok := program.Statements[0].(*ast.ChanDecl)
	if !ok {
		t.Fatalf("expected ChanDecl, got %T", program.Statements[0])
	}

	if decl.ElemType != "INT" {
		t.Errorf("expected element type INT, got %s", decl.ElemType)
	}

	if len(decl.Names) != 1 || decl.Names[0] != "c" {
		t.Errorf("expected name 'c', got %v", decl.Names)
	}
}

func TestSend(t *testing.T) {
	input := `c ! 42
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	send, ok := program.Statements[0].(*ast.Send)
	if !ok {
		t.Fatalf("expected Send, got %T", program.Statements[0])
	}

	if send.Channel != "c" {
		t.Errorf("expected channel 'c', got %s", send.Channel)
	}

	intLit, ok := send.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral, got %T", send.Value)
	}

	if intLit.Value != 42 {
		t.Errorf("expected value 42, got %d", intLit.Value)
	}
}

func TestReceive(t *testing.T) {
	input := `c ? x
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	recv, ok := program.Statements[0].(*ast.Receive)
	if !ok {
		t.Fatalf("expected Receive, got %T", program.Statements[0])
	}

	if recv.Channel != "c" {
		t.Errorf("expected channel 'c', got %s", recv.Channel)
	}

	if recv.Variable != "x" {
		t.Errorf("expected variable 'x', got %s", recv.Variable)
	}
}

func TestAltBlock(t *testing.T) {
	input := `ALT
  c1 ? x
    SKIP
  c2 ? y
    SKIP
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	alt, ok := program.Statements[0].(*ast.AltBlock)
	if !ok {
		t.Fatalf("expected AltBlock, got %T", program.Statements[0])
	}

	if len(alt.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(alt.Cases))
	}

	if alt.Cases[0].Channel != "c1" {
		t.Errorf("expected channel 'c1', got %s", alt.Cases[0].Channel)
	}

	if alt.Cases[0].Variable != "x" {
		t.Errorf("expected variable 'x', got %s", alt.Cases[0].Variable)
	}

	if alt.Cases[1].Channel != "c2" {
		t.Errorf("expected channel 'c2', got %s", alt.Cases[1].Channel)
	}

	if alt.Cases[1].Variable != "y" {
		t.Errorf("expected variable 'y', got %s", alt.Cases[1].Variable)
	}
}

func TestAltBlockWithGuard(t *testing.T) {
	input := `ALT
  TRUE & c1 ? x
    SKIP
  FALSE & c2 ? y
    SKIP
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	alt, ok := program.Statements[0].(*ast.AltBlock)
	if !ok {
		t.Fatalf("expected AltBlock, got %T", program.Statements[0])
	}

	if len(alt.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(alt.Cases))
	}

	// First case should have TRUE guard
	if alt.Cases[0].Guard == nil {
		t.Error("expected guard on first case")
	}

	// Second case should have FALSE guard
	if alt.Cases[1].Guard == nil {
		t.Error("expected guard on second case")
	}
}

func TestWhileLoop(t *testing.T) {
	input := `WHILE x > 0
  x := x - 1
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	loop, ok := program.Statements[0].(*ast.WhileLoop)
	if !ok {
		t.Fatalf("expected WhileLoop, got %T", program.Statements[0])
	}

	if loop.Condition == nil {
		t.Error("expected condition")
	}

	if loop.Body == nil {
		t.Error("expected body")
	}
}

func TestIfStatement(t *testing.T) {
	input := `IF
  x > 0
    y := 1
  x = 0
    y := 0
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	ifStmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", program.Statements[0])
	}

	if len(ifStmt.Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(ifStmt.Choices))
	}

	if ifStmt.Choices[0].Condition == nil {
		t.Error("expected condition on first choice")
	}

	if ifStmt.Choices[0].Body == nil {
		t.Error("expected body on first choice")
	}

	if ifStmt.Choices[1].Condition == nil {
		t.Error("expected condition on second choice")
	}

	if ifStmt.Choices[1].Body == nil {
		t.Error("expected body on second choice")
	}
}

func TestReplicatedSeq(t *testing.T) {
	input := `SEQ i = 0 FOR 5
  print.int(i)
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	seq, ok := program.Statements[0].(*ast.SeqBlock)
	if !ok {
		t.Fatalf("expected SeqBlock, got %T", program.Statements[0])
	}

	if seq.Replicator == nil {
		t.Fatal("expected replicator on SEQ block")
	}

	if seq.Replicator.Variable != "i" {
		t.Errorf("expected variable 'i', got %s", seq.Replicator.Variable)
	}

	startLit, ok := seq.Replicator.Start.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for start, got %T", seq.Replicator.Start)
	}
	if startLit.Value != 0 {
		t.Errorf("expected start 0, got %d", startLit.Value)
	}

	countLit, ok := seq.Replicator.Count.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for count, got %T", seq.Replicator.Count)
	}
	if countLit.Value != 5 {
		t.Errorf("expected count 5, got %d", countLit.Value)
	}
}

func TestReplicatedPar(t *testing.T) {
	input := `PAR i = 0 FOR 3
  SKIP
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	par, ok := program.Statements[0].(*ast.ParBlock)
	if !ok {
		t.Fatalf("expected ParBlock, got %T", program.Statements[0])
	}

	if par.Replicator == nil {
		t.Fatal("expected replicator on PAR block")
	}

	if par.Replicator.Variable != "i" {
		t.Errorf("expected variable 'i', got %s", par.Replicator.Variable)
	}
}

func TestArrayDecl(t *testing.T) {
	input := `[5]INT arr:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	decl, ok := program.Statements[0].(*ast.ArrayDecl)
	if !ok {
		t.Fatalf("expected ArrayDecl, got %T", program.Statements[0])
	}

	if decl.Type != "INT" {
		t.Errorf("expected type INT, got %s", decl.Type)
	}

	sizeLit, ok := decl.Size.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for size, got %T", decl.Size)
	}
	if sizeLit.Value != 5 {
		t.Errorf("expected size 5, got %d", sizeLit.Value)
	}

	if len(decl.Names) != 1 || decl.Names[0] != "arr" {
		t.Errorf("expected name 'arr', got %v", decl.Names)
	}
}

func TestArrayDeclMultipleNames(t *testing.T) {
	input := `[10]INT a, b:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	decl, ok := program.Statements[0].(*ast.ArrayDecl)
	if !ok {
		t.Fatalf("expected ArrayDecl, got %T", program.Statements[0])
	}

	if decl.Type != "INT" {
		t.Errorf("expected type INT, got %s", decl.Type)
	}

	expected := []string{"a", "b"}
	if len(decl.Names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(decl.Names))
	}
	for i, name := range expected {
		if decl.Names[i] != name {
			t.Errorf("expected name %s at position %d, got %s", name, i, decl.Names[i])
		}
	}
}

func TestIndexedAssignment(t *testing.T) {
	input := `arr[2] := 10
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	if assign.Name != "arr" {
		t.Errorf("expected name 'arr', got %s", assign.Name)
	}

	if assign.Index == nil {
		t.Fatal("expected index expression, got nil")
	}

	indexLit, ok := assign.Index.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for index, got %T", assign.Index)
	}
	if indexLit.Value != 2 {
		t.Errorf("expected index 2, got %d", indexLit.Value)
	}

	valLit, ok := assign.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for value, got %T", assign.Value)
	}
	if valLit.Value != 10 {
		t.Errorf("expected value 10, got %d", valLit.Value)
	}
}

func TestIndexExpression(t *testing.T) {
	input := `x := arr[0] + 1
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	// Should be: arr[0] + 1 -> BinaryExpr with IndexExpr on left
	binExpr, ok := assign.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}

	if binExpr.Operator != "+" {
		t.Errorf("expected +, got %s", binExpr.Operator)
	}

	indexExpr, ok := binExpr.Left.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr on left, got %T", binExpr.Left)
	}

	ident, ok := indexExpr.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier in IndexExpr, got %T", indexExpr.Left)
	}
	if ident.Value != "arr" {
		t.Errorf("expected 'arr', got %s", ident.Value)
	}

	idxLit, ok := indexExpr.Index.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for index, got %T", indexExpr.Index)
	}
	if idxLit.Value != 0 {
		t.Errorf("expected index 0, got %d", idxLit.Value)
	}
}

func TestFuncDeclIS(t *testing.T) {
	input := `INT FUNCTION square(VAL INT x)
  IS x * x
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	fn, ok := program.Statements[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Statements[0])
	}

	if fn.ReturnType != "INT" {
		t.Errorf("expected return type INT, got %s", fn.ReturnType)
	}

	if fn.Name != "square" {
		t.Errorf("expected name 'square', got %s", fn.Name)
	}

	if len(fn.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(fn.Params))
	}

	if fn.Params[0].Name != "x" || fn.Params[0].Type != "INT" || !fn.Params[0].IsVal {
		t.Errorf("expected VAL INT x, got %+v", fn.Params[0])
	}

	if fn.ResultExpr == nil {
		t.Fatal("expected result expression, got nil")
	}

	if len(fn.Body) != 0 {
		t.Errorf("expected empty body for IS form, got %d statements", len(fn.Body))
	}
}

func TestFuncDeclValof(t *testing.T) {
	input := `INT FUNCTION max(VAL INT a, VAL INT b)
  INT result:
  VALOF
    IF
      a > b
        result := a
      TRUE
        result := b
    RESULT result
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	fn, ok := program.Statements[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected FuncDecl, got %T", program.Statements[0])
	}

	if fn.ReturnType != "INT" {
		t.Errorf("expected return type INT, got %s", fn.ReturnType)
	}

	if fn.Name != "max" {
		t.Errorf("expected name 'max', got %s", fn.Name)
	}

	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}

	if fn.ResultExpr == nil {
		t.Fatal("expected result expression, got nil")
	}

	// Body should contain local var decl and the IF statement
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 statement in body")
	}
}

func TestCaseStatement(t *testing.T) {
	input := `CASE x
  1
    y := 10
  2
    y := 20
  ELSE
    y := 0
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	caseStmt, ok := program.Statements[0].(*ast.CaseStatement)
	if !ok {
		t.Fatalf("expected CaseStatement, got %T", program.Statements[0])
	}

	if caseStmt.Selector == nil {
		t.Fatal("expected selector expression")
	}

	ident, ok := caseStmt.Selector.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier selector, got %T", caseStmt.Selector)
	}
	if ident.Value != "x" {
		t.Errorf("expected selector 'x', got %s", ident.Value)
	}

	if len(caseStmt.Choices) != 3 {
		t.Fatalf("expected 3 choices, got %d", len(caseStmt.Choices))
	}

	// First choice: value 1
	if caseStmt.Choices[0].IsElse {
		t.Error("first choice should not be ELSE")
	}
	if len(caseStmt.Choices[0].Values) != 1 {
		t.Fatalf("expected 1 value in first choice, got %d", len(caseStmt.Choices[0].Values))
	}
	if caseStmt.Choices[0].Body == nil {
		t.Error("expected body on first choice")
	}

	// Second choice: value 2
	if caseStmt.Choices[1].IsElse {
		t.Error("second choice should not be ELSE")
	}
	if caseStmt.Choices[1].Body == nil {
		t.Error("expected body on second choice")
	}

	// Third choice: ELSE
	if !caseStmt.Choices[2].IsElse {
		t.Error("third choice should be ELSE")
	}
	if caseStmt.Choices[2].Body == nil {
		t.Error("expected body on ELSE choice")
	}
}

func TestTimerDecl(t *testing.T) {
	input := `TIMER tim:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	decl, ok := program.Statements[0].(*ast.TimerDecl)
	if !ok {
		t.Fatalf("expected TimerDecl, got %T", program.Statements[0])
	}

	if len(decl.Names) != 1 || decl.Names[0] != "tim" {
		t.Errorf("expected name 'tim', got %v", decl.Names)
	}
}

func TestTimerRead(t *testing.T) {
	input := `SEQ
  TIMER tim:
  INT t:
  tim ? t
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	seq, ok := program.Statements[0].(*ast.SeqBlock)
	if !ok {
		t.Fatalf("expected SeqBlock, got %T", program.Statements[0])
	}

	if len(seq.Statements) != 3 {
		t.Fatalf("expected 3 statements in SEQ, got %d", len(seq.Statements))
	}

	_, ok = seq.Statements[0].(*ast.TimerDecl)
	if !ok {
		t.Errorf("expected TimerDecl, got %T", seq.Statements[0])
	}

	tr, ok := seq.Statements[2].(*ast.TimerRead)
	if !ok {
		t.Fatalf("expected TimerRead, got %T", seq.Statements[2])
	}

	if tr.Timer != "tim" {
		t.Errorf("expected timer 'tim', got %s", tr.Timer)
	}

	if tr.Variable != "t" {
		t.Errorf("expected variable 't', got %s", tr.Variable)
	}
}

func TestAfterExpression(t *testing.T) {
	input := `x := t2 AFTER t1
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	binExpr, ok := assign.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}

	if binExpr.Operator != "AFTER" {
		t.Errorf("expected operator 'AFTER', got %s", binExpr.Operator)
	}

	left, ok := binExpr.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier on left, got %T", binExpr.Left)
	}
	if left.Value != "t2" {
		t.Errorf("expected 't2', got %s", left.Value)
	}

	right, ok := binExpr.Right.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier on right, got %T", binExpr.Right)
	}
	if right.Value != "t1" {
		t.Errorf("expected 't1', got %s", right.Value)
	}
}

func TestChanParam(t *testing.T) {
	input := `PROC worker(CHAN OF INT input)
  SEQ
    INT x:
    input ? x
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	proc, ok := program.Statements[0].(*ast.ProcDecl)
	if !ok {
		t.Fatalf("expected ProcDecl, got %T", program.Statements[0])
	}

	if len(proc.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(proc.Params))
	}

	param := proc.Params[0]
	if !param.IsChan {
		t.Errorf("expected IsChan=true")
	}
	if param.ChanElemType != "INT" {
		t.Errorf("expected ChanElemType=INT, got %s", param.ChanElemType)
	}
	if param.Name != "input" {
		t.Errorf("expected Name=input, got %s", param.Name)
	}
}

func TestChanParamMixed(t *testing.T) {
	input := `PROC foo(CHAN OF INT c, VAL INT x)
  SEQ
    c ! x
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	proc, ok := program.Statements[0].(*ast.ProcDecl)
	if !ok {
		t.Fatalf("expected ProcDecl, got %T", program.Statements[0])
	}

	if len(proc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(proc.Params))
	}

	// First param: CHAN OF INT c
	p0 := proc.Params[0]
	if !p0.IsChan {
		t.Errorf("param 0: expected IsChan=true")
	}
	if p0.ChanElemType != "INT" {
		t.Errorf("param 0: expected ChanElemType=INT, got %s", p0.ChanElemType)
	}
	if p0.Name != "c" {
		t.Errorf("param 0: expected Name=c, got %s", p0.Name)
	}

	// Second param: VAL INT x
	p1 := proc.Params[1]
	if p1.IsChan {
		t.Errorf("param 1: expected IsChan=false")
	}
	if !p1.IsVal {
		t.Errorf("param 1: expected IsVal=true")
	}
	if p1.Type != "INT" {
		t.Errorf("param 1: expected Type=INT, got %s", p1.Type)
	}
	if p1.Name != "x" {
		t.Errorf("param 1: expected Name=x, got %s", p1.Name)
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors:", len(errors))
	for _, msg := range errors {
		t.Errorf("  parser error: %s", msg)
	}
	t.FailNow()
}
