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

// ArrayDecl represents an array declaration: [5]INT arr: or [5][3]INT arr:
type ArrayDecl struct {
	Token lexer.Token  // the [ token
	Sizes []Expression // array sizes (one per dimension)
	Type  string       // element type ("INT", "BYTE", "BOOL", etc.)
	Names []string     // variable names
}

func (a *ArrayDecl) statementNode()       {}
func (a *ArrayDecl) TokenLiteral() string { return a.Token.Literal }

// Assignment represents an assignment: x := 5 or arr[i] := 5 or arr[i][j] := 5 or [arr FROM n FOR m] := value
type Assignment struct {
	Token       lexer.Token  // the := token
	Name        string       // variable name
	Indices     []Expression // optional: index expressions for arr[i][j] := x (nil/empty for simple assignments)
	SliceTarget *SliceExpr   // optional: slice target for [arr FROM n FOR m] := value
	Value       Expression   // the value being assigned
}

func (a *Assignment) statementNode()       {}
func (a *Assignment) TokenLiteral() string { return a.Token.Literal }

// MultiAssignTarget represents one target in a multi-assignment.
// Name is always set. Indices is non-empty for indexed targets like arr[i] or arr[i][j].
type MultiAssignTarget struct {
	Name    string       // variable name
	Indices []Expression // optional: index expressions for arr[i][j] (nil/empty for simple ident)
}

// MultiAssignment represents a multi-target assignment: a, b := func(x)
type MultiAssignment struct {
	Token   lexer.Token         // the := token
	Targets []MultiAssignTarget // targets on the left side
	Values  []Expression        // expressions on the right side
}

func (m *MultiAssignment) statementNode()       {}
func (m *MultiAssignment) TokenLiteral() string { return m.Token.Literal }

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

// Replicator represents a replication spec: i = start FOR count [STEP step]
type Replicator struct {
	Variable string     // loop variable name
	Start    Expression // start value
	Count    Expression // number of iterations
	Step     Expression // optional step value (nil means step of 1)
}

// Skip represents the SKIP statement (no-op)
type Skip struct {
	Token lexer.Token
}

func (s *Skip) statementNode()       {}
func (s *Skip) TokenLiteral() string { return s.Token.Literal }

// Stop represents the STOP statement (deadlock/halt)
type Stop struct {
	Token lexer.Token
}

func (s *Stop) statementNode()       {}
func (s *Stop) TokenLiteral() string { return s.Token.Literal }

// ProcDecl represents a procedure declaration
type ProcDecl struct {
	Token  lexer.Token // the PROC token
	Name   string
	Params []ProcParam
	Body   []Statement // local declarations + body process
}

func (p *ProcDecl) statementNode()       {}
func (p *ProcDecl) TokenLiteral() string { return p.Token.Literal }

// ProcParam represents a procedure parameter
type ProcParam struct {
	IsVal        bool   // VAL parameter (pass by value)
	Type         string // INT, BYTE, BOOL, etc.
	Name         string
	IsChan       bool   // true if this is a CHAN OF <type> parameter
	ChanArrayDims int   // number of [] dimensions for []CHAN, [][]CHAN, etc. (0 = not a chan array)
	OpenArrayDims int   // number of [] dimensions for []TYPE, [][]TYPE, etc. (0 = not an open array)
	ChanElemType string // element type when IsChan (e.g., "INT")
	ChanDir      string // "?" for input, "!" for output, "" for bidirectional
	ArraySize    string // non-empty for fixed-size array params like [2]INT
}

// ProcCall represents a procedure call
type ProcCall struct {
	Token lexer.Token // the procedure name token
	Name  string
	Args  []Expression
}

func (p *ProcCall) statementNode()       {}
func (p *ProcCall) TokenLiteral() string { return p.Token.Literal }

// FuncDecl represents a function declaration (single or multi-result)
type FuncDecl struct {
	Token       lexer.Token    // the return type token
	ReturnTypes []string       // return types: ["INT"], ["INT", "INT"], etc.
	Name        string
	Params      []ProcParam
	Body        []Statement    // local decls + body statements (VALOF form), empty for IS form
	ResultExprs []Expression   // return expressions (from IS or RESULT)
}

func (f *FuncDecl) statementNode()       {}
func (f *FuncDecl) TokenLiteral() string { return f.Token.Literal }

// FuncCall represents a function call expression
type FuncCall struct {
	Token lexer.Token // the function name token
	Name  string
	Args  []Expression
}

func (f *FuncCall) expressionNode()      {}
func (f *FuncCall) TokenLiteral() string { return f.Token.Literal }

// WhileLoop represents a WHILE loop
type WhileLoop struct {
	Token     lexer.Token // the WHILE token
	Condition Expression
	Body      []Statement
}

func (w *WhileLoop) statementNode()       {}
func (w *WhileLoop) TokenLiteral() string { return w.Token.Literal }

// IfStatement represents an IF statement
type IfStatement struct {
	Token      lexer.Token // the IF token
	Choices    []IfChoice
	Replicator *Replicator // optional replicator for IF i = start FOR count
}

