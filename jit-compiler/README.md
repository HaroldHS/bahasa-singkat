# Bahasa Singkat Interpreter

Bahasa Singkat Compiler (BaSiCl) is a small compiler (JIT compiler) that compile Bahasa Singkat programming language from a given input/file.
Bahasa Singkat Compiler works by parsing the input/file according to the BNF of Bahasa Singkat programming language, create bytecode/intermediate representation (IR) of it.
Later on, the result of the IR/Bytecode generation could be executed by an engine which is written in Golang.

### Software Requirements
* GHCI
* Golang

### Supported systems
* Linux x86-64

### How to run

```sh
# 0. Change/Give permission for 'build.sh' file
chmod +x ./build.sh

# 1. Build the project with 'build.sh'
./build.sh build

# 2. Compile a BaSing file with compiler artefact
./artefacts/compiler program.basing result.bsr

# 3. Run the bytecode file (.bsr file) with engine artefact
./artefacts/engine result.bsr

# 4. Clean all artefacts
./build.sh clean
```

> NOTE: As BaSing only support integer number, not floating number, there should be no arithmetic operation which result in zero division(e.g. 4/5) or fractional number. If so, the program could crash or print invalid number
