package src

func CompileBytecodeToAssembly (instruction string, value string) ([]byte) {
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
		assembly_code  = append(assembly_code, NumberToLittleEndian(value)...)
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
