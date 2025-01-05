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

func numberToLittleEndian (number string) ([]byte) {
	intNumber, _ := strconv.Atoi(number)

	// NOTE: After several debugging process, integer that is more than 127 needs zero padding (32 bit).
	if intNumber <= 127 {
		byteIntNumber := byte(intNumber)
		result := make([]byte, 1)
		result[0] = byteIntNumber
		return result
	} else if intNumber <= 4294967295 {
		byteIntNumber := uint32(intNumber)
		result := make([]byte, 4)
		binary.LittleEndian.PutUint32(result, byteIntNumber)
		return result
	} else {
		byteIntNumber := uint64(intNumber)
		result := make([]byte, 8)
		binary.LittleEndian.PutUint64(result, byteIntNumber)
		return result
	}
}

/* functions for JIT compiler */
func compileBytecodeToAssembly (instruction string, value string) ([]byte) {
	var assembly_code []byte

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
	 * ;assembly code here;
	 *
	 * ```
	 * 
	 */
	if instruction == "PUSH" {
		assembly_code  = append(assembly_code, []byte{0x6a}...) // machine code of 'push'
		assembly_code  = append(assembly_code, numberToLittleEndian(value)...)
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
		assembly_code  = append(assembly_code, []byte{0x90}...) // machine code of 'nop'
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
	result                 := assemblyFunction()

	// Unmap the mapped memories
	syscall.Munmap(assemblyFunctionMmap)

	return result
}

//type printFunction func(data *[]byte, length int)
type printFunction func()
func assemblyPrintFunction (stringData string) {

	/*
	 * NOTE: As golang mmap() is different from C mmap(), string is inserted inside the bytecode instead
	 *       of accessing the string address from RSI register directly when calling the mmap function.
	 *
	 *
	 * C example:
	 *
	 * typedef void(*printFunction)(char *msg, size_t msg_len)
	 *
	 * void assemblyPrintFunction(char *msg) {
	 *   char opcode[]     = ".....";   <- contains instruction with accessing RSI and RDX
	 *   size_t opcode_len = strlen(opcode);
	 *   size_t msg_len    = strlen(msg);
	 *
	 *   void *code = mmap(NULL, opcode_len, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
	 *
	 *   if (code == MAP_FAILED) { return 1; }
	 *
	 *   memcpy(code, opcode, opcode_len);
	 *   ((printFunction)code)(msg, msg_len);
	 *
	 * }
	 *
	 */

	stringLengthLE32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(stringLengthLE32, uint32(len(stringData)))

	/*
	 * Assembly of machine code below:
	 *
	 * xor rax, rax
	 * inc rax      ; rax = 1
	 * xor rdi, rdi
	 * inc rdi      ; rdi = 1
	 *
	 */
	customBytecodes := []byte{0x48, 0x31, 0xc0, 0x48, 0xff, 0xc0, 0x48, 0x31, 0xff, 0x48, 0xff, 0xc7}
	customBytecodes = append(customBytecodes, []byte{0x48, 0xc7, 0xc2}...)                         // mov rdx, .... <- 32 bit value
	customBytecodes = append(customBytecodes, stringLengthLE32...)                                 // 32 bit value / length of stringData
	customBytecodes = append(customBytecodes, []byte{0x48, 0x8d, 0x35, 0x03, 0x00, 0x00, 0x00}...) // lea rsi, [rip+0x03]
	/*
	 * Assembly of machine code below:
	 *
	 * syscall
	 * ret
	 *
	 */
	customBytecodes = append(customBytecodes, []byte{0x0f, 0x05, 0xc3}...)

	// Insert the string into the machine code
	for i := range stringData {
		customBytecodes = append(customBytecodes, []byte{byte(stringData[i])}...)
	}
	// Append carriage return + new line + null terminator after the string
	customBytecodes = append(customBytecodes, []byte{0x0d, 0x0a, 0x00}...) // \r\n\00

	assemblyFunctionMmap, err := syscall.Mmap(
		-1, 0, len(customBytecodes),
		syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC,
		syscall.MAP_PRIVATE | syscall.MAP_ANONYMOUS,
	)

	if err != nil {
		fmt.Println("[-] ERROR: Can't create memory mapping for constructed machine code in assemblyPrintFunction()")
		os.Exit(1)
	}
	
	for i := range customBytecodes {
		assemblyFunctionMmap[i] = customBytecodes[i]
	}

	unsafeAssemblyFunction := (uintptr)(unsafe.Pointer(&assemblyFunctionMmap))
	assemblyFunction       := *(*printFunction)(unsafe.Pointer(&unsafeAssemblyFunction))
	assemblyFunction()

	syscall.Munmap(assemblyFunctionMmap)

	// TODO: Debug the issue of new line doesn't appear on screen even though it has been added in customBytecodes
	fmt.Println("")
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

					nextInstructionStatus, nextInstructionValue := virtualStack.Pop()
					if nextInstructionStatus && (nextInstructionValue == "TAMPILKAN_FROM_STACK"){
						assemblyPrintFunction(result)
						currentBytecodes = currentBytecodes[:0]
					} else {
						// Push the result into stack with "PUSH" bytecode
						virtualStack.Push("PUSH " + result)
					}
				} else if status && (instruction == "TAMPILKAN"){
					stringValue := value[1:len(value)-1] // take the string value inside '
					assemblyPrintFunction(stringValue)
					currentBytecodes = currentBytecodes[:0]
				} else {
					currentBytecodes = append(currentBytecodes, compileBytecodeToAssembly(instruction, value)...)
				}
			}
		}
	}
}
