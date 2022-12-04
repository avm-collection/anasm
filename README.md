<p align="center">
	<img width="370px" src="res/logo.png"/>
	<p align="center">An assembler for avm in Go</p>
</p>
<p align="center">
	<a href="./LICENSE">
		<img alt="License" src="https://img.shields.io/badge/license-GPL-blue?color=7aca00"/>
	</a>
	<a href="https://github.com/LordOfTrident/anasm/issues">
		<img alt="Issues" src="https://img.shields.io/github/issues/LordOfTrident/anasm?color=0088ff"/>
	</a>
	<a href="https://github.com/LordOfTrident/anasm/pulls">
		<img alt="GitHub pull requests" src="https://img.shields.io/github/issues-pr/LordOfTrident/anasm?color=0088ff"/>
	</a>
	<br><br><br>
</p>

## Table of contents
* [Introduction](#introduction)
* [Quickstart](#quickstart)
* [Milestones](#milestones)
* [Documentation](#documentation)
* [Bugs](#bugs)
* [Make](#make)

## Introduction
An assembler for the [avm virtual machine](https://github.com/LordOfTrident/avm) written in Go

## Quickstart
```sh
$ make
$ make install
$ anasm -o=<OUTPUT PATH> <PROGRAM>
```

See [the `examples` folder](./examples) for example programs

## Milestones
- [X] Lexer
- [X] Compiling basic instructions
- [X] Labels
- [X] Instruction argument safety
- [ ] Macros

## Documentation
Coming soon.

## Bugs
If you find any bugs, please create an issue and report them.

## Make
Run `make all` to see all the make rules.