type IfChoice struct {
	Condition Expression
	Body      []Statement
	NestedIf  *IfStatement // non-nil when this choice is a nested/replicated IF
}

func (i *IfStatement) statementNode()       {}
func (i *IfStatement) TokenLiteral() string { return i.Token.Literal }

// CaseStatement represents a CASE statement
type CaseStatement struct {
	Token    lexer.Token  // the CASE token
	Selector Expression   // the selector expression
	Choices  []CaseChoice
}

type CaseChoice struct {
	Values []Expression // nil/empty for ELSE
	IsElse bool
	Body   []Statement
}

func (c *CaseStatement) statementNode()       {}
func (c *CaseStatement) TokenLiteral() string { return c.Token.Literal }

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

// StringLiteral represents a string literal: "hello"
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// ByteLiteral represents a byte literal: 'A', '*n', etc.
type ByteLiteral struct {
	Token lexer.Token
	Value byte
}

func (bl *ByteLiteral) expressionNode()      {}
func (bl *ByteLiteral) TokenLiteral() string { return bl.Token.Literal }

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

// TypeConversion represents a type conversion expression: INT x, BYTE n, etc.
type TypeConversion struct {
	Token      lexer.Token // the type token (INT, BYTE, etc.)
	TargetType string      // "INT", "BYTE", "BOOL", "REAL"
	Qualifier  string      // "" (none), "ROUND", or "TRUNC"
	Expr       Expression  // the expression to convert
}

func (tc *TypeConversion) expressionNode()      {}
func (tc *TypeConversion) TokenLiteral() string { return tc.Token.Literal }

// SizeExpr represents a SIZE expression: SIZE arr
type SizeExpr struct {
	Token lexer.Token // the SIZE token
	Expr  Expression  // the array/string expression
}

func (se *SizeExpr) expressionNode()      {}
func (se *SizeExpr) TokenLiteral() string { return se.Token.Literal }

// MostExpr represents MOSTNEG/MOSTPOS type expressions: MOSTNEG INT, MOSTPOS BYTE, etc.
type MostExpr struct {
	Token    lexer.Token // the MOSTNEG or MOSTPOS token
	ExprType string      // "INT", "BYTE", "REAL32", "REAL64", etc.
	IsNeg    bool        // true for MOSTNEG, false for MOSTPOS
}

func (me *MostExpr) expressionNode()      {}
func (me *MostExpr) TokenLiteral() string { return me.Token.Literal }

// ParenExpr represents a parenthesized expression
type ParenExpr struct {
	Token lexer.Token
	Expr  Expression
}

func (pe *ParenExpr) expressionNode()      {}
func (pe *ParenExpr) TokenLiteral() string { return pe.Token.Literal }

// IndexExpr represents an array index expression: arr[i]
type IndexExpr struct {
	Token lexer.Token // the [ token
	Left  Expression  // the array expression
	Index Expression  // the index expression
}

func (ie *IndexExpr) expressionNode()      {}
func (ie *IndexExpr) TokenLiteral() string { return ie.Token.Literal }

// ChanDecl represents a channel declaration: CHAN OF INT c: or [n]CHAN OF INT cs: or [n][m]CHAN OF INT cs:
type ChanDecl struct {
	Token    lexer.Token  // the CHAN token
	ElemType string       // the element type (INT, BYTE, etc.)
	Names    []string     // channel names
	Sizes    []Expression // array sizes per dimension (empty = scalar channel)
}

func (c *ChanDecl) statementNode()       {}
func (c *ChanDecl) TokenLiteral() string { return c.Token.Literal }

// Send represents a channel send: c ! x or c ! x ; y or c ! tag ; x
type Send struct {
	Token          lexer.Token  // the ! token
	Channel        string       // channel name
	ChannelIndices []Expression // non-empty for cs[i] ! value or cs[i][j] ! value
	Value          Expression   // value to send (simple send, backward compat)
	Values         []Expression // additional values for sequential sends (c ! x ; y)
	VariantTag     string       // variant tag name for variant sends (c ! tag ; x)
}

func (s *Send) statementNode()       {}
func (s *Send) TokenLiteral() string { return s.Token.Literal }

// Receive represents a channel receive: c ? x or c ? x ; y
type Receive struct {
	Token          lexer.Token  // the ? token
	Channel        string       // channel name
	ChannelIndices []Expression // non-empty for cs[i] ? x or cs[i][j] ? x
	Variable       string       // variable to receive into (simple receive)
	Variables      []string     // additional variables for sequential receives (c ? x ; y)
}

func (r *Receive) statementNode()       {}
func (r *Receive) TokenLiteral() string { return r.Token.Literal }

// AltBlock represents an ALT block (alternation/select)
// If Replicator is non-nil, this is a replicated ALT (ALT i = 0 FOR n)
type AltBlock struct {
	Token      lexer.Token // the ALT token
	Cases      []AltCase
	Replicator *Replicator // optional replicator
}

