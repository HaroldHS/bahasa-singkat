package src

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type executableFunction func() (int)
func ExecuteAssembly (bytecodes []byte) (int) {
	assemblyFunctionMmap, err := syscall.Mmap(
		-1, 0, len(bytecodes),
		syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC,
		syscall.MAP_PRIVATE | syscall.MAP_ANONYMOUS,
	)

	if err != nil {
		fmt.Println("[-] ERROR: error in ExecuteAssembly()")
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

type printFunction func()
func AssemblyPrintFunction (stringData string) {

	/*
	 * NOTE: As golang mmap() is different from C mmap(), string is inserted inside the bytecode instead
	 *       of accessing the string address from RSI register directly when calling the mmap function.
	 *
	 * C example:
	 *
	 * typedef void(*printFunction)(char *msg, size_t msg_len)
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
		fmt.Println("[-] ERROR: Can't create memory mapping for constructed machine code in AssemblyPrintFunction()")
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

type comparisonFunction func() (int)
func AssemblyComparisonFunction(command string, previousBytecode []byte) (int) {
	// Previous byte code (e.g. PUSH 1, GET_VARIABEL_BILANGAN "test")
	customBytecodes := previousBytecode

	// Logic of assembly code below -> if (condition met) then return 1 else return 0
	//
	// As a result, in Engine.go, if condition not met, skip instructions until first END_IF bytecode detected

	/*
	 * Assembly of machine code below:
	 *
	 * pop rax
	 * pop rbx
	 * lea rsi, [rip+0x18] <- for jumping to code 'mov rax,1 ret' because the conditional code below is 24 bytes from this code
	 * lea rdi, [rip+0x09] <- for jumping to code 'mov rax, 0 ret' because the conditional code below is 9 bytes from this code
	 *
	 */
	customBytecodes = append(customBytecodes, []byte{0x58, 0x5b, 0x48, 0x8d, 0x35, 0x18, 0x00, 0x00, 0x00, 0x48, 0x8d, 0x3d, 0x09, 0x00, 0x00, 0x00}...)
	/*
	 * Assembly of machine code below:
	 *
	 * cmp rax, rbx
	 * cmovne/cmovnl/cmovng rsi, rdi ; if condition not met, set rsi = rdi. So, when 'jmp rsi' executed, it will jump to the address that previously rdi has
	 * jmp rsi
	 * mov rax, 0
	 * ret
	 * 
	 */
	if command == "SAMA_DENGAN" { // cmovne
		customBytecodes = append(customBytecodes, []byte{0x48, 0x39, 0xd8, 0x48, 0x0f, 0x45, 0xf7, 0xff, 0xe6, 0x48, 0xc7, 0xc0, 0x00, 0x00, 0x00, 0x00, 0xc3}...)
	} else if command == "LEBIH_KECIL" { // cmovl
		customBytecodes = append(customBytecodes, []byte{0x48, 0x39, 0xd8, 0x48, 0x0f, 0x4d, 0xf7, 0xff, 0xe6, 0x48, 0xc7, 0xc0, 0x00, 0x00, 0x00, 0x00, 0xc3}...)
	} else if command == "LEBIH_BESAR" { // cmovg
		customBytecodes = append(customBytecodes, []byte{0x48, 0x39, 0xd8, 0x48, 0x0f, 0x4e, 0xf7, 0xff, 0xe6, 0x48, 0xc7, 0xc0, 0x00, 0x00, 0x00, 0x00, 0xc3}...)
	} else {
		return -1;
	}
	/*
	 * Assembly of machine code below:
	 *
	 * mov rax, 1
	 * ret
	 *
	 */
	customBytecodes = append(customBytecodes, []byte{0x48, 0xc7, 0xc0, 0x01, 0x00, 0x00, 0x00, 0xc3}...)

	assemblyFunctionMmap, err := syscall.Mmap(
		-1, 0, len(customBytecodes),
		syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC,
		syscall.MAP_PRIVATE | syscall.MAP_ANONYMOUS,
	)

	if err != nil {
		fmt.Println("[-] ERROR: Can't create memory mapping for constructed machine code in AssemblyComparisonFunction()")
		os.Exit(1)
	}

	for i := range customBytecodes {
		assemblyFunctionMmap[i] = customBytecodes[i]
	}
	
	unsafeAssemblyFunction := (uintptr)(unsafe.Pointer(&assemblyFunctionMmap))
	assemblyFunction       := *(*executableFunction)(unsafe.Pointer(&unsafeAssemblyFunction))
	result                 := assemblyFunction()

	// Unmap the mapped memories
	syscall.Munmap(assemblyFunctionMmap)

	return result
}
