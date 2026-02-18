package lexer

type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	NEWLINE
	INDENT
	DEDENT

	// Literals
	IDENT     // variable names, procedure names
	INT       // integer literal
	STRING    // string literal
	BYTE_LIT  // byte literal: 'A', '*n', etc.

	// Operators
	ASSIGN   // :=
	PLUS     // +
	MINUS    // -
	MULTIPLY // *
	DIVIDE   // /
	MODULO   // \ (backslash in Occam)
	EQ       // =
	NEQ      // <>
	LT       // <
	GT       // >
	LE       // <=
	GE       // >=
	SEND      // !
	RECEIVE   // ?
	AMPERSAND // & (guard separator in ALT)
	BITAND    // /\  (bitwise AND)
	BITOR     // \/  (bitwise OR)
	BITXOR    // ><  (bitwise XOR)
	BITNOT    // ~   (bitwise NOT)
	LSHIFT    // <<  (left shift)
	RSHIFT    // >>  (right shift)

	// Delimiters
	LPAREN    // (
	RPAREN    // )
	LBRACKET  // [
	RBRACKET  // ]
	COMMA     // ,
	COLON     // :
	SEMICOLON // ;

	// Keywords
	keyword_beg
	SEQ
	PAR
	ALT
	IF
	CASE
	ELSE
	WHILE
	FOR
	FROM
	PROC
	FUNC
	FUNCTION
	VALOF
	RESULT
	IS
	CHAN
	OF
	TRUE
	FALSE
	NOT
	AND
	OR
	SKIP
	STOP
	INT_TYPE
	BYTE_TYPE
	BOOL_TYPE
	REAL_TYPE
	REAL32_TYPE
	REAL64_TYPE
	TIMER
	AFTER
	VAL
	PROTOCOL
	RECORD
	SIZE_KW
	STEP
	keyword_end
)

var tokenNames = map[TokenType]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	NEWLINE: "NEWLINE",
	INDENT:  "INDENT",
	DEDENT:  "DEDENT",

	IDENT:    "IDENT",
	INT:      "INT",
	STRING:   "STRING",
	BYTE_LIT: "BYTE_LIT",

	ASSIGN:   ":=",
	PLUS:     "+",
	MINUS:    "-",
	MULTIPLY: "*",
	DIVIDE:   "/",
	MODULO:   "\\",
	EQ:       "=",
	NEQ:      "<>",
	LT:       "<",
	GT:       ">",
	LE:       "<=",
	GE:       ">=",
	SEND:      "!",
	RECEIVE:   "?",
	AMPERSAND: "&",
	BITAND:    "/\\",
	BITOR:     "\\/",
	BITXOR:    "><",
	BITNOT:    "~",
	LSHIFT:    "<<",
	RSHIFT:    ">>",

	LPAREN:    "(",
	RPAREN:    ")",
	LBRACKET:  "[",
	RBRACKET:  "]",
	COMMA:     ",",
	COLON:     ":",
	SEMICOLON: ";",

	SEQ:       "SEQ",
	PAR:       "PAR",
	ALT:       "ALT",
	IF:        "IF",
	CASE:      "CASE",
	ELSE:      "ELSE",
	WHILE:     "WHILE",
	FOR:       "FOR",
	FROM:      "FROM",
	PROC:      "PROC",
	FUNC:      "FUNC",
	FUNCTION:  "FUNCTION",
	VALOF:     "VALOF",
	RESULT:    "RESULT",
	IS:        "IS",
	CHAN:      "CHAN",
	OF:        "OF",
	TRUE:      "TRUE",
	FALSE:     "FALSE",
	NOT:       "NOT",
	AND:       "AND",
	OR:        "OR",
	SKIP:      "SKIP",
	STOP:      "STOP",
	INT_TYPE:  "INT",
	BYTE_TYPE: "BYTE",
	BOOL_TYPE: "BOOL",
	REAL_TYPE:   "REAL",
	REAL32_TYPE: "REAL32",
	REAL64_TYPE: "REAL64",
	TIMER:       "TIMER",
	AFTER:    "AFTER",
	VAL:       "VAL",
	PROTOCOL:  "PROTOCOL",
	RECORD:    "RECORD",
	SIZE_KW:   "SIZE",
	STEP:      "STEP",
}

var keywords = map[string]TokenType{
	"SEQ":   SEQ,
	"PAR":   PAR,
	"ALT":   ALT,
	"IF":    IF,
	"CASE":  CASE,
	"ELSE":  ELSE,
	"WHILE": WHILE,
	"FOR":   FOR,
	"FROM":  FROM,
	"PROC":  PROC,
	"FUNC":     FUNC,
	"FUNCTION": FUNCTION,
	"VALOF":    VALOF,
	"RESULT":   RESULT,
	"IS":       IS,
	"CHAN":     CHAN,
	"OF":    OF,
	"TRUE":  TRUE,
	"FALSE": FALSE,
	"NOT":   NOT,
	"AND":   AND,
	"OR":    OR,
	"SKIP":  SKIP,
	"STOP":  STOP,
	"INT":   INT_TYPE,
	"BYTE":  BYTE_TYPE,
	"BOOL":  BOOL_TYPE,
	"REAL":   REAL_TYPE,
	"REAL32": REAL32_TYPE,
	"REAL64": REAL64_TYPE,
	"TIMER":  TIMER,
	"AFTER": AFTER,
	"VAL":      VAL,
	"PROTOCOL": PROTOCOL,
	"RECORD":   RECORD,
	"SIZE":     SIZE_KW,
	"STEP":     STEP,
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}
