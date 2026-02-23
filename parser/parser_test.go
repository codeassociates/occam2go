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

func TestAltBlockWithGuardedSkip(t *testing.T) {
	input := `ALT
  ready & SKIP
    some.proc()
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

	if len(alt.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(alt.Cases))
	}

	c := alt.Cases[0]
	if !c.IsSkip {
		t.Error("expected IsSkip to be true")
	}
	if c.Guard == nil {
		t.Error("expected guard expression, got nil")
	}
}

func TestPriAltBlock(t *testing.T) {
	input := `PRI ALT
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

	if !alt.Priority {
		t.Error("expected Priority to be true for PRI ALT")
	}

	if len(alt.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(alt.Cases))
	}

	if alt.Cases[0].Channel != "c1" {
		t.Errorf("expected channel 'c1', got %s", alt.Cases[0].Channel)
	}

	if alt.Cases[1].Channel != "c2" {
		t.Errorf("expected channel 'c2', got %s", alt.Cases[1].Channel)
	}
}

func TestPriParBlock(t *testing.T) {
	input := `PRI PAR
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

	if !par.Priority {
		t.Error("expected Priority to be true for PRI PAR")
	}

	if len(par.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(par.Statements))
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

	if len(loop.Body) == 0 {
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

	if len(ifStmt.Choices[0].Body) == 0 {
		t.Error("expected body on first choice")
	}

	if ifStmt.Choices[1].Condition == nil {
		t.Error("expected condition on second choice")
	}

	if len(ifStmt.Choices[1].Body) == 0 {
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

func TestReplicatedSeqWithStep(t *testing.T) {
	input := `SEQ i = 0 FOR 5 STEP 2
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

	countLit2, ok := seq.Replicator.Count.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for count, got %T", seq.Replicator.Count)
	}
	if countLit2.Value != 5 {
		t.Errorf("expected count 5, got %d", countLit2.Value)
	}

	if seq.Replicator.Step == nil {
		t.Fatal("expected step on replicator")
	}
	stepLit, ok := seq.Replicator.Step.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for step, got %T", seq.Replicator.Step)
	}
	if stepLit.Value != 2 {
		t.Errorf("expected step 2, got %d", stepLit.Value)
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

func TestReplicatedIf(t *testing.T) {
	input := `IF i = 0 FOR 5
  i = 3
    SKIP
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

	if ifStmt.Replicator == nil {
		t.Fatal("expected replicator on IF statement")
	}

	if ifStmt.Replicator.Variable != "i" {
		t.Errorf("expected variable 'i', got %s", ifStmt.Replicator.Variable)
	}

	if len(ifStmt.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(ifStmt.Choices))
	}

	if ifStmt.Choices[0].Condition == nil {
		t.Error("expected condition on choice")
	}

	if len(ifStmt.Choices[0].Body) == 0 {
		t.Error("expected body on choice")
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

	if len(decl.Sizes) != 1 {
		t.Fatalf("expected 1 size dimension, got %d", len(decl.Sizes))
	}
	sizeLit, ok := decl.Sizes[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for size, got %T", decl.Sizes[0])
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

	if len(assign.Indices) != 1 {
		t.Fatalf("expected 1 index, got %d", len(assign.Indices))
	}

	indexLit, ok := assign.Indices[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for index, got %T", assign.Indices[0])
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

	if len(fn.ReturnTypes) != 1 || fn.ReturnTypes[0] != "INT" {
		t.Errorf("expected return types [INT], got %v", fn.ReturnTypes)
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

	if len(fn.ResultExprs) != 1 {
		t.Fatalf("expected 1 result expression, got %d", len(fn.ResultExprs))
	}

	if len(fn.Body) != 0 {
		t.Errorf("expected empty body for IS form, got %d statements", len(fn.Body))
	}
}

func TestInlineFuncDecl(t *testing.T) {
	input := `INT INLINE FUNCTION seconds(VAL INT s)
  INT ticks:
  VALOF
    ticks := s * 1000000
    RESULT ticks
:
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

	if len(fn.ReturnTypes) != 1 || fn.ReturnTypes[0] != "INT" {
		t.Errorf("expected return types [INT], got %v", fn.ReturnTypes)
	}

	if fn.Name != "seconds" {
		t.Errorf("expected name 'seconds', got %s", fn.Name)
	}

	if len(fn.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(fn.Params))
	}

	if fn.Params[0].Name != "s" || fn.Params[0].Type != "INT" || !fn.Params[0].IsVal {
		t.Errorf("expected VAL INT s, got %+v", fn.Params[0])
	}
}

func TestInlineFuncDeclIS(t *testing.T) {
	input := `INT INLINE FUNCTION double(VAL INT x)
  IS x * 2
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

	if fn.Name != "double" {
		t.Errorf("expected name 'double', got %s", fn.Name)
	}

	if len(fn.ResultExprs) != 1 {
		t.Fatalf("expected 1 result expression, got %d", len(fn.ResultExprs))
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

	if len(fn.ReturnTypes) != 1 || fn.ReturnTypes[0] != "INT" {
		t.Errorf("expected return types [INT], got %v", fn.ReturnTypes)
	}

	if fn.Name != "max" {
		t.Errorf("expected name 'max', got %s", fn.Name)
	}

	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}

	if len(fn.ResultExprs) != 1 {
		t.Fatalf("expected 1 result expression, got %d", len(fn.ResultExprs))
	}

	// Body should contain local var decl and the IF statement
	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 statement in body")
	}
}

func TestMultiResultFuncDecl(t *testing.T) {
	input := `INT, INT FUNCTION swap(VAL INT a, VAL INT b)
  INT x, y:
  VALOF
    SEQ
      x := b
      y := a
    RESULT x, y
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

	if len(fn.ReturnTypes) != 2 || fn.ReturnTypes[0] != "INT" || fn.ReturnTypes[1] != "INT" {
		t.Errorf("expected return types [INT, INT], got %v", fn.ReturnTypes)
	}

	if fn.Name != "swap" {
		t.Errorf("expected name 'swap', got %s", fn.Name)
	}

	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}

	if len(fn.ResultExprs) != 2 {
		t.Fatalf("expected 2 result expressions, got %d", len(fn.ResultExprs))
	}

	if len(fn.Body) < 1 {
		t.Fatal("expected at least 1 statement in body")
	}
}

func TestMultiAssignment(t *testing.T) {
	input := `a, b := swap(1, 2)
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	ma, ok := program.Statements[0].(*ast.MultiAssignment)
	if !ok {
		t.Fatalf("expected MultiAssignment, got %T", program.Statements[0])
	}

	if len(ma.Targets) != 2 || ma.Targets[0].Name != "a" || ma.Targets[1].Name != "b" {
		t.Errorf("expected targets [a, b], got %v", ma.Targets)
	}
	if len(ma.Targets[0].Indices) != 0 || len(ma.Targets[1].Indices) != 0 {
		t.Errorf("expected no index on targets")
	}

	if len(ma.Values) != 1 {
		t.Fatalf("expected 1 value expression, got %d", len(ma.Values))
	}

	fc, ok := ma.Values[0].(*ast.FuncCall)
	if !ok {
		t.Fatalf("expected FuncCall value, got %T", ma.Values[0])
	}

	if fc.Name != "swap" {
		t.Errorf("expected function name 'swap', got %s", fc.Name)
	}
}

func TestMultiAssignmentIndexed(t *testing.T) {
	input := `x[0], x[1] := x[1], x[0]
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	ma, ok := program.Statements[0].(*ast.MultiAssignment)
	if !ok {
		t.Fatalf("expected MultiAssignment, got %T", program.Statements[0])
	}

	if len(ma.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(ma.Targets))
	}
	if ma.Targets[0].Name != "x" || ma.Targets[1].Name != "x" {
		t.Errorf("expected target names [x, x], got [%s, %s]", ma.Targets[0].Name, ma.Targets[1].Name)
	}
	if len(ma.Targets[0].Indices) == 0 || len(ma.Targets[1].Indices) == 0 {
		t.Fatalf("expected indexed targets")
	}

	if len(ma.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(ma.Values))
	}
}

