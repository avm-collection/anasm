package main

import (
	"os"
	"fmt"
	"flag"
	"path/filepath"
	"strings"

	"github.com/avm-collection/anasm/internal/config"
	"github.com/avm-collection/anasm/internal/token"
	"github.com/avm-collection/anasm/internal/compiler"
	"github.com/avm-collection/anasm/internal/disasm"
)

var (
	out  = flag.String("o",        "",      "Path of the output binary")
	v    = flag.Bool("version",    false,   "Show the version")
	e    = flag.Bool("executable", true,    "Make the output file executable")
	d    = flag.Bool("disasm",     false,   "Run the disassembler")
	noW  = flag.Bool("noW",        false,   "Dont show warnings")
	maxE = flag.Int("maxE",        8,       "Max compiler errors count")

	args []string
)

func printError(format string, args... interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Sprintf(format, args...))
}

func printTry(arg string) {
	fmt.Fprintf(os.Stderr, "Try '%v %v'\n", os.Args[0], arg)
}

func usage() {
	fmt.Printf("Github: %v\n", config.GithubLink)
	fmt.Printf("Usage: %v [FILE] [OPTIONS]\n", os.Args[0])
	fmt.Println("Options:")

	flag.PrintDefaults()
}

func version() {
	fmt.Printf("%v %v.%v.%v\n", config.AppName,
	           config.VersionMajor, config.VersionMinor, config.VersionPatch)
}

func init() {
	token.AllTokensCoveredTest()

	flag.Usage = usage

	// Aliases
	flag.BoolVar(v, "v", *v, "Alias for -version")
	flag.BoolVar(e, "e", *e, "Alias for -executable")
	flag.BoolVar(d, "d", *d, "Alias for -disasm")

	flag.Parse()

	args = flag.Args()
	for i := 0; i < len(flag.Args()); i ++ {
		if len(flag.Args()[i]) == 0 {
			continue
		}

		if flag.Args()[i][0] != '-' {
			continue
		}

		args = flag.Args()[:i]
		flag.CommandLine.Parse(flag.Args()[i:])

		break
	}
}

func assemble(input, path string) {
	if len(*out) == 0 {
		if len(filepath.Ext(path)) == 0 {
			*out = path + ".out"
		} else {
			*out = strings.TrimSuffix(path, filepath.Ext(path))
		}

		*out = filepath.Base(*out)
	}

	c := compiler.New(input, path)

	if errs := c.CompileToBinary(*out, *e, *maxE); len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}
}

func disassemble(input []byte, path string) {
	if len(*out) == 0 {
		if filepath.Ext(path) == ".anasm" {
			*out = path + ".out"
		} else {
			*out = path + ".anasm"
		}

		*out = filepath.Base(*out)
	}

	d := disasm.New(input, path)

	if err := d.Disassemble(*out, *noW); err != nil {
		printError(err.Error())
	}
}

func main() {
	if *v {
		version()

		return
	}

	if len(args) == 0 {
		printError("No input file")
		printTry("-h")

		os.Exit(1)
	} else if len(args) > 1 {
		printError("Unexpected argument '%v'", args[1])
		printTry("-h")

		os.Exit(1)
	}

	path      := args[0]
	data, err := os.ReadFile(path)
	if err != nil {
		printError("Could not open file '%v'", path)
		printTry("-h")

		os.Exit(1)
	}

	if *d {
		disassemble(data, path)
	} else {
		assemble(string(data), path)
	}
}
