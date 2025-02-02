package scanner

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

// An ErrorHandler may be provided to [Scanner.New]. If a syntax error is
// encountered and a handler was installed, the handler is called with a
// position and an error message. The position points to the beginning of
// the offending token.
type ErrorHandler func(pos Pos, msg string)

type Scanner struct {
	text       []byte        // source text
	rdoffset   int           // reading offset (position after current character)
	current    rune          // current character
	pos        Pos           // value start position
	end        Pos           // value end position
	val        *bytes.Buffer // value buffer
	errHandler ErrorHandler  // error reporting; or nil
	line       *bytes.Buffer // line buffer
	lines      []string      // list of lines
	done       bool          // there is nothing more to scan
}

const (
	bom       = 0xFEFF // byte order mark, only permitted as very first character
	endmarker = -1     // end of file
)

func New(text []byte, errHandler ErrorHandler) *Scanner {
	scanner := &Scanner{
		text,
		0,
		endmarker,
		Pos{0, 0},
		Pos{0, 0},
		bytes.NewBuffer(nil),
		errHandler,
		bytes.NewBuffer(nil),
		nil,
		false,
	}

	scanner.load()
	if scanner.current == bom {
		// ignore BOM at file beginning
		scanner.load()
	}

	return scanner
}

// peek returns the byte following the most recently read character without
// advancing the scanner. If the scanner is at EOF, peek returns 0.
func (s *Scanner) peek() byte {
	if s.rdoffset < len(s.text) {
		return s.text[s.rdoffset]
	}
	return 0
}

// Read the next Unicode char into s.current and write it into value buffer.
func (s *Scanner) load() {
	if s.rdoffset < len(s.text) {
		r, w := rune(s.text[s.rdoffset]), 1
		switch {
		case r == 0:
			s.err(s.end, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.text[s.rdoffset:])
			if r == utf8.RuneError && w == 1 {
				s.err(s.end, "illegal UTF-8 encoding")
			} else if r == bom && s.rdoffset > 0 {
				s.err(s.pos, "illegal byte order mark")
			}
		}

		s.rdoffset += w
		s.current = r
	} else {
		s.current = endmarker
	}
}

func (s *Scanner) next() {
	switch s.current {
	case endmarker:
		if !s.done {
			// last line
			if s.line.Len() > 0 {
				s.lines = append(s.lines, s.line.String())
				s.line.Reset()
			}
			s.done = true
		}
	case '\n':
		s.lines = append(s.lines, s.line.String())
		s.line.Reset()
		s.end = Pos{s.end.Line + 1, 0}
	default:
		s.line.WriteRune(s.current)
		s.end = Pos{s.end.Line, s.end.Column + 1}
	}

	s.val.WriteRune(s.current)
	s.load()
}

func (s *Scanner) err(pos Pos, msg string) {
	if s.errHandler != nil {
		s.errHandler(pos, msg)
	}
}

func (s *Scanner) errf(pos Pos, format string, args ...any) {
	s.err(pos, fmt.Sprintf(format, args...))
}

func (s *Scanner) makeToken(kind TokenKind) Token {
	var value string
	switch kind {
	case NEWLINE:
		value = "newline"
	case ENDMARKER:
		value = "endmarker"
	default:
		value = s.val.String()
	}

	return Token{kind, s.pos, value}
}

func (s *Scanner) skipWhitespace() {
	for {
		switch s.current {
		case ' ', '\t', '\r':
			s.next()
		default:
			return
		}
	}
}

func (s *Scanner) switch2(ch rune, k0, k1 TokenKind) TokenKind {
	if s.current == ch {
		s.next()
		return k0
	}
	return k1
}

func digitValue(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16 // larger than any legal digit val
}

func lower(ch rune) rune     { return ('a' - 'A') | ch }
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

func isIdentifierBeginning(ch rune) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch == '_'
}

func isIdentifierMiddle(ch rune) bool {
	return isIdentifierBeginning(ch) || isDecimal(ch)
}

func (s *Scanner) scanComment() Token {
	for s.current != '\n' && s.current != endmarker {
		s.next()
	}
	return s.makeToken(COMMENT)
}

var keywords = map[string]TokenKind{
	"as":        AS,
	"const":     CONST,
	"embed":     EMBED,
	"false":     FALSE,
	"import":    IMPORT,
	"null":      NULL,
	"interface": INTERFACE,
	"struct":    STRUCT,
	"true":      TRUE,
	"type":      TYPE,
	"func":      FUNC,
}

func (s *Scanner) scanIdentifier() Token {
	for isIdentifierMiddle(s.current) {
		s.next()
	}
	kind, isKeyword := keywords[s.val.String()]
	if !isKeyword {
		kind = IDENTIFIER
	}

	return s.makeToken(kind)
}

const (
	decimal     = 10
	binary      = 2
	octal       = 8
	hexadecimal = 16
)

func litname(base int) string {
	switch base {
	case 2:
		return "binary"
	case 8:
		return "octal"
	case 16:
		return "hexadecimal"
	}
	return "decimal"
}

const (
	noDigits = 1 << iota
	invalidDigitSep
	leadingZero
)

func (s *Scanner) digits(base int, lead bool) int {
	flags := 0
	digits, n := 0, 0
	digsep := lead
	for {
		if digitValue(s.current) < base {
			digits++
			digsep = true
			s.next()
		} else if s.current == '_' {
			if digsep {
				digsep = false
			} else {
				flags |= invalidDigitSep
			}
			s.next()
		} else {
			break
		}
		n++
	}

	if digits == 0 && n > 0 || digits > 0 && !digsep {
		// trailing or single separator is not allowed
		flags |= invalidDigitSep
	}

	if digits == 0 {
		flags |= noDigits
	}

	return flags
}

