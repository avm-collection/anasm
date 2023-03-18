package node

import (
	"fmt"

	"github.com/avm-collection/agen"

	"github.com/avm-collection/anasm/internal/token"
)

type Int struct {
	Token token.Token

	Value int64
}

func (n *Int) expr() {}
func (n *Int) GetToken() token.Token {return n.Token}
func (n *Int) String()   string      {return fmt.Sprintf("%v", n.Value)}

type Float struct {
	Token token.Token

	Value float64
}

func (n *Float) expr() {}
func (n *Float) GetToken() token.Token {return n.Token}
func (n *Float) String()   string      {return fmt.Sprintf("%v", n.Value)}

type String struct {
	Token token.Token

	Value string
}

func (n *String) expr() {}
func (n *String) GetToken() token.Token {return n.Token}
func (n *String) String()   string      {return fmt.Sprintf("'%v'", n.Value)}

type Id struct {
	Token token.Token

	Value string
}

func (n *Id) expr() {}
func (n *Id) GetToken() token.Token {return n.Token}
func (n *Id) String()   string      {return n.Value}

type Type struct {
	Token token.Token

	Type agen.Type
}

func (n *Type) expr() {}
func (n *Type) GetToken() token.Token {return n.Token}
func (n *Type) String()   string      {return n.Token.Data}

type BinOp struct {
	Token token.Token

	Op   string
	Args []Expr
}

func (n *BinOp) expr() {}
func (n *BinOp) GetToken() token.Token {return n.Token}
func (n *BinOp) String()   (s string) {
	s += "(" + n.Op
	for _, arg := range n.Args {
		s += " " + arg.String()
	}
	s += ")"

	return
}

type SizeOf struct {
	Token token.Token

	Id   *Id
	Type *Type
}

func (n *SizeOf) expr() {}
func (n *SizeOf) GetToken() token.Token {return n.Token}
func (n *SizeOf) String()   string {
	if n.Id == nil {
		return fmt.Sprintf("(sizeof %v)", n.Type)
	} else {
		return fmt.Sprintf("(sizeof %v)", n.Id)
	}
}

type Fill struct {
	Token token.Token

	Value Expr
	Count Expr
}

func (n *Fill) expr() {}
func (n *Fill) GetToken() token.Token {return n.Token}
func (n *Fill) String()   string {
	return fmt.Sprintf("(.. %v %v)", n.Value, n.Count)
}
