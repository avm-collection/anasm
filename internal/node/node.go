package node

import "github.com/avm-collection/anasm/internal/token"

type Node interface {
	GetToken() token.Token
	String()   string
}

type Statement interface {
	Node
	statement()
}

type Expr interface {
	Node
	expr()
}
