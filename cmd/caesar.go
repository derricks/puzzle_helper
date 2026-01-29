package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

// CaesarShiftResult represents a single shifted string, including the shift amount.
type CaesarShiftResult struct {
	ShiftedText string
	Shift       int
}

// String implements fmt.Stringer for CaesarShiftResult.
func (csr CaesarShiftResult) String() string {
	return fmt.Sprintf("%d. %s", csr.Shift, csr.ShiftedText)
}

// PerformCaesarShifts contains the core logic for generating all Caesar shifts.
func PerformCaesarShifts(inputText string) []CaesarShiftResult {
	var results []CaesarShiftResult

	for shift := 1; shift <= 25; shift++ {
		var shiftedString strings.Builder
		for _, curByte := range []byte(inputText) {
			shiftedString.WriteByte(ShiftByte(curByte, shift))
		}
		results = append(results, CaesarShiftResult{
			ShiftedText: shiftedString.String(),
			Shift:       shift,
		})
	}
	return results
}

// ShiftByte shifts a single byte by the given amount.
func ShiftByte(byteToShift byte, shiftAmount int) byte {
	var startByte byte
	var endByte byte
	if IsUppercaseAscii(byteToShift) {
		startByte = 'A'
		endByte = 'Z'
	} else if IsLowercaseAscii(byteToShift) {
		startByte = 'a'
		endByte = 'z'
	} else {
		return byteToShift
	}

	newByte := byteToShift + byte(shiftAmount)
	if newByte > endByte {
		newByte = startByte + (newByte - endByte - byte(1))
	}
	return newByte
}

// IsUppercaseAscii checks if a byte is an uppercase ASCII letter.
func IsUppercaseAscii(char byte) bool {
	return 'A' <= char && char <= 'Z'
}

// IsLowercaseAscii checks if a byte is a lowercase ASCII letter.
func IsLowercaseAscii(char byte) bool {
	return 'a' <= char && char <= 'z'
}

// printCaesarShifts handles the cobra command for Caesar cipher.
func printCaesarShifts(command *cobra.Command, args []string) {
	fullString := strings.Join(args, " ")
	results := PerformCaesarShifts(fullString)

	// Convert results to fmt.Stringer slice for outputResponse
	var output []interface{}
	for _, res := range results {
		output = append(output, res)
	}
	outputResponse(output)
}
