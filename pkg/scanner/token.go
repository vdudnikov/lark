package scanner

import "strconv"

type TokenKind int

const (
	ILLEGAL TokenKind = iota
	ENDMARKER
	NEWLINE

	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACK
	RIGHT_BRACK
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	COLON
	SEMICOLON
	ARROW
	QMARK
	ASSIGN
	AT

	PLUS
	MINUS
	MULT
	DIV
	MOD
	AND
	OR
	EQ
	GE
	GT
	LE
	LT
	NEQ
	NOT

	COMMENT
	IDENTIFIER
	STRING
	INTEGER
	FLOAT

	AS
	CONST
	EMBED
	FALSE
	IMPORT
	INTERFACE
	NULL
	STRUCT
	TRUE
	TYPE
	FUNC
)

var tokens = [...]string{
	ILLEGAL:   "ILLEGAL",
	ENDMARKER: "ENDMARKER",
	NEWLINE:   "NEWLINE",

	LEFT_PAREN:  "(",
	RIGHT_PAREN: ")",
	LEFT_BRACK:  "[",
	RIGHT_BRACK: "]",
	LEFT_BRACE:  "{",
	RIGHT_BRACE: "}",
	COMMA:       ",",
	DOT:         ".",
	COLON:       ":",
	SEMICOLON:   ";",
	ARROW:       "->",
	QMARK:       "?",
	ASSIGN:      "=",
	AT:          "@",

	PLUS:  "+",
	MINUS: "-",
	MULT:  "*",
	DIV:   "/",
	MOD:   "%",
	AND:   "&&",
	OR:    "||",
	EQ:    "==",
	GE:    ">=",
	GT:    ">",
	LE:    "<=",
	LT:    "<",
	NEQ:   "!=",
	NOT:   "!",

	COMMENT:    "COMMENT",
	IDENTIFIER: "IDENTIFIER",
	STRING:     "STRING",
	INTEGER:    "INTEGER",
	FLOAT:      "FLOAT",

	AS:        "as",
	CONST:     "const",
	EMBED:     "embed",
	FALSE:     "false",
	IMPORT:    "import",
	INTERFACE: "interface",
	NULL:      "null",
	STRUCT:    "struct",
	TRUE:      "true",
	TYPE:      "type",
	FUNC:      "func",
}

func (kind TokenKind) String() string {
	if 0 <= kind && kind < TokenKind(len(tokens)) {
		return tokens[kind]
	}
	return "token(" + strconv.Itoa(int(kind)) + ")"
}

type Pos struct {
	Line, Column int
}

type Token struct {
	Kind  TokenKind
	Pos   Pos
	Value string
}
