package errors

import (
	"fmt"
	"os"
	"strings"
)

type Where interface {
	AtRow()   int
	AtCol()   int
	GetLen()  int
	InFile()  string
	GetLine() string
	String()  string
}

var (
	Max   = 8
	count = 0

	first = true
)

type typeData struct {
	name, attr string
}

func typeError() typeData {
	return typeData{name: "Error", attr: "\x1b[1;91m"}
}

func typeWarning() typeData {
	return typeData{name: "Warning", attr: "\x1b[1;93m"}
}

func typeNote() typeData {
	return typeData{name: "Note", attr: "\x1b[1;96m"}
}

type data struct {
	type_ typeData

	where Where
	msg   string

	lineNum, part1, main, part2 string
}

func newData(where Where, msg string, type_ typeData) data {
	d := data{
		type_:   type_,
		where:   where,
		msg:     msg,
		lineNum: fmt.Sprintf("%v", where.AtRow()),
	}

	idx := where.AtCol() - 1

	d.part1 = strings.Replace(where.GetLine()[:idx],                     "\t", "    ", -1)
	d.main  = strings.Replace(where.GetLine()[idx:idx + where.GetLen()], "\t", "    ", -1)
	d.part2 = strings.Replace(where.GetLine()[idx + where.GetLen():],    "\t", "    ", -1)

	return d
}

func printNiceHead(where Where, msg string, type_ typeData) {
	fmt.Fprintf(os.Stderr, "%v:\x1b[0;1m %v\x1b[0m: %v\n", type_.attr + type_.name, where, msg)
}

func printNiceLine(lineNum, line string) {
	fmt.Fprintf(os.Stderr, "    %v | %v\n", lineNum, line)
}

func printNice(d data) {
	separator()

	printNiceHead(d.where, d.msg, d.type_)
	printNiceLine(d.lineNum, fmt.Sprintf("%v%v\x1b[0m%v", d.part1, d.type_.attr + d.main, d.part2))
}

func NoOutput() bool {
	return first
}

func separator() {
	if !first {
		fmt.Println()
	} else {
		first = false
	}
}

func newError() {
	count ++

	if count > Max {
		fmt.Fprintf(os.Stderr, "...\nCompilation aborted\n")
		os.Exit(1)
	}
}

func Error(where Where, format string, any... interface{}) {
	newError()
	printNice(newData(where, fmt.Sprintf(format, any...), typeError()))
}

func simple(type_ typeData, msg string) {
	separator()
	fmt.Fprintf(os.Stderr, "%v%v:\x1b[0m %v\n", type_.attr, type_.name, msg)
}

func SimpleError(format string, any... interface{}) {
	newError()
	simple(typeError(), fmt.Sprintf(format, any...))
}

func SimpleWarning(format string, any... interface{}) {
	newError()
	simple(typeWarning(), fmt.Sprintf(format, any...))
}

func SimpleNote(format string, any... interface{}) {
	newError()
	simple(typeNote(), fmt.Sprintf(format, any...))
}

func Warning(where Where, format string, any... interface{}) {
	printNice(newData(where, fmt.Sprintf(format, any...), typeWarning()))
}

func Note(where Where, format string, any... interface{}) {
	printNice(newData(where, fmt.Sprintf(format, any...), typeNote()))
}

func Happened() bool {
	return count > 0
}

func Reset() {
	first = true
	count = 0
}

// Specific functions
func NoteSuggestName(where Where, name string) {
	d := newData(where, fmt.Sprintf("Did you mean '%v'?", name), typeNote())
	d.main = name

	printNice(d)
}

func NoteSuggestNewCode(where Where, msg string, code []string) {
	if len(code) == 0 {
		panic("NoteSuggestNewCode() expects at least 1 code line")
	}

	printNiceHead(where, msg, typeNote())

	longest := len(fmt.Sprintf("%v", where.AtRow() + len(code) - 1))
	for i, line := range code {
		lineNum := fmt.Sprintf("%v", where.AtRow() + i)
		if len(lineNum) < longest {
			lineNum = strings.Repeat(" ", longest - len(lineNum)) + lineNum
		}

		printNiceLine(fmt.Sprintf("%v+\x1b[0m %v", typeNote().attr, lineNum), line)
	}
}