func TestMultiAssignmentMixed(t *testing.T) {
	input := `a, x[i] := 1, 2
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	ma, ok := program.Statements[0].(*ast.MultiAssignment)
	if !ok {
		t.Fatalf("expected MultiAssignment, got %T", program.Statements[0])
	}

	if len(ma.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(ma.Targets))
	}
	if ma.Targets[0].Name != "a" || len(ma.Targets[0].Indices) != 0 {
		t.Errorf("expected simple target 'a', got %v", ma.Targets[0])
	}
	if ma.Targets[1].Name != "x" || len(ma.Targets[1].Indices) == 0 {
		t.Errorf("expected indexed target 'x[i]', got %v", ma.Targets[1])
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
	if len(caseStmt.Choices[0].Body) == 0 {
		t.Error("expected body on first choice")
	}

	// Second choice: value 2
	if caseStmt.Choices[1].IsElse {
		t.Error("second choice should not be ELSE")
	}
	if len(caseStmt.Choices[1].Body) == 0 {
		t.Error("expected body on second choice")
	}

	// Third choice: ELSE
	if !caseStmt.Choices[2].IsElse {
		t.Error("third choice should be ELSE")
	}
	if len(caseStmt.Choices[2].Body) == 0 {
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

func TestTypeConversion(t *testing.T) {
	input := `x := INT y
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

	tc, ok := assign.Value.(*ast.TypeConversion)
	if !ok {
		t.Fatalf("expected TypeConversion, got %T", assign.Value)
	}

	if tc.TargetType != "INT" {
		t.Errorf("expected TargetType INT, got %s", tc.TargetType)
	}

	ident, ok := tc.Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier inside TypeConversion, got %T", tc.Expr)
	}
	if ident.Value != "y" {
		t.Errorf("expected 'y', got %s", ident.Value)
	}
}

func TestTypeConversionWithQualifier(t *testing.T) {
	tests := []struct {
		input     string
		target    string
		qualifier string
	}{
		{"x := INT ROUND y\n", "INT", "ROUND"},
		{"x := INT TRUNC y\n", "INT", "TRUNC"},
		{"x := INT64 ROUND y\n", "INT64", "ROUND"},
		{"x := REAL32 TRUNC y\n", "REAL32", "TRUNC"},
		{"x := INT y\n", "INT", ""},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("input %q: expected 1 statement, got %d", tt.input, len(program.Statements))
		}

		assign, ok := program.Statements[0].(*ast.Assignment)
		if !ok {
			t.Fatalf("input %q: expected Assignment, got %T", tt.input, program.Statements[0])
		}

		tc, ok := assign.Value.(*ast.TypeConversion)
		if !ok {
			t.Fatalf("input %q: expected TypeConversion, got %T", tt.input, assign.Value)
		}

		if tc.TargetType != tt.target {
			t.Errorf("input %q: expected TargetType %s, got %s", tt.input, tt.target, tc.TargetType)
		}

		if tc.Qualifier != tt.qualifier {
			t.Errorf("input %q: expected Qualifier %q, got %q", tt.input, tt.qualifier, tc.Qualifier)
		}
	}
}

func TestTypeConversionInExpression(t *testing.T) {
	input := `x := INT y + 1
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	// Should be: (INT y) + 1 due to PREFIX precedence
	binExpr, ok := assign.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}

	if binExpr.Operator != "+" {
		t.Errorf("expected +, got %s", binExpr.Operator)
	}

	tc, ok := binExpr.Left.(*ast.TypeConversion)
	if !ok {
		t.Fatalf("expected TypeConversion on left, got %T", binExpr.Left)
	}

	if tc.TargetType != "INT" {
		t.Errorf("expected TargetType INT, got %s", tc.TargetType)
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

func TestStringLiteral(t *testing.T) {
	input := `x := "hello world"
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
		t.Errorf("expected Name=x, got %s", assign.Name)
	}

	strLit, ok := assign.Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", assign.Value)
	}

	if strLit.Value != "hello world" {
		t.Errorf("expected Value='hello world', got '%s'", strLit.Value)
	}
}

func TestStringEscapeConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`x := "hello*n"` + "\n", "hello\n"},
		{`x := "hello*c*n"` + "\n", "hello\r\n"},
		{`x := "*t*s"` + "\n", "\t "},
		{`x := "a**b"` + "\n", "a*b"},
		{`x := "it*'s"` + "\n", "it's"},
		{`x := "no escapes"` + "\n", "no escapes"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("input %q: expected 1 statement, got %d", tt.input, len(program.Statements))
		}

		assign, ok := program.Statements[0].(*ast.Assignment)
		if !ok {
			t.Fatalf("input %q: expected Assignment, got %T", tt.input, program.Statements[0])
		}

		strLit, ok := assign.Value.(*ast.StringLiteral)
		if !ok {
			t.Fatalf("input %q: expected StringLiteral, got %T", tt.input, assign.Value)
		}

		if strLit.Value != tt.expected {
			t.Errorf("input %q: expected Value=%q, got %q", tt.input, tt.expected, strLit.Value)
		}
	}
}

func TestByteLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected byte
	}{
		{"x := 'A'\n", 'A'},
		{"x := '0'\n", '0'},
		{"x := ' '\n", ' '},
		{"x := '*n'\n", '\n'},
		{"x := '*c'\n", '\r'},
		{"x := '*t'\n", '\t'},
		{"x := '*s'\n", ' '},
		{"x := '**'\n", '*'},
		{"x := '*''\n", '\''},
		{"x := '*\"'\n", '"'},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("input %q: expected 1 statement, got %d", tt.input, len(program.Statements))
		}

		assign, ok := program.Statements[0].(*ast.Assignment)
		if !ok {
			t.Fatalf("input %q: expected Assignment, got %T", tt.input, program.Statements[0])
		}

		byteLit, ok := assign.Value.(*ast.ByteLiteral)
		if !ok {
			t.Fatalf("input %q: expected ByteLiteral, got %T", tt.input, assign.Value)
		}

		if byteLit.Value != tt.expected {
			t.Errorf("input %q: expected Value=%d (%c), got %d (%c)", tt.input, tt.expected, tt.expected, byteLit.Value, byteLit.Value)
		}
	}
}

func TestStringLiteralInProcCall(t *testing.T) {
	input := `print.string("hello")
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	call, ok := program.Statements[0].(*ast.ProcCall)
	if !ok {
		t.Fatalf("expected ProcCall, got %T", program.Statements[0])
	}

	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}

	strLit, ok := call.Args[0].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral arg, got %T", call.Args[0])
	}

	if strLit.Value != "hello" {
		t.Errorf("expected Value='hello', got '%s'", strLit.Value)
	}
}

func TestSimpleProtocolDecl(t *testing.T) {
	input := `PROTOCOL SIGNAL IS INT
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	proto, ok := program.Statements[0].(*ast.ProtocolDecl)
	if !ok {
		t.Fatalf("expected ProtocolDecl, got %T", program.Statements[0])
	}

	if proto.Name != "SIGNAL" {
		t.Errorf("expected name 'SIGNAL', got %s", proto.Name)
	}

	if proto.Kind != "simple" {
		t.Errorf("expected kind 'simple', got %s", proto.Kind)
	}

	if len(proto.Types) != 1 || proto.Types[0] != "INT" {
		t.Errorf("expected types [INT], got %v", proto.Types)
	}
}

