package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// Implementations for the substitution command object
var substitutionCommand = regexp.MustCompile("[A-Z]=[a-z]")

// substitutionShell creates a loop which lets you interactively solve a substitution cipher.
// It will prompt for commands and show the current state of cipher text and plain text.
// Command reference:
//   A=z will replace A in ciphertext with a z in plaintext
func substitutionShell(cmd *cobra.Command, args []string) {
	cipherString := strings.Join(args, " ")
	cipherToPlain := make(map[byte]byte)

	reader := bufio.NewReader(os.Stdin)

	for {
		plainString := ""
		for _, cipherByte := range []byte(cipherString) {
	     if isUppercaseAscii(cipherByte) {
				 plainByte, solved := cipherToPlain[cipherByte]
				 if solved {
					 plainString += string(plainByte)
				 } else {
					 plainString += "_"
				 }
			 } else {
				 plainString += string(cipherByte)
			 }
		}

		fmt.Println(cipherString)
		fmt.Println(plainString)

    fmt.Print("? ")
		command, _ := reader.ReadString('\n')
		commandAsBytes := []byte(command)

		if substitutionCommand.Match(commandAsBytes) {
			  // 0 will be cipher character, 1 will be = and 2 will be plaintext
        cipherToPlain[commandAsBytes[0]] = commandAsBytes[2]
				continue
		}
	}

}
