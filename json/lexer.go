package json

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

type (
	Lexer struct {
		currtoken Token
		currdata  rune
		currline  int
		reader    *bufio.Reader
		svalue    string
		nvalue    float64
		isused    bool
	}

	LexerSyntaxError struct {
		line int
		err  error
	}
)

var (
	eof rune = -1

	ErrNotTerminatedToken      = errors.New("Not terminated token")
	ErrNotTerminatedString     = errors.New("Not terminated string")
	ErrInvalidUnicodeCharacter = errors.New("Invalid unicode character")
	ErrInvalidEscapeCharacter  = errors.New("Invalid escape character")
	ErrInvalidNumberLiteral    = errors.New("Invalid number literal")
	ErrInvalidKeyword          = errors.New("Invalid keyword")
)

// NewLexer creates and initializes a new Lexer using r as its initial reader.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		currtoken: NO_TOKEN,
		currdata:  eof,
		currline:  0,
		reader:    bufio.NewReader(r),
		svalue:    "",
		nvalue:    0.,
		isused:    false,
	}
}

func (l *Lexer) genSyntaxError(e error) error {
	return &LexerSyntaxError{
		line: l.currline,
		err:  e,
	}
}

func (e *LexerSyntaxError) Error() string {
	return fmt.Sprintf("json.Lexer: line %d, %s", e.line, e.err.Error())
}

func (l *Lexer) next() error {
	if r, _, err := l.reader.ReadRune(); err != nil {
		if err == io.EOF {
			l.currdata = eof
		} else {
			return err
		}
	} else {
		l.currdata = r
	}
	return nil
}

func (l *Lexer) resetValues() {
	l.svalue = ""
	l.nvalue = 0.
}

func (l *Lexer) Lex() (Token, error) {
	// At first call Lex(), must call next method.
	if !l.isused {
		l.isused = true
		if err := l.next(); err != nil {
			return NO_TOKEN, err
		}
	}
	l.resetValues()
	for l.currdata != eof {
		switch l.currdata {
		case '\n':
			l.currline++
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			continue
		case '\t', '\r', ' ':
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			continue
		case '{':
			l.currtoken = TK_LBRACE
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			return TK_LBRACE, nil
		case '}':
			l.currtoken = TK_RBRACE
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			return TK_RBRACE, nil
		case '[':
			l.currtoken = TK_LBRACKET
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			return TK_LBRACKET, nil
		case ']':
			l.currtoken = TK_RBRACKET
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			return TK_RBRACKET, nil
		case ':':
			l.currtoken = TK_COLON
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			return TK_COLON, nil
		case ',':
			l.currtoken = TK_COMMA
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			return TK_COMMA, nil
		case '"':
			l.currtoken = TK_STRING_LITERAL
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			if err := l.readString(); err != nil {
				return NO_TOKEN, err
			}
			if err := l.next(); err != nil {
				return NO_TOKEN, err
			}
			return TK_STRING_LITERAL, nil
		default:
			if unicode.IsDigit(l.currdata) || l.currdata == '-' {
				l.currtoken = TK_NUMBER_LITERAL
				if err := l.readNumber(); err != nil {
					return NO_TOKEN, err
				}
				return TK_NUMBER_LITERAL, nil
			} else if unicode.IsLetter(l.currdata) {
				return l.readKeyword()
			} else {
				s := fmt.Sprintf("Unexpected character %x", l.currdata)
				return NO_TOKEN, l.genSyntaxError(errors.New(s))
			}
		}
	}
	return NO_TOKEN, nil
}

func isXDigit(r rune) bool {
	if unicode.IsDigit(r) {
		return true
	}
	switch r {
	case 'a', 'A', 'b', 'B', 'c', 'C', 'd', 'D', 'e', 'E', 'f', 'F':
		return true
	default:
		return false
	}
}

func (l *Lexer) readString() error {
	buf := new(bytes.Buffer)
	for l.currdata != '"' {
		switch l.currdata {
		case eof:
			return l.genSyntaxError(ErrNotTerminatedString)
		case '\n':
			return l.genSyntaxError(ErrNotTerminatedString)
		case '\\':
			l.next()
			switch l.currdata {
			case 'u':
				l.next()
				b := new(bytes.Buffer)
				for i := 0; i < 4; i++ {
					if isXDigit(l.currdata) {
						b.WriteRune(l.currdata)
						l.next()
					} else {
						return l.genSyntaxError(ErrInvalidUnicodeCharacter)
					}
				}
				if _, err := buf.Write(b.Bytes()); err != nil {
					return err
				}
			case '"', '/':
				buf.WriteRune(l.currdata)
				if err := l.next(); err != nil {
					return err
				}
			case 'b':
				buf.WriteRune('\b')
				if err := l.next(); err != nil {
					return err
				}
			case 'f':
				buf.WriteRune('\f')
				if err := l.next(); err != nil {
					return err
				}
			case 'n':
				buf.WriteRune('\n')
				if err := l.next(); err != nil {
					return err
				}
			case 'r':
				buf.WriteRune('\r')
				if err := l.next(); err != nil {
					return err
				}
			case 't':
				buf.WriteRune('\t')
				if err := l.next(); err != nil {
					return err
				}
			case '\\':
				buf.WriteRune('\\')
				if err := l.next(); err != nil {
					return err
				}
			default:
				return l.genSyntaxError(ErrInvalidEscapeCharacter)
			}
		default:
			buf.WriteRune(l.currdata)
			if err := l.next(); err != nil {
				return err
			}
		}
	}
	l.svalue = buf.String()
	return nil
}

func isExponent(r rune) bool {
	switch r {
	case 'e', 'E':
		return true
	default:
		return false
	}
}

func (l *Lexer) readNumber() error {
	buf := new(bytes.Buffer)
	buf.WriteRune(l.currdata)
	if err := l.next(); err != nil {
		return err
	}
	for l.currdata == '.' || unicode.IsDigit(l.currdata) || isExponent(l.currdata) {
		if isExponent(l.currdata) {
			buf.WriteRune(l.currdata)
			if err := l.next(); err != nil {
				return err
			}
			if l.currdata == '+' || l.currdata == '-' {
				buf.WriteRune(l.currdata)
				if err := l.next(); err != nil {
					return err
				}
				continue
			} else {
				return l.genSyntaxError(ErrInvalidNumberLiteral)
			}
		}
		buf.WriteRune(l.currdata)
		if err := l.next(); err != nil {
			return err
		}
	}
	res, err := strconv.ParseFloat(buf.String(), 64)
	if err != nil {
		l.genSyntaxError(ErrInvalidNumberLiteral)
	}
	l.nvalue = res
	return nil
}

func (l *Lexer) readKeyword() (Token, error) {
	buf := new(bytes.Buffer)
	buf.WriteRune(l.currdata)
	if err := l.next(); err != nil {
		return NO_TOKEN, err
	}
	for unicode.IsLetter(l.currdata) {
		buf.WriteRune(l.currdata)
		if err := l.next(); err != nil {
			return NO_TOKEN, err
		}
	}
	switch buf.String() {
	case "true":
		l.currtoken = TK_TRUE
		return TK_TRUE, nil
	case "false":
		l.currtoken = TK_FALSE
		return TK_FALSE, nil
	case "null":
		l.currtoken = TK_NULL
		return TK_NULL, nil
	default:
		return NO_TOKEN, ErrInvalidKeyword
	}
}