func TestSequentialProtocolDecl(t *testing.T) {
	input := `PROTOCOL PAIR IS INT ; BYTE
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	proto, ok := program.Statements[0].(*ast.ProtocolDecl)
	if !ok {
		t.Fatalf("expected ProtocolDecl, got %T", program.Statements[0])
	}

	if proto.Name != "PAIR" {
		t.Errorf("expected name 'PAIR', got %s", proto.Name)
	}

	if proto.Kind != "sequential" {
		t.Errorf("expected kind 'sequential', got %s", proto.Kind)
	}

	if len(proto.Types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(proto.Types))
	}
	if proto.Types[0] != "INT" || proto.Types[1] != "BYTE" {
		t.Errorf("expected types [INT, BYTE], got %v", proto.Types)
	}
}

func TestVariantProtocolDecl(t *testing.T) {
	input := `PROTOCOL MSG
  CASE
    text; INT
    number; INT; INT
    quit
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	proto, ok := program.Statements[0].(*ast.ProtocolDecl)
	if !ok {
		t.Fatalf("expected ProtocolDecl, got %T", program.Statements[0])
	}

	if proto.Name != "MSG" {
		t.Errorf("expected name 'MSG', got %s", proto.Name)
	}

	if proto.Kind != "variant" {
		t.Errorf("expected kind 'variant', got %s", proto.Kind)
	}

	if len(proto.Variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(proto.Variants))
	}

	// text; INT
	if proto.Variants[0].Tag != "text" {
		t.Errorf("expected tag 'text', got %s", proto.Variants[0].Tag)
	}
	if len(proto.Variants[0].Types) != 1 || proto.Variants[0].Types[0] != "INT" {
		t.Errorf("expected types [INT] for text, got %v", proto.Variants[0].Types)
	}

	// number; INT; INT
	if proto.Variants[1].Tag != "number" {
		t.Errorf("expected tag 'number', got %s", proto.Variants[1].Tag)
	}
	if len(proto.Variants[1].Types) != 2 {
		t.Errorf("expected 2 types for number, got %d", len(proto.Variants[1].Types))
	}

	// quit (no payload)
	if proto.Variants[2].Tag != "quit" {
		t.Errorf("expected tag 'quit', got %s", proto.Variants[2].Tag)
	}
	if len(proto.Variants[2].Types) != 0 {
		t.Errorf("expected 0 types for quit, got %d", len(proto.Variants[2].Types))
	}
}

func TestVariantProtocolDeclDottedTags(t *testing.T) {
	input := `PROTOCOL BAR.PROTO
  CASE
    bar.data; INT
    bar.terminate
    bar.blank; INT
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	proto, ok := program.Statements[0].(*ast.ProtocolDecl)
	if !ok {
		t.Fatalf("expected ProtocolDecl, got %T", program.Statements[0])
	}

	if proto.Name != "BAR.PROTO" {
		t.Errorf("expected name 'BAR.PROTO', got %s", proto.Name)
	}

	if proto.Kind != "variant" {
		t.Errorf("expected kind 'variant', got %s", proto.Kind)
	}

	if len(proto.Variants) != 3 {
		t.Fatalf("expected 3 variants, got %d", len(proto.Variants))
	}

	// bar.data; INT
	if proto.Variants[0].Tag != "bar.data" {
		t.Errorf("expected tag 'bar.data', got %s", proto.Variants[0].Tag)
	}
	if len(proto.Variants[0].Types) != 1 || proto.Variants[0].Types[0] != "INT" {
		t.Errorf("expected types [INT] for bar.data, got %v", proto.Variants[0].Types)
	}

	// bar.terminate (no payload)
	if proto.Variants[1].Tag != "bar.terminate" {
		t.Errorf("expected tag 'bar.terminate', got %s", proto.Variants[1].Tag)
	}
	if len(proto.Variants[1].Types) != 0 {
		t.Errorf("expected 0 types for bar.terminate, got %d", len(proto.Variants[1].Types))
	}

	// bar.blank; INT
	if proto.Variants[2].Tag != "bar.blank" {
		t.Errorf("expected tag 'bar.blank', got %s", proto.Variants[2].Tag)
	}
	if len(proto.Variants[2].Types) != 1 || proto.Variants[2].Types[0] != "INT" {
		t.Errorf("expected types [INT] for bar.blank, got %v", proto.Variants[2].Types)
	}
}

func TestChanDeclWithProtocol(t *testing.T) {
	input := `PROTOCOL SIGNAL IS INT
CHAN OF SIGNAL c:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}

	chanDecl, ok := program.Statements[1].(*ast.ChanDecl)
	if !ok {
		t.Fatalf("expected ChanDecl, got %T", program.Statements[1])
	}

	if chanDecl.ElemType != "SIGNAL" {
		t.Errorf("expected element type 'SIGNAL', got %s", chanDecl.ElemType)
	}

	if len(chanDecl.Names) != 1 || chanDecl.Names[0] != "c" {
		t.Errorf("expected name 'c', got %v", chanDecl.Names)
	}
}

func TestSequentialSend(t *testing.T) {
	input := `c ! 42 ; 65
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

	// First value
	intLit, ok := send.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for Value, got %T", send.Value)
	}
	if intLit.Value != 42 {
		t.Errorf("expected value 42, got %d", intLit.Value)
	}

	// Additional values
	if len(send.Values) != 1 {
		t.Fatalf("expected 1 additional value, got %d", len(send.Values))
	}
	intLit2, ok := send.Values[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for Values[0], got %T", send.Values[0])
	}
	if intLit2.Value != 65 {
		t.Errorf("expected value 65, got %d", intLit2.Value)
	}
}

func TestRecordDecl(t *testing.T) {
	input := `RECORD POINT
  INT x:
  INT y:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	rec, ok := program.Statements[0].(*ast.RecordDecl)
	if !ok {
		t.Fatalf("expected RecordDecl, got %T", program.Statements[0])
	}

	if rec.Name != "POINT" {
		t.Errorf("expected name 'POINT', got %s", rec.Name)
	}

	if len(rec.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(rec.Fields))
	}

	if rec.Fields[0].Type != "INT" || rec.Fields[0].Name != "x" {
		t.Errorf("expected field 0: INT x, got %s %s", rec.Fields[0].Type, rec.Fields[0].Name)
	}

	if rec.Fields[1].Type != "INT" || rec.Fields[1].Name != "y" {
		t.Errorf("expected field 1: INT y, got %s %s", rec.Fields[1].Type, rec.Fields[1].Name)
	}
}

func TestRecordDeclMultipleFieldNames(t *testing.T) {
	input := `RECORD R
  INT a, b:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	rec, ok := program.Statements[0].(*ast.RecordDecl)
	if !ok {
		t.Fatalf("expected RecordDecl, got %T", program.Statements[0])
	}

	if len(rec.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(rec.Fields))
	}

	if rec.Fields[0].Type != "INT" || rec.Fields[0].Name != "a" {
		t.Errorf("expected field 0: INT a, got %s %s", rec.Fields[0].Type, rec.Fields[0].Name)
	}

	if rec.Fields[1].Type != "INT" || rec.Fields[1].Name != "b" {
		t.Errorf("expected field 1: INT b, got %s %s", rec.Fields[1].Type, rec.Fields[1].Name)
	}
}

func TestRecordVarDecl(t *testing.T) {
	input := `RECORD POINT
  INT x:
  INT y:
POINT p:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}

	varDecl, ok := program.Statements[1].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected VarDecl, got %T", program.Statements[1])
	}

	if varDecl.Type != "POINT" {
		t.Errorf("expected type 'POINT', got %s", varDecl.Type)
	}

	if len(varDecl.Names) != 1 || varDecl.Names[0] != "p" {
		t.Errorf("expected name 'p', got %v", varDecl.Names)
	}
}

func TestRecordFieldAssignment(t *testing.T) {
	// Record field assignment uses bracket syntax, same as indexed assignment
	input := `p[x] := 5
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

	if assign.Name != "p" {
		t.Errorf("expected name 'p', got %s", assign.Name)
	}

	if len(assign.Indices) == 0 {
		t.Fatal("expected index expression, got nil")
	}

	ident, ok := assign.Indices[0].(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for index, got %T", assign.Indices[0])
	}
	if ident.Value != "x" {
		t.Errorf("expected index 'x', got %s", ident.Value)
	}
}

func TestRecordFieldAccess(t *testing.T) {
	// Record field access in expression uses bracket syntax
	input := `val := p[x] + 1
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

	indexExpr, ok := binExpr.Left.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr on left, got %T", binExpr.Left)
	}

	left, ok := indexExpr.Left.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier in IndexExpr, got %T", indexExpr.Left)
	}
	if left.Value != "p" {
		t.Errorf("expected 'p', got %s", left.Value)
	}

	idx, ok := indexExpr.Index.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for index, got %T", indexExpr.Index)
	}
	if idx.Value != "x" {
		t.Errorf("expected 'x', got %s", idx.Value)
	}
}

