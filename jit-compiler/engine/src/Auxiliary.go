package src

import (
	"encoding/binary"
	"regexp"
	"strconv"
	"strings"
)

func GetInstruction (line string) (string, string) {
	matchOneKeywordOnly, _ := regexp.MatchString(`^[a-zA-Z\_]+$`, line)

	if !matchOneKeywordOnly {
		result := strings.SplitN(line, " ", 2)
		return result[0], result[1]
	}

	return line, ""
}

func NumberToLittleEndian (number string) ([]byte) {
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
