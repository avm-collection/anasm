package compiler

import (
	"fmt"
	"os"
	"strconv"
	"bytes"
	"math"
	"encoding/binary"

	"github.com/avm-collection/anasm/internal/token"
	"github.com/avm-collection/anasm/internal/lexer"
)

type Word uint64

const (
	VersionMajor = 1
	VersionMinor = 10
	VersionPatch // Not keeping track of the patch
)

type Var struct {
	Addr Word
	Size Word
}

type Compiler struct {
	toks []token.Token
	pos  Word
	tok  token.Token

	programSize, memorySize, entryPoint Word

	labels map[string]Word
	vars   map[string]Var
	macros map[string]Word

	memory  bytes.Buffer
	program bytes.Buffer

	l *lexer.Lexer

	errs []error
	maxE int
}

func New(input, path string) *Compiler {
	return &Compiler{l: lexer.New(input, path),
	                 labels: make(map[string]Word),
	                 vars:   make(map[string]Var),
	                 macros: make(map[string]Word)}
}

func fileWriteWord(f *os.File, word Word) error {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, word)

	_, err := f.Write(buf.Bytes())

	return err
}

func (c *Compiler) errorAt(where token.Where, format string, args... interface{}) error {
	return fmt.Errorf("Error at %v: %v", where, fmt.Sprintf(format, args...))
}

func (c *Compiler) errorHere(format string, args... interface{}) error {
	return c.errorAt(c.tok.Where, fmt.Sprintf(format, args...))
}

func (c *Compiler) isTokInst(tok token.Token) bool {
	_, ok := Insts[tok.Data]
	return ok
}

func (c *Compiler) isTokVar(tok token.Token) bool {
	_, ok := c.vars[tok.Data]
	return ok
}

func (c *Compiler) isTokLabel(tok token.Token) bool {
	_, ok := c.labels[tok.Data]
	return ok
}

func (c *Compiler) isTokMacro(tok token.Token) bool {
	_, ok := c.macros[tok.Data]
	return ok
}

func (c *Compiler) isTokExprStart(tok token.Token) bool {
	switch tok.Type {
	case token.Dec,  token.Hex,  token.Oct,   token.LParen,
	     token.Bin,  token.Char, token.Float, token.RParen: return true

	case token.Word: return !c.isTokInst(tok)

	default: return false
	}
}

func (c *Compiler) isArithIntrinsic(tok token.Token) bool {
	switch tok.Type {
	case token.Add,    token.Sub,   token.Mult,      token.Div,      token.Mod,
	     token.BitAnd, token.BitOr, token.BitSRight, token.BitSLeft, token.Pow: return true

	default: return false
	}
}

func (c *Compiler) writeInst(op byte, data Word) {
	binary.Write(&c.program, binary.BigEndian, op)
	binary.Write(&c.program, binary.BigEndian, data)
}

func (c *Compiler) CompileToBinary(path string, executable bool, maxE int) []error {
	c.maxE = maxE

	if !c.preproc() {
		return c.errs
	}
	c.programSize = c.pos // Program size (in instructions)

	c.compile()

	entry, ok := c.labels["entry"]
	if !ok {
		c.errs = append(c.errs, fmt.Errorf("Error: Program entry point label 'entry' not found"))
	}
	c.entryPoint = Word(entry)

	if len(c.errs) > 0 {
		if len(c.errs) > c.maxE {
			c.errs = c.errs[:c.maxE]
			c.errs = append(c.errs, fmt.Errorf("..."))
		}

		return c.errs
	}

	f, err := os.Create(path)
	if err != nil {
		return []error{fmt.Errorf("Error: %v", err.Error())}
	}
	defer f.Close()

	if executable {
		// Shebang
		f.Write([]byte("#!/usr/bin/avm\n"))

		os.Chmod(path, 0777)
	}

	// Metadata
	f.Write([]byte{'A', 'V', 'M'})
	f.Write([]byte{VersionMajor, VersionMinor, VersionPatch})

	fileWriteWord(f, c.programSize)
	fileWriteWord(f, c.memorySize)
	fileWriteWord(f, c.entryPoint)

	// Memory
	_, err = f.Write(c.memory.Bytes())
	if err != nil {
		return []error{fmt.Errorf("Error: %v", err.Error())}
	}

	// Program
	_, err = f.Write(c.program.Bytes())
	if err != nil {
		return []error{fmt.Errorf("Error: %v", err.Error())}
	}

	return []error{}
}

func (c *Compiler) compile() {
	c.pos = 0
	c.tok = c.toks[c.pos]

	for c.tok.Type != token.EOF {
		switch c.toks[c.pos].Type {
		case token.Word:
			if err := c.compileInst(); err != nil {
				c.errs = append(c.errs, err)
				c.next()
			}

		case token.Let:
			if err := c.compileLet(); err != nil {
				c.errs = append(c.errs, err)
				c.next()
			}

		case token.Macro:
			if err := c.compileMacro(); err != nil {
				c.errs = append(c.errs, err)
				c.next()
			}

		default:
			c.errs = append(c.errs, c.errorHere("Unexpected %v", c.tok))
			c.next()
		}

		if len(c.errs) >= c.maxE {
			return
		}
	}
}

