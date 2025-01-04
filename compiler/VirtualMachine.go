package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

/* Stack implementation */
type Stack struct {
	items []string
}

func (s *Stack) Push(bytecode string) {
	s.items = append(s.items, bytecode)
}

func (s *Stack) Pop() (bool, string) {
	if s.isEmpty() {
		return false, ""
	}

	element := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return true, element
}

func (s *Stack) isEmpty() bool {
	if len(s.items) == 0 {
		return true
	}
	return false
}

/* Virtual Memory implementation for variable assignment */


/* Auxiliary function */
func getInstruction (line string) (string, string) {
	matchOneKeywordOnly, _ := regexp.MatchString(`^[a-zA-Z\_]+$`, line)

	if !matchOneKeywordOnly {
		result := strings.SplitN(line, " ", 2)
		return result[0], result[1]
	}

	return line, ""
}

func numberToLittleEndian32 (number string) ([]byte) {
	intNumber, _ := strconv.Atoi(number)

	// NOTE: After several debugging process, integer that is more than 127 needs zero padding
	if intNumber <= 127 {
		byteIntNumber := byte(intNumber)
		result := make([]byte, 1)
		result[0] = byteIntNumber
		return result
	} else {
		byteIntNumber := uint32(intNumber)
		result := make([]byte, 4)
		binary.LittleEndian.PutUint32(result, byteIntNumber)
		return result
	}
}

/* functions for JIT compiler */
func compileBytecodeToAssembly (instruction string, value string) ([]byte) {
	var assembly_code []byte

	if instruction == "TAMPILKAN_FROM_STACK" {
		/*
		 * NOTE: The x86_64 machine code is obtained from the assembly code below. In case of verification, do the steps below:
		 * 
		 *       1. Compile the assembly code below with `nasm -f bin [assembly file that contains the code below] -o [output file name]`
		 *       2. Obtain the hex byte with `cat [output file name] | hd -b`
		 * 
		 * Assembly code
		 *      |
		 *     \|/
		 *      v
		 * ```
		 * BITS 64
		 * 
		 * mov rdx, rsi ; rdx/3rd argument for write() syscall, string length, come from rsi/2nd argument of assemblyPrintFunction()
		 * mov rsi, rdi ; rsi/2nd argument for write() syscall, string address, come from rdi/1st argument of assemblyPrintFunction()
		 * 
		 * ; prevent null-byte when inserting 1 into rdi & rax with INC instruction 
		 * xor rdi, rdi ; empty rdi
		 * inc rdi      ; rdi = 1, stdout
		 * xor rax, rax ; empty rax
		 * inc rax      ; rax = 1, write()
		 * syscall
		 * ret
		 * ```
		 *
		 */
		assembly_code = append(assembly_code, []byte{
			0x48, 0x89, 0xf2, 0x48, 0x89, 0xfe, 0x48,
			0x31, 0xff, 0x48, 0xff, 0xc7, 0x48, 0x31,
			0xc0, 0x48, 0xff, 0xc0, 0x0f, 0x05, 0xc3,
		}...)
	} else if instruction == "PUSH" {
		assembly_code  = append(assembly_code, []byte{0x6a}...) // machine code of 'push'
		assembly_code  = append(assembly_code, numberToLittleEndian32(value)...)
	} else if instruction == "TAMBAH" {
		/*
		 * BITS 64
		 * 
		 * pop rax
		 * pop rbx
		 * add rax, rbx
		 * push rax
		 * 
		 */
		assembly_code  = append(assembly_code, []byte{0x58, 0x5b, 0x48, 0x01, 0xd8, 0x50}...)
	} else if instruction == "KURANG" {
		/*
		 * BITS 64
		 * 
		 * pop rax
		 * pop rbx
		 * sub rax, rbx
		 * push rax
		 * 
		 */
		assembly_code  = append(assembly_code, []byte{0x58, 0x5b, 0x48, 0x29, 0xd8, 0x50}...)
	} else if instruction == "KALI" {
		/*
		 * BITS 64
		 * 
		 * pop rax
		 * pop rbx
		 * mul rbx
		 * push rax
		 * 
		 */
		assembly_code  = append(assembly_code, []byte{0x58, 0x5b, 0x48, 0xf7, 0xe3, 0x50}...)
	} else if instruction == "BAGI" {
		/*
		 * BITS 64
		 * 
		 * pop rax
		 * pop rbx
		 * div rbx
		 * push rax
		 * 
		 */
		assembly_code  = append(assembly_code, []byte{0x58, 0x5b, 0x48, 0xf7, 0xf3, 0x50}...)
	} else if instruction == "RETURN" {
		// for return, value to be returned is stored in rax, so the code will be 'pop rax \n ret'
		assembly_code  = append(assembly_code, []byte{0x58, 0xc3}...)
	} else {
		assembly_code  = append(assembly_code, []byte{0x90}...) // machine code of 'pop'
	}

	return assembly_code
}

