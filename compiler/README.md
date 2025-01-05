# Bahasa Singkat Interpreter

Bahasa Singkat Compiler (BaSiCl) is a small compiler (JIT compiler) that compile Bahasa Singkat programming language from a given input/file.
Bahasa Singkat Compiler works by parsing the input/file according to the BNF of Bahasa Singkat programming language, create bytecode/intermediate representation (IR) of it.
Later on, the result of the IR/Bytecode generation could be executed by virtual machine which is written in Golang.

### Sytem / Software Requirements

* GHCI
* Golang

### How to run

```sh
# 1. Run Main.hs with GHCI
ghci Main.hs

# 2. Inside GHCI, call main function
ghci> main

# 3. Give the file name (relative path)
[*] File name: ./example/program.basing
[*] IR file name: ./example/result.bsr

# 4. Quit from GHCI
ghci>:q

# 5. Run virtual machine
go run VirtualMachine.go ./example/result.bsr
```

> NOTE: As BaSing only support integer number, not floating number, there should be no arithmetic operation which result in zero division(e.g. 4/5) or fractional number . If so, the program could crash or print invalid number
