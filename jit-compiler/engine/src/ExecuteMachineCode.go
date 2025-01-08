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

	/*
	 * Assembly code logic for printing function
	 *
	 * mov rax, 1
	 * mov rdi, 1
	 * mov rdx, ...           <- len of string + 1 (null byteccode / 0x00)
	 * lea rsi, [rip + 0x0b]  <- 2 (len of 'syscall' machine code) + 7 (len of 'lea, rdx, [rip + ...]') + 2 (len of 'jmp rdx' machine code)
	 * syscall
	 * lea rdx, [rip + ...]   <- 2 (len of 'jmp rdx' machine code) + len of string + 3 (carriage return + line feed + null terminator)
	 * jmp rdx
	 * ... string byte here ...
	 * ret
	 *
	 */

	jumpToReturnAfterStringLE32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(jumpToReturnAfterStringLE32, uint32(2 + len(stringData) + 3))

	stringLengthLE32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(stringLengthLE32, uint32(len(stringData) + 3))

	customBytecodes := []byte{
		0x48, 0xc7, 0xc0, 0x01, 0x00, 0x00, 0x00,  // mov rax, 1
		0x48, 0xc7, 0xc7, 0x01, 0x00, 0x00, 0x00,  // mov rdi, 1
		0x48, 0xc7, 0xc2, /* stringLengthLE32 */   // mov rdx, ...
	}
	customBytecodes = append(customBytecodes, stringLengthLE32...)
	customBytecodes = append(customBytecodes, []byte{
		0x48, 0x8d, 0x35, 0x0b, 0x00, 0x00, 0x00,            // lea rsi, [rip + 0x0b]
		0x0f, 0x05,                                          // syscall
		0x48, 0x8d, 0x15, /* jumpToReturnAfterStringLE32 */  // lea rdx, [rip + ...]
	}...)
	customBytecodes = append(customBytecodes, jumpToReturnAfterStringLE32...)
	customBytecodes = append(customBytecodes, []byte{
		0xff, 0xe2,  // jmp rdx
	}...)

	// Insert the string into the machine code
	for i := range stringData {
		customBytecodes = append(customBytecodes, []byte{byte(stringData[i])}...)
	}

	// Append null terminator + return instruction after the string
	customBytecodes = append(customBytecodes, []byte{
		0x0d, // carriage return
		0x0a, // line feed
		0x00, // null terminator for string
		0xc3, // ret
	}...)

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
}

type comparisonFunction func() (int)
func AssemblyComparisonFunction(command string, previousBytecodes []byte) (int) {
	// Previous byte code (e.g. PUSH 1, GET_VARIABEL_BILANGAN "test")
	customBytecodes := previousBytecodes

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
	 * cmovne/cmovnl/cmovng rsi, rdi ; if condition not met, set rsi = rdi. So, when 'jmp rsi' executed, it will jump to the address that previously rdi has set
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

	syscall.Munmap(assemblyFunctionMmap)

	return result
}

/*
 * NOTE: This is an additional function in order to support looping operation by getting the bytecode of the bytecodes inside loop statement
 * 
 */
func GetAssemblyOfPrintFunction(stringData string) ([]byte) {
	/* This code is similar to AssemblyPrintFunction */
	jumpToReturnAfterStringLE32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(jumpToReturnAfterStringLE32, uint32(2 + len(stringData) + 3))
	stringLengthLE32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(stringLengthLE32, uint32(len(stringData) + 3))
	customBytecodes := []byte{
		0x48, 0xc7, 0xc0, 0x01, 0x00, 0x00, 0x00,
		0x48, 0xc7, 0xc7, 0x01, 0x00, 0x00, 0x00,
		0x48, 0xc7, 0xc2,
	}
	customBytecodes = append(customBytecodes, stringLengthLE32...)
	customBytecodes = append(customBytecodes, []byte{
		0x48, 0x8d, 0x35, 0x0b, 0x00, 0x00, 0x00,
		0x0f, 0x05,
		0x48, 0x8d, 0x15,
	}...)
	customBytecodes = append(customBytecodes, jumpToReturnAfterStringLE32...)
	customBytecodes = append(customBytecodes, []byte{
		0xff, 0xe2,
	}...)
	for i := range stringData {
		customBytecodes = append(customBytecodes, []byte{byte(stringData[i])}...)
	}
	customBytecodes = append(customBytecodes, []byte{
		0x0d,
		0x0a,
		0x00,
		/* no return value in order to prevent loop break */
	}...)
	/* END */

	return customBytecodes
}
/* END */

type loopingFunction func()
func AssemblyLoopingFunction(previousMachineCodeForCounter []byte, machineCodeInsideLoop []byte) {

	/*
	 * Assembly code logic for looping function
	 *
	 * ...all previous machine codes which ends with 'PUSH' instruction...
	 *
	 * pop rbx
	 * lea r10, [rip] <- jump to 'all machine codes inside loop' if rbx is not 0
	 *
	 * ...all machine codes inside loop...
	 *
	 * dec rbx
	 * cmp rbx, rb0
	 * cmovg r10, r10
	 * lea r9, [rip+0x0b]
	 * cmp rbx, 0
	 * cmove r10, r9
	 * jmp r10
	 * ret
	 *
	 */
	
	customBytecodes := previousMachineCodeForCounter
	customBytecodes = append(customBytecodes, []byte{
		0x5b,                                     // pop rbx
		0x4c, 0x8d, 0x15, 0x00, 0x00, 0x00, 0x00, // lea r10, [rip]
	}...)
	customBytecodes = append(customBytecodes, machineCodeInsideLoop...)
	customBytecodes = append(customBytecodes, []byte{
		0x48, 0xff, 0xcb,                         // dec rbx
		0x48, 0x83, 0xfb, 0x00,                   // cmp rbx, 0
		0x4d, 0x0f, 0x4f, 0xd2,                   // cmovg r10, r10
		0x4c, 0x8d, 0x0d, 0x0b, 0x00, 0x00, 0x00, // lea r9, [rip+0x0b]
		0x48, 0x83, 0xfb, 0x00,                   // cmp rbx, 0
		0x4d, 0x0f, 0x44, 0xd1,                   // cmove r10, r9
		0x41, 0xff, 0xe2,                         // jmp r10
		0xc3,                                     // ret
	}...)

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
	assemblyFunction       := *(*loopingFunction)(unsafe.Pointer(&unsafeAssemblyFunction))
	assemblyFunction()

	syscall.Munmap(assemblyFunctionMmap)
}

