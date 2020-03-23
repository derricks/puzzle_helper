package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

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

type substitutionWordMatches struct {
	word           string
	cryptPattern   string
	patternMatches []string
}

func (matchData *substitutionWordMatches) addMatch(word string) {
	matchData.patternMatches = append(matchData.patternMatches, word)
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

	resultsChannel := make(chan map[byte]byte)
	go func() {
		for validMap := range resultsChannel {
			printDecodedString(oneString, validMap)
		}
	}()
	partitionMapCollection(matchesData, resultsChannel)
	// ensure the channel has time to be cleared
	time.Sleep(2 * time.Second)
}

// partitionMapCollection splits up matchesData so that the work can
// be partitioned among goroutines that push their results to resultsChannel.
// it returns when waitGroup.Wait() finishes.
func partitionMapCollection(matchData []*substitutionWordMatches, resultsChannel chan map[byte]byte) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func(matches []*substitutionWordMatches, currentMap map[byte]byte) {
		collectValidMaps(matchData, currentMap, resultsChannel)
		waitGroup.Done()
	}(matchData, make(map[byte]byte))
	waitGroup.Wait()
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
func collectValidMaps(matches []*substitutionWordMatches, currentMap map[byte]byte, resultsChannel chan map[byte]byte) {
	if len(matches) == 0 {
		// we've reached the end of the matches to check, which means the currentMap is valid
		resultsChannel <- currentMap
		return
	}

	if len(matches[0].patternMatches) == 0 {
		// no matches were found for this word
		return
	}

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

		// at this point, every byte in the current match doesn't conflict with the existing byte map, so we can gather up the results
		// from the recursive call
		collectValidMaps(matches[1:], copyMap, resultsChannel)
	}
}

// copyByteMap allows the solving code to make a copy of the current byte map so it can pop old copies
// off the stack
func copyByteMap(input map[byte]byte) map[byte]byte {
	output := make(map[byte]byte)
	for key, value := range input {
		output[key] = value
	}
	return output
}

// buildSubstitutionData creates the full data needed to try and solve the substitution.
// When dictionaryFile is parsed, it's no longer needed and can be closed.
func buildSubstitutionData(solveString, dictionaryFile string) []*substitutionWordMatches {
	words := strings.Split(solveString, " ")

	wordMatches := make([]*substitutionWordMatches, 0, len(words))
	for _, curWord := range words {
		wordMatches = append(wordMatches, &substitutionWordMatches{curWord, substitutionPattern(curWord), make([]string, 0, 1)})
	}

	// for each line in the input, see if its pattern equals any pattern in the list
	// if so, add it to the list of matches
	var input *bufio.Reader
	var err error
	if dictionaryFile == "-" {
		input = bufio.NewReader(os.Stdin)
	} else {
		file, err := os.Open(dictionaryFile)
		if err != nil {
			fmt.Printf("Could not access file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		input = bufio.NewReader(file)
	}

	err = findMatchesFromDictionary(wordMatches, input)
	if err != nil {
		fmt.Printf("Error accessing dictionary: %v\n", err)
		os.Exit(1)
	}
	return wordMatches
}

// findMatchesFromDictionary populates each item in substitutionWordMatches with matching
// entries from the passed-in Reader. This mutates the structures that are passed in
func findMatchesFromDictionary(matchSets []*substitutionWordMatches, reader *bufio.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		entry := strings.ToUpper(scanner.Text())
		pattern := substitutionPattern(entry)
		for _, testMatch := range matchSets {
			if testMatch.cryptPattern == pattern {
				testMatch.addMatch(entry)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// substitutionPattern takes in a string and creates the pattern of its letters.
// For instance, substitutionPattern("HELLO") produces "ABCCD"
func substitutionPattern(input string) string {
	returnBytes := make([]byte, 0, len(input))
	textToPattern := make(map[byte]byte)
	maxByte := 65 // capital A ascii

	for _, inputByte := range []byte(input) {
		// if we don't already have a mapping, create one
		if _, exists := textToPattern[inputByte]; !exists {
			textToPattern[inputByte] = byte(maxByte)
			maxByte++
		}
		returnBytes = append(returnBytes, textToPattern[inputByte])
	}
	return string(returnBytes)
}
