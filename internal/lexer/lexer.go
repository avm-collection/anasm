package lexer

import "github.com/avm-collection/anasm/internal/token"

const EOF = '\x00'

type Lexer struct {
	input string
	pos   int
	ch    byte

	where token.Where
}

var Keywords = map[string]token.Type{
	"let": token.Let,
	"mac": token.Macro,

	"byte": token.TypeByte,
	"char": token.TypeChar,
	"i16":  token.TypeInt16,
	"i32":  token.TypeInt32,
	"i64":  token.TypeInt64,
	"f64":  token.TypeFloat64,

	"sizeof": token.SizeOf,

	"+": token.Add,
	"-": token.Sub,
	"*": token.Mult,
	"/": token.Div,
	"%": token.Mod,
	"^": token.Pow,

	"&":  token.BitAnd,
	"|":  token.BitOr,
	">>": token.BitSRight,
	"<<": token.BitSLeft,
}

func New(input, path string) *Lexer {
	token.AllTokensCoveredTest()

	l := &Lexer{input: input, pos: -1}
	l.next()

	l.where.Row  = 1
	l.where.Path = path

	return l
}

func isWordCh(ch byte) bool {
	switch ch {
	case '$', '_', '+', '-', '*', '/', '%', '>', '<', '&', '|', '^': return true

	default:
		return (ch >= 'a' && ch <= 'z') ||
		       (ch >= 'A' && ch <= 'Z') ||
		       (ch >= '0' && ch <= '9')
	}
}

func isDecDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isOctDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isBinDigit(ch byte) bool {
	return ch == '0' || ch == '1'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') ||
	       (ch >= 'a' && ch <= 'f') ||
	       (ch >= 'A' && ch <= 'F')
}

func isWhitespace(ch byte) bool {
	switch ch {
	case ' ', '\r', '\t', '\v', '\f', '\n': return true

	default: return false
	}
}

func (l *Lexer) NextToken() (tok token.Token) {
	for {
		start := l.where

		switch l.ch {
		case EOF: return token.NewEOF(l.where)

		case '#':
			l.skipComment()

			continue

		case '"':  tok = l.lexString()
		case '\'': tok = l.lexChar()

		case '.':
			if l.peek() == '.' {
				l.next()

				tok = token.Token{Type: token.Dots, Data: ".."}
				l.next()
			} else {
				tok = l.lexLabel()
			}

		case '-':
			if isDecDigit(l.peek()) {
				tok = l.lexNum()
			} else {
				tok = l.lexWord()
			}

		case '(':
			tok = token.Token{Type: token.LParen, Data: string(l.ch)}
			l.next()

		case ')':
			tok = token.Token{Type: token.RParen, Data: string(l.ch)}
			l.next()

		case ',':
			tok = token.Token{Type: token.Comma, Data: string(l.ch)}
			l.next()

		case '=':
			tok = token.Token{Type: token.Equals, Data: string(l.ch)}
			l.next()

		default:
			if isDecDigit(l.ch) {
				tok = l.lexNum()
			} else if isWordCh(l.ch) {
				tok = l.lexWord()
			} else if isWhitespace(l.ch) {
				l.next()

				continue
			} else {
				return token.NewError(l.where, "Unexpected character '%v'", string(l.ch))
			}
		}

		tok.Where = start

		break
	}

	return
}

func escapedCharToByte(ch byte) (byte, bool) {
	switch ch {
	case '0':  return 0,    true
	case 'a':  return '\a', true
	case 'b':  return '\b', true
	case 'e':  return 27,   true
	case 'f':  return '\f', true
	case 'n':  return '\n', true
	case 'r':  return '\r', true
	case 't':  return '\t', true
	case 'v':  return '\v', true
	case '\\': return '\\', true
	case '"':  return '"',  true
	case '\'': return '\'', true
	}

	return 0, false
}

func (l *Lexer) lexString() token.Token {
	str    := ""
	escape := false

	for l.next(); !(l.ch == '"' && !escape); l.next() {
		switch l.ch {
		case '\\':
			if escape {
				escape = false
				str   += "\\"
			} else {
				escape = true
			}

		case '\n': return token.NewError(l.where, "Expected '\"', got 'new line'")
		case EOF:  return token.NewError(l.where, "Expected '\"', got 'end of file'")

		default:
			if escape {
				ret, ok := escapedCharToByte(l.ch)
				if !ok {
					return token.NewError(l.where, "Unknown escape sequence '\\%v'", string(l.ch))
				}
				escape = false

				str += string(ret)
			} else {
				str += string(l.ch)
			}
		}
	}

	l.next()

	return token.Token{Type: token.String, Data: str}
}