func TestChanArrayDecl(t *testing.T) {
	input := `[5]CHAN OF INT cs:
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

	if len(decl.Sizes) == 0 {
		t.Error("expected IsArray=true")
	}

	sizeLit, ok := decl.Sizes[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for size, got %T", decl.Sizes[0])
	}
	if sizeLit.Value != 5 {
		t.Errorf("expected size 5, got %d", sizeLit.Value)
	}

	if decl.ElemType != "INT" {
		t.Errorf("expected element type INT, got %s", decl.ElemType)
	}

	if len(decl.Names) != 1 || decl.Names[0] != "cs" {
		t.Errorf("expected name 'cs', got %v", decl.Names)
	}
}

func TestIndexedSend(t *testing.T) {
	input := `cs[0] ! 42
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

	if send.Channel != "cs" {
		t.Errorf("expected channel 'cs', got %s", send.Channel)
	}

	if len(send.ChannelIndices) == 0 {
		t.Fatal("expected ChannelIndex, got nil")
	}

	idxLit, ok := send.ChannelIndices[0].(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for index, got %T", send.ChannelIndices[0])
	}
	if idxLit.Value != 0 {
		t.Errorf("expected index 0, got %d", idxLit.Value)
	}

	valLit, ok := send.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for value, got %T", send.Value)
	}
	if valLit.Value != 42 {
		t.Errorf("expected value 42, got %d", valLit.Value)
	}
}

func TestIndexedReceive(t *testing.T) {
	input := `cs[i] ? x
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

	if recv.Channel != "cs" {
		t.Errorf("expected channel 'cs', got %s", recv.Channel)
	}

	if len(recv.ChannelIndices) == 0 {
		t.Fatal("expected ChannelIndex, got nil")
	}

	idxIdent, ok := recv.ChannelIndices[0].(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for index, got %T", recv.ChannelIndices[0])
	}
	if idxIdent.Value != "i" {
		t.Errorf("expected index 'i', got %s", idxIdent.Value)
	}

	if recv.Variable != "x" {
		t.Errorf("expected variable 'x', got %s", recv.Variable)
	}
}

func TestChanArrayParam(t *testing.T) {
	input := `PROC worker([]CHAN OF INT cs, VAL INT n)
  SKIP
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

	if len(proc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(proc.Params))
	}

	p0 := proc.Params[0]
	if !p0.IsChan {
		t.Error("param 0: expected IsChan=true")
	}
	if p0.ChanArrayDims == 0 {
		t.Error("param 0: expected ChanArrayDims > 0")
	}
	if p0.ChanElemType != "INT" {
		t.Errorf("param 0: expected ChanElemType=INT, got %s", p0.ChanElemType)
	}
	if p0.Name != "cs" {
		t.Errorf("param 0: expected Name=cs, got %s", p0.Name)
	}

	p1 := proc.Params[1]
	if p1.IsChan || p1.ChanArrayDims > 0 {
		t.Error("param 1: expected IsChan=false, ChanArrayDims=0")
	}
	if !p1.IsVal {
		t.Error("param 1: expected IsVal=true")
	}
}

func TestChanDirParam(t *testing.T) {
	input := `PROC worker(CHAN OF INT input?, CHAN OF INT output!)
  SEQ
    INT x:
    input ? x
    output ! x
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

	if len(proc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(proc.Params))
	}

	// First param: CHAN OF INT input? (input direction)
	p0 := proc.Params[0]
	if !p0.IsChan {
		t.Error("param 0: expected IsChan=true")
	}
	if p0.ChanDir != "?" {
		t.Errorf("param 0: expected ChanDir=?, got %q", p0.ChanDir)
	}
	if p0.Name != "input" {
		t.Errorf("param 0: expected Name=input, got %s", p0.Name)
	}

	// Second param: CHAN OF INT output! (output direction)
	p1 := proc.Params[1]
	if !p1.IsChan {
		t.Error("param 1: expected IsChan=true")
	}
	if p1.ChanDir != "!" {
		t.Errorf("param 1: expected ChanDir=!, got %q", p1.ChanDir)
	}
	if p1.Name != "output" {
		t.Errorf("param 1: expected Name=output, got %s", p1.Name)
	}
}

func TestChanArrayDirParam(t *testing.T) {
	input := `PROC worker([]CHAN OF INT cs?)
  SKIP
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	proc, ok := program.Statements[0].(*ast.ProcDecl)
	if !ok {
		t.Fatalf("expected ProcDecl, got %T", program.Statements[0])
	}

	p0 := proc.Params[0]
	if p0.ChanArrayDims == 0 {
		t.Error("param 0: expected ChanArrayDims > 0")
	}
	if p0.ChanDir != "?" {
		t.Errorf("param 0: expected ChanDir=?, got %q", p0.ChanDir)
	}
}

func TestSequentialReceive(t *testing.T) {
	input := `c ? x ; y
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

	if len(recv.Variables) != 1 || recv.Variables[0] != "y" {
		t.Errorf("expected additional variables [y], got %v", recv.Variables)
	}
}

func TestSizeExpression(t *testing.T) {
	input := `x := SIZE arr
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

	sizeExpr, ok := assign.Value.(*ast.SizeExpr)
	if !ok {
		t.Fatalf("expected SizeExpr, got %T", assign.Value)
	}

	ident, ok := sizeExpr.Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier inside SizeExpr, got %T", sizeExpr.Expr)
	}
	if ident.Value != "arr" {
		t.Errorf("expected 'arr', got %s", ident.Value)
	}
}

func TestSizeExpressionInBinaryExpr(t *testing.T) {
	// SIZE has PREFIX precedence, so "SIZE arr + 1" parses as "(SIZE arr) + 1"
	input := `x := SIZE arr + 1
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	binExpr, ok := assign.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}

	_, ok = binExpr.Left.(*ast.SizeExpr)
	if !ok {
		t.Fatalf("expected SizeExpr as left of BinaryExpr, got %T", binExpr.Left)
	}
}

func TestMostNegExpression(t *testing.T) {
	input := `x := MOSTNEG INT
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

	mostExpr, ok := assign.Value.(*ast.MostExpr)
	if !ok {
		t.Fatalf("expected MostExpr, got %T", assign.Value)
	}

	if mostExpr.ExprType != "INT" {
		t.Errorf("expected ExprType 'INT', got %s", mostExpr.ExprType)
	}
	if !mostExpr.IsNeg {
		t.Error("expected IsNeg to be true for MOSTNEG")
	}
}

func TestMostPosExpression(t *testing.T) {
	input := `x := MOSTPOS BYTE
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	mostExpr, ok := assign.Value.(*ast.MostExpr)
	if !ok {
		t.Fatalf("expected MostExpr, got %T", assign.Value)
	}

	if mostExpr.ExprType != "BYTE" {
		t.Errorf("expected ExprType 'BYTE', got %s", mostExpr.ExprType)
	}
	if mostExpr.IsNeg {
		t.Error("expected IsNeg to be false for MOSTPOS")
	}
}

func TestMostNegInBinaryExpr(t *testing.T) {
	// MOSTNEG INT + 1 should parse as (MOSTNEG INT) + 1
	input := `x := MOSTNEG INT + 1
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assign, ok := program.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment, got %T", program.Statements[0])
	}

	binExpr, ok := assign.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}

	_, ok = binExpr.Left.(*ast.MostExpr)
	if !ok {
		t.Fatalf("expected MostExpr as left of BinaryExpr, got %T", binExpr.Left)
	}
}

