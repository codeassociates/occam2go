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
	lexer.AFTER:    LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.MULTIPLY: PRODUCT,
	lexer.DIVIDE:   PRODUCT,
	lexer.MODULO:   PRODUCT,
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
	case lexer.INT_TYPE, lexer.BYTE_TYPE, lexer.BOOL_TYPE, lexer.REAL_TYPE, lexer.REAL32_TYPE, lexer.REAL64_TYPE:
		if p.peekTokenIs(lexer.FUNCTION) || p.peekTokenIs(lexer.FUNC) {
			return p.parseFuncDecl()
		}
		return p.parseVarDecl()
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

func (p *Parser) parseArrayDecl() ast.Statement {
	lbracketToken := p.curToken

	// Parse size expression after [
	p.nextToken()
	size := p.parseExpression(LOWEST)

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

		// Expect OF
		if !p.expectPeek(lexer.OF) {
			return nil
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
			p.addError(fmt.Sprintf("expected type after CHAN OF, got %s", p.curToken.Type))
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

	// Expect OF
	if !p.expectPeek(lexer.OF) {
		return nil
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
		p.addError(fmt.Sprintf("expected type after CHAN OF, got %s", p.curToken.Type))
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

		// Parse a variant case: tag [; var]* \n INDENT body
		vc := ast.VariantCase{}

		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected variant tag name, got %s", p.curToken.Type))
			return stmt
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

		vc := ast.VariantCase{}

		if !p.curTokenIs(lexer.IDENT) {
			p.addError(fmt.Sprintf("expected variant tag name, got %s", p.curToken.Type))
			return stmt
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

		// Check for []CHAN OF <type> (channel array parameter)
		if p.curTokenIs(lexer.LBRACKET) && p.peekTokenIs(lexer.RBRACKET) {
			p.nextToken() // consume ]
			if !p.expectPeek(lexer.CHAN) {
				return params
			}
			param.IsChan = true
			param.IsChanArray = true
			if !p.expectPeek(lexer.OF) {
				return params
			}
			p.nextToken() // move to element type
			if p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BYTE_TYPE) ||
				p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.REAL_TYPE) ||
				p.curTokenIs(lexer.REAL32_TYPE) || p.curTokenIs(lexer.REAL64_TYPE) {
				param.ChanElemType = p.curToken.Literal
			} else if p.curTokenIs(lexer.IDENT) {
				param.ChanElemType = p.curToken.Literal
			} else {
				p.addError(fmt.Sprintf("expected type after []CHAN OF, got %s", p.curToken.Type))
				return params
			}
			p.nextToken()
		} else if p.curTokenIs(lexer.CHAN) {
			// Check for CHAN OF <type>
			param.IsChan = true
			if !p.expectPeek(lexer.OF) {
				return params
			}
			p.nextToken() // move to element type
			if p.curTokenIs(lexer.INT_TYPE) || p.curTokenIs(lexer.BYTE_TYPE) ||
				p.curTokenIs(lexer.BOOL_TYPE) || p.curTokenIs(lexer.REAL_TYPE) ||
				p.curTokenIs(lexer.REAL32_TYPE) || p.curTokenIs(lexer.REAL64_TYPE) {
				param.ChanElemType = p.curToken.Literal
			} else if p.curTokenIs(lexer.IDENT) {
				param.ChanElemType = p.curToken.Literal
			} else {
				p.addError(fmt.Sprintf("expected type after CHAN OF, got %s", p.curToken.Type))
				return params
			}
			p.nextToken()
		} else if p.curTokenIs(lexer.IDENT) && p.recordNames[p.curToken.Literal] {
			// Record type parameter
			param.Type = p.curToken.Literal
			p.nextToken()
		} else {
			// Expect scalar type
			if !p.curTokenIs(lexer.INT_TYPE) && !p.curTokenIs(lexer.BYTE_TYPE) &&
				!p.curTokenIs(lexer.BOOL_TYPE) && !p.curTokenIs(lexer.REAL_TYPE) &&
				!p.curTokenIs(lexer.REAL32_TYPE) && !p.curTokenIs(lexer.REAL64_TYPE) {
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

func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	fn := &ast.FuncDecl{
		Token:      p.curToken,
		ReturnType: p.curToken.Literal,
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
	p.nextToken() // consume INDENT
	p.nextToken() // move into body

	// IS form: simple expression return
	if p.curTokenIs(lexer.IS) {
		p.nextToken() // move past IS
		fn.ResultExpr = p.parseExpression(LOWEST)

		// Consume to DEDENT
		for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
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
	p.nextToken() // move into VALOF body

	// Parse the body statement (e.g., SEQ, IF, etc.)
	bodyStmt := p.parseStatement()
	if bodyStmt != nil {
		fn.Body = append(fn.Body, bodyStmt)
	}

	// Advance past nested DEDENTs/newlines to RESULT
	for !p.curTokenIs(lexer.RESULT) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	// Parse RESULT expression
	if p.curTokenIs(lexer.RESULT) {
		p.nextToken() // move past RESULT
		fn.ResultExpr = p.parseExpression(LOWEST)
	}

	// Consume to the function's DEDENT
	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return fn
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

			// Advance past the last token of the statement if needed
			if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
				p.nextToken()
			}
		}

		stmt.Choices = append(stmt.Choices, choice)
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
			choice.Body = p.parseStatement()

			// Advance past the last token of the statement if needed
			if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
				p.nextToken()
			}
		}

		stmt.Choices = append(stmt.Choices, choice)
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
	case lexer.STRING:
		left = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
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
	case lexer.BITNOT:
		token := p.curToken
		p.nextToken()
		left = &ast.UnaryExpr{
			Token:    token,
			Operator: "~",
			Right:    p.parseExpression(PREFIX),
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
