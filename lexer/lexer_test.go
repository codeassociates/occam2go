package lexer

import (
	"testing"
)

func TestBasicTokens(t *testing.T) {
	input := `INT x:
x := 5
`
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{INT_TYPE, "INT"},
		{IDENT, "x"},
		{COLON, ":"},
		{NEWLINE, "\\n"},
		{IDENT, "x"},
		{ASSIGN, ":="},
		{INT, "5"},
		{NEWLINE, "\\n"},
		{EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestIndentation(t *testing.T) {
	input := `SEQ
  INT x:
  x := 10
`
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{SEQ, "SEQ"},
		{NEWLINE, "\\n"},
		{INDENT, ""},
		{INT_TYPE, "INT"},
		{IDENT, "x"},
		{COLON, ":"},
		{NEWLINE, "\\n"},
		{IDENT, "x"},
		{ASSIGN, ":="},
		{INT, "10"},
		{NEWLINE, "\\n"},
		{DEDENT, ""},
		{EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := `x + y - z * a / b
x < y
x > y
x <= y
x >= y
x = y
x <> y
`
	l := New(input)

	expected := []TokenType{
		IDENT, PLUS, IDENT, MINUS, IDENT, MULTIPLY, IDENT, DIVIDE, IDENT, NEWLINE,
		IDENT, LT, IDENT, NEWLINE,
		IDENT, GT, IDENT, NEWLINE,
		IDENT, LE, IDENT, NEWLINE,
		IDENT, GE, IDENT, NEWLINE,
		IDENT, EQ, IDENT, NEWLINE,
		IDENT, NEQ, IDENT, NEWLINE,
		EOF,
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, exp, tok.Type)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := "SEQ PAR ALT IF WHILE PROC INT BYTE BOOL TRUE FALSE\n"
	expected := []TokenType{
		SEQ, PAR, ALT, IF, WHILE, PROC, INT_TYPE, BYTE_TYPE, BOOL_TYPE, TRUE, FALSE,
		NEWLINE, EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `INT x: -- this is a comment
x := 5
`
	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{INT_TYPE, "INT"},
		{IDENT, "x"},
		{COLON, ":"},
		{NEWLINE, "\\n"},
		{IDENT, "x"},
		{ASSIGN, ":="},
		{INT, "5"},
		{NEWLINE, "\\n"},
		{EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
	}
}

func TestBitwiseOperators(t *testing.T) {
	input := "a /\\ b\n"
	l := New(input)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{IDENT, "a"},
		{BITAND, "/\\"},
		{IDENT, "b"},
		{NEWLINE, "\\n"},
		{EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("bitand[%d] - type wrong. expected=%q, got=%q (literal=%q)",
				i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Fatalf("bitand[%d] - literal wrong. expected=%q, got=%q",
				i, exp.lit, tok.Literal)
		}
	}

	// Test all bitwise operators in sequence
	input2 := "a \\/ b >< c ~ d << e >> f\n"
	l2 := New(input2)
	expected2 := []struct {
		typ TokenType
		lit string
	}{
		{IDENT, "a"},
		{BITOR, "\\/"},
		{IDENT, "b"},
		{BITXOR, "><"},
		{IDENT, "c"},
		{BITNOT, "~"},
		{IDENT, "d"},
		{LSHIFT, "<<"},
		{IDENT, "e"},
		{RSHIFT, ">>"},
		{IDENT, "f"},
		{NEWLINE, "\\n"},
		{EOF, ""},
	}
	for i, exp := range expected2 {
		tok := l2.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("bitwise[%d] - type wrong. expected=%q, got=%q (literal=%q)",
				i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Fatalf("bitwise[%d] - literal wrong. expected=%q, got=%q",
				i, exp.lit, tok.Literal)
		}
	}
}

func TestBitwiseVsArithmetic(t *testing.T) {
	// Ensure / alone is still DIVIDE and \ alone is still MODULO
	input := "a / b \\ c\n"
	l := New(input)
	expected := []TokenType{IDENT, DIVIDE, IDENT, MODULO, IDENT, NEWLINE, EOF}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("tests[%d] - type wrong. expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestNestedIndentation(t *testing.T) {
	input := `SEQ
  INT x:
  PAR
    x := 1
    x := 2
  x := 3
`
	expected := []TokenType{
		SEQ, NEWLINE,
		INDENT, INT_TYPE, IDENT, COLON, NEWLINE,
		PAR, NEWLINE,
		INDENT, IDENT, ASSIGN, INT, NEWLINE,
		IDENT, ASSIGN, INT, NEWLINE,
		DEDENT, IDENT, ASSIGN, INT, NEWLINE,
		DEDENT, EOF,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, exp, tok.Type, tok.Literal)
		}
	}
}
