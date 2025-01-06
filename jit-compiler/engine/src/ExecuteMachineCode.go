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

