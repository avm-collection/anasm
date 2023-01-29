package token

import "fmt"

type Where struct {
	Row,  Col, Len  int
	Path, Line      string
}

func (w Where) AtRow()   int    {return w.Row}
func (w Where) AtCol()   int    {return w.Col}
func (w Where) GetLen()  int    {return w.Len}
func (w Where) InFile()  string {return w.Path}
func (w Where) GetLine() string {return w.Line}
func (w Where) String()  string {
	return fmt.Sprintf("%v:%v:%v", w.Path, w.Row, w.Col)
}

type Type int
const (
	EOF = Type(iota)

	Id
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

// TODO: Somehow make this compile-time
func AllTokensCoveredTest() {
	if count != 37 {
		panic("Cover all token types")
	}
}

func (t Type) String() string {
	switch t {
	case EOF: return "end of file"

	case Id:    return "identifier"
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

func (type_ Type) IsInt() bool {
	switch type_ {
	case Dec, Hex, Oct, Bin, Char: return true

	default: return false
	}
}

func (type_ Type) IsType() bool {
	switch type_ {
	case TypeByte, TypeChar, TypeInt16, TypeInt32, TypeInt64, TypeFloat64: return true

	default: return false
	}
}

func (type_ Type) IsBinOp() bool {
	switch type_ {
	case Add, Sub, Mult, Div, Mod, Pow: return true

	default: return false
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
