package main

import (
	"os"
	"fmt"
	"flag"

	"github.com/LordOfTrident/anasm/internal/compiler"
)

// 0.1.0: Can compile to avm version 0.2.0
// 0.2.0: Added instruction argument safety
// 0.3.0: Added an option to create an executable output file
// 0.4.0: Support avm 0.3.0
// 0.4.1: Parameter improvements, flags can now come after parameters

var (
	out = flag.String("o",        "a.out", "Path of the output binary")
	v   = flag.Bool("version",    false,   "Show the version")
	e   = flag.Bool("executable", true,    "Make the output file executable")

	args []string
)

const (
	appName = "anasm"

	versionMajor = 0
	versionMinor = 4
	versionPatch = 1
)

func printError(format string, args... interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", fmt.Sprintf(format, args...))
}

func printTry(arg string) {
	fmt.Fprintf(os.Stderr, "Try '%v %v'\n", os.Args[0], arg)
}

func usage() {
	fmt.Printf("Usage: %v [FILE] [OPTIONS]\n", os.Args[0])
	fmt.Println("Options:")

	flag.PrintDefaults()
}

func version() {
	fmt.Printf("%v %v.%v.%v\n", appName, versionMajor, versionMinor, versionPatch)
}

func init() {
	flag.Usage = usage

	// Aliases
	flag.BoolVar(v, "v", *v, "Alias for -version")
	flag.BoolVar(e, "e", *e, "Alias for -executable")

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

	path := args[0]

	data, err := os.ReadFile(path)
	if err != nil {
		printError("Could not open file '%v'", path)
		printTry("-h")

		os.Exit(1)
	}

	c := compiler.New(string(data), path)

	if err := c.CompileToBinary(*out, *e); err != nil {
		printError(err.Error())
	}
}
