package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codeassociates/occam2go/ast"
	"github.com/codeassociates/occam2go/lexer"
)

// Operator precedence levels
const (
	_ int = iota
	LOWEST
	OR_PREC      // OR
	AND_PREC     // AND
	EQUALS       // =, <>
	LESSGREATER  // <, >, <=, >=
	SUM          // +, -
	PRODUCT      // *, /, \
	PREFIX       // -x, NOT x
	INDEX        // arr[i]
)

var precedences = map[lexer.TokenType]int{
	lexer.OR:       OR_PREC,
	lexer.AND:      AND_PREC,
	lexer.EQ:       EQUALS,
	lexer.NEQ:      EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.LE:       LESSGREATER,
	lexer.GE:       LESSGREATER,
	lexer.AFTER:    LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.PLUS_KW:  SUM,
	lexer.MINUS_KW: SUM,
	lexer.MULTIPLY: PRODUCT,
	lexer.DIVIDE:   PRODUCT,
	lexer.MODULO:   PRODUCT,
	lexer.TIMES:    PRODUCT,
	lexer.BITAND:   PRODUCT,
	lexer.LSHIFT:   PRODUCT,
	lexer.RSHIFT:   PRODUCT,
	lexer.BITOR:    SUM,
	lexer.BITXOR:   SUM,
	lexer.LBRACKET: INDEX,
}

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	// Track current indentation level
	indentLevel int

	// Track timer names to distinguish timer reads from channel receives
	timerNames map[string]bool

	// Track protocol names and definitions
	protocolNames map[string]bool
	protocolDefs  map[string]*ast.ProtocolDecl

	// Track record type names and definitions
	recordNames map[string]bool
	recordDefs  map[string]*ast.RecordDecl
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:             l,
		errors:        []string{},
		timerNames:    make(map[string]bool),
		protocolNames: make(map[string]bool),
		protocolDefs:  make(map[string]*ast.ProtocolDecl),
		recordNames:   make(map[string]bool),
		recordDefs:    make(map[string]*ast.RecordDecl),
	}
	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d: %s", p.curToken.Line, msg))
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	// Track indentation level
	if p.curToken.Type == lexer.INDENT {
		p.indentLevel++
	} else if p.curToken.Type == lexer.DEDENT {
		p.indentLevel--
	}
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.addError(fmt.Sprintf("expected %s, got %s", t, p.peekToken.Type))
	return false
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	// Skip newlines
	for p.curTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	switch p.curToken.Type {
	case lexer.VAL:
		return p.parseAbbreviation()
	case lexer.INITIAL:
		return p.parseInitialDecl()
	case lexer.INT_TYPE, lexer.BYTE_TYPE, lexer.BOOL_TYPE, lexer.REAL_TYPE, lexer.REAL32_TYPE, lexer.REAL64_TYPE:
		if p.peekTokenIs(lexer.FUNCTION) || p.peekTokenIs(lexer.FUNC) || p.peekTokenIs(lexer.COMMA) {
			return p.parseFuncDecl()
		}
		return p.parseVarDeclOrAbbreviation()
	case lexer.LBRACKET:
		return p.parseArrayDecl()
	case lexer.CHAN:
		return p.parseChanDecl()
	case lexer.PROTOCOL:
		return p.parseProtocolDecl()
	case lexer.RECORD:
		return p.parseRecordDecl()
	case lexer.TIMER:
		return p.parseTimerDecl()
	case lexer.SEQ:
		return p.parseSeqBlock()
	case lexer.PAR:
		return p.parseParBlock()
	case lexer.ALT:
		return p.parseAltBlock()
	case lexer.SKIP:
		return &ast.Skip{Token: p.curToken}
	case lexer.STOP:
		return &ast.Stop{Token: p.curToken}
	case lexer.PROC:
		return p.parseProcDecl()
	case lexer.WHILE:
		return p.parseWhileLoop()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.CASE:
		return p.parseCaseStatement()
	case lexer.IDENT:
		// Check for record variable declaration: TYPENAME var:
		if p.recordNames[p.curToken.Literal] && p.peekTokenIs(lexer.IDENT) {
			return p.parseRecordVarDecl()
		}
		// Could be assignment, indexed assignment, indexed send/receive, send, receive, or procedure call
		if p.peekTokenIs(lexer.LBRACKET) {
			return p.parseIndexedOperation()
		}
		if p.peekTokenIs(lexer.ASSIGN) {
			return p.parseAssignment()
		}
		if p.peekTokenIs(lexer.COMMA) {
			return p.parseMultiAssignment()
		}
		if p.peekTokenIs(lexer.SEND) {
			return p.parseSend()
		}
		if p.peekTokenIs(lexer.RECEIVE) {
			if p.timerNames[p.curToken.Literal] {
				return p.parseTimerRead()
			}
			return p.parseReceive()
		}
		return p.parseProcCall()
	case lexer.INDENT, lexer.DEDENT, lexer.EOF:
		return nil
	default:
		p.addError(fmt.Sprintf("unexpected token: %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseVarDecl() *ast.VarDecl {
	decl := &ast.VarDecl{
		Token: p.curToken,
		Type:  p.curToken.Literal,
	}

	// Parse variable names
	for {
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		decl.Names = append(decl.Names, p.curToken.Literal)

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	return decl
}

// parseVarDeclOrAbbreviation parses either a variable declaration (INT x:)
// or a non-VAL abbreviation (INT x IS expr:). Called when current token is a type keyword.
func (p *Parser) parseVarDeclOrAbbreviation() ast.Statement {
	typeToken := p.curToken
	typeName := p.curToken.Literal

	// Consume the name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	name := p.curToken.Literal

	// Check if this is an abbreviation (next token is IS)
	if p.peekTokenIs(lexer.IS) {
		p.nextToken() // consume IS
		p.nextToken() // move to expression
		value := p.parseExpression(LOWEST)

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		return &ast.Abbreviation{
			Token: typeToken,
			IsVal: false,
			Type:  typeName,
			Name:  name,
			Value: value,
		}
	}

	// Otherwise, it's a regular variable declaration — continue parsing names
	decl := &ast.VarDecl{
		Token: typeToken,
		Type:  typeName,
		Names: []string{name},
	}

	// Parse additional comma-separated names
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		decl.Names = append(decl.Names, p.curToken.Literal)
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	return decl
}

// parseAbbreviation parses VAL abbreviations:
//   VAL INT x IS expr:          (typed VAL abbreviation)
//   VAL []BYTE x IS "string":   (open array abbreviation)
//   VAL x IS expr:              (untyped VAL abbreviation)
//   VAL INT X RETYPES X :       (RETYPES declaration)
//   VAL [n]INT X RETYPES X :    (array RETYPES declaration)
// Current token is VAL.
func (p *Parser) parseAbbreviation() ast.Statement {
	token := p.curToken // VAL token

	p.nextToken()

	// Check for []TYPE (open array abbreviation)
	isOpenArray := false
	if p.curTokenIs(lexer.LBRACKET) && p.peekTokenIs(lexer.RBRACKET) {
		isOpenArray = true
		p.nextToken() // consume ]
		p.nextToken() // move to type
	}

	// Check for [n]TYPE (fixed-size array, used in RETYPES)
	isArray := false
	var arraySize ast.Expression
	if !isOpenArray && p.curTokenIs(lexer.LBRACKET) {
		// Could be [n]TYPE name RETYPES ...
		isArray = true
		p.nextToken() // move past [
		arraySize = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
		p.nextToken() // move to type
	}

	// Check for untyped VAL abbreviation: VAL name IS expr :
	// Detect: curToken is IDENT and peekToken is IS (no type keyword)
	if !isOpenArray && !isArray && p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.IS) {
		name := p.curToken.Literal
		p.nextToken() // consume IS
		p.nextToken() // move to expression
		value := p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		return &ast.Abbreviation{
			Token: token,
			IsVal: true,
			Type:  "",
			Name:  name,
			Value: value,
		}
	}

	// Expect a type keyword
	if !isTypeToken(p.curToken.Type) {
		p.addError(fmt.Sprintf("expected type after VAL, got %s", p.curToken.Type))
		return nil
	}
	typeName := p.curToken.Literal

	// Expect name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	name := p.curToken.Literal

	// Check for RETYPES (instead of IS)
	if p.peekTokenIs(lexer.RETYPES) {
		p.nextToken() // consume RETYPES
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		source := p.curToken.Literal
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		return &ast.RetypesDecl{
			Token:      token,
			IsVal:      true,
			TargetType: typeName,
			IsArray:    isArray,
			ArraySize:  arraySize,
			Name:       name,
			Source:      source,
		}
	}

	// Expect IS
	if !p.expectPeek(lexer.IS) {
		return nil
	}

	// Parse expression
	p.nextToken()
	value := p.parseExpression(LOWEST)

	// Expect colon
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	return &ast.Abbreviation{
		Token:       token,
		IsVal:       true,
		IsOpenArray: isOpenArray,
		Type:        typeName,
		Name:        name,
		Value:       value,
	}
}

// parseInitialDecl parses an INITIAL declaration: INITIAL INT x IS expr:
// Current token is INITIAL.
func (p *Parser) parseInitialDecl() *ast.Abbreviation {
	token := p.curToken // INITIAL token

	// Expect a type keyword
	p.nextToken()
	if !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BYTE_TYPE) &&
		!p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.REAL_TYPE) &&
		!p.curTokenIs(lexer.REAL32_TYPE) && !p.curTokenIs(lexer.REAL64_TYPE) {
		p.addError(fmt.Sprintf("expected type after INITIAL, got %s", p.curToken.Type))
		return nil
	}
	typeName := p.curToken.Literal

	// Expect name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	name := p.curToken.Literal

	// Expect IS
	if !p.expectPeek(lexer.IS) {
		return nil
	}

	// Parse expression
	p.nextToken()
	value := p.parseExpression(LOWEST)

	// Expect colon
	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	return &ast.Abbreviation{
		Token:     token,
		IsInitial: true,
		Type:      typeName,
		Name:      name,
		Value:     value,
	}
}

