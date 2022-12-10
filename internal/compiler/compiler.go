package compiler

import (
	"fmt"
	"os"
	"strconv"
	"bytes"
	"math"
	"encoding/binary"

	"github.com/LordOfTrident/anasm/internal/token"
	"github.com/LordOfTrident/anasm/internal/lexer"
)

type Word uint64

const (
	VersionMajor = 0
	VersionMinor = 4
	VersionPatch // Not keeping track of the patch
)

type argType int
const (
	argNone = iota
	argNum
	argReg
)

func (t argType) String() string {
	switch t {
	case argNone: return "none"
	case argNum:  return "number"
	case argReg:  return "register"
	}

	panic("Unknown argType")
}

func tokTypeToArgType(tokType token.Type) argType {
	switch tokType {
	case token.Hex,   token.Dec, token.Oct,
	     token.Float, token.LabelRef: return argNum
	case token.Reg:                   return argReg

	default: return argNone
	}
}

func isTokTypeOfArgType(tokType token.Type, argType argType) bool {
	switch argType {
	case argNum: return tokType == token.Hex   || tokType == token.Dec || tokType == token.Oct ||
	                    tokType == token.Float || tokType == token.LabelRef
	case argReg: return tokType == token.Reg
	}

	return false
}

type Inst struct {
	Op   byte
	Args []argType

	FirstArgIsData bool
}

var (
	insts = map[string]Inst{
		"nop": Inst{Op: 0x00},

		"mov": Inst{Op: 0x10, Args: []argType{argReg, argNum}},
		"mor": Inst{Op: 0x11, Args: []argType{argReg, argReg}},

		"psh": Inst{Op: 0x12, Args: []argType{argNum}, FirstArgIsData: true},
		"psr": Inst{Op: 0x13, Args: []argType{argReg}},
		"pop": Inst{Op: 0x14},
		"por": Inst{Op: 0x15, Args: []argType{argReg}},

		"add": Inst{Op: 0x20},
		"sub": Inst{Op: 0x21},

		"mul": Inst{Op: 0x22},
		"div": Inst{Op: 0x23},
		"mod": Inst{Op: 0x24},

		"inc": Inst{Op: 0x25},
		"dec": Inst{Op: 0x26},

		"fad": Inst{Op: 0x27},
		"fsb": Inst{Op: 0x28},

		"fmu": Inst{Op: 0x29},
		"fdi": Inst{Op: 0x2a},

		"fin": Inst{Op: 0x2b},
		"fde": Inst{Op: 0x2c},

		"jmp": Inst{Op: 0x30, Args: []argType{argNum}, FirstArgIsData: true},
		"jnz": Inst{Op: 0x31, Args: []argType{argNum}, FirstArgIsData: true},

		"cal": Inst{Op: 0x38, Args: []argType{argNum}, FirstArgIsData: true},
		"ret": Inst{Op: 0x39},

		"equ": Inst{Op: 0x32},
		"neq": Inst{Op: 0x33},
		"grt": Inst{Op: 0x34},
		"geq": Inst{Op: 0x35},
		"les": Inst{Op: 0x36},
		"leq": Inst{Op: 0x37},

		"ueq": Inst{Op: 0x3a},
		"une": Inst{Op: 0x3b},
		"ugr": Inst{Op: 0x3c},
		"ugq": Inst{Op: 0x3d},
		"ule": Inst{Op: 0x3e},
		"ulq": Inst{Op: 0x3f},

		"feq": Inst{Op: 0x40},
		"fne": Inst{Op: 0x41},
		"fgr": Inst{Op: 0x42},
		"fgq": Inst{Op: 0x43},
		"fle": Inst{Op: 0x44},
		"flq": Inst{Op: 0x45},

		"dup": Inst{Op: 0x50},
		"swp": Inst{Op: 0x51},
		"emp": Inst{Op: 0x52},

		"dmp": Inst{Op: 0xF0},
		"prt": Inst{Op: 0xF1},
		"fpr": Inst{Op: 0xF2},

		"hlt": Inst{Op: 0xFF},
	}

	regs = map[string]byte{
		"r1":  0x00,
		"r2":  0x01,
		"r3":  0x02,
		"r4":  0x03,
		"r5":  0x04,
		"r6":  0x05,
		"r7":  0x06,
		"r8":  0x07,
		"r9":  0x08,
		"r10": 0x09,
		"r11": 0x0a,
		"r12": 0x0b,
		"r13": 0x0c,
		"r14": 0x0d,
		"r15": 0x0e,
		"r16": 0x0f,

		"ip": 0x10,
		"sp": 0x11,
		"sb": 0x12,
		"ex": 0x13,
	}
)

type Compiler struct {
	toks []token.Token
	pos  int
	tok  token.Token

	size, entry Word

	labels map[string]int

	out bytes.Buffer

	l *lexer.Lexer
}

func New(input, path string) *Compiler {
	return &Compiler{l: lexer.New(input, path), labels: make(map[string]int)}
}

func (c *Compiler) Error(format string, args... interface{}) error {
	return fmt.Errorf("At %v: %v", c.tok.Where, fmt.Sprintf(format, args...))
}

