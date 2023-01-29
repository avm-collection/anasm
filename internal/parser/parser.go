package parser

import (
	"os"
	"strconv"
	"path/filepath"

	"github.com/avm-collection/anasm/pkg/errors"
	"github.com/avm-collection/anasm/pkg/agen"
	"github.com/avm-collection/anasm/internal/lexer"
	"github.com/avm-collection/anasm/internal/token"
	"github.com/avm-collection/anasm/internal/node"
)

type Parser struct {
	statements *node.Statements

	tok token.Token
	l  *lexer.Lexer

	input, path string
}

func New(input, path string) *Parser {
	return &Parser{input: input, path: path}
}

func (p *Parser) Parse() *node.Statements {
	p.statements = &node.Statements{}
	p.parseFile(p.input, p.path)

	return p.statements
}

func (p *Parser) next() {
	if p.tok.Type == token.EOF {
		return
	}

	if p.tok = p.l.NextToken(); p.tok.Type == token.Error {
		errors.Error(p.tok.Where, p.tok.Data)
		os.Exit(1)
	}
}

func (p *Parser) parseFile(input, path string) {
	prevLexer := p.l
	prevTok   := p.tok

	p.l = lexer.New(input, path)

	if p.tok = p.l.NextToken(); p.tok.Type == token.Error {
		errors.Error(p.tok.Where, p.tok.Data)
		os.Exit(1)
	}

	for p.tok.Type != token.EOF {
		var s node.Statement

		switch p.tok.Type {
		case token.Id:    s = p.parseInst()
		case token.Label: s = p.parseLabel()
		case token.Let:   s = p.parseLet()
		case token.Embed: s = p.parseEmbed()
		case token.Macro: s = p.parseMacro()

		case token.Include:
			p.evalInclude()
			continue

		default: s = p.parseImplicitPush()
		}

		p.statements.List = append(p.statements.List, s)
	}

	p.l   = prevLexer
	p.tok = prevTok
}

func (p *Parser) evalInclude() {
	p.next()
	path := p.parseString()

	toInclude := path.Value
	if toInclude[0] == '.' {
		toInclude = filepath.Dir(p.path) + toInclude[1:]
	}

	data, err := os.ReadFile(toInclude)
	if err != nil {
		errors.Error(path.GetToken().Where, "Could not open file '%v'", toInclude)
		return
	}

	p.parseFile(string(data), path.Value)
}

func (p *Parser) parseImplicitPush() *node.Inst {
	return &node.Inst{Token: p.tok, Name: "psh", Arg: p.parseExpr()}
}

func (p *Parser) parseMacro() *node.Macro {
	n := &node.Macro{Token: p.tok}
	p.next()

	n.Name = p.parseId()
	if p.tok.Type != token.Equals {
		errors.Error(p.tok.Where, "Expected assignment with '%v', got %v",
		             token.Equals, token.Dots, p.tok)
		p.next()
		return nil
	}
	p.next()

	n.Value = p.parseExpr()
	return n
}

func (p *Parser) parseLet() node.Statement {
	n := &node.Let{Token: p.tok}
	p.next()

	n.Name = p.parseId()
	n.Type = p.parseType()

	if p.tok.Type != token.Equals {
		errors.Error(p.tok.Where, "Expected assignment with '%v' or size with '%v', got %v",
		             token.Equals, token.Dots, p.tok)
		p.next()
		return nil
	}

	p.next()

	for {
		val := p.parseExpr()
		if p.tok.Type == token.Dots {
			fill := &node.Fill{Token: p.tok}
			p.next()

			fill.Value = val
			fill.Count = p.parseExpr()

			n.Values = append(n.Values, fill)
		} else {
			n.Values = append(n.Values, val)
		}

		if p.tok.Type != token.Comma {
			break
		} else {
			p.next()
		}
	}

	return n
}

func (p *Parser) parseEmbed() *node.Embed {
	n := &node.Embed{Token: p.tok}
	p.next()

	n.Name = p.parseId()
	n.Path = p.parseString()
	return n
}

func (p *Parser) parseLabel() *node.Label {
	n := &node.Label{Token: p.tok}

	n.Name = &node.Id{Token: p.tok, Value: p.tok.Data}
	p.next()
	return n
}

func (p *Parser) parseInst() *node.Inst {
	n := &node.Inst{Token: p.tok}

	inst, ok := agen.Insts[p.tok.Data]
	if !ok {
		errors.Error(p.tok.Where, "'%v' is not a valid instruction", p.tok.Data)
		p.next()
		return nil
	}
	n.Name = p.tok.Data

	p.next()
	if inst.HasArg {
		n.Arg = p.parseExpr()
	}

	return n
}