func (p *Parser) parseAssignment() *ast.Assignment {
	stmt := &ast.Assignment{
		Name: p.curToken.Literal,
	}

	p.nextToken() // move to :=
	stmt.Token = p.curToken

	p.nextToken() // move past :=
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseMultiAssignment() *ast.MultiAssignment {
	firstTarget := ast.MultiAssignTarget{Name: p.curToken.Literal}
	return p.parseMultiAssignmentFrom(firstTarget)
}

// parseMultiAssignmentFrom parses a multi-assignment given the first target already parsed.
// The current token should be on the first target's last token (ident or ']').
// Peek token should be COMMA.
func (p *Parser) parseMultiAssignmentFrom(firstTarget ast.MultiAssignTarget) *ast.MultiAssignment {
	stmt := &ast.MultiAssignment{
		Targets: []ast.MultiAssignTarget{firstTarget},
	}

	// Parse comma-separated targets: a, b[i], c
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next target
		target := ast.MultiAssignTarget{Name: p.curToken.Literal}
		if p.peekTokenIs(lexer.LBRACKET) {
			p.nextToken() // move to [
			p.nextToken() // move past [
			target.Index = p.parseExpression(LOWEST)
			if !p.expectPeek(lexer.RBRACKET) {
				return nil
			}
		}
		stmt.Targets = append(stmt.Targets, target)
	}

	p.nextToken() // move to :=
	stmt.Token = p.curToken

	p.nextToken() // move past :=

	// Parse comma-separated values
	stmt.Values = []ast.Expression{p.parseExpression(LOWEST)}
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next expression
		stmt.Values = append(stmt.Values, p.parseExpression(LOWEST))
	}

	return stmt
}

func (p *Parser) parseArrayDecl() ast.Statement {
	lbracketToken := p.curToken

	// Parse size expression after [
	p.nextToken()
	size := p.parseExpression(LOWEST)

	// Check if this is a slice assignment: [arr FROM start FOR length] := value
	// Also handles [arr FOR length] shorthand (FROM 0)
	if p.peekTokenIs(lexer.FROM) || p.peekTokenIs(lexer.FOR) {
		return p.parseSliceAssignment(lbracketToken, size)
	}

	// Expect ]
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	// Check if this is a channel array: [n]CHAN OF TYPE
	if p.peekTokenIs(lexer.CHAN) {
		p.nextToken() // move to CHAN
		chanDecl := &ast.ChanDecl{
			Token:   p.curToken,
			IsArray: true,
			Size:    size,
		}

		// Expect OF (optional — CHAN BYTE is shorthand for CHAN OF BYTE)
		if p.peekTokenIs(lexer.OF) {
			p.nextToken() // consume OF
		}

		// Expect type (INT, BYTE, BOOL, etc.) or protocol name (IDENT)
		p.nextToken()
		if p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BYTE_TYPE) ||
			p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.REAL_TYPE) ||
			p.curTokenIs(lexer.REAL32_TYPE) || p.curTokenIs(lexer.REAL64_TYPE) {
			chanDecl.ElemType = p.curToken.Literal
		} else if p.curTokenIs(lexer.IDENT) {
			chanDecl.ElemType = p.curToken.Literal
		} else {
			p.addError(fmt.Sprintf("expected type after CHAN, got %s", p.curToken.Type))
			return nil
		}

		// Parse channel names
		for {
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			chanDecl.Names = append(chanDecl.Names, p.curToken.Literal)

			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume comma
			} else {
				break
			}
		}

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		return chanDecl
	}

	// Regular array declaration
	decl := &ast.ArrayDecl{Token: lbracketToken, Size: size}

	// Expect type (INT, BYTE, BOOL, REAL, REAL32, REAL64)
	p.nextToken()
	if !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BYTE_TYPE) &&
		!p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.REAL_TYPE) &&
		!p.curTokenIs(lexer.REAL32_TYPE) && !p.curTokenIs(lexer.REAL64_TYPE) {
		p.addError(fmt.Sprintf("expected type after array size, got %s", p.curToken.Type))
		return nil
	}
	decl.Type = p.curToken.Literal

	// Parse variable names
	for {
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		decl.Names = append(decl.Names, p.curToken.Literal)

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	return decl
}

// parseSliceAssignment parses [arr FROM start FOR length] := value
// Also handles [arr FOR length] shorthand (start defaults to 0).
// Called from parseArrayDecl when FROM or FOR is detected after the array expression.
// lbracketToken is the [ token, arrayExpr is the already-parsed array expression.
func (p *Parser) parseSliceAssignment(lbracketToken lexer.Token, arrayExpr ast.Expression) ast.Statement {
	var startExpr ast.Expression
	if p.peekTokenIs(lexer.FOR) {
		// [arr FOR length] shorthand — start is 0
		startExpr = &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "0"}, Value: 0}
	} else {
		p.nextToken() // consume FROM
		p.nextToken() // move to start expression
		startExpr = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.FOR) {
		return nil
	}
	p.nextToken() // move to length expression
	lengthExpr := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	assignToken := p.curToken
	p.nextToken() // move past :=

	value := p.parseExpression(LOWEST)

	return &ast.Assignment{
		Token: assignToken,
		SliceTarget: &ast.SliceExpr{
			Token:  lbracketToken,
			Array:  arrayExpr,
			Start:  startExpr,
			Length: lengthExpr,
		},
		Value: value,
	}
}

