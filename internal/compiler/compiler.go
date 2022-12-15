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

func (c *Compiler) writeInst(op byte, data Word) {
	binary.Write(&c.out, binary.BigEndian, op)
	binary.Write(&c.out, binary.BigEndian, data)
}

func (c *Compiler) compileInst() error {
	tok := c.tok

	inst, ok := Insts[tok.Data]
	if !ok {
		return c.Error("'%v' is not an instruction", tok.Data)
	}

	c.next()
	if !c.tok.IsArg() {
		if inst.HasArg {
			return c.ErrorFrom(tok.Where, "Instruction '%v' expects an argument", tok.Data)
		}

		c.writeInst(inst.Op, 0)

		return nil
	} else if !inst.HasArg {
		return c.ErrorFrom(tok.Where, "Instruction '%v' expects no arguments", tok.Data)
	}

	if !c.tok.IsArg() {
		return c.ErrorFrom(c.tok.Where, "Expected argument, got '%v'", c.tok.Type)
	}

	data, err := c.argToWord(c.tok)
	if err != nil {
		return err
	}
	c.next()

	c.writeInst(inst.Op, data)

	return nil
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
			if _, ok := Insts[tok.Data]; ok {
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
