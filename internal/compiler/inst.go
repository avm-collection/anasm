package compiler

type Inst struct {
	Op     byte
	HasArg bool
}

var (
	Insts = map[string]Inst{
		"nop": Inst{Op: 0x00},

		"psh": Inst{Op: 0x10, HasArg: true},
		"pop": Inst{Op: 0x11},

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

		"neg": Inst{Op: 0x2d},
		"not": Inst{Op: 0x2e},

		"jmp": Inst{Op: 0x30, HasArg: true},
		"jnz": Inst{Op: 0x31, HasArg: true},

		"cal": Inst{Op: 0x38, HasArg: true},
		"ret": Inst{Op: 0x39},

		"and": Inst{Op: 0x46},
		"orr": Inst{Op: 0x47},

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

		"dup": Inst{Op: 0x50, HasArg: true},
		"swp": Inst{Op: 0x51, HasArg: true},
		"emp": Inst{Op: 0x52},
		"set": Inst{Op: 0x53},
		"cpy": Inst{Op: 0x54},

		"r08": Inst{Op: 0x60},
		"r16": Inst{Op: 0x61},
		"r32": Inst{Op: 0x62},
		"r64": Inst{Op: 0x63},

		"w08": Inst{Op: 0x64},
		"w16": Inst{Op: 0x65},
		"w32": Inst{Op: 0x66},
		"w64": Inst{Op: 0x67},

		"ope": Inst{Op: 0x70},
		"clo": Inst{Op: 0x71},
		"wrf": Inst{Op: 0x72},
		"rdf": Inst{Op: 0x73},
		"szf": Inst{Op: 0x74},

		"ban": Inst{Op: 0x80},
		"bor": Inst{Op: 0x81},
		"bsr": Inst{Op: 0x82},
		"bsl": Inst{Op: 0x83},

		"dmp": Inst{Op: 0xF0},
		"prt": Inst{Op: 0xF1},
		"fpr": Inst{Op: 0xF2},

		"hlt": Inst{Op: 0xFF},
	}
)