func (p *Parser) parseIndexedOperation() ast.Statement {
	name := p.curToken.Literal

	p.nextToken() // move to [
	p.nextToken() // move past [
	index := p.parseExpression(LOWEST)

	// Expect ]
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	// Check what follows ]
	if p.peekTokenIs(lexer.COMMA) {
		// Multi-assignment starting with indexed target: name[index], ... := ...
		firstTarget := ast.MultiAssignTarget{Name: name, Index: index}
		return p.parseMultiAssignmentFrom(firstTarget)
	}
	if p.peekTokenIs(lexer.ASSIGN) {
		// Indexed assignment: name[index] := value
		p.nextToken() // move to :=
		stmt := &ast.Assignment{
			Name:  name,
			Token: p.curToken,
			Index: index,
		}
		p.nextToken() // move past :=
		stmt.Value = p.parseExpression(LOWEST)
		return stmt
	}

	if p.peekTokenIs(lexer.SEND) {
		// Indexed channel send: cs[i] ! value
		p.nextToken() // move to !
		sendToken := p.curToken
		p.nextToken() // move past !

		stmt := &ast.Send{
			Token:        sendToken,
			Channel:      name,
			ChannelIndex: index,
		}

		// Check if this is a variant send: first token is an identifier that is a variant tag
		if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.SEMICOLON) {
			possibleTag := p.curToken.Literal
			if p.isVariantTag(possibleTag) {
				stmt.VariantTag = possibleTag
				p.nextToken() // move to ;
				for p.curTokenIs(lexer.SEMICOLON) {
					p.nextToken() // move past ;
					val := p.parseExpression(LOWEST)
					stmt.Values = append(stmt.Values, val)
				}
				return stmt
			}
		}

		stmt.Value = p.parseExpression(LOWEST)

		// Check for sequential send
		for p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // move to ;
			p.nextToken() // move past ;
			val := p.parseExpression(LOWEST)
			stmt.Values = append(stmt.Values, val)
		}

		return stmt
	}

	if p.peekTokenIs(lexer.RECEIVE) {
		// Indexed channel receive: cs[i] ? x or cs[i] ? CASE ...
		p.nextToken() // move to ?
		recvToken := p.curToken

		// Check for variant receive: cs[i] ? CASE
		if p.peekTokenIs(lexer.CASE) {
			p.nextToken() // move to CASE
			return p.parseVariantReceiveWithIndex(name, index, recvToken)
		}

		stmt := &ast.Receive{
			Token:        recvToken,
			Channel:      name,
			ChannelIndex: index,
		}

		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.Variable = p.curToken.Literal

		// Check for sequential receive
		for p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // move to ;
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			stmt.Variables = append(stmt.Variables, p.curToken.Literal)
		}

		return stmt
	}

	// Default: treat as indexed assignment (shouldn't reach here normally)
	p.addError(fmt.Sprintf("expected :=, !, or ? after %s[...], got %s", name, p.peekToken.Type))
	return nil
}

func (p *Parser) parseIndexExpression(left ast.Expression) *ast.IndexExpr {
	expr := &ast.IndexExpr{
		Token: p.curToken,
		Left:  left,
	}

	p.nextToken() // move past [
	expr.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	return expr
}

func (p *Parser) parseChanDecl() *ast.ChanDecl {
	decl := &ast.ChanDecl{Token: p.curToken}

	// Expect OF (optional — CHAN BYTE is shorthand for CHAN OF BYTE)
	if p.peekTokenIs(lexer.OF) {
		p.nextToken() // consume OF
	}

	// Expect type (INT, BYTE, BOOL, etc.) or protocol name (IDENT)
	p.nextToken()
	if p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BYTE_TYPE) ||
		p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.REAL_TYPE) ||
		p.curTokenIs(lexer.REAL32_TYPE) || p.curTokenIs(lexer.REAL64_TYPE) {
		decl.ElemType = p.curToken.Literal
	} else if p.curTokenIs(lexer.IDENT) {
		decl.ElemType = p.curToken.Literal
	} else {
		p.addError(fmt.Sprintf("expected type after CHAN, got %s", p.curToken.Type))
		return nil
	}

	// Parse channel names
	for {
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		decl.Names = append(decl.Names, p.curToken.Literal)

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	return decl
}

func (p *Parser) parseProtocolDecl() *ast.ProtocolDecl {
	decl := &ast.ProtocolDecl{Token: p.curToken}

	// Expect protocol name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal

	// Check if this is IS form (simple/sequential) or CASE form (variant)
	if p.peekTokenIs(lexer.NEWLINE) || p.peekTokenIs(lexer.INDENT) {
		// Could be variant: PROTOCOL NAME \n INDENT CASE ...
		// Skip newlines
		for p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.peekTokenIs(lexer.INDENT) {
			p.nextToken() // consume INDENT
			p.nextToken() // move into block

			if p.curTokenIs(lexer.CASE) {
				// Variant protocol
				decl.Kind = "variant"
				decl.Variants = p.parseProtocolVariants()
				p.protocolNames[decl.Name] = true
				p.protocolDefs[decl.Name] = decl
				return decl
			}
		}

		p.addError("expected IS or CASE in protocol declaration")
		return nil
	}

	// IS form: PROTOCOL NAME IS TYPE [; TYPE]*
	if !p.expectPeek(lexer.IS) {
		return nil
	}

	// Parse type list
	p.nextToken()
	typeName := p.parseProtocolTypeName()
	if typeName == "" {
		return nil
	}
	decl.Types = append(decl.Types, typeName)

	// Check for sequential: ; TYPE
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // move to ;
		p.nextToken() // move past ;
		typeName = p.parseProtocolTypeName()
		if typeName == "" {
			return nil
		}
		decl.Types = append(decl.Types, typeName)
	}

	if len(decl.Types) == 1 {
		decl.Kind = "simple"
	} else {
		decl.Kind = "sequential"
	}

	p.protocolNames[decl.Name] = true
	p.protocolDefs[decl.Name] = decl
	return decl
}

func (p *Parser) parseProtocolTypeName() string {
	switch p.curToken.Type {
	case lexer.INT_TYPE:
		return "INT"
	case lexer.BYTE_TYPE:
		return "BYTE"
	case lexer.BOOL_TYPE:
		return "BOOL"
	case lexer.REAL_TYPE:
		return "REAL"
	case lexer.REAL32_TYPE:
		return "REAL32"
	case lexer.REAL64_TYPE:
		return "REAL64"
	case lexer.IDENT:
		return p.curToken.Literal
	default:
		p.addError(fmt.Sprintf("expected type name in protocol, got %s", p.curToken.Type))
		return ""
	}
}

func (p *Parser) parseProtocolVariants() []ast.ProtocolVariant {
	var variants []ast.ProtocolVariant

	// Skip to next line after CASE
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after CASE in protocol")
		return variants
	}
	p.nextToken() // consume INDENT
	startLevel := p.indentLevel
	p.nextToken() // move into block

	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENT tokens
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				return variants
			}
			p.nextToken()
		}

		// Skip any more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		if p.indentLevel < startLevel {
			break
		}

		// Parse variant: tag [; TYPE]*
		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected variant tag name, got %s", p.curToken.Type))
			return variants
		}

		v := ast.ProtocolVariant{
			Tag: p.curToken.Literal,
		}

		// Parse optional types after semicolons
		for p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // move to ;
			p.nextToken() // move past ;
			typeName := p.parseProtocolTypeName()
			if typeName == "" {
				return variants
			}
			v.Types = append(v.Types, typeName)
		}

		variants = append(variants, v)

		// Advance past newline if needed
		if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}

	return variants
}

func (p *Parser) parseRecordDecl() *ast.RecordDecl {
	decl := &ast.RecordDecl{Token: p.curToken}

	// Expect record name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	decl.Name = p.curToken.Literal

	// Skip newlines
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT for field block
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after RECORD declaration")
		return nil
	}
	p.nextToken() // consume INDENT
	startLevel := p.indentLevel
	p.nextToken() // move into block

	// Parse field declarations: TYPE name[, name]*:
	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENT tokens
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				p.recordNames[decl.Name] = true
				p.recordDefs[decl.Name] = decl
				return decl
			}
			p.nextToken()
		}

		// Skip any more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		if p.indentLevel < startLevel {
			break
		}

		// Expect a type keyword (INT, BYTE, BOOL, REAL, REAL32, REAL64)
		if !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BYTE_TYPE) &&
			!p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.REAL_TYPE) &&
			!p.curTokenIs(lexer.REAL32_TYPE) && !p.curTokenIs(lexer.REAL64_TYPE) {
			p.addError(fmt.Sprintf("expected type in record field, got %s", p.curToken.Type))
			return nil
		}
		fieldType := p.curToken.Literal

		// Parse field names (comma-separated)
		for {
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			decl.Fields = append(decl.Fields, ast.RecordField{
				Type: fieldType,
				Name: p.curToken.Literal,
			})

			if p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume comma
			} else {
				break
			}
		}

		// Expect colon
		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		// Advance past newline if needed
		if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}

	p.recordNames[decl.Name] = true
	p.recordDefs[decl.Name] = decl
	return decl
}

