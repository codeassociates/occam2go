package parser

import (
	"fmt"
	"strconv"

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
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.MULTIPLY: PRODUCT,
	lexer.DIVIDE:   PRODUCT,
	lexer.MODULO:   PRODUCT,
	lexer.LBRACKET: INDEX,
}

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	// Track current indentation level
	indentLevel int
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
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
	case lexer.INT_TYPE, lexer.BYTE_TYPE, lexer.BOOL_TYPE, lexer.REAL_TYPE:
		return p.parseVarDecl()
	case lexer.LBRACKET:
		return p.parseArrayDecl()
	case lexer.CHAN:
		return p.parseChanDecl()
	case lexer.SEQ:
		return p.parseSeqBlock()
	case lexer.PAR:
		return p.parseParBlock()
	case lexer.ALT:
		return p.parseAltBlock()
	case lexer.SKIP:
		return &ast.Skip{Token: p.curToken}
	case lexer.PROC:
		return p.parseProcDecl()
	case lexer.WHILE:
		return p.parseWhileLoop()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.IDENT:
		// Could be assignment, indexed assignment, send, receive, or procedure call
		if p.peekTokenIs(lexer.LBRACKET) {
			return p.parseIndexedAssignment()
		}
		if p.peekTokenIs(lexer.ASSIGN) {
			return p.parseAssignment()
		}
		if p.peekTokenIs(lexer.SEND) {
			return p.parseSend()
		}
		if p.peekTokenIs(lexer.RECEIVE) {
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

func (p *Parser) parseArrayDecl() *ast.ArrayDecl {
	decl := &ast.ArrayDecl{Token: p.curToken}

	// Parse size expression after [
	p.nextToken()
	decl.Size = p.parseExpression(LOWEST)

	// Expect ]
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	// Expect type (INT, BYTE, BOOL, REAL)
	p.nextToken()
	if !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BYTE_TYPE) &&
		!p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.REAL_TYPE) {
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

func (p *Parser) parseIndexedAssignment() *ast.Assignment {
	stmt := &ast.Assignment{
		Name: p.curToken.Literal,
	}

	p.nextToken() // move to [
	p.nextToken() // move past [
	stmt.Index = p.parseExpression(LOWEST)

	// Expect ]
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	// Expect :=
	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}
	stmt.Token = p.curToken

	p.nextToken() // move past :=
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
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

	// Expect OF
	if !p.expectPeek(lexer.OF) {
		return nil
	}

	// Expect type (INT, BYTE, BOOL, etc.)
	p.nextToken()
	if !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BYTE_TYPE) &&
		!p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.REAL_TYPE) {
		p.addError(fmt.Sprintf("expected type after CHAN OF, got %s", p.curToken.Type))
		return nil
	}
	decl.ElemType = p.curToken.Literal

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

func (p *Parser) parseSend() *ast.Send {
	stmt := &ast.Send{
		Channel: p.curToken.Literal,
	}

	p.nextToken() // move to !
	stmt.Token = p.curToken

	p.nextToken() // move past !
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseReceive() *ast.Receive {
	stmt := &ast.Receive{
		Channel: p.curToken.Literal,
	}

	p.nextToken() // move to ?
	stmt.Token = p.curToken

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Variable = p.curToken.Literal

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

// parseReplicator parses: variable = start FOR count
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

		// Parse an ALT case: [guard &] channel ? var
		altCase := p.parseAltCase()
		if altCase != nil {
			cases = append(cases, *altCase)
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
	// If next token is ? then it's a channel receive
	if p.peekTokenIs(lexer.RECEIVE) {
		// Simple case: channel ? var
		altCase.Channel = p.curToken.Literal
		p.nextToken() // move to ?
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

		// Now expect channel ? var
		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected channel name after guard, got %s", p.curToken.Type))
			return nil
		}
		altCase.Channel = p.curToken.Literal

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

	// Parse the body (first statement)
	altCase.Body = p.parseStatement()

	// Skip to end of body block
	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

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

	// Parse the procedure body (first statement in the indented block)
	p.nextToken()
	proc.Body = p.parseStatement()

	// Consume remaining statements at this level and the DEDENT
	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return proc
}

func (p *Parser) parseProcParams() []ast.ProcParam {
	var params []ast.ProcParam

	if p.peekTokenIs(lexer.RPAREN) {
		return params
	}

	p.nextToken()

	for {
		param := ast.ProcParam{}

		// Check for VAL keyword
		if p.curTokenIs(lexer.VAL) {
			param.IsVal = true
			p.nextToken()
		}

		// Expect type
		if !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BYTE_TYPE) &&
			!p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.REAL_TYPE) {
			p.addError(fmt.Sprintf("expected type in parameter, got %s", p.curToken.Type))
			return params
		}
		param.Type = p.curToken.Literal
		p.nextToken()

		// Expect identifier
		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected parameter name, got %s", p.curToken.Type))
			return params
		}
		param.Name = p.curToken.Literal

		params = append(params, param)

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

	loop.Body = p.parseStatement()

	// Consume until DEDENT
	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return loop
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

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
	p.nextToken() // move into block

	// Parse if choices (condition -> body pairs)
	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		// Skip newlines
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.curTokenIs(lexer.DEDENT) {
			break
		}

		choice := ast.IfChoice{}
		choice.Condition = p.parseExpression(LOWEST)

		// Skip newlines and expect INDENT for body
		for p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}

		if p.peekTokenIs(lexer.INDENT) {
			p.nextToken() // consume INDENT
			p.nextToken() // move to body
			choice.Body = p.parseStatement()

			// Consume until DEDENT
			for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
				p.nextToken()
			}
		}

		stmt.Choices = append(stmt.Choices, choice)
		p.nextToken()
	}

	return stmt
}

// Expression parsing using Pratt parsing

func (p *Parser) parseExpression(precedence int) ast.Expression {
	var left ast.Expression

	switch p.curToken.Type {
	case lexer.IDENT:
		left = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.INT:
		val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
		if err != nil {
			p.addError(fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
			return nil
		}
		left = &ast.IntegerLiteral{Token: p.curToken, Value: val}
	case lexer.TRUE:
		left = &ast.BooleanLiteral{Token: p.curToken, Value: true}
	case lexer.FALSE:
		left = &ast.BooleanLiteral{Token: p.curToken, Value: false}
	case lexer.LPAREN:
		p.nextToken()
		left = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	case lexer.MINUS:
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
	default:
		p.addError(fmt.Sprintf("unexpected token in expression: %s", p.curToken.Type))
		return nil
	}

	// Parse infix expressions
	for !p.peekTokenIs(lexer.NEWLINE) && !p.peekTokenIs(lexer.EOF) &&
		precedence < p.peekPrecedence() {

		switch p.peekToken.Type {
		case lexer.PLUS, lexer.MINUS, lexer.MULTIPLY, lexer.DIVIDE, lexer.MODULO,
			lexer.EQ, lexer.NEQ, lexer.LT, lexer.GT, lexer.LE, lexer.GE,
			lexer.AND, lexer.OR:
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
