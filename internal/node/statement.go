package node

import (
	"fmt"

	"github.com/avm-collection/anasm/internal/token"
)

type Statements struct {
	List []Statement
}

func (n *Statements) statement() {}
func (n *Statements) GetToken() token.Token {return n.List[0].GetToken()}
func (n *Statements) String()   (s string) {
	s += "(\n"
	for _, statement := range n.List {
		s += statement.String() + "\n"
	}
	s += ")"

	return
}

type Inst struct {
	Token token.Token

	Name string
	Arg  Expr
}

func (n *Inst) statement() {}
func (n *Inst) GetToken() token.Token {return n.Token}
func (n *Inst) String()   string {
	if n.Arg == nil {
		return fmt.Sprintf("(%v)", n.Name)
	} else {
		return fmt.Sprintf("(%v %v)", n.Name, n.Arg)
	}
}

type Label struct {
	Token token.Token

	Name *Id
}

func (n *Label) statement() {}
func (n *Label) GetToken() token.Token {return n.Token}
func (n *Label) String()   string      {return fmt.Sprintf("(label %v)", n.Name)}

type Embed struct {
	Token token.Token

	Name *Id
	Path *String
}

func (n *Embed) statement() {}
func (n *Embed) GetToken() token.Token {return n.Token}
func (n *Embed) String()   string      {return fmt.Sprintf("(embed %v %v)", n.Name, n.Path)}

type Macro struct {
	Token token.Token

	Name *Id
	Value Expr
}

func (n *Macro) statement() {}
func (n *Macro) GetToken() token.Token {return n.Token}
func (n *Macro) String()   string      {return fmt.Sprintf("(macro %v %v)", n.Name, n.Value)}

type Let struct {
	Token token.Token

	Name  *Id
	Type  *Type
	Values []Expr
}

func (n *Let) statement() {}
func (n *Let) GetToken() token.Token {return n.Token}
func (n *Let) String()   (s string) {
	s += fmt.Sprintf("(let %v %v", n.Name, n.Type)
	for _, val := range n.Values {
		s += fmt.Sprintf(" %v", val)
	}
	s += ")"

	return
}
