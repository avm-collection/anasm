package agen

import (
	"os"
	"bytes"
	"encoding/binary"
)

const (
	VersionMajor = 1
	VersionMinor = 13
	VersionPatch // Not keeping track of the patch
)

type InstInfo struct {
	Op     byte
	HasArg bool
}

func InstByOpcode(op byte) (InstInfo, bool) {
	for _, inst := range Insts {
		if inst.Op == op {
			return inst, true
		}
	}

	return InstInfo{}, false
}

var (
	Insts = map[string]InstInfo{
		"nop": InstInfo{Op: 0x00},

		"psh": InstInfo{Op: 0x10, HasArg: true},
		"pop": InstInfo{Op: 0x11},

		"add": InstInfo{Op: 0x20},
		"sub": InstInfo{Op: 0x21},

		"mul": InstInfo{Op: 0x22},
		"div": InstInfo{Op: 0x23},
		"mod": InstInfo{Op: 0x24},

		"inc": InstInfo{Op: 0x25},
		"dec": InstInfo{Op: 0x26},

		"fad": InstInfo{Op: 0x27},
		"fsb": InstInfo{Op: 0x28},

		"fmu": InstInfo{Op: 0x29},
		"fdi": InstInfo{Op: 0x2a},

		"fin": InstInfo{Op: 0x2b},
		"fde": InstInfo{Op: 0x2c},

		"neg": InstInfo{Op: 0x2d},
		"not": InstInfo{Op: 0x2e},

		"jmp": InstInfo{Op: 0x30, HasArg: true},
		"jnz": InstInfo{Op: 0x31, HasArg: true},

		"cal": InstInfo{Op: 0x38, HasArg: true},
		"ret": InstInfo{Op: 0x39},

		"and": InstInfo{Op: 0x46},
		"orr": InstInfo{Op: 0x47},

		"equ": InstInfo{Op: 0x32},
		"neq": InstInfo{Op: 0x33},
		"grt": InstInfo{Op: 0x34},
		"geq": InstInfo{Op: 0x35},
		"les": InstInfo{Op: 0x36},
		"leq": InstInfo{Op: 0x37},

		"ueq": InstInfo{Op: 0x3a},
		"une": InstInfo{Op: 0x3b},
		"ugr": InstInfo{Op: 0x3c},
		"ugq": InstInfo{Op: 0x3d},
		"ule": InstInfo{Op: 0x3e},
		"ulq": InstInfo{Op: 0x3f},

		"feq": InstInfo{Op: 0x40},
		"fne": InstInfo{Op: 0x41},
		"fgr": InstInfo{Op: 0x42},
		"fgq": InstInfo{Op: 0x43},
		"fle": InstInfo{Op: 0x44},
		"flq": InstInfo{Op: 0x45},

		"dup": InstInfo{Op: 0x50, HasArg: true},
		"swp": InstInfo{Op: 0x51, HasArg: true},
		"emp": InstInfo{Op: 0x52},
		"set": InstInfo{Op: 0x53},
		"cpy": InstInfo{Op: 0x54},

		"r08": InstInfo{Op: 0x60},
		"r16": InstInfo{Op: 0x61},
		"r32": InstInfo{Op: 0x62},
		"r64": InstInfo{Op: 0x63},

		"w08": InstInfo{Op: 0x64},
		"w16": InstInfo{Op: 0x65},
		"w32": InstInfo{Op: 0x66},
		"w64": InstInfo{Op: 0x67},

		"ope": InstInfo{Op: 0x70},
		"clo": InstInfo{Op: 0x71},
		"wrf": InstInfo{Op: 0x72},
		"rdf": InstInfo{Op: 0x73},
		"szf": InstInfo{Op: 0x74},
		"flu": InstInfo{Op: 0x75},

		"ban": InstInfo{Op: 0x80},
		"bor": InstInfo{Op: 0x81},
		"bsr": InstInfo{Op: 0x82},
		"bsl": InstInfo{Op: 0x83},

		"lol": InstInfo{Op: 0x90},
		"cll": InstInfo{Op: 0x91},
		"llf": InstInfo{Op: 0x92},
		"ulf": InstInfo{Op: 0x93},
		"clf": InstInfo{Op: 0x94},

		"dmp": InstInfo{Op: 0xF0},
		"prt": InstInfo{Op: 0xF1},
		"fpr": InstInfo{Op: 0xF2},

		"hlt": InstInfo{Op: 0xFF},
	}
)

type Word uint64

func (w Word) WriteIntoFile(f *os.File) error {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, w)

	_, err := f.Write(buf.Bytes())
	return err
}

type Inst struct {
	Op   byte
	Data Word
}
const InstSize = 9

func (i Inst) WriteIntoFile(f *os.File) error {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, i.Data)

	bytes := []byte{i.Op}
	bytes  = append(bytes, buf.Bytes()...)

	_, err := f.Write(bytes);
	return err
}

type Type int
const (
	I8 = Type(iota)
	I16
	I32
	I64
)

type AGEN struct {
	memory     bytes.Buffer
	program    []Inst
	entryPoint Word
}

func New() *AGEN {
	a := &AGEN{}
	a.memory.Write([]byte{0})

	return a
}

func (a *AGEN) MemorySize() Word {
	return Word(len(a.memory.Bytes()))
}

func (a *AGEN) ProgramSize() Word {
	return Word(len(a.program))
}

func (a *AGEN) EntryPoint() Word {
	return a.entryPoint
}

func (a *AGEN) SetEntryHere() {
	a.entryPoint = a.ProgramSize()
}

func (a *AGEN) SetEntry(addr Word) {
	a.entryPoint = addr
}

func (a *AGEN) CreateExecAVM(path string, executable bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if executable {
		f.Write([]byte("#!/usr/bin/avm\n"))
		os.Chmod(path, 0777)
	}

	// Metadata
	f.Write([]byte("AVM"))
	f.Write([]byte{VersionMajor, VersionMinor, VersionPatch})

	a.ProgramSize().WriteIntoFile(f)
	a.MemorySize().WriteIntoFile(f)
	a.EntryPoint().WriteIntoFile(f)

	_, err = f.Write(a.memory.Bytes())
	if err != nil {
		return err
	}

	for _, inst := range a.program {
		if err := inst.WriteIntoFile(f); err != nil {
			return err
		}
	}

	return nil
}

func (a *AGEN) Label() Word {
	return Word(len(a.program))
}

func (a *AGEN) AddInstWith(str string, data Word) *Inst {
	a.program = append(a.program, Inst{Op: Insts[str].Op, Data: data})
	return &a.program[len(a.program) - 1]
}

func (a *AGEN) AddInst(str string) *Inst {
	return a.AddInstWith(str, 0)
}

func (a *AGEN) AddMemoryInt(list []Word, type_ Type) Word {
	addr := Word(a.MemorySize())
	for _, data := range list {
		switch type_ {
		case I8:  binary.Write(&a.memory, binary.BigEndian, uint8(data))
		case I16: binary.Write(&a.memory, binary.BigEndian, uint16(data))
		case I32: binary.Write(&a.memory, binary.BigEndian, uint32(data))
		case I64: binary.Write(&a.memory, binary.BigEndian, uint64(data))
		}
	}
	return addr
}

func (a *AGEN) AddMemoryString(str string) Word {
	addr := Word(a.MemorySize())
	a.memory.Write([]byte(str))
	return addr
}
