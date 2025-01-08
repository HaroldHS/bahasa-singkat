package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"engine/src"
)

/* Virtual Stack implementation */
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

/* 
 * Virtual Memory implementation with simple dictionary for variable assignment.
 *
 * TODO: Add concurrency handler for preventing race condition.
 *
 */
type VirtualMemory struct {
	bilangan map[string]int
	untaian  map[string]string
}

func (m *VirtualMemory) InsertBilangan(key string, value string) {
	intValue, _ := strconv.Atoi(value)
	m.bilangan[key] = intValue
}

func (m *VirtualMemory) InsertUntaian(key string, value string) {
	m.untaian[key] = value
}

func (m *VirtualMemory) DeleteBilangan(key string) {
	delete(m.bilangan, key)
}

func (m *VirtualMemory) DeleteUntaian(key string) {
	delete(m.untaian, key)
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
	var virtualMemory VirtualMemory

	// Initialize virtual memory
	virtualMemory.bilangan = make(map[string]int)
	virtualMemory.untaian  = make(map[string]string)

	// Flag for conditional statement and loop statement
	bytecodeIsInsideIfStatementFlag := false
	isInsideLoop := false

	for _, content := range fileContents {
		// Insert bytecode into virtual memory
		if (content != "") && (content != "DO_NOTHING") {
			virtualStack.Push(content)
		}
		// When empty line encountered, it means the end of a single line BaSing code, so execute the line
		if content == "" {
			
			var currentBytecodes []byte
			numOfIterationBytecodes := []byte{} // for loop statement

			for {
				if virtualStack.isEmpty() {
					break
				}

				status, result := virtualStack.Pop()
				instruction, value := src.GetInstruction(result)
				if status && !bytecodeIsInsideIfStatementFlag && (instruction == "RETURN") {

					// TODO: handle 'tampilkan <arithmetic>' when inside a loop statement

					returnBytecode   := src.CompileBytecodeToAssembly(instruction, "")
					currentBytecodes := append(currentBytecodes, returnBytecode...)
					result           := strconv.Itoa(src.ExecuteAssembly(currentBytecodes))
					currentBytecodes = currentBytecodes[:0] // reset bytcode container after executing it

					nextInstructionStatus, nextInstructionValue := virtualStack.Pop()
					if nextInstructionStatus && (nextInstructionValue == "TAMPILKAN_FROM_STACK") {
						
						if isInsideLoop {
							currentBytecodes = append(currentBytecodes, src.GetAssemblyOfPrintFunction(result)...)
						} else {
							src.AssemblyPrintFunction(result)
							currentBytecodes = currentBytecodes[:0]
						}

					} else {
						// Push the result into stack with "PUSH" bytecode
						virtualStack.Push("PUSH " + result)
					}

				} else if status && !bytecodeIsInsideIfStatementFlag && (instruction == "TAMPILKAN") {

					stringValue := value[1:len(value)-1] // take the string value inside '
					if isInsideLoop {
						currentBytecodes = append(currentBytecodes, src.GetAssemblyOfPrintFunction(stringValue)...)
					} else {
						src.AssemblyPrintFunction(stringValue)
						currentBytecodes = currentBytecodes[:0]
					}

				} else if status && !bytecodeIsInsideIfStatementFlag && ((instruction == "SET_VARIABEL_BILANGAN") || (instruction == "SET_VARIABEL_UNTAIAN")) {
					
					pair := strings.SplitN(value, " ", 2)
					pairKey, pairValue := pair[0], pair[1]
					pairKey = pairKey[1:len(pairKey)-1]

					if instruction == "SET_VARIABEL_BILANGAN" {
						virtualMemory.InsertBilangan(pairKey, pairValue)
					}

				} else if status && !bytecodeIsInsideIfStatementFlag && ((instruction == "GET_VARIABEL_BILANGAN") || (instruction == "GET_VARIABEL_UNTAIAN")) {
					
					namaVariabel := value[1:len(value)-1]

					if instruction == "GET_VARIABEL_BILANGAN" {
						virtualStack.Push("PUSH " + strconv.Itoa(virtualMemory.bilangan[namaVariabel]))
					}

				} else if status && !bytecodeIsInsideIfStatementFlag && ((instruction == "LEBIH_KECIL") || (instruction == "LEBIH_BESAR") || (instruction == "SAMA_DENGAN")) {
					
					result := src.AssemblyComparisonFunction(instruction, currentBytecodes)

					if result == -1 {
						fmt.Println("[-] Error: invalid comparison operation")
						return
					} else if result == 0 { // condition not met, so skip/don't execute all bytecodes until END_BLOCK is found
						bytecodeIsInsideIfStatementFlag = true
					} else if result == 1 {
						continue
					} else {
						fmt.Println("[-] Error: unexpected error when performing comparison")
						return
					}

					currentBytecodes = currentBytecodes[:0]

				} else if status && (instruction == "PENGULANGAN") {

					numOfIterationBytecodes = append(numOfIterationBytecodes, currentBytecodes...) // previous bytecode is for the counter
					currentBytecodes = currentBytecodes[:0]     // empty current bytecodes container
					isInsideLoop = true
					_ = numOfIterationBytecodes                 // NOTE: this statement is just to bypass/remove 'decalred and not used' error warning

				} else if status && (instruction == "END_BLOCK") { // end of block, such as 'jika' statement, reset the flag
					
					if isInsideLoop { // execute loop statement
						isInsideLoop = false
						src.AssemblyLoopingFunction(numOfIterationBytecodes, currentBytecodes)
						numOfIterationBytecodes = numOfIterationBytecodes[:0]
						currentBytecodes = currentBytecodes[:0]
						continue
					}

					bytecodeIsInsideIfStatementFlag = false

				} else if status && (instruction == "ERROR") {
					
					stringValue := value[1:len(value)-1]
					fmt.Printf("[-] Error: %s\n", stringValue)
					return

				} else {

					if !bytecodeIsInsideIfStatementFlag {
						currentBytecodes = append(currentBytecodes, src.CompileBytecodeToAssembly(instruction, value)...)
					} else {
						continue
					}

				}
			}
		}
	}
}
