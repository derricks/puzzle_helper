package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

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

// SubstitutionResult represents a found key and the deciphered text.
type SubstitutionResult struct {
	Key          map[byte]byte
	DecipheredText string
	CipherText     string // Keep original ciphertext for display
}

// String implements fmt.Stringer for SubstitutionResult.
func (sr SubstitutionResult) String() string {
	var builder strings.Builder

	// Display key (cipher to plain)
	builder.WriteString("Cipher: ")
	for curByte := byte('A'); curByte <= byte('Z'); curByte++ {
		builder.WriteByte(curByte)
	}
	builder.WriteString("\nPlain:  ")
	for curByte := byte('A'); curByte <= byte('Z'); curByte++ {
		plainChar, mapped := sr.Key[curByte]
		if mapped {
			builder.WriteByte(plainChar)
		} else {
			builder.WriteByte('_')
		}
	}
	builder.WriteString("\n\n")

	// Display deciphered text
	builder.WriteString("Deciphered:\n")
	builder.WriteString(sr.CipherText)
	builder.WriteString("\n")
	builder.WriteString(sr.DecipheredText)
	builder.WriteString("\n")

	return builder.String()
}

// PerformSubstitutionSolve takes a ciphertext, dictionary, and ngram frequency map
// to find possible substitution cipher solutions.
func PerformSubstitutionSolve(
	cipherText string,
	dictionary *TrieNode,
	ngramFrequencyMap map[string]float64,
	ngramSize int,
	concurrency int,
) []SubstitutionResult {
	oneString := strings.ToUpper(cipherText)
	matchesData := BuildSubstitutionData(oneString, dictionary) // Use dictionary instead of file

	// sort such that items with shorter lists are evaluated first to prune earlier
	sort.Slice(matchesData, func(i, j int) bool {
		return len(matchesData[i].PatternMatches) < len(matchesData[j].PatternMatches)
	})

	resultsChannel := make(chan map[byte]byte)
	solutions := make([]SubstitutionResult, 0)
	var wg sync.WaitGroup

	// Collector goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for validMap := range resultsChannel {
			deciphered := DecodeString(oneString, validMap)
			solutions = append(solutions, SubstitutionResult{
				Key:            validMap,
				DecipheredText: deciphered,
				CipherText:     oneString,
			})
		}
	}()

	// Partition and solve
	PartitionMapCollection(matchesData, resultsChannel, concurrency)
	close(resultsChannel)
	wg.Wait()

	// You might want to filter/sort solutions based on fitness if a frequency map is used
	// For simplicity, this example just returns all found solutions.
	// If the ngramFrequencyMap is passed, you could calculate fitness here and sort.

	return solutions
}

