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

	Dec
	Hex
	Oct
	Float

	LabelRef // Label reference, for example 'jmp @label'

	Error
	count // Count of all token types
)

func (t Type) String() string {
	if count != 10 {
		panic("Cover all token types")
	}

	switch t {
	case EOF: return "end of file"

	case Word:  return "word"
	case Label: return "label declaration"

	case Dec:   return "decimal integer"
	case Hex:   return "hexadecimal integer"
	case Oct:   return "octal integer"
	case Float: return "float"

	case LabelRef: return "label reference"

	case Error: return "error"

	default: panic("Unreachable")
	}
}

type Token struct {
	Type Type
	Data string

	Where Where
}

func (t Token) String() string {
	switch t.Type {
	case EOF: return "'end of file'"

	default: return fmt.Sprintf("'%v' of type '%v'", t.Data, t.Type)
	}
}

func NewEOF(where Where) Token {
	return Token{Type: EOF, Where: where}
}

func NewError(where Where, format string, args... interface{}) Token {
	return Token{Type: Error, Where: where, Data: fmt.Sprintf(format, args...)}
}

func (t Token) IsArg() bool {
	switch t.Type {
	case Dec, Hex, Oct, Float, LabelRef: return true

	default: return false
	}
}
