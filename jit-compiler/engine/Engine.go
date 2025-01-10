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
	fileName         := os.Args[1]
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

	// Flags for conditional statement and loop statement
	ifConditionNotMet := false
	isInsideIfBlock   := false
	isInsideLoopBlock := false

	// Variables to save all bytecodes regarding loop statement as if bytecodes in 'if content == ""' would be removed due to out of scope
	loopStatementBytecodes  := []byte{}
	numOfIterationBytecodes := []byte{}

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
				instruction, value := src.GetInstruction(result)
				if status && !isInsideIfBlock && !isInsideLoopBlock {
					if instruction == "RETURN" {
						returnBytecode   := src.CompileBytecodeToAssembly(instruction, "")
						currentBytecodes := append(currentBytecodes, returnBytecode...)
						result           := strconv.Itoa(src.ExecuteAssembly(currentBytecodes))
						currentBytecodes  = currentBytecodes[:0] // reset bytecode slice after executing it

						nextInstructionStatus, nextInstructionValue := virtualStack.Pop()
						if nextInstructionStatus && (nextInstructionValue == "TAMPILKAN_FROM_STACK") {
							src.AssemblyPrintFunction(result)
						} else if nextInstructionStatus {
							virtualStack.Push("PUSH " + result) // push the result onto the stack with "PUSH" bytecode
						}
					} else if instruction == "TAMPILKAN" {

						stringValue := value[1 : len(value)-1]  // take the string value inside '
						src.AssemblyPrintFunction(stringValue)
						currentBytecodes = currentBytecodes[:0]

					} else if (instruction == "SET_VARIABEL_BILANGAN") || (instruction == "SET_VARIABEL_UNTAIAN") {
						pair               := strings.SplitN(value, " ", 2)
						pairKey, pairValue := pair[0], pair[1]
						pairKey             = pairKey[1 : len(pairKey)-1] // obtain key inside '

						if instruction == "SET_VARIABEL_BILANGAN" {
							virtualMemory.InsertBilangan(pairKey, pairValue)
						}
						if instruction == "SET_VARIABEL_UNTAIAN" {
							virtualMemory.InsertUntaian(pairKey, pairValue[1 : len(pairValue)-1])
						}
					} else if (instruction == "GET_VARIABEL_BILANGAN") || (instruction == "GET_VARIABEL_UNTAIAN") {
						namaVariabel := value[1 : len(value)-1]

						if instruction == "GET_VARIABEL_BILANGAN" {
							virtualStack.Push("PUSH " + strconv.Itoa(virtualMemory.bilangan[namaVariabel]))
						}
						if instruction == "GET_VARIABEL_UNTAIAN" {
							namaVariabel := value[1 : len(value)-1]
							nextInstructionStatus, nextInstructionValue := virtualStack.Pop()
							nextNextInstructionStatus, nextNextInstructionValue := virtualStack.Pop()
							
							if (nextInstructionStatus && nextNextInstructionStatus) && (nextInstructionValue == "RETURN" && nextNextInstructionValue == "TAMPILKAN_FROM_STACK") {
								stringValue     := virtualMemory.untaian[namaVariabel]
								src.AssemblyPrintFunction(stringValue)
							} else {
								virtualStack.Push(nextNextInstructionValue)
								virtualStack.Push(nextInstructionValue)
							}
						}
					} else if (instruction == "LEBIH_KECIL") || (instruction == "LEBIH_BESAR") || (instruction == "SAMA_DENGAN") {
						nextInstructionStatus, nextInstructionValue := virtualStack.Pop()

						if nextInstructionStatus && (nextInstructionValue == "START_JIKA_BLOCK") {
							result := src.AssemblyComparisonFunction(instruction, currentBytecodes)

							if result == 0 {
								// if condition not met, then skip/don't execute all bytecodes until END_JIKA_BLOCK is found
								ifConditionNotMet = true
								isInsideIfBlock   = true
							} else if result == 1{
								isInsideIfBlock = true
							} else if result == -1 {
								fmt.Println("[-] Error: invalid comparison operation")
								return
							} else {
								fmt.Println("[-] Error: unexpected error when performing comparison")
								return
							}
						} else {
							virtualStack.Push(nextInstructionValue)
						}

						currentBytecodes = currentBytecodes[:0]
					} else if instruction == "PENGULANGAN"{
						nextInstructionStatus, nextInstructionValue := virtualStack.Pop()
						
						if nextInstructionStatus && (nextInstructionValue == "START_PENGULANGAN_BLOCK") {
							isInsideLoopBlock = true
						
							// change previous bytecodes as bytecodes for counter
							numOfIterationBytecodes = append(numOfIterationBytecodes, currentBytecodes...)
							_ = numOfIterationBytecodes // NOTE: this statement is intended to bypass/remove 'declared and not used' error/warning
							currentBytecodes = currentBytecodes[:0]
							continue
						} else {
							virtualStack.Push(nextInstructionValue)
						}
					} else {
						currentBytecodes = append(currentBytecodes, src.CompileBytecodeToAssembly(instruction, value)...)
					}
				}

				if status && isInsideIfBlock && !isInsideLoopBlock {
					if instruction == "END_JIKA_BLOCK" {
						isInsideIfBlock = false
						continue
					}

					if !ifConditionNotMet {
						if instruction == "TAMPILKAN" {
							stringValue := value[1 : len(value)-1]  // take the string value inside '
							src.AssemblyPrintFunction(stringValue)
							currentBytecodes = currentBytecodes[:0]
						}
					}
				}

				if status && !isInsideIfBlock && isInsideLoopBlock {
					if instruction == "END_PENGULANGAN_BLOCK" {
						src.AssemblyLoopingFunction(numOfIterationBytecodes, loopStatementBytecodes)
						isInsideLoopBlock       = false
						numOfIterationBytecodes = numOfIterationBytecodes[:0]
						loopStatementBytecodes  = loopStatementBytecodes[:0]
					}

					// cases for bytecodes inside loop block
					if instruction == "TAMPILKAN" {
						loopStatementBytecodes = append(loopStatementBytecodes, src.GetAssemblyOfPrintFunction(value[1 : len(value)-1])...)
					} else if instruction == "GET_VARIABEL_UNTAIAN" { // tampilkan variabel untaian '...'
						namaVariabel := value[1 : len(value)-1]
						nextInstructionStatus, nextInstructionValue := virtualStack.Pop()
						nextNextInstructionStatus, nextNextInstructionValue := virtualStack.Pop()

						if (nextInstructionStatus && nextNextInstructionStatus) && (nextInstructionValue == "RETURN" && nextNextInstructionValue == "TAMPILKAN_FROM_STACK") {
							stringValue     := virtualMemory.untaian[namaVariabel]
							loopStatementBytecodes = append(loopStatementBytecodes, src.GetAssemblyOfPrintFunction(stringValue)...)
						}
					} else {
						loopStatementBytecodes = append(loopStatementBytecodes, src.CompileBytecodeToAssembly(instruction, value)...)
					}
				}

				if status && (instruction == "ERROR") {
					fmt.Printf("[-] Error: %s\n", value[1 : len(value)-1])
					return
				}

				if status && (instruction == "DO_NOTHING") {
					continue
				}

				if !status {
					fmt.Println("[-] Error: unexpected error")
					return
				}
			}
		}
	}
}
