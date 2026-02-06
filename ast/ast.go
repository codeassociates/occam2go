package ast

import (
	"github.com/codeassociates/occam2go/lexer"
)

// Node is the base interface for all AST nodes
type Node interface {
	TokenLiteral() string
}

// Statement represents a statement node
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// VarDecl represents a variable declaration: INT x:
type VarDecl struct {
	Token lexer.Token // the type token (INT, BYTE, BOOL)
	Type  string      // "INT", "BYTE", "BOOL", etc.
	Names []string    // variable names (can declare multiple: INT x, y, z:)
}

func (v *VarDecl) statementNode()       {}
func (v *VarDecl) TokenLiteral() string { return v.Token.Literal }

// Assignment represents an assignment: x := 5
type Assignment struct {
	Token lexer.Token // the := token
	Name  string      // variable name
	Value Expression  // the value being assigned
}

func (a *Assignment) statementNode()       {}
func (a *Assignment) TokenLiteral() string { return a.Token.Literal }

// SeqBlock represents a SEQ block (sequential execution)
// If Replicator is non-nil, this is a replicated SEQ (SEQ i = 0 FOR n)
type SeqBlock struct {
	Token      lexer.Token // the SEQ token
	Statements []Statement
	Replicator *Replicator // optional replicator
}

func (s *SeqBlock) statementNode()       {}
func (s *SeqBlock) TokenLiteral() string { return s.Token.Literal }

// ParBlock represents a PAR block (parallel execution)
// If Replicator is non-nil, this is a replicated PAR (PAR i = 0 FOR n)
type ParBlock struct {
	Token      lexer.Token // the PAR token
	Statements []Statement
	Replicator *Replicator // optional replicator
}

func (p *ParBlock) statementNode()       {}
func (p *ParBlock) TokenLiteral() string { return p.Token.Literal }

// Replicator represents a replication spec: i = start FOR count
type Replicator struct {
	Variable string     // loop variable name
	Start    Expression // start value
	Count    Expression // number of iterations
}

// Skip represents the SKIP statement (no-op)
type Skip struct {
	Token lexer.Token
}

func (s *Skip) statementNode()       {}
func (s *Skip) TokenLiteral() string { return s.Token.Literal }

// ProcDecl represents a procedure declaration
type ProcDecl struct {
	Token  lexer.Token // the PROC token
	Name   string
	Params []ProcParam
	Body   Statement // usually a SEQ block
}

func (p *ProcDecl) statementNode()       {}
func (p *ProcDecl) TokenLiteral() string { return p.Token.Literal }

// ProcParam represents a procedure parameter
type ProcParam struct {
	IsVal bool   // VAL parameter (pass by value)
	Type  string // INT, BYTE, BOOL, etc.
	Name  string
}

// ProcCall represents a procedure call
type ProcCall struct {
	Token lexer.Token // the procedure name token
	Name  string
	Args  []Expression
}

func (p *ProcCall) statementNode()       {}
func (p *ProcCall) TokenLiteral() string { return p.Token.Literal }

// WhileLoop represents a WHILE loop
type WhileLoop struct {
	Token     lexer.Token // the WHILE token
	Condition Expression
	Body      Statement
}

func (w *WhileLoop) statementNode()       {}
func (w *WhileLoop) TokenLiteral() string { return w.Token.Literal }

// IfStatement represents an IF statement
type IfStatement struct {
	Token   lexer.Token // the IF token
	Choices []IfChoice
}

type IfChoice struct {
	Condition Expression
	Body      Statement
}

func (i *IfStatement) statementNode()       {}
func (i *IfStatement) TokenLiteral() string { return i.Token.Literal }

// Expressions

// Identifier represents a variable reference
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// IntegerLiteral represents an integer literal
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// BooleanLiteral represents TRUE or FALSE
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }

// BinaryExpr represents a binary operation: x + y
type BinaryExpr struct {
	Token    lexer.Token // the operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpr) expressionNode()      {}
func (be *BinaryExpr) TokenLiteral() string { return be.Token.Literal }

// UnaryExpr represents a unary operation: NOT x, -x
type UnaryExpr struct {
	Token    lexer.Token // the operator token
	Operator string
	Right    Expression
}

func (ue *UnaryExpr) expressionNode()      {}
func (ue *UnaryExpr) TokenLiteral() string { return ue.Token.Literal }

// ParenExpr represents a parenthesized expression
type ParenExpr struct {
	Token lexer.Token
	Expr  Expression
}

func (pe *ParenExpr) expressionNode()      {}
func (pe *ParenExpr) TokenLiteral() string { return pe.Token.Literal }

// ChanDecl represents a channel declaration: CHAN OF INT c:
type ChanDecl struct {
	Token    lexer.Token // the CHAN token
	ElemType string      // the element type (INT, BYTE, etc.)
	Names    []string    // channel names
}

func (c *ChanDecl) statementNode()       {}
func (c *ChanDecl) TokenLiteral() string { return c.Token.Literal }

// Send represents a channel send: c ! x
type Send struct {
	Token   lexer.Token // the ! token
	Channel string      // channel name
	Value   Expression  // value to send
}

func (s *Send) statementNode()       {}
func (s *Send) TokenLiteral() string { return s.Token.Literal }

// Receive represents a channel receive: c ? x
type Receive struct {
	Token    lexer.Token // the ? token
	Channel  string      // channel name
	Variable string      // variable to receive into
}

func (r *Receive) statementNode()       {}
func (r *Receive) TokenLiteral() string { return r.Token.Literal }

// AltBlock represents an ALT block (alternation/select)
type AltBlock struct {
	Token lexer.Token // the ALT token
	Cases []AltCase
}

func (a *AltBlock) statementNode()       {}
func (a *AltBlock) TokenLiteral() string { return a.Token.Literal }

// AltCase represents a single case in an ALT block
type AltCase struct {
	Guard    Expression // optional guard condition (nil if no guard)
	Channel  string     // channel name
	Variable string     // variable to receive into
	Body     Statement  // the body to execute
}
