package main

import (
	"os"
	"fmt"
	"flag"

	"github.com/LordOfTrident/anasm/internal/compiler"
)

// 1.0.0: First release, supports avm version 1.0.0
// 1.1.0: Added instruction argument safety
// 1.2.0: Added an option to create an executable output file

var out = flag.String("o",        "a.out", "Path of the output binary")
var v   = flag.Bool("version",    false,   "Show the version")
var e   = flag.Bool("executable", true,    "Make the output file executable")

const (
	appName = "anasm"

	versionMajor = 1
	versionMinor = 2
	versionPatch = 0
)

func usage() {
	fmt.Printf("Usage: %v [OPTIONS] [FILE]\n", os.Args[0])
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
}

func main() {
	if *v {
		version()

		return
	}

	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No input file\nTry '%v -h'\n", os.Args[0])

		os.Exit(1)
	}

	path := flag.Args()[0]

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not open file '%v'\n", path)

		os.Exit(1)
	}

	c := compiler.New(string(data), path)

	if err := c.CompileToBinary(*out, *e); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err.Error())
	}
}
