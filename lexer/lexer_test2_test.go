package lexer

import "testing"

func TestAllKeywords(t *testing.T) {
	// Test all keywords that aren't covered in TestKeywords
	// Note: AND and OR are continuation operators, so they can't be at line end
	// (the lexer would suppress the following NEWLINE). Put them mid-line.
	input := "CASE ELSE FUNC FUNCTION VALOF RESULT IS CHAN OF SKIP STOP VAL PROTOCOL RECORD SIZE STEP MOSTNEG MOSTPOS INITIAL RETYPES PLUS MINUS TIMES TIMER AFTER FOR FROM REAL REAL32 REAL64 NOT AND OR WHILE\n"
	expected := []TokenType{
		CASE, ELSE, FUNC, FUNCTION, VALOF, RESULT, IS, CHAN, OF,
		SKIP, STOP, VAL, PROTOCOL, RECORD, SIZE_KW, STEP,
		MOSTNEG_KW, MOSTPOS_KW, INITIAL, RETYPES,
		PLUS_KW, MINUS_KW, TIMES, TIMER, AFTER, FOR, FROM,
		REAL_TYPE, REAL32_TYPE, REAL64_TYPE,
		NOT, AND, OR, WHILE,
		NEWLINE, EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("tests[%d] - expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestParenDepthSuppressesIndent(t *testing.T) {
	// Inside parentheses, INDENT/DEDENT/NEWLINE should be suppressed
	input := `x := (1
    + 2
    + 3)
`
	expected := []TokenType{
		IDENT,    // x
		ASSIGN,   // :=
		LPAREN,   // (
		INT,      // 1
		PLUS,     // +
		INT,      // 2
		PLUS,     // +
		INT,      // 3
		RPAREN,   // )
		NEWLINE,  // after closing paren
		EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("paren[%d] - expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestBracketDepthSuppressesIndent(t *testing.T) {
	// Inside brackets, INDENT/DEDENT/NEWLINE should be suppressed
	input := `x := [1,
    2,
    3]
`
	expected := []TokenType{
		IDENT,    // x
		ASSIGN,   // :=
		LBRACKET, // [
		INT,      // 1
		COMMA,    // ,
		INT,      // 2
		COMMA,    // ,
		INT,      // 3
		RBRACKET, // ]
		NEWLINE,  // after closing bracket
		EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("bracket[%d] - expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestContinuationOperator(t *testing.T) {
	// Operator at end of line causes NEWLINE/INDENT/DEDENT suppression on next line
	input := `x := a +
    b
`
	expected := []TokenType{
		IDENT,   // x
		ASSIGN,  // :=
		IDENT,   // a
		PLUS,    // +
		IDENT,   // b
		NEWLINE, // after b
		EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("continuation[%d] - expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestContinuationAND(t *testing.T) {
	// AND at end of line should continue
	input := `(x > 0) AND
    (x < 10)
`
	expected := []TokenType{
		LPAREN, IDENT, GT, INT, RPAREN, // (x > 0)
		AND,                             // AND
		LPAREN, IDENT, LT, INT, RPAREN, // (x < 10)
		NEWLINE,
		EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("cont_and[%d] - expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"hello world"` + "\n"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Fatalf("expected STRING, got %q", tok.Type)
	}
	if tok.Literal != "hello world" {
		t.Fatalf("expected literal %q, got %q", "hello world", tok.Literal)
	}
}

func TestStringEscapeSequences(t *testing.T) {
	// The lexer preserves raw occam escapes (*n, *t, etc.) in string literals.
	// Escape conversion (*n → \n) happens in the parser, not the lexer.
	input := `"a*nb"` + "\n"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Fatalf("expected STRING, got %q", tok.Type)
	}
	if tok.Literal != "a*nb" {
		t.Fatalf("expected literal %q, got %q", "a*nb", tok.Literal)
	}
}

func TestByteLiteralToken(t *testing.T) {
	input := "'A'\n"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != BYTE_LIT {
		t.Fatalf("expected BYTE_LIT, got %q (literal=%q)", tok.Type, tok.Literal)
	}
	if tok.Literal != "A" {
		t.Fatalf("expected literal %q, got %q", "A", tok.Literal)
	}
}

func TestByteLiteralEscapeToken(t *testing.T) {
	// The lexer preserves raw occam escape (*n) in byte literals.
	// Escape conversion happens in the parser.
	input := "'*n'\n"
	l := New(input)
	tok := l.NextToken()
	if tok.Type != BYTE_LIT {
		t.Fatalf("expected BYTE_LIT, got %q (literal=%q)", tok.Type, tok.Literal)
	}
	if tok.Literal != "*n" {
		t.Fatalf("expected literal %q, got %q", "*n", tok.Literal)
	}
}

func TestSendReceiveTokens(t *testing.T) {
	input := "c ! 42\nc ? x\n"
	expected := []struct {
		typ TokenType
		lit string
	}{
		{IDENT, "c"},
		{SEND, "!"},
		{INT, "42"},
		{NEWLINE, "\\n"},
		{IDENT, "c"},
		{RECEIVE, "?"},
		{IDENT, "x"},
		{NEWLINE, "\\n"},
		{EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("send_recv[%d] - expected type=%q, got=%q (literal=%q)",
				i, exp.typ, tok.Type, tok.Literal)
		}
	}
}

func TestAmpersandToken(t *testing.T) {
	// & used as guard separator in ALT
	input := "TRUE & c ? x\n"
	expected := []TokenType{TRUE, AMPERSAND, IDENT, RECEIVE, IDENT, NEWLINE, EOF}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("ampersand[%d] - expected=%q, got=%q", i, exp, tok.Type)
		}
	}
}

func TestSemicolonToken(t *testing.T) {
	input := "c ! 10 ; 20\n"
	expected := []TokenType{IDENT, SEND, INT, SEMICOLON, INT, NEWLINE, EOF}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("semicolon[%d] - expected=%q, got=%q", i, exp, tok.Type)
		}
	}
}

func TestNestedParenDepth(t *testing.T) {
	// Nested parens: depth tracks correctly
	input := `x := ((1
    + 2)
    + 3)
`
	expected := []TokenType{
		IDENT, ASSIGN,
		LPAREN, LPAREN, INT,
		PLUS, INT, RPAREN,
		PLUS, INT, RPAREN,
		NEWLINE, EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("nested_paren[%d] - expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestMixedParenBracketDepth(t *testing.T) {
	// Mix of parens and brackets, both should suppress indent
	input := `x := arr[(1
    + 2)]
`
	expected := []TokenType{
		IDENT, ASSIGN,
		IDENT, LBRACKET, LPAREN, INT,
		PLUS, INT, RPAREN, RBRACKET,
		NEWLINE, EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("mixed[%d] - expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestLineAndColumnTracking(t *testing.T) {
	input := "INT x:\nx := 5\n"
	l := New(input)

	// INT at line 1, col 1
	tok := l.NextToken()
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("INT: expected line=1 col=1, got line=%d col=%d", tok.Line, tok.Column)
	}

	// x at line 1, col 5
	tok = l.NextToken()
	if tok.Line != 1 || tok.Column != 5 {
		t.Errorf("x: expected line=1 col=5, got line=%d col=%d", tok.Line, tok.Column)
	}
}