func intDataToWord(data string, base int) Word {
	word, err := strconv.ParseInt(data, base, 64)
	if err != nil {
		panic(err)
	}

	return Word(word)
}

func charDataToWord(data string) Word {
	if len(data) > 1 {
		panic("len(data) > 1")
	}

	return Word(data[0])
}

func floatDataToWord(data string) Word {
	word, err := strconv.ParseFloat(data, 8)
	if err != nil {
		panic(err)
	}

	return Word(math.Float64bits(word))
}

func (c *Compiler) evalExpr() (data Word, err error) {
	switch c.tok.Type {
	case token.Dec:   data = intDataToWord(c.tok.Data, 10)
	case token.Hex:   data = intDataToWord(c.tok.Data, 16)
	case token.Oct:   data = intDataToWord(c.tok.Data, 8)
	case token.Bin:   data = intDataToWord(c.tok.Data, 2)
	case token.Char:  data = charDataToWord(c.tok.Data)
	case token.Float: data = floatDataToWord(c.tok.Data)

	case token.LParen: data, err = c.evalParen()

	case token.Word:
		if c.isTokInst(c.tok) {
			return 0, c.errorHere("Unexpected instruction '%v'", c.tok.Data)
		}

		switch {
		case c.isTokMacro(c.tok): data = c.macros[c.tok.Data]
		case c.isTokLabel(c.tok): data = c.labels[c.tok.Data]
		case c.isTokVar(c.tok):   data = c.vars[c.tok.Data].Addr

		default: return 0, c.errorHere("Undefined identifier '%v'", c.tok.Data)
		}
	}

	return
}

func (c *Compiler) evalParen() (Word, error) {
	c.next()
	switch c.tok.Type {
	case token.SizeOf: return c.evalSizeOf()

	default:
		if c.isArithIntrinsic(c.tok) {
			return c.evalArith()
		} else {
			return 0, c.errorHere("Expected intrinsic name, got %v", c.tok)
		}
	}
}

var typeToSize = map[token.Type]int{
	token.TypeByte:    1,
	token.TypeChar:    1,
	token.TypeInt16:   2,
	token.TypeInt32:   4,
	token.TypeInt64:   8,
	token.TypeFloat64: 8,
}

func (c *Compiler) evalSizeOf() (size Word, err error) {
	c.next()
	if c.tok.Type == token.Word {
		switch {
		case c.isTokVar(c.tok): size = c.vars[c.tok.Data].Size
		case c.isTokLabel(c.tok):
			return 0, c.errorHere("Expected a variable identifier, got label '%v'", c.tok.Data)
		case c.isTokMacro(c.tok):
			return 0, c.errorHere("Expected a variable identifier, got macro '%v'", c.tok.Data)

		default: return 0, c.errorHere("Undefined identifier '%v'", c.tok.Data)
		}
	} else {
		typeSize, ok := typeToSize[c.tok.Type]
		if !ok {
			return 0, c.errorHere("Expected a type, got %v", c.tok)
		}

		size = Word(typeSize)
	}

	if c.next(); c.tok.Type != token.RParen {
		return 0, c.errorHere("Expected matching ')', got %v", c.tok)
	}

	return
}

func (c *Compiler) evalArith() (res Word, err error) {
	instrinsic := c.tok
	firstArg   := true

	for c.next(); c.tok.Type != token.RParen; c.next() {
		if c.tok.Type == token.EOF {
			return 0, c.errorHere("Expected matching ')', got %v", c.tok)
		}

		data, err := c.evalExpr()
		if err != nil {
			return 0, err
		}

		if firstArg {
			res      = data
			firstArg = false
		} else {
			switch instrinsic.Type {
			case token.Add:  res += data
			case token.Sub:  res -= data
			case token.Mult: res *= data
			case token.Div:  res /= data
			case token.Mod:  res %= data
			case token.Pow:  res  = Word(math.Pow(float64(res), float64(data)))

			case token.BitAnd:    res &=  data
			case token.BitOr:     res |=  data
			case token.BitSRight: res >>= data
			case token.BitSLeft:  res <<= data

			default: panic("Unknown intrinsic.Type")
			}
		}
	}

	return
}