func TestValAbbreviation(t *testing.T) {
	input := `VAL INT x IS 42:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	abbr, ok := program.Statements[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", program.Statements[0])
	}

	if !abbr.IsVal {
		t.Error("expected IsVal to be true")
	}
	if abbr.Type != "INT" {
		t.Errorf("expected type INT, got %s", abbr.Type)
	}
	if abbr.Name != "x" {
		t.Errorf("expected name 'x', got %s", abbr.Name)
	}
	if abbr.Value == nil {
		t.Fatal("expected non-nil Value")
	}
	lit, ok := abbr.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral, got %T", abbr.Value)
	}
	if lit.Value != 42 {
		t.Errorf("expected value 42, got %d", lit.Value)
	}
}

func TestNonValAbbreviation(t *testing.T) {
	input := `INT y IS z:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	abbr, ok := program.Statements[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", program.Statements[0])
	}

	if abbr.IsVal {
		t.Error("expected IsVal to be false")
	}
	if abbr.Type != "INT" {
		t.Errorf("expected type INT, got %s", abbr.Type)
	}
	if abbr.Name != "y" {
		t.Errorf("expected name 'y', got %s", abbr.Name)
	}
	ident, ok := abbr.Value.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier, got %T", abbr.Value)
	}
	if ident.Value != "z" {
		t.Errorf("expected value 'z', got %s", ident.Value)
	}
}

func TestValBoolAbbreviation(t *testing.T) {
	input := `VAL BOOL flag IS TRUE:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	abbr, ok := program.Statements[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", program.Statements[0])
	}

	if !abbr.IsVal {
		t.Error("expected IsVal to be true")
	}
	if abbr.Type != "BOOL" {
		t.Errorf("expected type BOOL, got %s", abbr.Type)
	}
	if abbr.Name != "flag" {
		t.Errorf("expected name 'flag', got %s", abbr.Name)
	}
}

func TestAbbreviationWithExpression(t *testing.T) {
	input := `VAL INT sum IS (a + b):
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	abbr, ok := program.Statements[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", program.Statements[0])
	}

	if abbr.Name != "sum" {
		t.Errorf("expected name 'sum', got %s", abbr.Name)
	}

	// Value should be a binary expression (parens are stripped during parsing)
	_, ok = abbr.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", abbr.Value)
	}
}

func TestInitialDecl(t *testing.T) {
	input := `INITIAL INT x IS 42:
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	abbr, ok := program.Statements[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", program.Statements[0])
	}

	if !abbr.IsInitial {
		t.Error("expected IsInitial to be true")
	}
	if abbr.IsVal {
		t.Error("expected IsVal to be false")
	}
	if abbr.Type != "INT" {
		t.Errorf("expected type 'INT', got %s", abbr.Type)
	}
	if abbr.Name != "x" {
		t.Errorf("expected name 'x', got %s", abbr.Name)
	}
}

func TestInitialDeclWithExpression(t *testing.T) {
	input := `INITIAL INT left IS (a + b):
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	abbr, ok := program.Statements[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", program.Statements[0])
	}

	if !abbr.IsInitial {
		t.Error("expected IsInitial to be true")
	}
	if abbr.Name != "left" {
		t.Errorf("expected name 'left', got %s", abbr.Name)
	}

	_, ok = abbr.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", abbr.Value)
	}
}

func TestOpenArrayParam(t *testing.T) {
	input := `PROC sum.array(VAL []INT arr, INT result)
  SKIP
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

	if len(proc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(proc.Params))
	}

	p0 := proc.Params[0]
	if !p0.IsVal {
		t.Error("param 0: expected IsVal=true")
	}
	if p0.OpenArrayDims == 0 {
		t.Error("param 0: expected OpenArrayDims > 0")
	}
	if p0.Type != "INT" {
		t.Errorf("param 0: expected Type=INT, got %s", p0.Type)
	}
	if p0.Name != "arr" {
		t.Errorf("param 0: expected Name=arr, got %s", p0.Name)
	}

	p1 := proc.Params[1]
	if p1.OpenArrayDims > 0 {
		t.Error("param 1: expected OpenArrayDims=0")
	}
	if p1.IsVal {
		t.Error("param 1: expected IsVal=false")
	}
	if p1.Type != "INT" {
		t.Errorf("param 1: expected Type=INT, got %s", p1.Type)
	}
}

func TestChanDeclShorthand(t *testing.T) {
	input := `CHAN BYTE c:
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

	if decl.ElemType != "BYTE" {
		t.Errorf("expected element type BYTE, got %s", decl.ElemType)
	}

	if len(decl.Names) != 1 || decl.Names[0] != "c" {
		t.Errorf("expected name 'c', got %v", decl.Names)
	}
}

func TestChanArrayDeclShorthand(t *testing.T) {
	input := `[5]CHAN INT cs:
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

	if len(decl.Sizes) == 0 {
		t.Error("expected IsArray=true")
	}

	if decl.ElemType != "INT" {
		t.Errorf("expected element type INT, got %s", decl.ElemType)
	}

	if len(decl.Names) != 1 || decl.Names[0] != "cs" {
		t.Errorf("expected name 'cs', got %v", decl.Names)
	}
}

func TestChanParamShorthand(t *testing.T) {
	input := `PROC worker(CHAN BYTE input?, []CHAN INT cs)
  SKIP
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

	if len(proc.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(proc.Params))
	}

	// First param: CHAN BYTE input?
	p0 := proc.Params[0]
	if !p0.IsChan {
		t.Error("param 0: expected IsChan=true")
	}
	if p0.ChanElemType != "BYTE" {
		t.Errorf("param 0: expected ChanElemType=BYTE, got %s", p0.ChanElemType)
	}
	if p0.ChanDir != "?" {
		t.Errorf("param 0: expected ChanDir=?, got %q", p0.ChanDir)
	}

	// Second param: []CHAN INT cs
	p1 := proc.Params[1]
	if !p1.IsChan || p1.ChanArrayDims == 0 {
		t.Error("param 1: expected IsChan=true, ChanArrayDims > 0")
	}
	if p1.ChanElemType != "INT" {
		t.Errorf("param 1: expected ChanElemType=INT, got %s", p1.ChanElemType)
	}
}

func TestHexIntegerLiteral(t *testing.T) {
	input := `x := #FF
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

	intLit, ok := assign.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral, got %T", assign.Value)
	}

	if intLit.Value != 255 {
		t.Errorf("expected value 255, got %d", intLit.Value)
	}
}

func TestHexIntegerLiteralLarge(t *testing.T) {
	input := `x := #80000000
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

	intLit, ok := assign.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral, got %T", assign.Value)
	}

	if intLit.Value != 0x80000000 {
		t.Errorf("expected value %d, got %d", int64(0x80000000), intLit.Value)
	}
}

func TestNestedProcDecl(t *testing.T) {
	input := `PROC outer(VAL INT n)
  INT x:
  PROC inner(VAL INT y)
    x := y
  :
  SEQ
    inner(n)
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

	if proc.Name != "outer" {
		t.Errorf("expected name 'outer', got %s", proc.Name)
	}

	// Body should have 3 statements: VarDecl, nested ProcDecl, SeqBlock
	if len(proc.Body) != 3 {
		t.Fatalf("expected 3 body statements, got %d", len(proc.Body))
	}

	// First: INT x:
	if _, ok := proc.Body[0].(*ast.VarDecl); !ok {
		t.Errorf("expected VarDecl as first body statement, got %T", proc.Body[0])
	}

	// Second: nested PROC inner
	nestedProc, ok := proc.Body[1].(*ast.ProcDecl)
	if !ok {
		t.Fatalf("expected nested ProcDecl, got %T", proc.Body[1])
	}
	if nestedProc.Name != "inner" {
		t.Errorf("expected nested proc name 'inner', got %s", nestedProc.Name)
	}

	// Third: SEQ block
	if _, ok := proc.Body[2].(*ast.SeqBlock); !ok {
		t.Errorf("expected SeqBlock as third body statement, got %T", proc.Body[2])
	}
}

func TestNestedFuncDecl(t *testing.T) {
	input := `PROC compute(VAL INT n)
  INT FUNCTION double(VAL INT x)
    IS x * 2
  SEQ
    print.int(double(n))
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

	// Body should have 2 statements: nested FuncDecl, SeqBlock
	if len(proc.Body) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(proc.Body))
	}

	fn, ok := proc.Body[0].(*ast.FuncDecl)
	if !ok {
		t.Fatalf("expected nested FuncDecl, got %T", proc.Body[0])
	}
	if fn.Name != "double" {
		t.Errorf("expected nested func name 'double', got %s", fn.Name)
	}

	if _, ok := proc.Body[1].(*ast.SeqBlock); !ok {
		t.Errorf("expected SeqBlock as second body statement, got %T", proc.Body[1])
	}
}