type executableFunction func() (int)
func executeAssembly (bytecodes []byte) (int) {
	assemblyFunctionMmap, err := syscall.Mmap(
		-1, 0, len(bytecodes),
		syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC,
		syscall.MAP_PRIVATE | syscall.MAP_ANONYMOUS,
	)

	if err != nil {
		fmt.Println("[-] ERROR: error in executeAssembly()")
		os.Exit(1)
	}

	for i := range bytecodes {
		assemblyFunctionMmap[i] = bytecodes[i]
	}

	unsafeAssemblyFunction := (uintptr)(unsafe.Pointer(&assemblyFunctionMmap))
	assemblyFunction       := *(*executableFunction)(unsafe.Pointer(&unsafeAssemblyFunction))
	return assemblyFunction()
}

type printFunction func(data *[]byte, length int)
func assemblyPrintFunction (bytecodes []byte, stringData string, stringLength int) {

	// This mmap is used for storing stringData into memory address that could be accessed by assemblyFunctionMmap
	stringDataMmap, dataErr := syscall.Mmap(
		-1, 0, len(stringData),
		syscall.PROT_READ | syscall.PROT_WRITE,
		syscall.MAP_PRIVATE | syscall.MAP_ANONYMOUS,
	)

	assemblyFunctionMmap, err := syscall.Mmap(
		-1, 0, len(bytecodes),
		syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC,
		syscall.MAP_PRIVATE | syscall.MAP_ANONYMOUS,
	)

	if (dataErr != nil) || (err != nil) {
		fmt.Println("[-] ERROR: error in executeAssembly()")
		os.Exit(1)
	}

	for i := range bytecodes {
		assemblyFunctionMmap[i] = bytecodes[i]
	}

	for i := range stringData {
		stringDataMmap[i] = stringData[i]
	}

	unsafeAssemblyFunction := (uintptr)(unsafe.Pointer(&assemblyFunctionMmap))
	assemblyFunction       := *(*printFunction)(unsafe.Pointer(&unsafeAssemblyFunction))
	assemblyFunction(&stringDataMmap, stringLength)
}

/* Program entry point */
func main () {
	fileName := os.Args[1]
	filePointer, err := os.Open(fileName)
	defer filePointer.Close() // auto close file when main returns/done

	if err != nil {
		fmt.Println("[-] Error: Invalid file")
		return
	}

	fileScanner := bufio.NewScanner(filePointer)
	fileScanner.Split(bufio.ScanLines)
	var fileContents []string
	for fileScanner.Scan() {
		fileContents = append(fileContents, fileScanner.Text())
	}

	var virtualStack Stack

	for _, content := range fileContents {
		// Insert bytecode into virtual memory
		if (content != "") && (content != "DO_NOTHING") {
			virtualStack.Push(content)
		}
		// When empty line encountered, it means the end of a single line BaSing code, so execute the line
		if content == "" {
			var currentBytecodes []byte
			for {
				if virtualStack.isEmpty() {
					break
				}

				status, result := virtualStack.Pop()
				instruction, value := getInstruction(result)
				if status && (instruction == "RETURN") {
					returnBytecode   := compileBytecodeToAssembly(instruction, "")
					currentBytecodes := append(currentBytecodes, returnBytecode...)
					result           := strconv.Itoa(executeAssembly(currentBytecodes))
					currentBytecodes = currentBytecodes[:0] // reset bytcode container after executing it
					fmt.Println(result)

					nextInstructionStatus, nextInstructionValue := virtualStack.Pop()
					if nextInstructionStatus && (nextInstructionValue == "TAMPILKAN_FROM_STACK"){
						currentBytecodes = compileBytecodeToAssembly(nextInstructionValue, "")
						assemblyPrintFunction(currentBytecodes, result, len(result))
						currentBytecodes = currentBytecodes[:0]
					} else {
						// Push the result into stack with "PUSH" bytecode
						virtualStack.Push("PUSH " + result)
					}
				} else {
					currentBytecodes = append(currentBytecodes, compileBytecodeToAssembly(instruction, value)...)
				}
			}
		}
	}
}