func (p *Parser) parseRecordVarDecl() *ast.VarDecl {
	decl := &ast.VarDecl{
		Token: p.curToken,
		Type:  p.curToken.Literal,
	}

	// Parse variable names
	for {
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		decl.Names = append(decl.Names, p.curToken.Literal)

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	return decl
}

func (p *Parser) parseTimerDecl() *ast.TimerDecl {
	decl := &ast.TimerDecl{Token: p.curToken}

	// Parse timer names
	for {
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		decl.Names = append(decl.Names, p.curToken.Literal)
		p.timerNames[p.curToken.Literal] = true

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	return decl
}

func (p *Parser) parseTimerRead() *ast.TimerRead {
	stmt := &ast.TimerRead{
		Timer: p.curToken.Literal,
	}

	p.nextToken() // move to ?
	stmt.Token = p.curToken

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Variable = p.curToken.Literal

	return stmt
}

func (p *Parser) parseSend() *ast.Send {
	stmt := &ast.Send{
		Channel: p.curToken.Literal,
	}

	p.nextToken() // move to !
	stmt.Token = p.curToken

	p.nextToken() // move past !

	// Check if this is a variant send: first token is an identifier that is a variant tag
	// We detect this by checking if the identifier is followed by SEMICOLON
	// and the identifier is NOT followed by an operator (i.e., it's a bare tag name)
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.SEMICOLON) {
		// Could be variant send (tag ; values) or expression ; values
		// Check if the identifier is a known protocol variant tag
		// For simplicity, if IDENT is followed by SEMICOLON and the ident is
		// not followed by an operator, treat it as a variant tag
		// We save the ident and check further
		possibleTag := p.curToken.Literal
		// Check if this identifier is a protocol variant tag
		if p.isVariantTag(possibleTag) {
			stmt.VariantTag = possibleTag
			p.nextToken() // move to ;
			// Parse remaining values after the tag
			for p.curTokenIs(lexer.SEMICOLON) {
				p.nextToken() // move past ;
				val := p.parseExpression(LOWEST)
				stmt.Values = append(stmt.Values, val)
			}
			return stmt
		}
	}

	stmt.Value = p.parseExpression(LOWEST)

	// Check for sequential send: c ! expr ; expr ; ...
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // move to ;
		p.nextToken() // move past ;
		val := p.parseExpression(LOWEST)
		stmt.Values = append(stmt.Values, val)
	}

	return stmt
}

func (p *Parser) isVariantTag(name string) bool {
	for _, proto := range p.protocolDefs {
		if proto.Kind == "variant" {
			for _, v := range proto.Variants {
				if v.Tag == name {
					return true
				}
			}
		}
	}
	return false
}

func (p *Parser) parseReceive() ast.Statement {
	channel := p.curToken.Literal

	p.nextToken() // move to ?
	recvToken := p.curToken

	// Check for variant receive: c ? CASE
	if p.peekTokenIs(lexer.CASE) {
		p.nextToken() // move to CASE
		return p.parseVariantReceive(channel, recvToken)
	}

	stmt := &ast.Receive{
		Channel: channel,
		Token:   recvToken,
	}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Variable = p.curToken.Literal

	// Check for sequential receive: c ? x ; y ; z
	for p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken() // move to ;
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.Variables = append(stmt.Variables, p.curToken.Literal)
	}

	return stmt
}

func (p *Parser) parseVariantReceive(channel string, token lexer.Token) *ast.VariantReceive {
	stmt := &ast.VariantReceive{
		Token:   token,
		Channel: channel,
	}

	// Skip to next line
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after ? CASE")
		return stmt
	}
	p.nextToken() // consume INDENT
	startLevel := p.indentLevel
	p.nextToken() // move into block

	// Parse variant cases (similar to parseCaseStatement pattern)
	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENT tokens
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				return stmt
			}
			p.nextToken()
		}

		// Skip any more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		if p.indentLevel < startLevel {
			break
		}

		// Safety guard: record position before parsing to detect no-progress
		prevToken := p.curToken
		prevPeek := p.peekToken

		// Parse a variant case: tag [; var]* \n INDENT body
		vc := ast.VariantCase{}

		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected variant tag name, got %s", p.curToken.Type))
			p.nextToken() // skip unrecognized token to avoid infinite loop
			continue
		}
		vc.Tag = p.curToken.Literal

		// Parse optional variables after semicolons: tag ; x ; y
		for p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // move to ;
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			vc.Variables = append(vc.Variables, p.curToken.Literal)
		}

		// Skip newlines and expect INDENT for body
		for p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.peekTokenIs(lexer.INDENT) {
			p.nextToken() // consume INDENT
			p.nextToken() // move to body
			vc.Body = p.parseStatement()

			// Advance past the last token of the statement if needed
			if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
				p.nextToken()
			}
		}

		stmt.Cases = append(stmt.Cases, vc)

		// No-progress guard: if we haven't moved, break to prevent infinite loop
		if p.curToken == prevToken && p.peekToken == prevPeek {
			p.nextToken() // force progress
			if p.curToken == prevToken {
				break
			}
		}
	}

	return stmt
}

func (p *Parser) parseVariantReceiveWithIndex(channel string, channelIndex ast.Expression, token lexer.Token) *ast.VariantReceive {
	stmt := &ast.VariantReceive{
		Token:        token,
		Channel:      channel,
		ChannelIndex: channelIndex,
	}

	// Skip to next line
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after ? CASE")
		return stmt
	}
	p.nextToken() // consume INDENT
	startLevel := p.indentLevel
	p.nextToken() // move into block

	for !p.curTokenIs(lexer.EOF) {
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				return stmt
			}
			p.nextToken()
		}

		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		if p.indentLevel < startLevel {
			break
		}

		// Safety guard: record position before parsing to detect no-progress
		prevToken := p.curToken
		prevPeek := p.peekToken

		vc := ast.VariantCase{}

		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected variant tag name, got %s", p.curToken.Type))
			p.nextToken() // skip unrecognized token to avoid infinite loop
			continue
		}
		vc.Tag = p.curToken.Literal

		for p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken() // move to ;
			if !p.expectPeek(lexer.IDENT) {
				return stmt
			}
			vc.Variables = append(vc.Variables, p.curToken.Literal)
		}

		for p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.peekTokenIs(lexer.INDENT) {
			p.nextToken() // consume INDENT
			p.nextToken() // move to body
			vc.Body = p.parseStatement()

			if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
				p.nextToken()
			}
		}

		stmt.Cases = append(stmt.Cases, vc)

		// No-progress guard: if we haven't moved, break to prevent infinite loop
		if p.curToken == prevToken && p.peekToken == prevPeek {
			p.nextToken() // force progress
			if p.curToken == prevToken {
				break
			}
		}
	}

	return stmt
}

func (p *Parser) parseSeqBlock() *ast.SeqBlock {
	block := &ast.SeqBlock{Token: p.curToken}

	// Check for replicator: SEQ i = start FOR count
	if p.peekTokenIs(lexer.IDENT) {
		// Save position to check if it's a replicator
		p.nextToken() // move to identifier
		if p.peekTokenIs(lexer.EQ) {
			// This is a replicator
			block.Replicator = p.parseReplicator()
		} else {
			// Not a replicator, this shouldn't happen in valid Occam
			// (SEQ followed by identifier at same indentation level)
			p.addError("unexpected identifier after SEQ")
			return block
		}
	}

	// Skip to next line
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after SEQ")
		return block
	}
	p.nextToken() // consume INDENT

	block.Statements = p.parseBlockStatements()

	return block
}

