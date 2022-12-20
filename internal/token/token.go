package token

import "fmt"

type Where struct {
	Row, Col int
	Path     string
}

func (w Where) String() string {
	return fmt.Sprintf("%v:%v:%v", w.Path, w.Row, w.Col)
}

type Type int
const (
	EOF = iota

	Word
	Label
	Comma

	Dec
	Hex
	Oct
	Bin
	Char
	Float
	String

	Addr // Variables, label references, for example 'jmp @label'

	Let
	Macro

	Size8
	Size16
	Size32
	Size64

	Add
	Sub
	Mult
	Div
	Mod
	SizeOf

	Dots

	LParen
	RParen

	Error
	count // Count of all token types
)

func AllTokensCoveredTest() {
	if count != 28 {
		panic("Cover all token types")
	}
}

func (t Type) String() string {
	switch t {
	case EOF: return "end of file"

	case Word:  return "word"
	case Label: return "label declaration"
	case Comma: return ","

	case Dec:    return "decimal integer"
	case Hex:    return "hexadecimal integer"
	case Oct:    return "octal integer"
	case Bin:    return "binary integer"
	case Char:   return "character"
	case Float:  return "float"
	case String: return "string"

	case Addr: return "address"

	case Let:   return "let"
	case Macro: return "mac"

	case Size8:  return "sz8"
	case Size16: return "sz16"
	case Size32: return "sz32"
	case Size64: return "sz64"

	case Add:    return "+"
	case Sub:    return "-"
	case Mult:   return "*"
	case Div:    return "/"
	case Mod:    return "%"
	case SizeOf: return "szof"

	case Dots: return ".."

	case LParen: return "("
	case RParen: return ")"

	case Error: return "error"

	default: panic("Unreachable")
	}
}

type Token struct {
	Type Type
	Data string

	Where Where
}

func (tok Token) String() string {
	switch tok.Type {
	case EOF: return "'end of file'"
	case Size8, Size16, Size32, Size64, SizeOf, Add, Sub, Macro,
	     Mult, Div, Mod, Comma, Let, Dots, LParen, RParen: return "'" + tok.Type.String() +  "'"

	default: return fmt.Sprintf("'%v' of type '%v'", tok.Data, tok.Type)
	}
}

func NewEOF(where Where) Token {
	return Token{Type: EOF, Where: where}
}

func NewError(where Where, format string, args... interface{}) Token {
	return Token{Type: Error, Where: where, Data: fmt.Sprintf(format, args...)}
}

func (t Token) IsConstExprSimple() bool {
	switch t.Type {
	case Dec, Hex, Oct, Bin, Char, Float, Addr, Word: return true

	default: return false
	}
}
