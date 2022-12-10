package lexer

import "github.com/LordOfTrident/anasm/internal/token"

const EOF = '\x00'

type Lexer struct {
	input string
	pos   int
	ch    byte

	where token.Where
}

func New(input, path string) *Lexer {
	l := &Lexer{input: input, pos: -1}
	l.next()

	l.where.Row  = 1
	l.where.Path = path

	return l
}

func isWordCh(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || ch == '$' ||
	       (ch >= 'A' && ch <= 'Z') || ch == '_' ||
	       (ch >= '0' && ch <= '9')
}

func isDecDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isOctDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
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

		case '@': tok = l.lexLabelRef()
		case '&': tok = l.lexReg()
		case '.': tok = l.lexLabel()

		case ':':
			tok = token.Token{Type: token.Colon, Data: ":"}
			l.next()

		case ',':
			tok = token.Token{Type: token.Comma, Data: ","}
			l.next()

		case '-': tok = l.lexNum()

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

func (l *Lexer) lexNum() token.Token {
	if l.ch == '0' && (l.peek() == 'x' || l.peek() == 'X') {
		l.next()
		l.next()

		return l.lexHex()
	} else if l.ch == '0' && (l.peek() == 'o' || l.peek() == 'O') {
		l.next()
		l.next()

		return l.lexOct()
	} else {
		return l.lexDec()
	}
}

func (l *Lexer) lexHex() token.Token {
	str := ""

	for !isWhitespace(l.ch) && l.ch != ',' && l.ch != ':' {
		if !isHexDigit(l.ch) {
			return token.NewError(l.where, "Unexpected character '%v' in hexadecimal number",
			                      string(l.ch))
		}

		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Hex, Data: str}
}

func (l *Lexer) lexOct() token.Token {
	str := ""

	for !isWhitespace(l.ch) && l.ch != ',' && l.ch != ':' {
		if !isHexDigit(l.ch) {
			return token.NewError(l.where, "Unexpected character '%v' in octal number",
			                      string(l.ch))
		}

		str += string(l.ch)

		l.next()
	}

	return token.Token{Type: token.Oct, Data: str}
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
		} else if !isDecDigit(l.ch) || l.ch == '-' {
			if !(l.ch == '-' && atStart) {
				return token.NewError(l.where, "Unexpected character '%v' in decimal number",
				                      string(l.ch))
			}
		}

		str += string(l.ch)

		l.next()

		if atStart {
			atStart = false
		}
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

func (l *Lexer) lexReg() token.Token {
	if l.next(); !isWordCh(l.ch) {
		return token.NewError(l.where, "Unexpected character '%v' in register name",
		                      string(l.ch))
	}

	return token.Token{Type: token.Reg, Data: l.readWord()}
}

func (l *Lexer) lexLabelRef() token.Token {
	if l.next(); !isWordCh(l.ch) {
		return token.NewError(l.where, "Unexpected character '%v' in label name",
		                      string(l.ch))
	}

	return token.Token{Type: token.LabelRef, Data: l.readWord()}
}

func (l *Lexer) lexWord() token.Token {
	return token.Token{Type: token.Word, Data: l.readWord()}
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

