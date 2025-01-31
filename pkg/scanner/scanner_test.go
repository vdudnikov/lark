package scanner

import (
	"strings"
	"testing"
)

func scan(input string, t *testing.T) Token {
	s := New([]byte(input), func(pos Pos, msg string) {
		t.Errorf("error at %d:%d, %s\n", pos.Line+1, pos.Column+1, msg)
	})
	return s.Scan()
}

func TestNonLiteral(t *testing.T) {
	type testCase struct {
		input string
		want  TokenKind
	}

	tests := []testCase{
		{"\n", NEWLINE},
		{"(", LEFT_PAREN},
		{")", RIGHT_PAREN},
		{"[", LEFT_BRACK},
		{"]", RIGHT_BRACK},
		{"{", LEFT_BRACE},
		{"}", RIGHT_BRACE},
		{",", COMMA},
		{".", DOT},
		{":", COLON},
		{";", SEMICOLON},
		{"->", ARROW},
		{"@", AT},
		{"?", QMARK},
		{"=", ASSIGN},
		{"+", PLUS},
		{"-", MINUS},
		{"*", MULT},
		{"/", DIV},
		{"%", MOD},
		{"&&", AND},
		{"||", OR},
		{"==", EQ},
		{">=", GE},
		{">", GT},
		{"<=", LE},
		{"<", LT},
		{"!=", NEQ},
		{"!", NOT},
		{"as", AS},
		{"const", CONST},
		{"embed", EMBED},
		{"false", FALSE},
		{"import", IMPORT},
		{"interface", INTERFACE},
		{"null", NULL},
		{"struct", STRUCT},
		{"true", TRUE},
		{"type", TYPE},
		{"func", FUNC},
	}

	for _, test := range tests {
		s := New([]byte(test.input), func(pos Pos, msg string) {
			t.Errorf("error at %d:%d, %s\n", pos.Line+1, pos.Column+1, msg)
		})

		token := s.Scan()
		if token.Kind != test.want {
			t.Errorf("%q: got token %s; want %s", test.input, token.Kind, test.want)
		}
	}
}

func TestIdentifier(t *testing.T) {
	tests := []string{
		"_",
		"foobar",
		"a0123456789",
	}

	for _, input := range tests {
		token := scan(input, t)
		if token.Kind != IDENTIFIER {
			t.Errorf("expected IDENTIFIER but found '%s'", token.Kind)
		} else if token.Value != input {
			t.Errorf("expected '%s' but found '%s'", input, token.Value)
		}

	}
}

func TestString(t *testing.T) {
	tests := []string{
		"\"\\a\"",
		"\"\\b\"",
		"\"\\f\"",
		"\"\\n\"",
		"\"\\r\"",
		"\"\\t\"",
		"\"\\v\"",
		"\"\\\\\"",
		"\"\\\"\"",
		"\"\\xff\"",
		"\"\\xFF\"",
		"\"\\uFFFF\"",
		"\"\\U0010FFFF\"",
	}

	for _, input := range tests {
		token := scan(input, t)
		if token.Kind != STRING {
			t.Errorf("expected STRING but found %s", token.Kind)
		} else if token.Value != input {
			t.Errorf("expected '%s' but found '%s'", input, token.Value)
		}
	}
}