func (p *Parser) parseParBlock() *ast.ParBlock {
	block := &ast.ParBlock{Token: p.curToken}

	// Check for replicator: PAR i = start FOR count
	if p.peekTokenIs(lexer.IDENT) {
		// Save position to check if it's a replicator
		p.nextToken() // move to identifier
		if p.peekTokenIs(lexer.EQ) {
			// This is a replicator
			block.Replicator = p.parseReplicator()
		} else {
			// Not a replicator, this shouldn't happen in valid Occam
			// (PAR followed by identifier at same indentation level)
			p.addError("unexpected identifier after PAR")
			return block
		}
	}

	// Skip to next line
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after PAR")
		return block
	}
	p.nextToken() // consume INDENT

	block.Statements = p.parseBlockStatements()

	return block
}

// parseReplicator parses: variable = start FOR count [STEP step]
// Assumes the variable identifier has already been consumed and is in curToken
func (p *Parser) parseReplicator() *ast.Replicator {
	rep := &ast.Replicator{
		Variable: p.curToken.Literal,
	}

	// Expect =
	if !p.expectPeek(lexer.EQ) {
		return nil
	}

	// Parse start expression
	p.nextToken()
	rep.Start = p.parseExpression(LOWEST)

	// Expect FOR
	if !p.expectPeek(lexer.FOR) {
		return nil
	}

	// Parse count expression
	p.nextToken()
	rep.Count = p.parseExpression(LOWEST)

	// Optional STEP
	if p.peekTokenIs(lexer.STEP) {
		p.nextToken() // consume STEP
		p.nextToken() // move to step expression
		rep.Step = p.parseExpression(LOWEST)
	}

	return rep
}

func (p *Parser) parseAltBlock() *ast.AltBlock {
	block := &ast.AltBlock{Token: p.curToken}

	// Skip to next line
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after ALT")
		return block
	}
	p.nextToken() // consume INDENT

	block.Cases = p.parseAltCases()

	return block
}

func (p *Parser) parseAltCases() []ast.AltCase {
	var cases []ast.AltCase
	startLevel := p.indentLevel

	p.nextToken() // move past INDENT

	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENT tokens
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				return cases
			}
			p.nextToken()
		}

		// Skip any more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		if p.indentLevel < startLevel {
			break
		}

		// Safety guard: record position before parsing to detect no-progress
		prevToken := p.curToken
		prevPeek := p.peekToken

		// Parse an ALT case: [guard &] channel ? var
		altCase := p.parseAltCase()
		if altCase != nil {
			cases = append(cases, *altCase)
		}

		// No-progress guard: if we haven't moved, break to prevent infinite loop
		if p.curToken == prevToken && p.peekToken == prevPeek {
			break
		}
	}

	return cases
}

func (p *Parser) parseAltCase() *ast.AltCase {
	altCase := &ast.AltCase{}

	// Check for guard: expression & channel ? var
	// For now, we expect: channel ? var (no guard support yet)
	// or: guard & channel ? var

	// First token should be identifier (channel name or guard start)
	if !p.curTokenIs(lexer.IDENT) && !p.curTokenIs(lexer.TRUE) && !p.curTokenIs(lexer.FALSE) {
		p.addError(fmt.Sprintf("expected channel name or guard in ALT case, got %s", p.curToken.Type))
		return nil
	}

	// Look ahead to determine if this is a guard or channel
	// If next token is & then we have a guard
	// If next token is ? then it's a channel/timer receive
	if p.peekTokenIs(lexer.RECEIVE) {
		name := p.curToken.Literal
		if p.timerNames[name] {
			// Timer case: tim ? AFTER deadline
			altCase.IsTimer = true
			altCase.Timer = name
			p.nextToken() // move to ?
			if !p.expectPeek(lexer.AFTER) {
				return nil
			}
			p.nextToken() // move past AFTER
			altCase.Deadline = p.parseExpression(LOWEST)
		} else {
			// Simple case: channel ? var
			altCase.Channel = name
			p.nextToken() // move to ?
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			altCase.Variable = p.curToken.Literal
		}
	} else if p.peekTokenIs(lexer.LBRACKET) {
		// Indexed channel case: cs[i] ? var
		name := p.curToken.Literal
		altCase.Channel = name
		p.nextToken() // move to [
		p.nextToken() // move past [
		altCase.ChannelIndex = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
		if !p.expectPeek(lexer.RECEIVE) {
			return nil
		}
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		altCase.Variable = p.curToken.Literal
	} else {
		// Could be a guard followed by & channel ? var
		// For simplicity, parse expression until we hit &
		// For now, only support simple TRUE/FALSE or identifier guards
		guard := p.parseExpression(LOWEST)
		altCase.Guard = guard

		// Expect &
		if !p.peekTokenIs(lexer.AMPERSAND) {
			p.addError("expected & after guard in ALT case")
			return nil
		}
		p.nextToken() // move to &
		p.nextToken() // move past &

		// Now expect channel ? var or channel[index] ? var
		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected channel name after guard, got %s", p.curToken.Type))
			return nil
		}
		altCase.Channel = p.curToken.Literal

		if p.peekTokenIs(lexer.LBRACKET) {
			// Indexed channel with guard: guard & cs[i] ? var
			p.nextToken() // move to [
			p.nextToken() // move past [
			altCase.ChannelIndex = p.parseExpression(LOWEST)
			if !p.expectPeek(lexer.RBRACKET) {
				return nil
			}
		}

		if !p.expectPeek(lexer.RECEIVE) {
			return nil
		}
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		altCase.Variable = p.curToken.Literal
	}

	// Skip to next line for the body
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT for body
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented body after ALT case")
		return altCase
	}
	p.nextToken() // consume INDENT
	p.nextToken() // move into body

	altCase.Body = p.parseBodyStatements()

	return altCase
}

func (p *Parser) parseBlockStatements() []ast.Statement {
	var statements []ast.Statement
	startLevel := p.indentLevel

	p.nextToken() // move past INDENT

	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENT tokens
		// If we're at a DEDENT and indentLevel has dropped below startLevel,
		// this DEDENT ends our block - stop parsing
		// If indentLevel >= startLevel, this DEDENT is from a nested block - skip it
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				return statements
			}
			p.nextToken() // skip nested block's DEDENT
		}

		// Skip any more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		// Double-check we haven't gone below our level
		if p.indentLevel < startLevel {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}

		// After parsing a statement, we need to advance.
		// But if we're already at NEWLINE/DEDENT/EOF, the next iteration will handle it.
		// Only advance if we're still on the last token of the statement.
		if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}

	return statements
}

// parseBodyStatements parses multiple statements inside a branch body
// (IF choice, CASE choice, ALT case, WHILE). Called after the caller has
// consumed the INDENT token and advanced into the body.
// Returns all statements found at this indentation level.
func (p *Parser) parseBodyStatements() []ast.Statement {
	var statements []ast.Statement
	startLevel := p.indentLevel

	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENT tokens
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				return statements
			}
			p.nextToken()
		}

		// Skip any more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		if p.indentLevel < startLevel {
			break
		}

		// Safety guard: record position before parsing to detect no-progress
		prevToken := p.curToken
		prevPeek := p.peekToken

		stmt := p.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}

		// Advance past the last token of the statement if needed
		if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}

		// No-progress guard: if we haven't moved, break to prevent infinite loop
		if p.curToken == prevToken && p.peekToken == prevPeek {
			break
		}
	}

	return statements
}

func (p *Parser) parseProcDecl() *ast.ProcDecl {
	proc := &ast.ProcDecl{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	proc.Name = p.curToken.Literal

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	proc.Params = p.parseProcParams()

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// Skip to next line and expect indented body
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented body after PROC declaration")
		return proc
	}
	p.nextToken() // consume INDENT

	// Parse all statements in the body (local declarations + body process)
	bodyLevel := p.indentLevel
	p.nextToken()

	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENTs
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < bodyLevel {
				goto procBodyDone
			}
			p.nextToken()
		}

		// Skip more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) || p.indentLevel < bodyLevel {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			proc.Body = append(proc.Body, stmt)
		}

		if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
			p.nextToken()
		}
	}
