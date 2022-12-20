<p align="center">
	<img width="250px" src="res/logo.png">
</p>
<p align="center">An assembler for avm</p>

<p align="center">
	<a href="./LICENSE">
		<img alt="License" src="https://img.shields.io/badge/license-GPL-blue?color=26d374"/>
	</a>
	<a href="https://github.com/avm-collection/anasm/issues">
		<img alt="Issues" src="https://img.shields.io/github/issues/avm-collection/anasm?color=4f79e4"/>
	</a>
	<a href="https://github.com/avm-collection/anasm/pulls">
		<img alt="GitHub pull requests" src="https://img.shields.io/github/issues-pr/avm-collection/anasm?color=4f79e4"/>
	</a>
</p>

An assembler for the [AVM virtual machine](https://github.com/avm-collection/avm) written in Go

## Table of contents
* [Quickstart](#quickstart)
* [Milestones](#milestones)
* [Editors](#editors)
* [Documentation](#documentation)
* [Bugs](#bugs)
* [Make](#make)

## Quickstart
```sh
$ make
$ make install
$ anasm ./examples/fib.anasm
$ ./fib
```
`anasm ./examples/fib.anasm` compiles the fibonacci sequence example into an AVM binary `./fib`.

See [the `./examples` folder](./examples) for example programs

## Milestones
- [X] Lexer
- [X] Compiling basic instructions
- [X] Labels
- [X] Instruction argument safety
- [X] Macros

## Editors
Syntax highlighting configs for text editors are in the [`./editors`](./editors) folder

## Documentation
Hosted [here](https://avm-collection.github.io/anasm/documentation)

## Bugs
If you find any bugs, please create an issue and report them.

## Make
Run `make all` to see all the make rules.
