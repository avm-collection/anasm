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
- `1.0.0`: Support avm 1.1 - remove registers, commas, improved dup and swap instructions - syntax
           is stabilized
- `1.1.0`: Remove colons, make the default output path be the basename of the input without the
           extension. If there was no extension, add '.out'
- `1.2.0`: Add a disassembler
- `1.2.1`: Support avm 1.2 (Just an internal avm update)
- `1.3.1`: Support avm 1.5
- `1.4.1`: Support avm 1.6
- `1.4.2`: Fix octal integer lexing
- `1.5.2`: Add binary integer and character literals support, prepare string support
- `1.6.2`: Support avm 1.7, add strings
- `1.7.2`: Support avm 1.8 (no changes to the assembler)