procBodyDone:

	// Optionally consume KRoC-style colon terminator
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
	}

	return proc
}

// isTypeToken returns true if the token type is a scalar type keyword.
func isTypeToken(t lexer.TokenType) bool {
	return t == lexer.INT_TYPE || t == lexer.BYTE_TYPE ||
		t == lexer.BOOL_TYPE || t == lexer.REAL_TYPE ||
		t == lexer.REAL32_TYPE || t == lexer.REAL64_TYPE
}

func (p *Parser) parseProcParams() []ast.ProcParam {
	var params []ast.ProcParam

	if p.peekTokenIs(lexer.RPAREN) {
		return params
	}

	p.nextToken()

	// Track the previous param's type info for shared-type parameters
	var prevParam *ast.ProcParam

	for {
		// Skip newlines inside parameter lists (multi-line params)
		// Note: INDENT/DEDENT/NEWLINE inside (...) are suppressed by the lexer
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		param := ast.ProcParam{}

		// Check if this is a shared-type parameter: after a comma, if current token
		// is an IDENT that is NOT a type keyword, record name, CHAN, VAL, RESULT, or [,
		// re-use the previous param's type/flags.
		if prevParam != nil && p.curTokenIs(lexer.IDENT) && !p.recordNames[p.curToken.Literal] {
			// This is a shared-type param — re-use type info from previous param
			param.IsVal = prevParam.IsVal
			param.Type = prevParam.Type
			param.IsChan = prevParam.IsChan
			param.IsChanArray = prevParam.IsChanArray
			param.IsOpenArray = prevParam.IsOpenArray
			param.ChanElemType = prevParam.ChanElemType
			param.ArraySize = prevParam.ArraySize
			param.Name = p.curToken.Literal

			// Check for channel direction marker (? or !)
			if (param.IsChan || param.IsChanArray) && (p.peekTokenIs(lexer.RECEIVE) || p.peekTokenIs(lexer.SEND)) {
				p.nextToken()
				param.ChanDir = p.curToken.Literal
			}

			params = append(params, param)
			prevParam = &params[len(params)-1]

			if !p.peekTokenIs(lexer.COMMA) {
				break
			}
			p.nextToken() // consume comma
			p.nextToken() // move to next param
			continue
		}

		// Check for VAL keyword
		if p.curTokenIs(lexer.VAL) {
			param.IsVal = true
			p.nextToken()
		}

		// Check for RESULT keyword (output-only parameter — maps to pointer like non-VAL)
		if p.curTokenIs(lexer.RESULT) {
			// RESULT is semantically like non-VAL (pointer param), just skip it
			p.nextToken()
		}

		// Check for []CHAN OF <type>, []TYPE (open array), or [n]TYPE (fixed-size array)
		if p.curTokenIs(lexer.LBRACKET) {
			if p.peekTokenIs(lexer.RBRACKET) {
				// Open array: []CHAN OF TYPE or []TYPE
				p.nextToken() // consume ]
				p.nextToken() // move past ]
				if p.curTokenIs(lexer.CHAN) {
					// []CHAN OF <type> or []CHAN <type> (channel array parameter)
					param.IsChan = true
					param.IsChanArray = true
					if p.peekTokenIs(lexer.OF) {
						p.nextToken() // consume OF
					}
					p.nextToken() // move to element type
					if isTypeToken(p.curToken.Type) || p.curTokenIs(lexer.IDENT) {
						param.ChanElemType = p.curToken.Literal
					} else {
						p.addError(fmt.Sprintf("expected type after []CHAN, got %s", p.curToken.Type))
						return params
					}
					p.nextToken()
				} else if isTypeToken(p.curToken.Type) {
					param.IsOpenArray = true
					param.Type = p.curToken.Literal
					p.nextToken()
				} else if p.curTokenIs(lexer.IDENT) && p.recordNames[p.curToken.Literal] {
					param.IsOpenArray = true
					param.Type = p.curToken.Literal
					p.nextToken()
				} else {
					p.addError(fmt.Sprintf("expected type after [], got %s", p.curToken.Type))
					return params
				}
			} else {
				// Fixed-size array: [n]TYPE
				p.nextToken() // move past [
				if !p.curTokenIs(lexer.INT) {
					p.addError(fmt.Sprintf("expected array size, got %s", p.curToken.Type))
					return params
				}
				param.ArraySize = p.curToken.Literal
				if !p.expectPeek(lexer.RBRACKET) {
					return params
				}
				p.nextToken() // move to type
				if isTypeToken(p.curToken.Type) {
					param.Type = p.curToken.Literal
				} else if p.curTokenIs(lexer.IDENT) && p.recordNames[p.curToken.Literal] {
					param.Type = p.curToken.Literal
				} else {
					p.addError(fmt.Sprintf("expected type after [%s], got %s", param.ArraySize, p.curToken.Type))
					return params
				}
				p.nextToken()
			}
		} else if p.curTokenIs(lexer.CHAN) {
			// Check for CHAN OF <type> or CHAN <type>
			param.IsChan = true
			if p.peekTokenIs(lexer.OF) {
				p.nextToken() // consume OF
			}
			p.nextToken() // move to element type
			if isTypeToken(p.curToken.Type) || p.curTokenIs(lexer.IDENT) {
				param.ChanElemType = p.curToken.Literal
			} else {
				p.addError(fmt.Sprintf("expected type after CHAN, got %s", p.curToken.Type))
				return params
			}
			p.nextToken()
		} else if p.curTokenIs(lexer.IDENT) && p.recordNames[p.curToken.Literal] {
			// Record type parameter
			param.Type = p.curToken.Literal
			p.nextToken()
		} else {
			// Expect scalar type
			if !isTypeToken(p.curToken.Type) {
				p.addError(fmt.Sprintf("expected type in parameter, got %s", p.curToken.Type))
				return params
			}
			param.Type = p.curToken.Literal
			p.nextToken()
		}

		// Expect identifier
		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected parameter name, got %s", p.curToken.Type))
			return params
		}
		param.Name = p.curToken.Literal

		// Check for channel direction marker (? or !)
		if (param.IsChan || param.IsChanArray) && (p.peekTokenIs(lexer.RECEIVE) || p.peekTokenIs(lexer.SEND)) {
			p.nextToken()
			param.ChanDir = p.curToken.Literal
		}

		params = append(params, param)
		prevParam = &params[len(params)-1]

		if !p.peekTokenIs(lexer.COMMA) {
			break
		}
		p.nextToken() // consume comma
		p.nextToken() // move to next param
	}

	return params
}

func (p *Parser) parseProcCall() *ast.ProcCall {
	call := &ast.ProcCall{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	if !p.peekTokenIs(lexer.LPAREN) {
		// No arguments
		return call
	}

	p.nextToken() // consume (

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken() // consume )
		return call
	}

	p.nextToken() // move to first arg
	call.Args = append(call.Args, p.parseExpression(LOWEST))
	// Consume optional channel direction annotation at call site (e.g., out!)
	if p.peekTokenIs(lexer.SEND) || p.peekTokenIs(lexer.RECEIVE) {
		p.nextToken()
	}

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next arg
		call.Args = append(call.Args, p.parseExpression(LOWEST))
		// Consume optional channel direction annotation at call site
		if p.peekTokenIs(lexer.SEND) || p.peekTokenIs(lexer.RECEIVE) {
			p.nextToken()
		}
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return call
}

