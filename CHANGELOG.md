# ANASM changelog
* [ANASM 0](#anasm-0)
* [ANASM 1](#anasm-1)

## ANASM 0
- `0.1.0`: Support avm version 0.2
- `0.2.0`: Added instruction argument safety
- `0.3.0`: Added an option to create an executable output file
- `0.4.0`: Support avm 0.3
- `0.4.1`: Parameter improvements, flags can now come after parameters
- `0.5.1`: Add octal and float instruction arguments
- `0.6.1`: Support avm 0.4

## ANASM 1
- `1.0.0`:  Support avm 1.1 - remove registers, commas, improved dup and swap instructions - syntax
            is stabilized
- `1.1.0`:  Remove colons, make the default output path be the basename of the input without the
            extension. If there was no extension, add '.out'
- `1.2.0`:  Add a disassembler
- `1.2.1`:  Support avm 1.2 (just an internal avm update)
- `1.3.1`:  Support avm 1.5
- `1.4.1`:  Support avm 1.6
- `1.4.2`:  Fix octal integer lexing
- `1.5.2`:  Add binary integer and character literals support, prepare string support
- `1.6.2`:  Support avm 1.7, add strings
- `1.7.2`:  Support avm 1.8 (no changes to the assembler)
- `1.7.3`:  Fix variable addresses
- `1.8.3`:  Support avm 1.9 (file IO)
- `1.9.3`:  Support avm 1.10 (file IO improvement)
- `1.9.4`:  Fix escape sequence lexing
- `1.10.4`: Add constant expressions (szof, +, -, *, /, %)
- `1.11.4`: Add macros, let statement filling
- `1.11.5`: Fix compiler bug with instruction arguments
- `1.12.5`: Syntax changes (remove '@' before var/label referemces, new type names...)
- `1.13.5`: Better error system
- `1.13.6`: Fix a bug with program size and with let arrays
- `1.13.7`: Fix the disassembler
- `1.13.8`: Fix constant expressions and lexer errors
- `1.14.8`: Add power, bit and, bit or and bit shifting operations to constant expressions
- `1.15.8`: Support avm 1.11
- `1.15.9`: Fix the compiler not erroring over invalid expressions
- `1.16.9`: Support avm 1.12 (loading shared libraries)
- `1.17.9`: Add include and emb keywords
- `1.18.9`: Support avm 1.13 (flush instruction)
- `1.19.9`: Support avm 1.14 (memory now starts at 0 and not 1)
- `1.20.9`: Implicit pushes
