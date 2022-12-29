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

	Let
	Macro
	Equals

	TypeByte
	TypeChar
	TypeInt16
	TypeInt32
	TypeInt64
	TypeFloat64

	Add
	Sub
	Mult
	Div
	Mod
	Pow

	BitAnd
	BitOr
	BitSRight
	BitSLeft

	SizeOf

	Dots

	LParen
	RParen

	Include
	Embed

	Error
	count // Count of all token types
)

func AllTokensCoveredTest() {
	if count != 37 {
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

	case Let:    return "let"
	case Macro:  return "mac"
	case Equals: return "="

	case TypeByte:    return "byte"
	case TypeChar:    return "char"
	case TypeInt16:   return "int16"
	case TypeInt32:   return "int32"
	case TypeInt64:   return "int64"
	case TypeFloat64: return "float64"

	case Add:  return "+"
	case Sub:  return "-"
	case Mult: return "*"
	case Div:  return "/"
	case Mod:  return "%"
	case Pow:  return "^"

	case BitAnd:    return "&"
	case BitOr:     return "|"
	case BitSRight: return ">>"
	case BitSLeft:  return "<<"

	case SizeOf: return "sizeof"

	case Dots: return ".."

	case LParen: return "("
	case RParen: return ")"

	case Include: return "include"
	case Embed:   return "embed"

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

	case Dec, Hex, Char, Float,
	     Oct, Bin, String: return fmt.Sprintf("'%v' of type '%v'", tok.Data, tok.Type)

	default: return "'" + tok.Type.String() +  "'"
	}
}

func NewEOF(where Where) Token {
	return Token{Type: EOF, Where: where}
}

func NewError(where Where, format string, args... interface{}) Token {
	return Token{Type: Error, Where: where, Data: fmt.Sprintf(format, args...)}
}
