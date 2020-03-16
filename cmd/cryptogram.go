/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// cryptogramCmd represents the cryptogram command
var cryptogramCmd = &cobra.Command{
	Use:   "cryptogram",
	Short: "Provides a variety of tools for doing cryptanalyis",
	Long: `The cryptogram subcommand is designed to help with puzzle-level cryptanalysis.

	Examples:
	   puzzles cryptogram freq TEXT: `,
}

var freqCmd = &cobra.Command{
	Use:   "freq",
	Short: "Provides frequency information for the uppercase letters in a string.",
	Long:  `Many common cryptograms require frequency analysis. This command provides single-character frequency as well as digraphs and trigraphs.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   printFrequencyTable,
}

var interactSubstitution = &cobra.Command{
	Use:   "interact-substitution",
	Short: "Creates an interactive session for solving substitution ciphers",
	Args:  cobra.MinimumNArgs(1),
	Run: substitutionShell,
}

func init() {
	cryptogramCmd.AddCommand(freqCmd)
	cryptogramCmd.AddCommand(interactSubstitution)
	rootCmd.AddCommand(cryptogramCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cryptogramCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cryptogramCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

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

// printFrequencyTable generates output about the frequency of characters, digraphs, and trigraphs in a string
func printFrequencyTable(cmd *cobra.Command, args []string) {
	totalString := strings.Join(args, " ")
	singleLetterCounts := frequencyCountInString(totalString)
	totalLetterCount := countTotalCharacters(totalString)
	fmt.Println("Frequency Table")
	fmt.Println("---------------")
	fmt.Printf("Total letters: %v\n", totalLetterCount)
	for curByte, count := range singleLetterCounts {
		fmt.Printf("%c: %v (%v%%)\n", curByte, count, fmt.Sprintf("%.2f", 100.0*(float32(count)/float32(totalLetterCount))))
	}
}

// countTotalCharacters counts the number of uppercase letters in the given string
func countTotalCharacters(toCount string) int {
	var totalCount = 0
	for _, curByte := range []byte(toCount) {
		if isUppercaseAscii(curByte) {
			totalCount += 1
		}
	}
	return totalCount
}

// frequencyCountInString takes in a string and returns the frequency of bytes in it
func frequencyCountInString(toCount string) map[byte]int {
	counts := make(map[byte]int)
	for _, curByte := range []byte(toCount) {
		if !isUppercaseAscii(curByte) {
			continue
		}

		curCount, exists := counts[curByte]
		if exists {
			counts[curByte] = curCount + 1
		} else {
			counts[curByte] = 1
		}
	}
	return counts
}

func isUppercaseAscii(check byte) bool {
	return check >= 65 && check < 91
}