func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	fn := &ast.FuncDecl{
		Token:       p.curToken,
		ReturnTypes: []string{p.curToken.Literal},
	}

	// Parse additional return types for multi-result functions: INT, INT FUNCTION
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next type
		fn.ReturnTypes = append(fn.ReturnTypes, p.curToken.Literal)
	}

	// Consume FUNCTION keyword
	p.nextToken()

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	fn.Name = p.curToken.Literal

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	fn.Params = p.parseProcParams()

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	// Force all params to IsVal = true (occam FUNCTION params are always VAL)
	for i := range fn.Params {
		fn.Params[i].IsVal = true
	}

	// Skip newlines, expect INDENT
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented body after FUNCTION declaration")
		return fn
	}
	funcLevel := p.indentLevel
	p.nextToken() // consume INDENT
	p.nextToken() // move into body

	// IS form: simple expression return
	if p.curTokenIs(lexer.IS) {
		p.nextToken() // move past IS
		fn.ResultExprs = []ast.Expression{p.parseExpression(LOWEST)}

		// Consume remaining tokens and DEDENTs back to function's indentation level
		for !p.curTokenIs(lexer.EOF) {
			if p.curTokenIs(lexer.DEDENT) && p.indentLevel <= funcLevel {
				break
			}
			p.nextToken()
		}

		// Optionally consume KRoC-style colon terminator
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken()
		}
		return fn
	}

	// VALOF form: local declarations, then VALOF keyword, then body, then RESULT
	// Parse local declarations (type keywords before VALOF)
	for p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BYTE_TYPE) ||
		p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.REAL_TYPE) ||
		p.curTokenIs(lexer.REAL32_TYPE) || p.curTokenIs(lexer.REAL64_TYPE) {
		stmt := p.parseVarDecl()
		if stmt != nil {
			fn.Body = append(fn.Body, stmt)
		}
		// Advance past NEWLINE
		for p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}
		p.nextToken()
	}

	// Expect VALOF keyword
	if !p.curTokenIs(lexer.VALOF) {
		p.addError(fmt.Sprintf("expected VALOF or IS in function body, got %s", p.curToken.Type))
		return fn
	}

	// Skip newlines and expect INDENT for VALOF body
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after VALOF")
		return fn
	}
	p.nextToken() // consume INDENT
	startLevel := p.indentLevel
	p.nextToken() // move into VALOF body

	// Parse the VALOF body — declarations and statements until RESULT
	for !p.curTokenIs(lexer.RESULT) && !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}
		// Handle DEDENTs
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				break
			}
			p.nextToken()
		}
		if p.curTokenIs(lexer.EOF) || p.curTokenIs(lexer.RESULT) {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			fn.Body = append(fn.Body, stmt)
		}
		if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) && !p.curTokenIs(lexer.RESULT) {
			p.nextToken()
		}
	}

	// Parse RESULT expression(s) — comma-separated for multi-result functions
	if p.curTokenIs(lexer.RESULT) {
		p.nextToken() // move past RESULT
		fn.ResultExprs = []ast.Expression{p.parseExpression(LOWEST)}
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to next expression
			fn.ResultExprs = append(fn.ResultExprs, p.parseExpression(LOWEST))
		}
	}

	// Consume remaining tokens and DEDENTs back to function's indentation level
	for !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.DEDENT) && p.indentLevel <= funcLevel {
			break
		}
		p.nextToken()
	}

	// Optionally consume KRoC-style colon terminator
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
	}

	return fn
}

// convertOccamStringEscapes converts occam escape sequences in string literals
// to their actual byte values. Occam uses *c, *n, *t, *s, **, *", *' as escapes.
func (p *Parser) convertOccamStringEscapes(raw string) string {
	var buf strings.Builder
	buf.Grow(len(raw))
	for i := 0; i < len(raw); i++ {
		if raw[i] == '*' && i+1 < len(raw) {
			i++
			switch raw[i] {
			case 'n':
				buf.WriteByte('\n')
			case 'c':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 's':
				buf.WriteByte(' ')
			case '*':
				buf.WriteByte('*')
			case '"':
				buf.WriteByte('"')
			case '\'':
				buf.WriteByte('\'')
			default:
				// Unknown escape: pass through as-is
				buf.WriteByte('*')
				buf.WriteByte(raw[i])
			}
		} else {
			buf.WriteByte(raw[i])
		}
	}
	return buf.String()
}

// parseByteLiteralValue processes the raw content of a byte literal (between single quotes),
// handling occam escape sequences (* prefix), and returns the resulting byte value.
func (p *Parser) parseByteLiteralValue(raw string) (byte, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("empty byte literal")
	}
	if raw[0] == '*' {
		if len(raw) != 2 {
			return 0, fmt.Errorf("invalid escape sequence in byte literal: '*%s'", raw[1:])
		}
		switch raw[1] {
		case 'n':
			return '\n', nil
		case 'c':
			return '\r', nil
		case 't':
			return '\t', nil
		case 's':
			return ' ', nil
		case '*':
			return '*', nil
		case '\'':
			return '\'', nil
		case '"':
			return '"', nil
		default:
			return 0, fmt.Errorf("unknown escape sequence in byte literal: '*%c'", raw[1])
		}
	}
	if len(raw) != 1 {
		return 0, fmt.Errorf("byte literal must be a single character, got %q", raw)
	}
	return raw[0], nil
}

func (p *Parser) parseFuncCallExpr() *ast.FuncCall {
	call := &ast.FuncCall{
		Token: p.curToken,
		Name:  p.curToken.Literal,
	}

	p.nextToken() // consume (

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken() // consume )
		return call
	}

	p.nextToken() // move to first arg
	call.Args = append(call.Args, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next arg
		call.Args = append(call.Args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return call
}

func (p *Parser) parseWhileLoop() *ast.WhileLoop {
	loop := &ast.WhileLoop{Token: p.curToken}

	p.nextToken()
	loop.Condition = p.parseExpression(LOWEST)

	// Skip to next line
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after WHILE condition")
		return loop
	}
	p.nextToken() // consume INDENT
	p.nextToken() // move to first statement

	loop.Body = p.parseBodyStatements()

	return loop
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	// Check for replicator: IF i = start FOR count
	if p.peekTokenIs(lexer.IDENT) {
		p.nextToken() // move to identifier
		if p.peekTokenIs(lexer.EQ) {
			stmt.Replicator = p.parseReplicator()
		} else {
			p.addError("unexpected identifier after IF")
			return stmt
		}
	}

	// Skip to next line
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after IF")
		return stmt
	}
	p.nextToken() // consume INDENT
	startLevel := p.indentLevel
	p.nextToken() // move into block

	// Parse if choices (condition -> body pairs)
	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENT tokens
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				return stmt
			}
			p.nextToken()
		}

		// Skip any more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		if p.indentLevel < startLevel {
			break
		}

		// Safety guard: record position before parsing to detect no-progress
		prevToken := p.curToken
		prevPeek := p.peekToken

		choice := ast.IfChoice{}

		// Nested IF (plain or replicated) used as a choice within this IF
		if p.curTokenIs(lexer.IF) {
			nestedIf := p.parseIfStatement()
			choice.NestedIf = nestedIf
		} else {
			choice.Condition = p.parseExpression(LOWEST)

			// Skip newlines and expect INDENT for body
			for p.peekTokenIs(lexer.NEWLINE) {
				p.nextToken()
			}

			if p.peekTokenIs(lexer.INDENT) {
				p.nextToken() // consume INDENT
				p.nextToken() // move to body
				choice.Body = p.parseBodyStatements()
			}
		}

		stmt.Choices = append(stmt.Choices, choice)

		// No-progress guard: if we haven't moved, break to prevent infinite loop
		if p.curToken == prevToken && p.peekToken == prevPeek {
			break
		}
	}

	return stmt
}