func TestProcLocalVarDecls(t *testing.T) {
	input := `PROC foo(VAL INT n)
  INT x:
  INT y:
  SEQ
    x := n
    y := n * 2
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

	// Body should have 3 statements: 2 VarDecls + SeqBlock
	if len(proc.Body) != 3 {
		t.Fatalf("expected 3 body statements, got %d", len(proc.Body))
	}

	for i := 0; i < 2; i++ {
		if _, ok := proc.Body[i].(*ast.VarDecl); !ok {
			t.Errorf("expected VarDecl at index %d, got %T", i, proc.Body[i])
		}
	}

	if _, ok := proc.Body[2].(*ast.SeqBlock); !ok {
		t.Errorf("expected SeqBlock at index 2, got %T", proc.Body[2])
	}
}

func TestCheckedArithmeticOperators(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"x := a PLUS b\n", "PLUS"},
		{"x := a MINUS b\n", "MINUS"},
		{"x := a TIMES b\n", "TIMES"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("[%s] expected 1 statement, got %d", tt.operator, len(program.Statements))
		}

		assign, ok := program.Statements[0].(*ast.Assignment)
		if !ok {
			t.Fatalf("[%s] expected Assignment, got %T", tt.operator, program.Statements[0])
		}

		binExpr, ok := assign.Value.(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("[%s] expected BinaryExpr, got %T", tt.operator, assign.Value)
		}

		if binExpr.Operator != tt.operator {
			t.Errorf("[%s] expected operator %q, got %q", tt.operator, tt.operator, binExpr.Operator)
		}
	}
}

func TestCheckedArithmeticPrecedence(t *testing.T) {
	// a PLUS b TIMES c should parse as a PLUS (b TIMES c)
	input := "x := a PLUS b TIMES c\n"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assign := program.Statements[0].(*ast.Assignment)
	binExpr := assign.Value.(*ast.BinaryExpr)

	if binExpr.Operator != "PLUS" {
		t.Errorf("expected top-level operator PLUS, got %s", binExpr.Operator)
	}

	rightBin, ok := binExpr.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right side to be BinaryExpr, got %T", binExpr.Right)
	}
	if rightBin.Operator != "TIMES" {
		t.Errorf("expected right operator TIMES, got %s", rightBin.Operator)
	}
}

func TestCheckedAndSymbolMixed(t *testing.T) {
	// a + (b TIMES c) should work with mixed operators
	input := "x := a + (b TIMES c)\n"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	assign := program.Statements[0].(*ast.Assignment)
	binExpr := assign.Value.(*ast.BinaryExpr)

	if binExpr.Operator != "+" {
		t.Errorf("expected top-level operator +, got %s", binExpr.Operator)
	}
}

func TestMultiStatementIfBody(t *testing.T) {
	input := `IF
  x > 0
    INT y:
    y := 42
    print.int(y)
  TRUE
    SKIP
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

	// First choice should have 3 body statements: VarDecl, Assignment, ProcCall
	if len(ifStmt.Choices[0].Body) != 3 {
		t.Fatalf("expected 3 body statements in first choice, got %d", len(ifStmt.Choices[0].Body))
	}
	if _, ok := ifStmt.Choices[0].Body[0].(*ast.VarDecl); !ok {
		t.Errorf("expected VarDecl as first body stmt, got %T", ifStmt.Choices[0].Body[0])
	}
	if _, ok := ifStmt.Choices[0].Body[1].(*ast.Assignment); !ok {
		t.Errorf("expected Assignment as second body stmt, got %T", ifStmt.Choices[0].Body[1])
	}
	if _, ok := ifStmt.Choices[0].Body[2].(*ast.ProcCall); !ok {
		t.Errorf("expected ProcCall as third body stmt, got %T", ifStmt.Choices[0].Body[2])
	}

	// Second choice should have 1 body statement: Skip
	if len(ifStmt.Choices[1].Body) != 1 {
		t.Fatalf("expected 1 body statement in second choice, got %d", len(ifStmt.Choices[1].Body))
	}
	if _, ok := ifStmt.Choices[1].Body[0].(*ast.Skip); !ok {
		t.Errorf("expected Skip, got %T", ifStmt.Choices[1].Body[0])
	}
}

func TestChannelDirAtCallSite(t *testing.T) {
	input := `foo(out!, in?)
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	call, ok := program.Statements[0].(*ast.ProcCall)
	if !ok {
		t.Fatalf("expected ProcCall, got %T", program.Statements[0])
	}

	if call.Name != "foo" {
		t.Errorf("expected proc name 'foo', got %q", call.Name)
	}

	if len(call.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(call.Args))
	}

	arg0, ok := call.Args[0].(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for arg 0, got %T", call.Args[0])
	}
	if arg0.Value != "out" {
		t.Errorf("expected arg 0 = 'out', got %q", arg0.Value)
	}

	arg1, ok := call.Args[1].(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for arg 1, got %T", call.Args[1])
	}
	if arg1.Value != "in" {
		t.Errorf("expected arg 1 = 'in', got %q", arg1.Value)
	}
}

func TestUntypedValAbbreviation(t *testing.T) {
	input := `VAL x IS 42 :
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	abbr, ok := program.Statements[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", program.Statements[0])
	}

	if !abbr.IsVal {
		t.Error("expected IsVal to be true")
	}
	if abbr.Type != "" {
		t.Errorf("expected empty type, got %q", abbr.Type)
	}
	if abbr.Name != "x" {
		t.Errorf("expected name 'x', got %s", abbr.Name)
	}
	if abbr.Value == nil {
		t.Fatal("expected non-nil Value")
	}
	lit, ok := abbr.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral, got %T", abbr.Value)
	}
	if lit.Value != 42 {
		t.Errorf("expected value 42, got %d", lit.Value)
	}
}

func TestArrayLiteral(t *testing.T) {
	input := `VAL x IS [1, 2, 3] :
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	abbr, ok := program.Statements[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", program.Statements[0])
	}

	if abbr.Value == nil {
		t.Fatal("expected non-nil Value")
	}

	arr, ok := abbr.Value.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral, got %T", abbr.Value)
	}

	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}

	for i, expected := range []int64{1, 2, 3} {
		lit, ok := arr.Elements[i].(*ast.IntegerLiteral)
		if !ok {
			t.Fatalf("element %d: expected IntegerLiteral, got %T", i, arr.Elements[i])
		}
		if lit.Value != expected {
			t.Errorf("element %d: expected %d, got %d", i, expected, lit.Value)
		}
	}
}

func TestRetypesDecl(t *testing.T) {
	input := `VAL INT X RETYPES Y :
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	rt, ok := program.Statements[0].(*ast.RetypesDecl)
	if !ok {
		t.Fatalf("expected RetypesDecl, got %T", program.Statements[0])
	}

	if !rt.IsVal {
		t.Error("expected IsVal to be true")
	}
	if rt.TargetType != "INT" {
		t.Errorf("expected TargetType 'INT', got %q", rt.TargetType)
	}
	if rt.Name != "X" {
		t.Errorf("expected Name 'X', got %q", rt.Name)
	}
	if rt.Source != "Y" {
		t.Errorf("expected Source 'Y', got %q", rt.Source)
	}
	if rt.IsArray {
		t.Error("expected IsArray to be false")
	}
}