// substitutionShell creates a loop which lets you interactively solve a substitution cipher.
// It will prompt for commands and show the current state of cipher text and plain text.
// Command reference:
//
//	A=z will replace A in ciphertext with a z in plaintext
//	cipher2Plain will list the cipher key in alphabetical order with the plain key underneath
//	plain2Cipher will list the plain key in alphabetical order with the cipher key underneath
//	clear will remove any mappings
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
			if IsUppercaseAscii(cipherByte) { // Use exported helper
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

// DecodeString decodes a ciphertext using the provided key.
func DecodeString(cipherText string, cipherToPlain map[byte]byte) string {
	var builder strings.Builder
	for _, cipherChar := range []byte(cipherText) {
		if IsUppercaseAscii(cipherChar) {
			plainChar, mapped := cipherToPlain[cipherChar]
			if mapped {
				builder.WriteByte(plainChar)
			} else {
				builder.WriteByte('_')
			}
		} else {
			builder.WriteByte(cipherChar)
		}
	}
	return builder.String()
}

func writeLines(writer *bufio.Writer, lines ...string) {
	for _, line := range lines {
		writer.Write([]byte(line))
		writer.Write([]byte{'\n'})
	}
	writer.Flush()
}

type SubstitutionWordMatches struct {
	Word           string
	CryptPattern   string
	PatternMatches []string
}

func (matchData *SubstitutionWordMatches) AddMatch(word string) {
	matchData.PatternMatches = append(matchData.PatternMatches, word)
}

// substitutionSolve uses a dictionary file to create possible matches for a substitution string.
// For each word in the string, the function finds a set of cryptographic matches. Then it tries
// combinations of those strings, updating a dictionary as it goes and rejecting possibilities
// where the dictionary conflicts.
var substitutionCmd = &cobra.Command{
	Use:   "substitution",
	Short: "Solve substitution ciphers",
	Long:  `This command solves substitution ciphers.`,
	Args:  cobra.MinimumNArgs(1),
	Run:   substitutionCmdHandler,
}

var concurrency int
var substitutionDictionaryFile string
var substitutionNgramFrequencyFile string
// substitutionNgramSize is now determined from the frequency file and passed as a parameter.

func substitutionCmdHandler(cmd *cobra.Command, args []string) {
	oneString := strings.ToUpper(strings.Join(args, " "))

	if substitutionDictionaryFile == "" {
		fmt.Println("A dictionary file is required for substitution solving")
		os.Exit(1)
	}
	if substitutionNgramFrequencyFile == "" {
		fmt.Println("An ngram frequency file is required for substitution solving")
		os.Exit(1)
	}

	// Load dictionary
	dictChannel := make(chan string)
	go func() {
		FeedDictionaryPaths(dictChannel, substitutionDictionaryFile)
	}()
	dictionary := ReadDictionaryToTrie(dictChannel)

	// Load ngram frequency map
	ngramReader, err := os.Open(substitutionNgramFrequencyFile)
	if err != nil {
		fmt.Printf("Error with ngram frequency file: %v\n", err)
		os.Exit(1)
	}
	defer ngramReader.Close()
	ngramFrequencyMap, detectedNgramSize := PopulateFrequencyMapFromReader(ngramReader)

	solutions := PerformSubstitutionSolve(
		oneString,
		dictionary,
		ngramFrequencyMap,
		detectedNgramSize, // Pass the detected ngram size
		concurrency,
	)

	var output []interface{}
	for _, res := range solutions {
		output = append(output, res)
	}
	outputResponse(output)
}

// PartitionMapCollection splits up matchesData so that the work can
// be partitioned among goroutines that push their results to resultsChannel.
// It returns when waitGroup.Wait() finishes.
func PartitionMapCollection(matchData []*SubstitutionWordMatches, resultsChannel chan map[byte]byte, concurrency int) {
	// build partitioned slices of SubstitutionWordMatches objects off of the first one
	// in the list. The matches in the head of the group will be split up to create
	// an object with a smaller set of matches that can be put into a goroutine
	// in other words:
	//    if the word matches for item 0 are HELLO and BOSSY, you'll end up with one slice
	//      where the first item only has HELLO as a match and another slice where the first
	//      item only has BOSSY as a match. Then the solution trees are independent
	partitionCount := concurrency

	if len(matchData[0].PatternMatches) < concurrency {
		partitionCount = len(matchData[0].PatternMatches)
	}

	var waitGroup sync.WaitGroup
	heads := PartitionMatches(partitionCount, matchData[0])
	for _, curHead := range heads {

		// build up the new set of SubstitutionWordMatches with this head (and its map of partitions)
		// as the first one
		newMatchData := make([]*SubstitutionWordMatches, 0, len(matchData))
		newMatchData = append(newMatchData, curHead)
		newMatchData = append(newMatchData, matchData[1:]...)

		waitGroup.Add(1)
		go func(matches []*SubstitutionWordMatches, currentMap map[byte]byte) {
			CollectValidMaps(matches, currentMap, resultsChannel)
			waitGroup.Done()
		}(newMatchData, make(map[byte]byte))
	}
	waitGroup.Wait()
}

// PartitionMatches creates count new *SubstitutionWordMatches where each
// contains a subset of the source's PatternMatches object. This allows solve efforts
// to be parallelized
func PartitionMatches(count int, source *SubstitutionWordMatches) []*SubstitutionWordMatches {
	partitions := make([]*SubstitutionWordMatches, 0, count)

	// create the initial set, with no pattern matches
	for index := 0; index < count; index++ {
		partitions = append(partitions, &SubstitutionWordMatches{source.Word, source.CryptPattern, make([]string, 0, 1)})
	}

	// now populate the PatternMatches for a given return struct by modding the index
	for index, curMatch := range source.PatternMatches {
		curPartition := partitions[index%count]
		curPartition.PatternMatches = append(curPartition.PatternMatches, curMatch)
	}
	return partitions
}

// CollectValidMaps builds a slice of valid byte -> byte mappings that work for all the
// matches it's looked at so far. This method is called recursively to build the list
func CollectValidMaps(matches []*SubstitutionWordMatches, currentMap map[byte]byte, resultsChannel chan map[byte]byte) {
	if len(matches) == 0 {
		// we've reached the end of the matches to check, which means the currentMap is valid
		resultsChannel <- currentMap
		return
	}

	if len(matches[0].PatternMatches) == 0 {
		// no matches were found for this word
		return
	}

	for _, currentMatch := range matches[0].PatternMatches {
		copyMap := CopyByteMap(currentMap)
		matchBytes := []byte(currentMatch)
		// now edit the map based on the letters in matches[0].Word
		// for a given byte in word, check to see the corresponding byte in
		// currentMatch. If that mapping is not in copyMap, add it. If the mapping
		// is in copyMap and is the same, keep going. Finally, if the mapping is in copyMap
		// but maps to a different byte, flag the word as not a match

		allBytesWorked := true
		for index, cryptByte := range []byte(matches[0].Word) {
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
		CollectValidMaps(matches[1:], copyMap, resultsChannel)
	}
}

// CopyByteMap allows the solving code to make a copy of the current byte map so it can pop old copies
// off the stack
func CopyByteMap(input map[byte]byte) map[byte]byte {
	output := make(map[byte]byte)
	for key, value := range input {
		output[key] = value
	}
	return output
}

// BuildSubstitutionData creates the full data needed to try and solve the substitution.
func BuildSubstitutionData(solveString string, dictionary *TrieNode) []*SubstitutionWordMatches {
	words := strings.Split(solveString, " ")

	wordMatches := make([]*SubstitutionWordMatches, 0, len(words))
	for _, curWord := range words {
		match := &SubstitutionWordMatches{curWord, SubstitutionPattern(curWord), make([]string, 0, 1)}
		// Instead of reading from a file, directly query the trie
		FindMatchesInTrie(match, dictionary)
		wordMatches = append(wordMatches, match)
	}
	return wordMatches
}

// FindMatchesInTrie populates matchData with matching entries from the trie.
func FindMatchesInTrie(matchData *SubstitutionWordMatches, dictionary *TrieNode) {
	// This function needs to iterate through the dictionary and find words that
	// match the pattern of matchData.Word. This is a placeholder; actual implementation
	// would involve traversing the trie and checking patterns.
	// For now, it's simplified. A full implementation would be more complex.
	// As a workaround, we'll iterate through all words in the dictionary (which can be slow for large dicts)
	// and check their patterns.
	tempChannel := make(chan TrieWord)
	go dictionary.FeedWordsToChannel(tempChannel) // Assuming FeedWordsToChannel is exported

	for entry := range tempChannel {
		if SubstitutionPattern(entry.Word) == matchData.CryptPattern { // Use entry.Word here
			matchData.AddMatch(entry.Word)
		}
	}
}

// SubstitutionPattern takes in a string and creates the pattern of its letters.
// For instance, SubstitutionPattern("HELLO") produces "ABCCD"
func SubstitutionPattern(input string) string {
	returnBytes := make([]byte, 0, len(input))
	textToPattern := make(map[byte]byte)
	maxByte := byte('A') // capital A ascii

	for _, inputByte := range []byte(input) {
		// if we don't already have a mapping, create one
		if _, exists := textToPattern[inputByte]; !exists {
			textToPattern[inputByte] = maxByte
			maxByte++
		}
		returnBytes = append(returnBytes, textToPattern[inputByte])
	}
	return string(returnBytes)
}

func init() {
	substitutionCmd.Flags().StringVarP(&substitutionDictionaryFile, "dictionary", "d", "", "Dictionary file to use, or - to use stdin")
	substitutionCmd.MarkFlagRequired("dictionary")

	substitutionCmd.Flags().StringVarP(&substitutionNgramFrequencyFile, "ngram-frequency-file", "n", "", "Ngram frequency file to use (e.g., tetragrams.txt)")
	substitutionCmd.MarkFlagRequired("ngram-frequency-file")

	// substitutionNgramSize is now determined from the frequency file, so no need for a flag.

	substitutionCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 4, "Number of concurrent goroutines to use for solving")
	rootCmd.AddCommand(substitutionCmd)
}