func (p *Parser) parseExpr() node.Expr {
	switch p.tok.Type {
	case token.Id:     return p.parseId()
	case token.LParen: return p.parseFunc()
	case token.String: return p.parseString()
	case token.Float:  return p.parseFloat()

	default:
		if p.tok.Type.IsInt() {
			return p.parseInt()
		} else if p.tok.Type.IsType() {
			return p.parseType()
		} else {
			errors.Error(p.tok.Where, "Unexpected %v in expression", p.tok)
			p.next()
			return nil
		}
	}
}

func (p *Parser) parseId() *node.Id {
	n := &node.Id{Token: p.tok}

	if p.tok.Type != token.Id {
		errors.Error(p.tok.Where, "Expected identifier, got %v", p.tok)
		p.next()
		return nil
	}

	if _, ok := agen.Insts[p.tok.Data]; ok {
		errors.Error(p.tok.Where, "Expected identifier, got instruction '%v'", p.tok.Data)
		p.next()
		return nil
	}

	n.Value = p.tok.Data
	p.next()
	return n
}

func (p *Parser) parseString() *node.String {
	n := &node.String{Token: p.tok}

	if p.tok.Type != token.String {
		errors.Error(p.tok.Where, "Expected string, got %v", p.tok)
		p.next()
		return nil
	}

	n.Value = p.tok.Data
	p.next()
	return n
}

func (p *Parser) parseInt() *node.Int {
	n := &node.Int{Token: p.tok}

	switch p.tok.Type {
	case token.Dec:  n.Value, _ = strconv.ParseInt(p.tok.Data, 10, 64)
	case token.Hex:  n.Value, _ = strconv.ParseInt(p.tok.Data, 16, 64)
	case token.Oct:  n.Value, _ = strconv.ParseInt(p.tok.Data, 8,  64)
	case token.Bin:  n.Value, _ = strconv.ParseInt(p.tok.Data, 2,  64)
	case token.Char: n.Value    = int64(p.tok.Data[0])

	default:
		errors.Error(p.tok.Where, "Expected an integer or a character, got %v", p.tok)
		p.next()
		return nil
	}

	p.next()
	return n
}

func (p *Parser) parseFloat() *node.Float {
	n := &node.Float{Token: p.tok}

	if p.tok.Type != token.Float {
		errors.Error(p.tok.Where, "Expected a float, got %v", p.tok)
		p.next()
		return nil
	}

	n.Value, _ = strconv.ParseFloat(p.tok.Data, 8)
	p.next()
	return n
}

func (p *Parser) parseType() *node.Type {
	n := &node.Type{Token: p.tok}

	switch p.tok.Type {
	case token.TypeByte, token.TypeChar:     n.Type = agen.I8
	case token.TypeInt16:                    n.Type = agen.I16
	case token.TypeInt32:                    n.Type = agen.I32
	case token.TypeInt64, token.TypeFloat64: n.Type = agen.I64

	default:
		errors.Error(p.tok.Where, "Expected a type (byte/char/i16/i32/i64/f164), got %v", p.tok)
		p.next()
		return nil
	}

	p.next()
	return n
}

func (p *Parser) parseFunc() node.Expr {
	start := p.tok
	p.next()

	if p.tok.Type == token.SizeOf {
		return p.parseSizeOf(start)
	} else if p.tok.Type.IsBinOp() {
		return p.parseBinOp(start)
	} else {
		errors.Error(p.tok.Where, "Expected function, got %v", p.tok)
		p.next()
		return nil
	}
}

func (p *Parser) parseSizeOf(start token.Token) *node.SizeOf {
	n := &node.SizeOf{Token: start}

	p.next()
	if p.tok.Type == token.Id {
		n.Id = p.parseId()
	} else if p.tok.Type.IsType() {
		n.Type = p.parseType()
	} else {
		errors.Error(p.tok.Where, "Expected an identifier or a type, got %v", p.tok)
		p.next()
		return nil
	}

	if p.tok.Type != token.RParen {
		errors.Error(p.tok.Where, "Expected matching '%v', got %v", token.RParen, p.tok)
		errors.Note(start.Where, "Opened here")
		return nil
	}
	p.next()

	return n
}

func (p *Parser) parseBinOp(start token.Token) *node.BinOp {
	n := &node.BinOp{Token: start}
	n.Op = p.tok.Data

	p.next()
	for p.tok.Type != token.RParen {
		n.Args = append(n.Args, p.parseExpr())
	}

	if p.tok.Type != token.RParen {
		errors.Error(p.tok.Where, "Expected matching '%v', got %v", token.RParen, p.tok)
		errors.Note(start.Where, "Opened here")
		return nil
	}
	p.next()

	return n
}