func (p *Parser) parseCaseStatement() *ast.CaseStatement {
	stmt := &ast.CaseStatement{Token: p.curToken}

	// Parse selector expression on the same line
	p.nextToken()
	stmt.Selector = p.parseExpression(LOWEST)

	// Skip to next line
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Expect INDENT
	if !p.peekTokenIs(lexer.INDENT) {
		p.addError("expected indented block after CASE")
		return stmt
	}
	p.nextToken() // consume INDENT
	startLevel := p.indentLevel
	p.nextToken() // move into block

	// Parse case choices
	for !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		// Handle DEDENT tokens
		for p.curTokenIs(lexer.DEDENT) {
			if p.indentLevel < startLevel {
				return stmt
			}
			p.nextToken()
		}

		// Skip any more newlines after DEDENT
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.EOF) {
			break
		}

		if p.indentLevel < startLevel {
			break
		}

		// Safety guard: record position before parsing to detect no-progress
		prevToken := p.curToken
		prevPeek := p.peekToken

		choice := ast.CaseChoice{}

		if p.curTokenIs(lexer.ELSE) {
			choice.IsElse = true
		} else {
			// Parse value expression
			choice.Values = append(choice.Values, p.parseExpression(LOWEST))
		}

		// Skip newlines and expect INDENT for body
		for p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.peekTokenIs(lexer.INDENT) {
			p.nextToken() // consume INDENT
			p.nextToken() // move to body
			choice.Body = p.parseBodyStatements()
		}

		stmt.Choices = append(stmt.Choices, choice)

		// No-progress guard: if we haven't moved, break to prevent infinite loop
		if p.curToken == prevToken && p.peekToken == prevPeek {
			break
		}
	}

	return stmt
}

// Expression parsing using Pratt parsing

func (p *Parser) parseExpression(precedence int) ast.Expression {
	var left ast.Expression

	switch p.curToken.Type {
	case lexer.IDENT:
		if p.peekTokenIs(lexer.LPAREN) {
			left = p.parseFuncCallExpr()
		} else {
			left = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		}
	case lexer.INT:
		base := 10
		literal := p.curToken.Literal
		if strings.HasPrefix(literal, "0x") || strings.HasPrefix(literal, "0X") {
			base = 16
			literal = literal[2:]
		}
		val, err := strconv.ParseInt(literal, base, 64)
		if err != nil {
			p.addError(fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
			return nil
		}
		left = &ast.IntegerLiteral{Token: p.curToken, Value: val}
	case lexer.TRUE:
		left = &ast.BooleanLiteral{Token: p.curToken, Value: true}
	case lexer.FALSE:
		left = &ast.BooleanLiteral{Token: p.curToken, Value: false}
	case lexer.STRING:
		left = &ast.StringLiteral{Token: p.curToken, Value: p.convertOccamStringEscapes(p.curToken.Literal)}
	case lexer.BYTE_LIT:
		b, err := p.parseByteLiteralValue(p.curToken.Literal)
		if err != nil {
			p.addError(err.Error())
			return nil
		}
		left = &ast.ByteLiteral{Token: p.curToken, Value: b}
	case lexer.LPAREN:
		p.nextToken()
		left = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	case lexer.MINUS, lexer.MINUS_KW:
		token := p.curToken
		p.nextToken()
		left = &ast.UnaryExpr{
			Token:    token,
			Operator: "-",
			Right:    p.parseExpression(PREFIX),
		}
	case lexer.NOT:
		token := p.curToken
		p.nextToken()
		left = &ast.UnaryExpr{
			Token:    token,
			Operator: "NOT",
			Right:    p.parseExpression(PREFIX),
		}
	case lexer.BITNOT:
		token := p.curToken
		p.nextToken()
		left = &ast.UnaryExpr{
			Token:    token,
			Operator: "~",
			Right:    p.parseExpression(PREFIX),
		}
	case lexer.LBRACKET:
		// Could be: [arr FROM start FOR length], [arr FOR length], or [expr, expr, ...] array literal
		lbracket := p.curToken
		p.nextToken() // move past [
		firstExpr := p.parseExpression(LOWEST)

		if p.peekTokenIs(lexer.COMMA) {
			// Array literal: [expr, expr, ...]
			elements := []ast.Expression{firstExpr}
			for p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // consume comma
				p.nextToken() // move to next element
				elements = append(elements, p.parseExpression(LOWEST))
			}
			if !p.expectPeek(lexer.RBRACKET) {
				return nil
			}
			left = &ast.ArrayLiteral{
				Token:    lbracket,
				Elements: elements,
			}
		} else if p.peekTokenIs(lexer.RBRACKET) {
			// Single-element array literal: [expr]
			p.nextToken() // consume ]
			left = &ast.ArrayLiteral{
				Token:    lbracket,
				Elements: []ast.Expression{firstExpr},
			}
		} else {
			// Slice expression: [arr FROM start FOR length] or [arr FOR length]
			var startExpr ast.Expression
			if p.peekTokenIs(lexer.FOR) {
				// [arr FOR length] shorthand — start is 0
				startExpr = &ast.IntegerLiteral{Token: lexer.Token{Type: lexer.INT, Literal: "0"}, Value: 0}
			} else {
				if !p.expectPeek(lexer.FROM) {
					return nil
				}
				p.nextToken() // move past FROM
				startExpr = p.parseExpression(LOWEST)
			}
			if !p.expectPeek(lexer.FOR) {
				return nil
			}
			p.nextToken() // move past FOR
			lengthExpr := p.parseExpression(LOWEST)
			if !p.expectPeek(lexer.RBRACKET) {
				return nil
			}
			left = &ast.SliceExpr{
				Token:  lbracket,
				Array:  firstExpr,
				Start:  startExpr,
				Length: lengthExpr,
			}
		}
	case lexer.SIZE_KW:
		token := p.curToken
		p.nextToken()
		left = &ast.SizeExpr{
			Token: token,
			Expr:  p.parseExpression(PREFIX),
		}
	case lexer.MOSTNEG_KW, lexer.MOSTPOS_KW:
		token := p.curToken
		isNeg := token.Type == lexer.MOSTNEG_KW
		// Expect a type name next
		if !p.peekTokenIs(lexer.INT_TYPE) && !p.peekTokenIs(lexer.BYTE_TYPE) &&
			!p.peekTokenIs(lexer.BOOL_TYPE) && !p.peekTokenIs(lexer.REAL_TYPE) &&
			!p.peekTokenIs(lexer.REAL32_TYPE) && !p.peekTokenIs(lexer.REAL64_TYPE) {
			p.addError(fmt.Sprintf("expected type after %s, got %s", token.Literal, p.peekToken.Type))
			return nil
		}
		p.nextToken()
		left = &ast.MostExpr{
			Token:    token,
			ExprType: p.curToken.Literal,
			IsNeg:    isNeg,
		}
	case lexer.INT_TYPE, lexer.BYTE_TYPE, lexer.BOOL_TYPE, lexer.REAL_TYPE, lexer.REAL32_TYPE, lexer.REAL64_TYPE:
		token := p.curToken
		p.nextToken()
		left = &ast.TypeConversion{
			Token:      token,
			TargetType: token.Literal,
			Expr:       p.parseExpression(PREFIX),
		}
	default:
		p.addError(fmt.Sprintf("unexpected token in expression: %s", p.curToken.Type))
		return nil
	}

	// Parse infix expressions
	for !p.peekTokenIs(lexer.NEWLINE) && !p.peekTokenIs(lexer.EOF) &&
		precedence < p.peekPrecedence() {

		switch p.peekToken.Type {
		case lexer.PLUS, lexer.MINUS, lexer.MULTIPLY, lexer.DIVIDE, lexer.MODULO,
			lexer.PLUS_KW, lexer.MINUS_KW, lexer.TIMES,
			lexer.EQ, lexer.NEQ, lexer.LT, lexer.GT, lexer.LE, lexer.GE,
			lexer.AND, lexer.OR, lexer.AFTER,
			lexer.BITAND, lexer.BITOR, lexer.BITXOR, lexer.LSHIFT, lexer.RSHIFT:
			p.nextToken()
			left = p.parseBinaryExpr(left)
		case lexer.LBRACKET:
			p.nextToken()
			left = p.parseIndexExpression(left)
		default:
			return left
		}
	}

	return left
}

func (p *Parser) parseBinaryExpr(left ast.Expression) ast.Expression {
	expr := &ast.BinaryExpr{
		Token:    p.curToken,
		Left:     left,
		Operator: p.curToken.Literal,
	}

	prec := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(prec)

	return expr
}