func (s *Scanner) scanNumber() Token {
	flags := 0
	kind := INTEGER
	base := decimal
	if s.current == '0' {
		s.next()
		switch s.current {
		case 'b', 'B':
			base = binary
			s.next()
		case 'o', 'O':
			base = octal
			s.next()
		case 'x', 'X':
			base = hexadecimal
			s.next()
		default:
			if isDecimal(s.current) {
				flags |= leadingZero
			}
		}
	}

	if s.current != '.' {
		if f := s.digits(base, true); f&invalidDigitSep != 0 {
			flags |= invalidDigitSep
		} else if f&noDigits != 0 && base != decimal {
			s.errf(s.pos, "%s literal has no digits", litname(base))
		}
	}

	if s.current == '.' {
		if base != decimal {
			s.errf(s.end, "invalid radix point in %s literal", litname(base))
		}
		s.next()
		kind = FLOAT
		if f := s.digits(decimal, false); f&invalidDigitSep != 0 {
			flags |= invalidDigitSep
		}
	}

	if s.current == 'e' || s.current == 'E' {
		s.next()
		kind = FLOAT
		if s.current == '+' || s.current == '-' {
			s.next()
		}

		if f := s.digits(base, false); f&invalidDigitSep != 0 {
			flags |= invalidDigitSep
		} else if f&noDigits != 0 {
			s.err(s.pos, "exponent has no digits")
		}
	}

	if flags&leadingZero != 0 && kind == INTEGER {
		s.err(s.pos, "leading zeros in decimal integer literals are not permitted")
	}

	if flags&invalidDigitSep != 0 {
		s.err(s.pos, "'_' must separate successive digits")
	}

	return s.makeToken(kind)
}

func (s *Scanner) escape() {
	pos := s.end
	current := s.current
	s.next()
	var n, max int
	switch current {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '"':
		return
	case 'x':
		n, max = 2, 255
	case 'u':
		n, max = 4, utf8.MaxRune
	case 'U':
		n, max = 8, utf8.MaxRune
	default:
		s.err(pos, "unknown escape sequence")
		return
	}

	x := 0
	for ; n > 0; n-- {
		current = s.current
		s.next()
		d := digitValue(current)
		if d == 16 {
			s.errf(pos, "illegal hexadecimal digit %#U in escape sequence", current)
			return
		}

		x = x*16 + d
	}

	if x > max || x >= 0xD800 && x < 0xE000 {
		s.err(pos, "escape sequence is invalid unicode code point")
	}
}

func (s *Scanner) scanString() Token {
	for s.current != '"' && s.current != endmarker && s.current != '\n' {
		current := s.current
		s.next()
		if current == '\\' {
			s.escape()
		}
	}

	if s.current == '"' {
		s.next()
	} else {
		s.err(s.pos, "unterminated string")
	}

	return s.makeToken(STRING)
}

func (s *Scanner) Scan() Token {
	// Prepare to scan a next token
	s.skipWhitespace()
	s.val.Reset()
	s.pos = s.end

	current := s.current
	switch {
	case isIdentifierBeginning(current):
		return s.scanIdentifier()
	case isDecimal(current) || current == '.' && isDecimal(rune(s.peek())):
		return s.scanNumber()
	case current == '/' && rune(s.peek()) == '/':
		return s.scanComment()
	default:
		s.next()
		switch current {
		case endmarker:
			return s.makeToken(ENDMARKER)
		case '\n':
			return s.makeToken(NEWLINE)
		case '(':
			return s.makeToken(LEFT_PAREN)
		case ')':
			return s.makeToken(RIGHT_PAREN)
		case '[':
			return s.makeToken(LEFT_BRACK)
		case ']':
			return s.makeToken(RIGHT_BRACK)
		case '{':
			return s.makeToken(LEFT_BRACE)
		case '}':
			return s.makeToken(RIGHT_BRACE)
		case ',':
			return s.makeToken(COMMA)
		case '.':
			return s.makeToken(DOT)
		case ':':
			return s.makeToken(COLON)
		case ';':
			return s.makeToken(SEMICOLON)
		case '?':
			return s.makeToken(QMARK)
		case '=':
			return s.makeToken(s.switch2('=', EQ, ASSIGN))
		case '@':
			return s.makeToken(AT)
		case '+':
			return s.makeToken(PLUS)
		case '-':
			return s.makeToken(s.switch2('>', ARROW, MINUS))
		case '*':
			return s.makeToken(MULT)
		case '/':
			return s.makeToken(DIV)
		case '%':
			return s.makeToken(MOD)
		case '&':
			kind := s.switch2('&', AND, ILLEGAL)
			if kind == AND {
				return s.makeToken(kind)
			}
		case '|':
			kind := s.switch2('|', OR, ILLEGAL)
			if kind == OR {
				return s.makeToken(kind)
			}
		case '<':
			return s.makeToken(s.switch2('=', LE, LT))
		case '>':
			return s.makeToken(s.switch2('=', GE, GT))
		case '!':
			return s.makeToken(s.switch2('=', NEQ, NOT))
		case '"':
			return s.scanString()
		}
	}

	s.err(s.pos, fmt.Sprintf("illegal character %#U", current))

	return s.makeToken(ILLEGAL)
}

// Returns a list of lines (without \n) that have already been processed.
func (s *Scanner) Lines() []string {
	return s.lines
}

func (s *Scanner) Done() bool {
	return s.done
}