func TestNumber(t *testing.T) {
	type testCase struct {
		kind   TokenKind
		input  string
		tokens string
		errMsg string
	}

	tests := []testCase{
		// binaries
		{INTEGER, "0b0", "0b0", ""},
		{INTEGER, "0b1010", "0b1010", ""},
		{INTEGER, "0B1110", "0B1110", ""},

		{INTEGER, "0b", "0b", "binary literal has no digits"},
		{INTEGER, "0b01a0", "0b01 a0", ""}, // only accept 0-9

		{FLOAT, "0b.", "0b.", "invalid radix point in binary literal"},
		{FLOAT, "0b.1", "0b.1", "invalid radix point in binary literal"},
		{FLOAT, "0b1.0", "0b1.0", "invalid radix point in binary literal"},

		// octals
		{INTEGER, "0o0", "0o0", ""},
		{INTEGER, "0o1234", "0o1234", ""},
		{INTEGER, "0O1234", "0O1234", ""},

		{INTEGER, "0o", "0o", "octal literal has no digits"},
		// {INTEGER, "0o8123", "0o8123", "invalid digit '8' in octal literal"},
		// {INTEGER, "0o1293", "0o1293", "invalid digit '9' in octal literal"},
		{INTEGER, "0o12a3", "0o12 a3", ""}, // only accept 0-9

		{FLOAT, "0o.", "0o.", "invalid radix point in octal literal"},
		{FLOAT, "0o.2", "0o.2", "invalid radix point in octal literal"},
		{FLOAT, "0o1.2", "0o1.2", "invalid radix point in octal literal"},

		// 0-octals not allowed
		{INTEGER, "0123", "0123", "leading zeros in decimal integer literals are not permitted"},

		// decimals
		{INTEGER, "0", "0", ""},
		{INTEGER, "1", "1", ""},
		{INTEGER, "1234", "1234", ""},

		{INTEGER, "1f", "1 f", ""}, // only accept 0-9

		// decimal floats
		{FLOAT, "0.", "0.", ""},
		{FLOAT, "123.", "123.", ""},
		{FLOAT, "0123.", "0123.", ""},

		{FLOAT, ".0", ".0", ""},
		{FLOAT, ".123", ".123", ""},
		{FLOAT, ".0123", ".0123", ""},

		{FLOAT, "0.0", "0.0", ""},
		{FLOAT, "123.123", "123.123", ""},
		{FLOAT, "0123.0123", "0123.0123", ""},

		{FLOAT, "0e0", "0e0", ""},
		{FLOAT, "123e+0", "123e+0", ""},
		{FLOAT, "0123E-1", "0123E-1", ""},

		{FLOAT, "0.e+1", "0.e+1", ""},
		{FLOAT, "123.E-10", "123.E-10", ""},
		{FLOAT, "0123.e123", "0123.e123", ""},

		{FLOAT, ".0e-1", ".0e-1", ""},
		{FLOAT, ".123E+10", ".123E+10", ""},
		{FLOAT, ".0123E123", ".0123E123", ""},

		{FLOAT, "0.0e1", "0.0e1", ""},
		{FLOAT, "123.123E-10", "123.123E-10", ""},
		{FLOAT, "0123.0123e+456", "0123.0123e+456", ""},

		{FLOAT, "0e", "0e", "exponent has no digits"},
		{FLOAT, "0E+", "0E+", "exponent has no digits"},
		{FLOAT, "1e+f", "1e+ f", "exponent has no digits"},

		// hexadecimals
		{INTEGER, "0x0", "0x0", ""},
		{INTEGER, "0x1234", "0x1234", ""},
		{INTEGER, "0xcafef00d", "0xcafef00d", ""},
		{INTEGER, "0XCAFEF00D", "0XCAFEF00D", ""},

		{INTEGER, "0x", "0x", "hexadecimal literal has no digits"},
		{INTEGER, "0x1g", "0x1 g", ""},

		{FLOAT, "0x.", "0x.", "invalid radix point in hexadecimal literal"},
		{FLOAT, "0x.1", "0x.1", "invalid radix point in hexadecimal literal"},
		{FLOAT, "0x1.0", "0x1.0", "invalid radix point in hexadecimal literal"},

		// separators
		{INTEGER, "0b_1000_0001", "0b_1000_0001", ""},
		{INTEGER, "0o_600", "0o_600", ""},
		{INTEGER, "0_466", "0_466", ""},
		{INTEGER, "1_000", "1_000", ""},
		{FLOAT, "1_000.000_1", "1_000.000_1", ""},
		{INTEGER, "0x_f00d", "0x_f00d", ""},

		{INTEGER, "0b__1000", "0b__1000", "'_' must separate successive digits"},
		{INTEGER, "0o60___0", "0o60___0", "'_' must separate successive digits"},
		{FLOAT, "1_.", "1_.", "'_' must separate successive digits"},
		{FLOAT, "0._1", "0._1", "'_' must separate successive digits"},
		{FLOAT, "2.7_e0", "2.7_e0", "'_' must separate successive digits"},
		{INTEGER, "0x___0", "0x___0", "'_' must separate successive digits"},
	}

	for _, test := range tests {
		errMsg := ""
		s := New([]byte(test.input), func(pos Pos, msg string) {
			if errMsg == "" {
				errMsg = msg
			}
		})

		for i, want := range strings.Split(test.tokens, " ") {
			token := s.Scan()
			if i == 0 {
				if token.Kind != test.kind {
					t.Errorf("%q: got token %s; want %s", test.input, token.Value, test.kind)
				}
				if errMsg != test.errMsg {
					t.Errorf("%q: got error %q; want %q", test.input, errMsg, test.errMsg)
				}
			}

			if token.Value != want {
				t.Errorf("%q: got literal %q (%s); want %s", test.input, token.Value, token.Kind, want)
			}
		}
	}
}