func (l *Lexer) lexChar() token.Token {
	str := ""

	if l.next(); l.ch == '\\' {
		l.next()
		ret, ok := escapedCharToByte(l.ch)
		if !ok {
			return token.NewError(l.where, "Unknown escape sequence '\\%v'", string(l.ch))
		}

		str += string(ret)
	} else {
		str += string(l.ch)
	}

	if l.next(); l.ch != '\'' {
		return token.NewError(l.where, "Character literal expected to be exactly 1 byte long")
	}

	l.next()

	return token.Token{Type: token.Char, Data: str}
}

func (l *Lexer) lexNum() token.Token {
	if l.ch == '0' && (l.peek() == 'x' || l.peek() == 'X') {
		l.next()
		l.next()

		return l.lexHex()
	} else if l.ch == '0' && (l.peek() == 'o' || l.peek() == 'O') {
		l.next()
		l.next()

		return l.lexOct()
	} else if l.ch == '0' && (l.peek() == 'b' || l.peek() == 'B') {
		l.next()
		l.next()

		return l.lexBin()
	} else {
		return l.lexDec()
	}
}

func (l *Lexer) lexHex() token.Token {
	str := ""

	for isHexDigit(l.ch) {
		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Hex, Data: str}
}

func (l *Lexer) lexOct() token.Token {
	str := ""

	for {
		if !isOctDigit(l.ch) {
			if isHexDigit(l.ch) {
				return token.NewError(l.where, "Unexpected character '%v' in octal number",
				                      string(l.ch))
			}

			break
		}

		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Oct, Data: str}
}

func (l *Lexer) lexBin() token.Token {
	str := ""

	for {
		if !isBinDigit(l.ch) {
			if isHexDigit(l.ch) {
				return token.NewError(l.where, "Unexpected character '%v' in binary number",
				                      string(l.ch))
			}

			break
		}

		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Bin, Data: str}
}

func (l *Lexer) lexDec() token.Token {
	str     := ""
	float   := false
	atStart := true

	for !isWhitespace(l.ch) && l.ch != ',' && l.ch != ':' {
		if l.ch == '.' {
			if float {
				return token.NewError(l.where, "Unexpected '.' in float number")
			}

			float = true
		} else if !isDecDigit(l.ch) && !(atStart && l.ch == '-') {
			if isHexDigit(l.ch) {
				return token.NewError(l.where, "Unexpected character '%v' in decimal number",
				                      string(l.ch))
			}

			break
		}

		str += string(l.ch)

		if atStart {
			atStart = false
		}

		l.next()
	}

	if float {
		return token.Token{Type: token.Float, Data: str}
	} else {
		return token.Token{Type: token.Dec, Data: str}
	}
}

func (l *Lexer) lexLabel() token.Token {
	if l.next(); !isWordCh(l.ch) {
		return token.NewError(l.where, "Unexpected character '%v' in label name",
		                      string(l.ch))
	}

	return token.Token{Type: token.Label, Data: l.readWord()}
}

func (l *Lexer) lexWord() token.Token {
	str := l.readWord()
	type_, ok := Keywords[str]
	if ok {
		return token.Token{Type: type_, Data: str}
	}

	return token.Token{Type: token.Word, Data: str}
}

func (l *Lexer) readWord() (str string) {
	for isWordCh(l.ch) {
		str += string(l.ch)

		l.next()
	}

	return str
}

func (l *Lexer) skipComment() {
	for l.ch != EOF && l.ch != '\n' {
		l.next()
	}
}

func (l *Lexer) next() {
	l.pos ++
	if l.pos >= len(l.input) {
		l.ch = EOF
	} else {
		l.ch = l.input[l.pos]
	}

	if l.ch == '\n' {
		l.where.Col = 0
		l.where.Row ++
	} else {
		l.where.Col ++
	}
}

func (l *Lexer) peek() byte {
	if l.pos + 1 >= len(l.input) {
		return EOF
	} else {
		return l.input[l.pos + 1]
	}
}

