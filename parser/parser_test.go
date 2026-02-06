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
