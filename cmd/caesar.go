package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

func printCaesarShifts(command *cobra.Command, args []string) {
	fullString := strings.Join(args, " ")
	// run each possible shift
	for shift := 1; shift <= 25; shift++ {
		fmt.Printf("%d. ", shift)
		for _, curByte := range []byte(fullString) {
			fmt.Printf("%c", shiftByte(curByte, shift))
		}
		fmt.Print("\n")
	}
}

func shiftByte(byteToShift byte, shiftAmount int) byte {
	var startByte byte
	var endByte byte
	if isUppercaseAscii(byteToShift) {
		startByte = 'A'
		endByte = 'Z'
	} else if isLowercaseAscii(byteToShift) {
		startByte = 'a'
		endByte = 'z'
	} else {
		return byteToShift
	}

	// at a mininum, shift the byte
	newByte := byteToShift + byte(shiftAmount)
	// but if it runs over the edge of the letters mod by startByte and add that to startByte
	if newByte > endByte {
		newByte = startByte + (newByte - endByte - byte(1))
	}
	return newByte
}
