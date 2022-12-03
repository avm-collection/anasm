package compiler

import (
	"fmt"
	"os"
	"strconv"
	"bytes"
	"encoding/binary"

	"github.com/LordOfTrident/anasm/internal/token"
	"github.com/LordOfTrident/anasm/internal/lexer"
)

// TODO: Implement instruction formats for safety
//       Right now, all instructions can take any arguments and there is no check if the type is
//       correct

type Word uint64

const (
	VersionMajor = 1
	VersionMinor = 0
	VersionPatch = 0
)

const (
	opNop = 0x00

	opMov = 0x10
	opMor = 0x11

	opPsh = 0x12
	opPsr = 0x13
	opPop = 0x14
	opPor = 0x15

	opAdd = 0x20
	opSub = 0x21

	opMul = 0x22
	opDiv = 0x23
	opMod = 0x24

	opInc = 0x25
	opDec = 0x26

	opJmp = 0x30
	opJnz = 0x31

	opEqu = 0x32
	opNeq = 0x33
	opGrt = 0x34
	opGeq = 0x35
	opLes = 0x36
	opLeq = 0x37

	opDup = 0x40
	opSwp = 0x41

	opDum = 0xF0
	opHlt = 0xFF
)

const (
	reg1 = iota
	reg2
	reg3
	reg4
	reg5
	reg6
	reg7
	reg8
	reg9
	reg10
	reg11
	reg12
	reg13
	reg14
	reg15
	reg16

	regIp // Instruction pointer
	regSp // Stack pointer
	regSb // Stack base pointer
	regEx // Exitcode
)

var (
	ops = map[string]byte{
		"nop": opNop,

		"mov": opMov,
		"mor": opMor,

		"psh": opPsh,
		"psr": opPsr,
		"pop": opPop,
		"por": opPor,

		"add": opAdd,
		"sub": opSub,

		"mul": opMul,
		"div": opDiv,
		"mod": opMod,

		"inc": opInc,
		"dec": opDec,

		"jmp": opJmp,
		"jnz": opJnz,

		"equ": opEqu,
		"neq": opNeq,
		"grt": opGrt,
		"geq": opGeq,
		"les": opLes,
		"leq": opLeq,

		"dup": opDup,
		"swp": opSwp,

		"dum": opDum,
		"hlt": opHlt,
	}

	regs = map[string]byte{
		"r1":  reg1,
		"r2":  reg2,
		"r3":  reg3,
		"r4":  reg4,
		"r5":  reg5,
		"r6":  reg6,
		"r7":  reg7,
		"r8":  reg8,
		"r9":  reg9,
		"r10": reg10,
		"r11": reg11,
		"r12": reg12,
		"r13": reg13,
		"r14": reg14,
		"r15": reg15,
		"r16": reg16,

		"ip": regIp,
		"sp": regSp,
		"sb": regSb,
		"ex": regEx,
	}
)

type Compiler struct {
	toks []token.Token
	tok  token.Token
	pos  int

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

func fileWriteWord(f *os.File, word Word) error {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, word)

	_, err := f.Write(buf.Bytes())

	return err
}

func (c *Compiler) CompileToBinary(path string) error {
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
	op, ok := ops[c.tok.Data]
	if !ok {
		return c.Error("'%v' is not an instruction", c.tok.Data)
	}

	if c.next(); c.tok.Type != token.Colon {
		c.writeInst(op, 0, 0)

		return nil
	}
	c.next()

	// Register
	word, err := c.argToWord()
	if err != nil {
		return err
	}
	reg := byte(word)

	if c.next(); c.tok.Type != token.Comma {
		c.writeInst(op, reg, 0)

		return nil
	}
	c.next()

	// Data
	word, err = c.argToWord()
	if err != nil {
		return err
	}
	data := word

	if c.next(); c.tok.Type == token.Comma {
		return c.Error("Instructions can have at max 2 arguments")
	}

	c.writeInst(op, reg, data)

	return nil
}

func (c *Compiler) argToWord() (Word, error) {
	switch c.tok.Type {
	case token.Dec:
		i64, err := strconv.ParseInt(c.tok.Data, 10, 64)
		if err != nil {
			panic(err) // This should never happen
		}

		return Word(i64), nil

	case token.Hex:
		i64, err := strconv.ParseInt(c.tok.Data, 16, 64)
		if err != nil {
			panic(err) // This should never happen
		}

		return Word(i64), nil

	case token.Reg:
		i64, ok := regs[c.tok.Data]
		if !ok {
			return 0, c.Error("'%v' is not a valid register", c.tok.Data)
		}

		return Word(i64), nil

	case token.LabelRef:
		i64, ok := c.labels[c.tok.Data]
		if !ok {
			return 0, c.Error("Label '%v' was not declared", c.tok.Data)
		}

		return Word(i64), nil

	default: return 0, c.Error("Expected argument, instead got %v", c.tok)
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
			if _, ok := ops[tok.Data]; ok {
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
