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
	INT16_TYPE
	INT32_TYPE
	INT64_TYPE
	TIMER
	AFTER
	VAL
	PROTOCOL
	RECORD
	SIZE_KW
	STEP
	MOSTNEG_KW
	MOSTPOS_KW
	INITIAL
	RETYPES  // RETYPES (bit-level type reinterpretation)
	INLINE   // INLINE (function modifier, ignored for transpilation)
	PLUS_KW  // PLUS (modular addition keyword, distinct from + symbol)
	MINUS_KW // MINUS (modular subtraction keyword, distinct from - symbol)
	TIMES    // TIMES (modular multiplication keyword)
	ROUND_KW // ROUND (type conversion qualifier)
	TRUNC_KW // TRUNC (type conversion qualifier)
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
	INT16_TYPE:  "INT16",
	INT32_TYPE:  "INT32",
	INT64_TYPE:  "INT64",
	TIMER:       "TIMER",
	AFTER:    "AFTER",
	VAL:       "VAL",
	PROTOCOL:  "PROTOCOL",
	RECORD:    "RECORD",
	SIZE_KW:    "SIZE",
	STEP:       "STEP",
	MOSTNEG_KW: "MOSTNEG",
	MOSTPOS_KW: "MOSTPOS",
	INITIAL:    "INITIAL",
	RETYPES:    "RETYPES",
	INLINE:     "INLINE",
	PLUS_KW:    "PLUS",
	MINUS_KW:   "MINUS",
	TIMES:      "TIMES",
	ROUND_KW:   "ROUND",
	TRUNC_KW:   "TRUNC",
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
	"INT16":  INT16_TYPE,
	"INT32":  INT32_TYPE,
	"INT64":  INT64_TYPE,
	"TIMER":  TIMER,
	"AFTER": AFTER,
	"VAL":      VAL,
	"PROTOCOL": PROTOCOL,
	"RECORD":   RECORD,
	"SIZE":     SIZE_KW,
	"STEP":     STEP,
	"MOSTNEG":  MOSTNEG_KW,
	"MOSTPOS":  MOSTPOS_KW,
	"INITIAL":  INITIAL,
	"RETYPES":  RETYPES,
	"INLINE":   INLINE,
	"PLUS":     PLUS_KW,
	"MINUS":    MINUS_KW,
	"TIMES":    TIMES,
	"ROUND":    ROUND_KW,
	"TRUNC":    TRUNC_KW,
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