func (c *Compiler) ErrorFrom(where token.Where, format string, args... interface{}) error {
	return fmt.Errorf("At %v: %v", where, fmt.Sprintf(format, args...))
}

func fileWriteWord(f *os.File, word Word) error {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, word)

	_, err := f.Write(buf.Bytes())

	return err
}

func (c *Compiler) CompileToBinary(path string, executable bool) error {
	if err := c.preproc(); err != nil {
		return err
	}

	if err := c.compile(); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
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

	fileWriteWord(f, c.size)
	fileWriteWord(f, c.entry)

	// Program
	_, err = f.Write(c.out.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *Compiler) compile() error {
	c.pos = 0
	c.tok = c.toks[c.pos]

	for c.tok.Type != token.EOF {
		switch c.toks[c.pos].Type {
		case token.Word:
			if err := c.compileInst(); err != nil {
				return err
			}

		default: return c.Error("Unexpected %v", c.tok)
		}
	}

	return nil
}

func (c *Compiler) writeInst(op byte, reg byte, data Word) {
	binary.Write(&c.out, binary.BigEndian, op)
	binary.Write(&c.out, binary.BigEndian, reg)
	binary.Write(&c.out, binary.BigEndian, data)
}

func (c *Compiler) compileInst() error {
	tok := c.tok

	inst, ok := insts[tok.Data]
	if !ok {
		return c.Error("'%v' is not an instruction", tok.Data)
	}

	c.next()
	args, err := c.getInstArgs()
	if err != nil {
		return err
	}

	if len(args) != len(inst.Args) {
		return c.ErrorFrom(tok.Where, "Instruction '%v' expected %v argument(s), got %v",
		                   tok.Data, len(inst.Args), len(args))
	}

	for i := 0; i < len(args); i ++ {
		if !isTokTypeOfArgType(args[i].Type, inst.Args[i]) {
			return c.ErrorFrom(args[i].Where, "Argument expected to be '%v', got '%v'",
			                   inst.Args[i], tokTypeToArgType(args[i].Type))
		}
	}

	var reg, data Word

	if len(args) > 0 {
		reg, err = c.argToWord(args[0])
		if err != nil {
			return err
		}
	}

	if len(args) > 1 {
		data, err = c.argToWord(args[0])
		if err != nil {
			return err
		}
	}

	if inst.FirstArgIsData {
		data = reg
		reg  = 0
	}

	c.writeInst(inst.Op, byte(reg), data)

	return nil
}

func (c *Compiler) getInstArgs() ([]token.Token, error) {
	args := []token.Token{}

	if c.tok.Type != token.Colon {
		return args, nil
	}
	c.next()

	for {
		if !c.tok.IsArg() {
			return args, c.Error("Expected instruction argument, got %v", c.tok)
		}

		args = append(args, c.tok)

		if c.next(); c.tok.Type != token.Comma {
			break
		}

		c.next()
	}

	return args, nil
}

func (c *Compiler) argToWord(tok token.Token) (Word, error) {
	switch tok.Type {
	case token.Dec:
		i64, err := strconv.ParseInt(tok.Data, 10, 64)
		if err != nil {
			panic(err) // This should never happen
		}

		return Word(i64), nil

	case token.Hex:
		i64, err := strconv.ParseInt(tok.Data, 16, 64)
		if err != nil {
			panic(err) // This should never happen
		}

		return Word(i64), nil

	case token.Oct:
		i64, err := strconv.ParseInt(tok.Data, 8, 64)
		if err != nil {
			panic(err) // This should never happen
		}

		return Word(i64), nil

	case token.Float:
		i64, err := strconv.ParseFloat(tok.Data, 8)
		if err != nil {
			panic(err) // This should never happen
		}

		return Word(math.Float64bits(i64)), nil

	case token.LabelRef:
		i64, ok := c.labels[tok.Data]
		if !ok {
			return 0, c.Error("Label '%v' was not declared", tok.Data)
		}

		return Word(i64), nil


	case token.Reg:
		i64, ok := regs[tok.Data]
		if !ok {
			return 0, c.Error("'%v' is not a valid register", tok.Data)
		}

		return Word(i64), nil

	default: return 0, c.Error("Expected register argument, instead got %v", tok)
	}
}

func (c *Compiler) next() {
	if c.tok.Type == token.EOF {
		return
	}

	c.pos ++
	c.tok = c.toks[c.pos]
}

func (c *Compiler) preproc() error {
	tok := c.l.NextToken()
	for ; tok.Type != token.EOF; tok = c.l.NextToken() {
		if tok.Type == token.Error {
			return fmt.Errorf("At %v: %v", tok.Where, tok.Data)
		}

		// Eat and evaluate the preprocessor, leave out the other tokens
		switch tok.Type {
		case token.Word:
			if _, ok := insts[tok.Data]; ok {
				c.pos ++
			}

		case token.Label:
			c.labels[tok.Data] = c.pos

			continue
		}

		c.toks = append(c.toks, tok)
	}

	entry, ok := c.labels["entry"]
	if !ok {
		return fmt.Errorf("Program entry point label 'entry' not found")
	}

	// Add the EOF token
	c.toks = append(c.toks, tok)

	// Program size (in instructions) and program entry point
	c.size  = Word(c.pos)
	c.entry = Word(entry)

	return nil
}
