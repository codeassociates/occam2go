package lexer

import (
	"strings"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position (after current char)
	ch           byte // current char under examination
	line         int
	column       int

	// Indentation tracking
	indentStack  []int // stack of indentation levels
	pendingTokens []Token // tokens to emit before reading more input
	atLineStart  bool
}

func New(input string) *Lexer {
	l := &Lexer{
		input:       input,
		line:        1,
		column:      0,
		indentStack: []int{0}, // start with base indentation of 0
		atLineStart: true,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() Token {
	// Return any pending tokens first (from indentation processing)
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	// Handle indentation at the start of a line
	if l.atLineStart {
		l.atLineStart = false
		indent := l.measureIndent()
		currentIndent := l.indentStack[len(l.indentStack)-1]

		if indent > currentIndent {
			l.indentStack = append(l.indentStack, indent)
			return Token{Type: INDENT, Literal: "", Line: l.line, Column: 1}
		} else if indent < currentIndent {
			// May need multiple DEDENTs
			for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indent {
				l.indentStack = l.indentStack[:len(l.indentStack)-1]
				l.pendingTokens = append(l.pendingTokens, Token{Type: DEDENT, Literal: "", Line: l.line, Column: 1})
			}
			if len(l.pendingTokens) > 0 {
				tok := l.pendingTokens[0]
				l.pendingTokens = l.pendingTokens[1:]
				return tok
			}
		}
	}

	l.skipWhitespace()

	var tok Token
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '(':
		tok = l.newToken(LPAREN, l.ch)
	case ')':
		tok = l.newToken(RPAREN, l.ch)
	case '[':
		tok = l.newToken(LBRACKET, l.ch)
	case ']':
		tok = l.newToken(RBRACKET, l.ch)
	case ',':
		tok = l.newToken(COMMA, l.ch)
	case ';':
		tok = l.newToken(SEMICOLON, l.ch)
	case '+':
		tok = l.newToken(PLUS, l.ch)
	case '*':
		tok = l.newToken(MULTIPLY, l.ch)
	case '/':
		if l.peekChar() == '\\' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: BITAND, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(DIVIDE, l.ch)
		}
	case '\\':
		if l.peekChar() == '/' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: BITOR, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(MODULO, l.ch)
		}
	case '~':
		tok = l.newToken(BITNOT, l.ch)
	case '=':
		tok = l.newToken(EQ, l.ch)
	case '!':
		tok = l.newToken(SEND, l.ch)
	case '?':
		tok = l.newToken(RECEIVE, l.ch)
	case '&':
		tok = l.newToken(AMPERSAND, l.ch)
	case ':':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ASSIGN, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(COLON, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LE, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NEQ, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LSHIFT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GE, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: RSHIFT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: BITXOR, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(GT, l.ch)
		}
	case '#':
		if isHexDigit(l.peekChar()) {
			tok.Type = INT
			tok.Literal = l.readHexNumber()
			tok.Line = l.line
			return tok
		} else {
			tok = l.newToken(ILLEGAL, l.ch)
		}
	case '-':
		if l.peekChar() == '-' {
			l.skipComment()
			return l.NextToken()
		} else {
			tok = l.newToken(MINUS, l.ch)
		}
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
	case '\'':
		tok.Type = BYTE_LIT
		tok.Literal = l.readByteLiteral()
		tok.Line = l.line
		tok.Column = l.column
	case '\n':
		tok = Token{Type: NEWLINE, Literal: "\\n", Line: l.line, Column: l.column}
		l.line++
		l.column = 0
		l.atLineStart = true
		l.readChar()
		// Skip blank lines (but not EOF)
		for l.ch != 0 && (l.ch == '\n' || l.isBlankLine()) {
			if l.ch == '\n' {
				l.line++
				l.column = 0
				l.readChar()
			} else {
				l.skipToEndOfLine()
			}
		}
		return tok
	case 0:
		// Emit any remaining DEDENTs before EOF
		if len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			return Token{Type: DEDENT, Literal: "", Line: l.line, Column: l.column}
		}
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			tok.Line = l.line
			return tok
		} else if isDigit(l.ch) {
			tok.Type = INT
			tok.Literal = l.readNumber()
			tok.Line = l.line
			return tok
		} else {
			tok = l.newToken(ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Literal: string(ch), Line: l.line, Column: l.column}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readHexNumber() string {
	// Current char is '#', skip it
	l.readChar()
	position := l.position
	for isHexDigit(l.ch) {
		l.readChar()
	}
	return "0x" + l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readByteLiteral() string {
	// Current char is the opening single quote.
	// Read content between single quotes, handling *' escape.
	// In occam, * is the escape character. ** means literal *, *' means literal '.
	position := l.position + 1
	escaped := false
	for {
		l.readChar()
		if l.ch == 0 {
			break
		}
		if escaped {
			// This char is the escaped character; consume it and clear flag
			escaped = false
			continue
		}
		if l.ch == '*' {
			escaped = true
			continue
		}
		if l.ch == '\'' {
			break
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	// Skip -- comment until end of line
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) measureIndent() int {
	indent := 0
	pos := l.position
	for pos < len(l.input) {
		ch := l.input[pos]
		if ch == ' ' {
			indent++
			pos++
		} else if ch == '\t' {
			indent += 2 // treat tabs as 2 spaces
			pos++
		} else {
			break
		}
	}
	return indent
}

func (l *Lexer) isBlankLine() bool {
	pos := l.position
	for pos < len(l.input) {
		ch := l.input[pos]
		if ch == '\n' {
			return true
		}
		if ch != ' ' && ch != '\t' && ch != '\r' {
			// Check for comment-only line
			if ch == '-' && pos+1 < len(l.input) && l.input[pos+1] == '-' {
				return true
			}
			return false
		}
		pos++
	}
	return false // EOF is not a blank line
}

func (l *Lexer) skipToEndOfLine() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '\n' {
		l.line++
		l.column = 0
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// Tokenize returns all tokens from the input
func Tokenize(input string) []Token {
	// Ensure input ends with newline for consistent processing
	if !strings.HasSuffix(input, "\n") {
		input = input + "\n"
	}

	l := New(input)
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}
	return tokens
}
