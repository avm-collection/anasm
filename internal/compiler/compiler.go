package compiler

import (
	"os"
	"math"

	"github.com/avm-collection/anasm/pkg/errors"
	"github.com/avm-collection/anasm/pkg/agen"
	"github.com/avm-collection/anasm/internal/token"
	"github.com/avm-collection/anasm/internal/parser"
	"github.com/avm-collection/anasm/internal/node"
)

const (
	VersionMajor = 1
	VersionMinor = 14
	VersionPatch // Not keeping track of the patch
)

const EntryLabel = "entry"

type Label struct {
	Token token.Token
	Addr  agen.Word
}

type Var struct {
	Token token.Token
	Size  agen.Word
	Addr  agen.Word
}

type Macro struct {
	Token token.Token
	Value agen.Word
}

type Compiler struct {
	a       *agen.AGEN
	program *node.Statements

	labels map[string]Label
	vars   map[string]Var
	macros map[string]Macro

	input, path string
}

func New(input, path string) *Compiler {
	return &Compiler{
		a: agen.New(), input: input, path: path,
		labels: make(map[string]Label),
		vars:   make(map[string]Var),
		macros: make(map[string]Macro),
	}
}

func (c *Compiler) Compile() bool {
	p := parser.New(c.input, c.path)
	if c.program = p.Parse(); errors.Happened() {
		return false
	}

	if c.preproc(); errors.Happened() {
		return false
	}

	if c.compile(); errors.Happened() {
		return false
	}

	if _, ok := c.labels[EntryLabel]; !ok {
		errors.Simple("Program entry point label '%v' not found", EntryLabel)
		return false
	}

	return true
}

func (c *Compiler) CreateExec(path string, executable bool) error {
	return c.a.CreateExecAVM(path, executable)
}

func (c *Compiler) preproc() {
	var addr agen.Word
	for _, s := range c.program.List {
		switch n := s.(type) {
		case *node.Label:
			if c.redefined(n.Name) {
				break
			}

			c.labels[n.Name.Value] = Label{Token: n.Token, Addr: addr}
			if n.Name.Value == EntryLabel {
				c.a.SetEntry(addr)
			}

		case *node.Inst: addr ++
		default:
		}
	}
}

func (c *Compiler) compile() {
	for _, s := range c.program.List {
		switch n := s.(type) {
		case *node.Label: continue;

		case *node.Macro: c.compileMacro(n)
		case *node.Embed: c.compileEmbed(n)
		case *node.Let:   c.compileLet(n)
		case *node.Inst:  c.compileInst(n)
		}
	}
}

func (c *Compiler) redefined(name *node.Id) bool {
	if prev, ok := c.labels[name.Value]; ok {
		errors.Error(name.Token.Where, "Label '%v' redefined", name.Value)
		errors.Note(prev.Token.Where, "Previously defined here")
		return true
	} else if prev, ok := c.vars[name.Value]; ok {
		errors.Error(name.Token.Where, "Variable '%v' redefined", name.Value)
		errors.Note(prev.Token.Where, "Previously defined here")
		return true
	} else if prev, ok := c.macros[name.Value]; ok {
		errors.Error(name.Token.Where, "Macro '%v' redefined", name.Value)
		errors.Note(prev.Token.Where, "Previously defined here")
		return true
	}

	return false
}

func (c *Compiler) compileMacro(n *node.Macro) {
	if c.redefined(n.Name) {
		return
	}

	c.macros[n.Name.Value] = Macro{Token: n.Token, Value: c.evalExpr(n.Value)}
}

func (c *Compiler) compileEmbed(n *node.Embed) {
	if c.redefined(n.Name) {
		return
	}

	data, err := os.ReadFile(n.Path.Value)
	if err != nil {
		errors.Error(n.Token.Where, "Could not embed file '%v'", n.Path.Value)
		return
	}

	size := c.a.MemorySize()
	addr := c.a.AddMemoryString(string(data))
	size  = c.a.MemorySize() - size

	c.vars[n.Name.Value] = Var{Token: n.Token, Addr: addr, Size: size}
}

func (c *Compiler) compileLet(n *node.Let) {
	if c.redefined(n.Name) {
		return
	}

	list := []agen.Word{}
	for _, expr := range n.Values {
		switch e := expr.(type) {
		case *node.Fill:
			count := c.evalExpr(e.Count)
			value := c.evalExpr(e.Value)
			for i := agen.Word(0); i < count; i ++ {
				list = append(list, value)
			}

		case *node.String:
			for _, ch := range e.Value {
				list = append(list, agen.Word(ch))
			}

		default: list = append(list, c.evalExpr(expr))
		}
	}

	size := c.a.MemorySize()
	addr := c.a.AddMemoryInt(list, n.Type.Type)
	size  = c.a.MemorySize() - size

	c.vars[n.Name.Value] = Var{Token: n.Token, Addr: addr, Size: size}
}

func (c *Compiler) compileInst(n *node.Inst) {
	if n.Arg == nil {
		c.a.AddInst(n.Name)
	} else {
		c.a.AddInstWith(n.Name, c.evalExpr(n.Arg))
	}
}

func (c *Compiler) evalExpr(e node.Expr) agen.Word {
	switch n := e.(type) {
	case *node.Int:   return agen.Word(n.Value)
	case *node.Float: return agen.Word(math.Float64bits(n.Value))
	case *node.Id:
		if label, ok := c.labels[n.Value]; ok {
			return label.Addr
		} else if var_, ok := c.vars[n.Value]; ok {
			return var_.Addr
		} else if macro, ok := c.macros[n.Value]; ok {
			return macro.Value
		}

	case *node.BinOp:  return c.evalBinOp(n)
	case *node.SizeOf: return c.evalSizeOf(n)

	case *node.Type:   errors.Error(n.Token.Where, "Unexpected type in constant expression")
	case *node.String: errors.Error(n.Token.Where, "Unexpected string in constant expression")
	case *node.Fill:   errors.Error(n.Token.Where, "Unexpected fill in constant expression")
	default: errors.Error(n.GetToken().Where, "Unexpected %v in constant expression", n.GetToken())
	}

	return 0;
}

func (c *Compiler) evalSizeOf(n *node.SizeOf) agen.Word {
	if n.Id == nil {
		switch n.Type.Type {
		case agen.I8:  return 1
		case agen.I16: return 2
		case agen.I32: return 4
		case agen.I64: return 8
		}
	} else {
		if _, ok := c.labels[n.Id.Value]; ok {
			errors.Error(n.Token.Where, "Cannot get size of label '%v'", n.Id.Value)
		} else if var_, ok := c.vars[n.Id.Value]; ok {
			return var_.Size
		} else if _, ok := c.macros[n.Id.Value]; ok {
			errors.Error(n.Token.Where, "Cannot get size of macro '%v'", n.Id.Value)
		}
	}

	return 0
}

func (c *Compiler) evalBinOp(n *node.BinOp) agen.Word {
	result := c.evalExpr(n.Args[0])
	for i, expr := range n.Args {
		if i == 0 {
			continue
		}

		switch n.Op {
		case "+": result += c.evalExpr(expr)
		case "-": result -= c.evalExpr(expr)
		case "*": result *= c.evalExpr(expr)
		case "/": result /= c.evalExpr(expr)
		case "%": result %= c.evalExpr(expr)
		case "^": result  = agen.Word(math.Pow(float64(result), float64(c.evalExpr(expr))))
		}
	}

	return result
}