func TestRetypesDeclArray(t *testing.T) {
	input := `VAL [2]INT X RETYPES Y :
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	rt, ok := program.Statements[0].(*ast.RetypesDecl)
	if !ok {
		t.Fatalf("expected RetypesDecl, got %T", program.Statements[0])
	}

	if !rt.IsVal {
		t.Error("expected IsVal to be true")
	}
	if rt.TargetType != "INT" {
		t.Errorf("expected TargetType 'INT', got %q", rt.TargetType)
	}
	if rt.Name != "X" {
		t.Errorf("expected Name 'X', got %q", rt.Name)
	}
	if rt.Source != "Y" {
		t.Errorf("expected Source 'Y', got %q", rt.Source)
	}
	if !rt.IsArray {
		t.Error("expected IsArray to be true")
	}
	if rt.ArraySize == nil {
		t.Fatal("expected non-nil ArraySize")
	}
	sizelit, ok := rt.ArraySize.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for ArraySize, got %T", rt.ArraySize)
	}
	if sizelit.Value != 2 {
		t.Errorf("expected ArraySize 2, got %d", sizelit.Value)
	}
}

func TestMultiLineBooleanExpression(t *testing.T) {
	input := `PROC test()
  INT x:
  IF
    TRUE AND
      TRUE
      x := 1
    TRUE
      x := 2
:
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

	// Body should have VarDecl + IfStatement
	if len(proc.Body) < 2 {
		t.Fatalf("expected at least 2 body statements, got %d", len(proc.Body))
	}

	if _, ok := proc.Body[0].(*ast.VarDecl); !ok {
		t.Errorf("expected VarDecl at index 0, got %T", proc.Body[0])
	}

	ifStmt, ok := proc.Body[1].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement at index 1, got %T", proc.Body[1])
	}

	if len(ifStmt.Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(ifStmt.Choices))
	}

	// First choice condition should be a BinaryExpr (TRUE AND TRUE)
	binExpr, ok := ifStmt.Choices[0].Condition.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr for first choice condition, got %T", ifStmt.Choices[0].Condition)
	}
	if binExpr.Operator != "AND" {
		t.Errorf("expected operator 'AND', got %q", binExpr.Operator)
	}
}

func TestAltReplicator(t *testing.T) {
	input := `ALT i = 0 FOR n
  BYTE ch:
  in[i] ? ch
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

	if alt.Replicator == nil {
		t.Fatal("expected replicator, got nil")
	}

	if alt.Replicator.Variable != "i" {
		t.Errorf("expected replicator variable 'i', got %q", alt.Replicator.Variable)
	}

	startLit, ok := alt.Replicator.Start.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for start, got %T", alt.Replicator.Start)
	}
	if startLit.Value != 0 {
		t.Errorf("expected start 0, got %d", startLit.Value)
	}

	countIdent, ok := alt.Replicator.Count.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier for count, got %T", alt.Replicator.Count)
	}
	if countIdent.Value != "n" {
		t.Errorf("expected count 'n', got %q", countIdent.Value)
	}

	if len(alt.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(alt.Cases))
	}

	c := alt.Cases[0]
	if len(c.Declarations) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(c.Declarations))
	}
	vd, ok := c.Declarations[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected VarDecl, got %T", c.Declarations[0])
	}
	if vd.Type != "BYTE" {
		t.Errorf("expected type 'BYTE', got %q", vd.Type)
	}
	if len(vd.Names) != 1 || vd.Names[0] != "ch" {
		t.Errorf("expected name 'ch', got %v", vd.Names)
	}

	if c.Channel != "in" {
		t.Errorf("expected channel 'in', got %q", c.Channel)
	}
	if len(c.ChannelIndices) == 0 {
		t.Fatal("expected channel index, got nil")
	}
	if c.Variable != "ch" {
		t.Errorf("expected variable 'ch', got %q", c.Variable)
	}
}

func TestAltReplicatorWithAbbreviation(t *testing.T) {
	input := `ALT j = 0 FOR s
  VAL INT X IS (j + 1):
  INT any:
  in[X] ? any
    SKIP
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	alt, ok := program.Statements[0].(*ast.AltBlock)
	if !ok {
		t.Fatalf("expected AltBlock, got %T", program.Statements[0])
	}

	if alt.Replicator == nil {
		t.Fatal("expected replicator, got nil")
	}

	c := alt.Cases[0]
	if len(c.Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(c.Declarations))
	}

	abbr, ok := c.Declarations[0].(*ast.Abbreviation)
	if !ok {
		t.Fatalf("expected Abbreviation, got %T", c.Declarations[0])
	}
	if abbr.Name != "X" {
		t.Errorf("expected abbreviation name 'X', got %q", abbr.Name)
	}

	vd, ok := c.Declarations[1].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected VarDecl, got %T", c.Declarations[1])
	}
	if vd.Type != "INT" || vd.Names[0] != "any" {
		t.Errorf("expected INT any, got %s %v", vd.Type, vd.Names)
	}

	if c.Channel != "in" {
		t.Errorf("expected channel 'in', got %q", c.Channel)
	}
	if c.Variable != "any" {
		t.Errorf("expected variable 'any', got %q", c.Variable)
	}
}

func TestInt16Int32Int64VarDecl(t *testing.T) {
	types := []struct {
		input    string
		expected string
	}{
		{"INT16 x:\n", "INT16"},
		{"INT32 x:\n", "INT32"},
		{"INT64 x:\n", "INT64"},
	}
	for _, tt := range types {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("for %q: expected 1 statement, got %d", tt.input, len(program.Statements))
		}
		decl, ok := program.Statements[0].(*ast.VarDecl)
		if !ok {
			t.Fatalf("for %q: expected VarDecl, got %T", tt.input, program.Statements[0])
		}
		if decl.Type != tt.expected {
			t.Errorf("for %q: expected type %s, got %s", tt.input, tt.expected, decl.Type)
		}
	}
}

func TestInt16Int32Int64TypeConversion(t *testing.T) {
	types := []struct {
		input    string
		convType string
	}{
		{"x := INT16 y\n", "INT16"},
		{"x := INT32 y\n", "INT32"},
		{"x := INT64 y\n", "INT64"},
	}
	for _, tt := range types {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("for %q: expected 1 statement, got %d", tt.input, len(program.Statements))
		}
		assign, ok := program.Statements[0].(*ast.Assignment)
		if !ok {
			t.Fatalf("for %q: expected Assignment, got %T", tt.input, program.Statements[0])
		}
		conv, ok := assign.Value.(*ast.TypeConversion)
		if !ok {
			t.Fatalf("for %q: expected TypeConversion, got %T", tt.input, assign.Value)
		}
		if conv.TargetType != tt.convType {
			t.Errorf("for %q: expected target type %s, got %s", tt.input, tt.convType, conv.TargetType)
		}
	}
}

func TestMostNegMostPosInt16Int32Int64(t *testing.T) {
	types := []string{"INT16", "INT32", "INT64"}
	for _, typ := range types {
		for _, kw := range []string{"MOSTNEG", "MOSTPOS"} {
			input := "x := " + kw + " " + typ + "\n"
			l := lexer.New(input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("for %q: expected 1 statement, got %d", input, len(program.Statements))
			}
			assign, ok := program.Statements[0].(*ast.Assignment)
			if !ok {
				t.Fatalf("for %q: expected Assignment, got %T", input, program.Statements[0])
			}
			most, ok := assign.Value.(*ast.MostExpr)
			if !ok {
				t.Fatalf("for %q: expected MostExpr, got %T", input, assign.Value)
			}
			if most.ExprType != typ {
				t.Errorf("for %q: expected type %s, got %s", input, typ, most.ExprType)
			}
			expectedNeg := kw == "MOSTNEG"
			if most.IsNeg != expectedNeg {
				t.Errorf("for %q: expected IsNeg=%v, got %v", input, expectedNeg, most.IsNeg)
			}
		}
	}
}

