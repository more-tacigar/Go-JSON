package json

type Token int

const (
	TK_LBRACE Token = iota
	TK_RBRACE
	TK_LBRACKET
	TK_RBRACKET
	TK_COMMA
	TK_COLON
	TK_EOF
	TK_NULL
	TK_FALSE
	TK_TRUE
	TK_STRING_LITERAL
	TK_NUMBER_LITERAL
	NO_TOKEN
)

func (t Token) String() string {
	switch t {
	case TK_LBRACE:
		return "{"
	case TK_RBRACE:
		return "}"
	case TK_LBRACKET:
		return "["
	case TK_RBRACKET:
		return "]"
	case TK_COMMA:
		return ","
	case TK_COLON:
		return ":"
	case TK_EOF:
		return "eof"
	case TK_NULL:
		return "null"
	case TK_FALSE:
		return "false"
	case TK_TRUE:
		return "true"
	case TK_STRING_LITERAL:
		return "<string literal>"
	case TK_NUMBER_LITERAL:
		return "<number literal>"
	case NO_TOKEN:
		return "<no token>"
	default:
		return "<invalid>"
	}
}
