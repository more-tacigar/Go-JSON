package json

import (
	"strings"
	"testing"
)

type lexTest struct {
	in  string
	out []Token
}

var lexTests = []lexTest{
	{in: `true`, out: []Token{TK_TRUE}},
	{in: `1`, out: []Token{TK_NUMBER_LITERAL}},
	{in: `1.4`, out: []Token{TK_NUMBER_LITERAL}},
	{in: `"x"`, out: []Token{TK_STRING_LITERAL}},
	{in: `true`, out: []Token{TK_TRUE}},
	{in: `{}`, out: []Token{TK_LBRACE, TK_RBRACE}},
	{in: `[]`, out: []Token{TK_LBRACKET, TK_RBRACKET}},
	{
		in:  `{"x":20}`,
		out: []Token{TK_LBRACE, TK_STRING_LITERAL, TK_COLON, TK_NUMBER_LITERAL, TK_RBRACE},
	},
}

func TestLex(t *testing.T) {
	for i, tst := range lexTests {
		l := NewLexer(strings.NewReader(tst.in))
		for j := 0; j < len(tst.out); j++ {
			tk := tst.out[j]
			if _, err := l.Lex(); err != nil {
				t.Errorf("#%d: %v", i, err)
			}
			if Token(tk) != l.currtoken {
				t.Errorf("#%d: %v want %v", i, l.currtoken, Token(tk))
			}
		}
	}
}
