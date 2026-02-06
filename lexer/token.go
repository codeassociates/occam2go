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
	IDENT  // variable names, procedure names
	INT    // integer literal
	STRING // string literal

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
	WHILE
	FOR
	PROC
	FUNC
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
	VAL
	keyword_end
)

var tokenNames = map[TokenType]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	NEWLINE: "NEWLINE",
	INDENT:  "INDENT",
	DEDENT:  "DEDENT",

	IDENT:  "IDENT",
	INT:    "INT",
	STRING: "STRING",

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
	WHILE:     "WHILE",
	FOR:       "FOR",
	PROC:      "PROC",
	FUNC:      "FUNC",
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
	REAL_TYPE: "REAL",
	VAL:       "VAL",
}

var keywords = map[string]TokenType{
	"SEQ":   SEQ,
	"PAR":   PAR,
	"ALT":   ALT,
	"IF":    IF,
	"WHILE": WHILE,
	"FOR":   FOR,
	"PROC":  PROC,
	"FUNC":  FUNC,
	"CHAN":  CHAN,
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
	"REAL":  REAL_TYPE,
	"VAL":   VAL,
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