func (a *AltBlock) statementNode()       {}
func (a *AltBlock) TokenLiteral() string { return a.Token.Literal }

// AltCase represents a single case in an ALT block
type AltCase struct {
	Guard          Expression   // optional guard condition (nil if no guard)
	Channel        string       // channel name
	ChannelIndices []Expression // non-empty for cs[i] ? x or cs[i][j] ? x in ALT
	Variable       string       // variable to receive into
	Body           []Statement  // the body to execute
	IsTimer        bool         // true if this is a timer AFTER case
	IsSkip         bool         // true if this is a guarded SKIP case (guard & SKIP)
	Timer          string       // timer name (when IsTimer)
	Deadline       Expression   // AFTER deadline expression (when IsTimer)
	Declarations   []Statement  // scoped declarations before channel input (e.g., BYTE ch:)
}

// TimerDecl represents a timer declaration: TIMER tim:
type TimerDecl struct {
	Token lexer.Token // the TIMER token
	Names []string    // timer variable names
}

func (td *TimerDecl) statementNode()       {}
func (td *TimerDecl) TokenLiteral() string { return td.Token.Literal }

// TimerRead represents a timer read: tim ? t
type TimerRead struct {
	Token    lexer.Token // the ? token
	Timer    string      // timer name
	Variable string      // variable to receive time into
}

func (tr *TimerRead) statementNode()       {}
func (tr *TimerRead) TokenLiteral() string { return tr.Token.Literal }

// ProtocolDecl represents a protocol declaration
type ProtocolDecl struct {
	Token    lexer.Token       // the PROTOCOL token
	Name     string            // protocol name
	Kind     string            // "simple", "sequential", or "variant"
	Types    []string          // element types (simple: len=1, sequential: len>1)
	Variants []ProtocolVariant // only for Kind="variant"
}

type ProtocolVariant struct {
	Tag   string   // tag name (e.g., "text", "quit")
	Types []string // associated types (empty for no-payload tags)
}

func (pd *ProtocolDecl) statementNode()       {}
func (pd *ProtocolDecl) TokenLiteral() string { return pd.Token.Literal }

// VariantReceive represents a variant protocol receive: c ? CASE ...
type VariantReceive struct {
	Token          lexer.Token  // the ? token
	Channel        string
	ChannelIndices []Expression // non-empty for cs[i] ? CASE ... or cs[i][j] ? CASE ...
	Cases          []VariantCase
}

type VariantCase struct {
	Tag       string    // variant tag name
	Variables []string  // variables to bind payload fields
	Body      Statement
}

func (vr *VariantReceive) statementNode()       {}
func (vr *VariantReceive) TokenLiteral() string { return vr.Token.Literal }

// RecordDecl represents a record type declaration: RECORD POINT { INT x: INT y: }
type RecordDecl struct {
	Token  lexer.Token   // the RECORD token
	Name   string        // record type name
	Fields []RecordField // named fields
}

type RecordField struct {
	Type string // "INT", "BYTE", "BOOL", "REAL"
	Name string
}

func (rd *RecordDecl) statementNode()       {}
func (rd *RecordDecl) TokenLiteral() string { return rd.Token.Literal }

// SliceExpr represents an array slice: [arr FROM start FOR length]
type SliceExpr struct {
	Token  lexer.Token // the [ token
	Array  Expression  // the array being sliced
	Start  Expression  // start index
	Length Expression  // number of elements
}

func (se *SliceExpr) expressionNode()      {}
func (se *SliceExpr) TokenLiteral() string { return se.Token.Literal }

// Abbreviation represents an abbreviation: VAL INT x IS 42:, INT y IS z:, or INITIAL INT x IS 42:
type Abbreviation struct {
	Token       lexer.Token // VAL, INITIAL, or type token
	IsVal       bool        // true for VAL abbreviations
	IsInitial   bool        // true for INITIAL declarations
	IsOpenArray bool        // true for []TYPE abbreviations (e.g. VAL []BYTE)
	Type        string      // "INT", "BYTE", "BOOL", etc.
	Name        string      // variable name
	Value       Expression  // the expression
}

func (a *Abbreviation) statementNode()       {}
func (a *Abbreviation) TokenLiteral() string { return a.Token.Literal }

// ArrayLiteral represents an array literal expression: [expr1, expr2, ...]
type ArrayLiteral struct {
	Token    lexer.Token  // the [ token
	Elements []Expression // the elements
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

// RetypesDecl represents a RETYPES declaration:
// VAL INT X RETYPES X : or VAL [2]INT X RETYPES X :
type RetypesDecl struct {
	Token      lexer.Token // the VAL token
	IsVal      bool        // always true for now (VAL ... RETYPES ...)
	TargetType string      // "INT", "REAL32", etc.
	IsArray    bool        // true for [n]TYPE
	ArraySize  Expression  // array size when IsArray
	Name       string      // target variable name
	Source     string      // source variable name
}

func (r *RetypesDecl) statementNode()       {}
func (r *RetypesDecl) TokenLiteral() string { return r.Token.Literal }