func TestMultiDimArrayDecl(t *testing.T) {
	input := `[3][4]INT grid:
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

	if len(decl.Sizes) != 2 {
		t.Fatalf("expected 2 dimensions, got %d", len(decl.Sizes))
	}

	s0, ok := decl.Sizes[0].(*ast.IntegerLiteral)
	if !ok || s0.Value != 3 {
		t.Errorf("expected first size 3, got %v", decl.Sizes[0])
	}

	s1, ok := decl.Sizes[1].(*ast.IntegerLiteral)
	if !ok || s1.Value != 4 {
		t.Errorf("expected second size 4, got %v", decl.Sizes[1])
	}

	if decl.Type != "INT" {
		t.Errorf("expected type INT, got %s", decl.Type)
	}

	if len(decl.Names) != 1 || decl.Names[0] != "grid" {
		t.Errorf("expected name 'grid', got %v", decl.Names)
	}
}

func TestMultiDimChanArrayDecl(t *testing.T) {
	input := `[2][3]CHAN OF INT links:
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

	if len(decl.Sizes) != 2 {
		t.Fatalf("expected 2 dimensions, got %d", len(decl.Sizes))
	}

	s0, ok := decl.Sizes[0].(*ast.IntegerLiteral)
	if !ok || s0.Value != 2 {
		t.Errorf("expected first size 2, got %v", decl.Sizes[0])
	}

	s1, ok := decl.Sizes[1].(*ast.IntegerLiteral)
	if !ok || s1.Value != 3 {
		t.Errorf("expected second size 3, got %v", decl.Sizes[1])
	}

	if decl.ElemType != "INT" {
		t.Errorf("expected ElemType INT, got %s", decl.ElemType)
	}

	if len(decl.Names) != 1 || decl.Names[0] != "links" {
		t.Errorf("expected name 'links', got %v", decl.Names)
	}
}

func TestMultiDimIndexedAssignment(t *testing.T) {
	input := `grid[i][j] := 42
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

	if assign.Name != "grid" {
		t.Errorf("expected name 'grid', got %s", assign.Name)
	}

	if len(assign.Indices) != 2 {
		t.Fatalf("expected 2 indices, got %d", len(assign.Indices))
	}

	idx0, ok := assign.Indices[0].(*ast.Identifier)
	if !ok || idx0.Value != "i" {
		t.Errorf("expected first index 'i', got %v", assign.Indices[0])
	}

	idx1, ok := assign.Indices[1].(*ast.Identifier)
	if !ok || idx1.Value != "j" {
		t.Errorf("expected second index 'j', got %v", assign.Indices[1])
	}
}

func TestMultiDimIndexedSend(t *testing.T) {
	input := `cs[i][j] ! 42
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

	if send.Channel != "cs" {
		t.Errorf("expected channel 'cs', got %s", send.Channel)
	}

	if len(send.ChannelIndices) != 2 {
		t.Fatalf("expected 2 indices, got %d", len(send.ChannelIndices))
	}
}

func TestMultiDimIndexedReceive(t *testing.T) {
	input := `cs[i][j] ? x
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

	if recv.Channel != "cs" {
		t.Errorf("expected channel 'cs', got %s", recv.Channel)
	}

	if len(recv.ChannelIndices) != 2 {
		t.Fatalf("expected 2 indices, got %d", len(recv.ChannelIndices))
	}

	if recv.Variable != "x" {
		t.Errorf("expected variable 'x', got %s", recv.Variable)
	}
}

func TestMultiDimOpenArrayParam(t *testing.T) {
	input := `PROC fill([][]CHAN OF INT grid)
  SKIP
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

	p0 := proc.Params[0]
	if p0.ChanArrayDims != 2 {
		t.Errorf("expected ChanArrayDims=2, got %d", p0.ChanArrayDims)
	}
	if !p0.IsChan {
		t.Error("expected IsChan=true")
	}
	if p0.ChanElemType != "INT" {
		t.Errorf("expected ChanElemType=INT, got %s", p0.ChanElemType)
	}
}

func TestVariantReceiveScopedDecl(t *testing.T) {
	input := `PROTOCOL CMD
  CASE
    data; INT
    evolve
    quit

PROC test(CHAN OF CMD ch)
  BOOL done:
  SEQ
    done := FALSE
    WHILE NOT done
      ch ? CASE
        data; done
          SKIP
        evolve
          BOOL flag:
          SEQ
            flag := TRUE
            done := flag
        quit
          done := TRUE
:
`
	l := lexer.New(input)
	pr := New(l)
	program := pr.ParseProgram()
	checkParserErrors(t, pr)

	// Find the PROC
	if len(program.Statements) < 2 {
		t.Fatalf("expected at least 2 statements, got %d", len(program.Statements))
	}
	proc, ok := program.Statements[1].(*ast.ProcDecl)
	if !ok {
		t.Fatalf("expected ProcDecl, got %T", program.Statements[1])
	}

	// Walk to the variant receive inside the WHILE
	// proc body: VarDecl(done), SeqBlock{ assign, WhileLoop{ VariantReceive } }
	seq, ok := proc.Body[1].(*ast.SeqBlock)
	if !ok {
		t.Fatalf("expected SeqBlock, got %T", proc.Body[1])
	}
	wl, ok := seq.Statements[1].(*ast.WhileLoop)
	if !ok {
		t.Fatalf("expected WhileLoop, got %T", seq.Statements[1])
	}
	if len(wl.Body) < 1 {
		t.Fatalf("expected at least 1 statement in while body, got %d", len(wl.Body))
	}
	vr, ok := wl.Body[0].(*ast.VariantReceive)
	if !ok {
		t.Fatalf("expected VariantReceive, got %T", wl.Body[0])
	}

	if len(vr.Cases) != 3 {
		t.Fatalf("expected 3 variant cases, got %d", len(vr.Cases))
	}

	// "evolve" case should have 2 body statements: VarDecl + SeqBlock
	evolveCase := vr.Cases[1]
	if evolveCase.Tag != "evolve" {
		t.Errorf("expected tag 'evolve', got %s", evolveCase.Tag)
	}
	if len(evolveCase.Body) != 2 {
		t.Fatalf("expected 2 body statements in 'evolve' case, got %d", len(evolveCase.Body))
	}
	if _, ok := evolveCase.Body[0].(*ast.VarDecl); !ok {
		t.Errorf("expected VarDecl as first body statement, got %T", evolveCase.Body[0])
	}
	if _, ok := evolveCase.Body[1].(*ast.SeqBlock); !ok {
		t.Errorf("expected SeqBlock as second body statement, got %T", evolveCase.Body[1])
	}
}

func TestReceiveIndexedVariable(t *testing.T) {
	input := `ch ? flags[0]
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

	if recv.Channel != "ch" {
		t.Errorf("expected channel 'ch', got %s", recv.Channel)
	}

	if recv.Variable != "flags" {
		t.Errorf("expected variable 'flags', got %s", recv.Variable)
	}

	if len(recv.VariableIndices) != 1 {
		t.Fatalf("expected 1 variable index, got %d", len(recv.VariableIndices))
	}
}

func TestReceiveMultiIndexedVariable(t *testing.T) {
	input := `ch ? grid[i][j]
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

	if recv.Variable != "grid" {
		t.Errorf("expected variable 'grid', got %s", recv.Variable)
	}

	if len(recv.VariableIndices) != 2 {
		t.Fatalf("expected 2 variable indices, got %d", len(recv.VariableIndices))
	}
}

func TestIndexedChannelReceiveIndexedVariable(t *testing.T) {
	input := `cs[0] ? flags[1]
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

	if recv.Channel != "cs" {
		t.Errorf("expected channel 'cs', got %s", recv.Channel)
	}

	if len(recv.ChannelIndices) != 1 {
		t.Fatalf("expected 1 channel index, got %d", len(recv.ChannelIndices))
	}

	if recv.Variable != "flags" {
		t.Errorf("expected variable 'flags', got %s", recv.Variable)
	}

	if len(recv.VariableIndices) != 1 {
		t.Fatalf("expected 1 variable index, got %d", len(recv.VariableIndices))
	}
}