func (c *Compiler) compileInst() error {
	tok      := c.tok
	inst, ok := Insts[tok.Data]
	if !ok {
		return c.errorHere("Unknown instruction '%v'", tok.Data)
	}

	c.next()
	if inst.HasArg {
		if !c.isTokExprStart(c.tok) {
			return c.errorHere("Instruction '%v' expected a parameter, got %v", tok.Data, c.tok)
		}

		data, err := c.evalExpr()
		if err != nil {
			return err
		}
		c.next()

		c.writeInst(inst.Op, data)
	} else {
		if c.isTokExprStart(c.tok) {
			return c.errorHere("Instruction '%v' expects no parameters, got %v", tok.Data, c.tok)
		}

		c.writeInst(inst.Op, 0)
	}

	return nil
}

func (c *Compiler) idExists(tok token.Token) error {
	switch {
	case c.isTokInst(tok):  return c.errorAt(tok.Where, "Expected identifier, got instruction '%v'",
	                                         tok.Data)
	case c.isTokLabel(tok): return c.errorAt(tok.Where, "Identifier '%v' redefined " +
	                                         "(previously a label)", tok.Data)
	case c.isTokMacro(tok): return c.errorAt(tok.Where, "Identifier '%v' redefined " +
	                                         "(previously a macro)", tok.Data)
	case c.isTokVar(tok):   return c.errorAt(tok.Where, "Identifier '%v' redefined " +
	                                         "(previously a variable)", tok.Data)

	default: return nil
	}
}

func (c *Compiler) compileMacro() error {
	c.next()
	id := c.tok
	if err := c.idExists(id); err != nil {
		return err
	}

	if c.next(); c.tok.Type != token.Equals {
		return c.errorHere("Expected macro assignment with '=', got %v", c.tok)
	}
	c.next()

	if !c.isTokExprStart(c.tok) {
		return c.errorHere("Expected macro value expression, got %v", c.tok)
	}

	data, err := c.evalExpr()
	if err != nil {
		return err
	}
	c.next()

	c.macros[id.Data] = data

	return nil
}

func (c *Compiler) writeMemory(data Word, size int) error {
	switch size {
	case 1: binary.Write(&c.memory, binary.BigEndian, uint8(data))
	case 2: binary.Write(&c.memory, binary.BigEndian, uint16(data))
	case 4: binary.Write(&c.memory, binary.BigEndian, uint32(data))
	case 8: binary.Write(&c.memory, binary.BigEndian, data)

	default: return fmt.Errorf("Got incorrect data element size %v", size)
	}

	c.memorySize += Word(size)

	return nil
}

func (c *Compiler) compileLet() error {
	c.next()
	id := c.tok
	if err := c.idExists(id); err != nil {
		return err
	}

	c.next()
	size, ok := typeToSize[c.tok.Type]
	if !ok {
		return c.errorHere("Expected a type, got %v", c.tok)
	}

	if c.next(); c.tok.Type != token.Equals {
		return c.errorHere("Expected variable assignment with '=', got %v", c.tok)
	}
	c.next()


	addr      := c.memorySize + 1
	startSize := c.memorySize

	for {
		if c.tok.Type == token.String {
			c.writeString(c.tok.Data, size)

			if c.next(); c.tok.Type == token.Dots {
				return c.errorHere("Unexpected '%v' after string (cannot fill with a string)",
				                   token.Dots)
			}
		} else {
			data, err := c.evalExpr()
			if err != nil {
				return err
			}

			if c.next(); c.tok.Type == token.Dots {
				c.next()

				count, err := c.evalExpr()
				if err != nil {
					return err
				}

				for i := Word(0); i < count; i ++ {
					c.writeMemory(data, size)
				}

				c.next()
			} else {
				c.writeMemory(data, size)
			}
		}

		if c.tok.Type == token.Comma {
			c.next()
		} else {
			break
		}
	}

	c.vars[id.Data] = Var{Addr: addr, Size: c.memorySize - startSize}

	return nil
}

func (c *Compiler) writeString(data string, charSize int) error {
	for _, ch := range data {
		if err := c.writeMemory(Word(ch), charSize); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) next() {
	if c.tok.Type == token.EOF {
		return
	}

	c.pos ++
	c.tok = c.toks[c.pos]
}

func (c *Compiler) preproc() bool {
	for c.tok = c.l.NextToken(); c.tok.Type != token.EOF; c.tok = c.l.NextToken() {
		// Eat and evaluate the preprocessor, leave out the other tokens
		switch c.tok.Type {
		case token.Error:
			c.errs = append(c.errs, c.errorAt(c.tok.Where, c.tok.Data))
			return false

		case token.Word:
			if c.isTokInst(c.tok) {
				c.pos ++
			}

		case token.Label:
			id := c.tok
			if err := c.idExists(id); err != nil {
				c.errs = append(c.errs, err)
			}

			c.labels[id.Data] = c.pos

			continue
		}

		if len(c.errs) > c.maxE {
			return true
		}

		c.toks = append(c.toks, c.tok)
	}

	// Add the EOF token
	c.toks = append(c.toks, c.tok)

	return true
}
