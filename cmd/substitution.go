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

// substitutionSolve uses a dictionary file to create possible matches for a substitution string.
// For each word in the string, the function finds a set of cryptographic matches. Then it tries
// combinations of those strings, updating a dictionary as it goes and rejecting possibilities
// where the dictionary conflicts.
func substitutionSolve(cmd *cobra.Command, args []string) {
	// the user could pass in "abcd efg" rather than ABCD EFG, so clean up the data
	oneString := strings.ToUpper(strings.Join(args, " "))
	matchesData := buildSubstitutionData(oneString, dictionaryFile)

	// sort such that items with shorter lists are evaluated first to prune earlier
	sort.Slice(matchesData, func(i, j int) bool {
		return len(matchesData[i].patternMatches) < len(matchesData[j].patternMatches)
	})

	validMaps := collectValidMaps(matchesData, make(map[byte]byte))
	for _, curMap := range validMaps {
		printDecodedString(oneString, curMap)
	}
}

// printDecodedString uses cipherToPlain to decode cipherText
func printDecodedString(cipherText string, cipherToPlain map[byte]byte) {
	for _, cipherChar := range []byte(cipherText) {
		plainChar, mapped := cipherToPlain[cipherChar]
		if !mapped {
			fmt.Printf("%c", cipherChar)
		} else {
			fmt.Printf("%c", plainChar)
		}
	}
	fmt.Print("\n")
}

// collectValidMaps builds a slice of valid byte -> byte mappings that work for all the
// matches it's looked at so far. this method is called  recursively to build the list
func collectValidMaps(matches []*substitutionWordMatches, currentMap map[byte]byte) []map[byte]byte {
	if len(matches) == 0 {
		// we've reached the end of the matches to check, which means the currentMap is valid
		return []map[byte]byte{currentMap}
	}

	if len(matches[0].patternMatches) == 0 {
		// no matches were found for this word
		return make([]map[byte]byte, 0)
	}

	results := make([]map[byte]byte, 0)
	for _, currentMatch := range matches[0].patternMatches {
		copyMap := copyByteMap(currentMap)
		matchBytes := []byte(currentMatch)
		// now edit the map based on the letters in matches[0].word
		// for a given byte in word, check to see the corresponding byte in
		// currentMatch. If that mapping is not in copyMap, add it. If the mapping
		// is in copyMap and is the same, keep going. Finally, if the mapping is in copyMap
		// but maps to a different byte, flag the word as not a match

		allBytesWorked := true
		for index, cryptByte := range []byte(matches[0].word) {
			plainTextByte, exists := copyMap[cryptByte]
			if !exists {
				copyMap[cryptByte] = matchBytes[index]
			} else {
				if plainTextByte != matchBytes[index] {
					allBytesWorked = false
					break
				}
			}
		}

		if !allBytesWorked {
			// move on to the next word; this one didn't work
			continue
		}

	}
}
