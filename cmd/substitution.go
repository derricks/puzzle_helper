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
var substitutionCommand = regexp.MustCompile("[A-Z]=[a-z_]")

type KeyDisplay int

const (
	cipher2Plain KeyDisplay = iota
	plain2Cipher
)

const (
	cipher2PlainCommand string = "cipher2Plain"
	plain2CipherCommand string = "plain2Cipher"
	clearCommand        string = "clear"
)

// substitutionShell creates a loop which lets you interactively solve a substitution cipher.
// It will prompt for commands and show the current state of cipher text and plain text.
// Command reference:
//   A=z will replace A in ciphertext with a z in plaintext
//   cipher2Plain will list the cipher key in alphabetical order with the plain key underneath
//   plain2Cipher will list the plain key in alphabetical order with the cipher key underneath
//   clear will remove any mappings
func substitutionShell(cmd *cobra.Command, args []string) {
	// whether to overwrite the text on the screen (will usually be true)
	// or just push lines onto the screen
	overwrite := false
	outWriter := bufio.NewWriter(os.Stdout)

	cipherString := strings.Join(args, " ")
	displayType := cipher2Plain

	cipherToPlain := make(map[byte]byte)
	plainToCipher := make(map[byte]byte)

	reader := bufio.NewReader(os.Stdin)

	for {
		if overwrite {
			outWriter.Write([]byte("\u001b[6A"))
			outWriter.Write([]byte("\u001b[100D"))
		} else {
			outWriter.Write([]byte{'\n'})
		}

		cipherKeyBytes := make([]byte, 0, 26)
		plainKeyBytes := make([]byte, 0, 26)
		switch displayType {
		case cipher2Plain:
			for curByte := byte('A'); curByte <= byte('Z'); curByte = byte(curByte + 1) {
				cipherKeyBytes = append(cipherKeyBytes, curByte)
				plainChar, mapped := cipherToPlain[curByte]
				if mapped {
					plainKeyBytes = append(plainKeyBytes, plainChar)
				} else {
					plainKeyBytes = append(plainKeyBytes, '_')
				}
			}
			writeLines(outWriter, string(cipherKeyBytes), string(plainKeyBytes))
		case plain2Cipher:
			for curByte := byte('a'); curByte <= byte('z'); curByte = byte(curByte + 1) {
				plainKeyBytes = append(plainKeyBytes, curByte)
				cipherChar, mapped := plainToCipher[curByte]
				if mapped {
					cipherKeyBytes = append(cipherKeyBytes, cipherChar)
				} else {
					cipherKeyBytes = append(cipherKeyBytes, '?')
				}
			}
			writeLines(outWriter, string(plainKeyBytes), string(cipherKeyBytes))
		default:
			// shouldn't get here
			writeLines(outWriter)
			fmt.Printf("Unknown display type: %v\n", displayType)
		}
		writeLines(outWriter, "")

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

		writeLines(outWriter, cipherString, plainString)

		outWriter.Write([]byte("? "))
		outWriter.Write([]byte("\u001b[0K"))
		outWriter.Flush()
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)
		commandAsBytes := []byte(command)
		if substitutionCommand.Match(commandAsBytes) {
			// 0 will be cipher character, 1 will be = and 2 will be plaintext
			if commandAsBytes[2] == '_' {
				// deleting from the plainToCipher map means figuring out the current cipherToPlain mapping
				delete(plainToCipher, cipherToPlain[commandAsBytes[0]])
				delete(cipherToPlain, commandAsBytes[0])
			} else {
				cipherToPlain[commandAsBytes[0]] = commandAsBytes[2]
				plainToCipher[commandAsBytes[2]] = commandAsBytes[0]
			}
		} else if command == cipher2PlainCommand {
			displayType = cipher2Plain
		} else if command == plain2CipherCommand {
			displayType = plain2Cipher
		} else if command == clearCommand {
			cipherToPlain = make(map[byte]byte)
			plainToCipher = make(map[byte]byte)
		}

		overwrite = true
	}
}

func writeLines(writer *bufio.Writer, lines ...string) {
	for _, line := range lines {
		writer.Write([]byte(line))
		writer.Write([]byte{'\n'})
	}
	writer.Flush()
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

	// build partitioned slices of substitutionWordMatches objects off of the first one
	// in the list. The matches in the head of the group will be split up to create
	// an object with a smaller set of matches that can be put into a goroutine
	// in other words:
	//    if the word matches for item 0 are HELLO and BOSSY, you'll end up with one slice
	//      where the first item only has HELLO as a match and another slice where the first
	//      item only has BOSSY as a match. Then the solution trees are independent
	partitionCount := concurrency

	if len(matchData[0].patternMatches) < concurrency {
		partitionCount = len(matchData[0].patternMatches)
	}

	var waitGroup sync.WaitGroup
	heads := partitionMatches(partitionCount, matchData[0])
	for _, curHead := range heads {

		// build up the new set of substitutionWordMatches with this head (and its map of partitions)
		// as the first one
		newMatchData := make([]*substitutionWordMatches, 0, len(matchData))
		newMatchData = append(newMatchData, curHead)
		newMatchData = append(newMatchData, matchData[1:]...)

		waitGroup.Add(1)
		go func(matches []*substitutionWordMatches, currentMap map[byte]byte) {
			collectValidMaps(matchData, currentMap, resultsChannel)
			waitGroup.Done()
		}(newMatchData, make(map[byte]byte))
	}
	waitGroup.Wait()
}

// partitionMatches creates count new *substitutionWordMatches where each
// contains a subset of the source's patternMatches object. This allows solve efforts
// to be parallelized
func partitionMatches(count int, source *substitutionWordMatches) []*substitutionWordMatches {
	partitions := make([]*substitutionWordMatches, 0, count)

	// create the initial set, with no pattern matches
	for index := 0; index < count; index++ {
		partitions = append(partitions, &substitutionWordMatches{source.word, source.cryptPattern, make([]string, 0, 1)})
	}

	// now populate the patternMatches for a given return struct by modding the index
	for index, curMatch := range source.patternMatches {
		curPartition := partitions[index%count]
		curPartition.patternMatches = append(curPartition.patternMatches, curMatch)
	}
	return partitions
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

	results := make(chan string)
	go func() {
		feedDictionaryPaths(results, dictionaryFile)
	}()

	findMatchesFromDictionary(wordMatches, results)
	return wordMatches
}

// findMatchesFromDictionary populates each item in substitutionWordMatches with matching
// entries from the passed-in Reader. This mutates the structures that are passed in
func findMatchesFromDictionary(matchSets []*substitutionWordMatches, feed chan string) {
	for entry := range feed {
		pattern := substitutionPattern(entry)
		for _, testMatch := range matchSets {
			if testMatch.cryptPattern == pattern {
				testMatch.addMatch(entry)
			}
		}
	}
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
