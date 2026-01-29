package cmd

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var minWordLength int
var maxWordLength int
var maxNumberOfWords int
var minNumberOfWords int

// TransposalResult represents a single solution to a transposal puzzle.
type TransposalResult struct {
	Words []string `json:"words"`
}

// String implements fmt.Stringer for TransposalResult.
func (tr TransposalResult) String() string {
	return strings.Join(tr.Words, " ")
}

// PerformTransposalSolve finds transposals of the given ciphertext using the provided dictionary.
func PerformTransposalSolve(
	cipherText string,
	dictionary *TrieNode,
	minWordLen int,
	maxWordLen int,
	minNumWords int,
	maxNumWords int,
) []TransposalResult {
	letterCounts := CreateLetterCountsMap(cipherText)
	solutionChannel := make(chan []string)
	done := make(chan struct{}) // Channel to signal when recursive calls are done

	var results []TransposalResult

	go func() {
		for solution := range solutionChannel {
			if len(solution) < minNumWords || len(solution) > maxNumWords {
				continue
			}

			// check that word lengths are within bounds
			isValid := true
			for _, word := range solution {
				if len(word) < minWordLen || len(word) > maxWordLen {
					isValid = false
					break
				}
			}
			if isValid {
				results = append(results, TransposalResult{Words: solution})
			}
		}
		close(done) // Signal that we're done processing solutions
	}()

	// Start recursive search
	for letter, _ := range letterCounts {
		childIndex := []byte(letter)[0] - ASCII_A
		if dictionary.children[childIndex] != nil {
			RecursiveFindTransposals(
				dictionary,
				dictionary.children[childIndex],
				DecrementLetterCounts(letter, letterCounts),
				make([]string, 0),
				letter,
				solutionChannel,
				minWordLen,
				maxWordLen,
				minNumWords,
				maxNumWords,
			)
		}
	}
	close(solutionChannel) // Close the solutions channel when all recursive calls are initiated

	<-done // Wait for all solutions to be processed

	return results
}

// transposalCmdHandler represents the transposal command.
var transposalCmd = &cobra.Command{
	Use:   "transposal",
	Short: "Finds transposals of the passed-in string",
	Long: `
	  Transposals, or anagrams, can be multiword responses. Use -m or --min-word-length
		to put a lower bound on the length of a given word. The default is 4. Use -w or --max-words
		to put an upper bound on the number of words that will be searched for. The default is 3.
		Lower word lengths or higher numbers of allowed strings will take longer
  `,
	Args: cobra.MinimumNArgs(1),
	Run:  transposalCmdHandler,
}

// transposalCmdHandler handles the cobra command for finding transposals.
func transposalCmdHandler(cmd *cobra.Command, args []string) {
	if dictionaryFile == "" {
		fmt.Println("A dictionary file is required for finding transposals")
		os.Exit(1)
	}

	// Load dictionary for CLI mode
	resultsChan := make(chan string)
	go func() {
		FeedDictionaryPaths(resultsChan, dictionaryFile)
	}()
	rootTrie := ReadDictionaryToTrie(resultsChan)

	fullString := strings.ToUpper(strings.Join(args, ""))
	transposalResults := PerformTransposalSolve(
		fullString,
		rootTrie,
		minWordLength,
		maxWordLength,
		minNumberOfWords,
		maxNumberOfWords,
	)

	var output []interface{}
	for _, res := range transposalResults {
		output = append(output, res)
	}
	outputResponse(output)
}

// RecursiveFindTransposals crawls tries and decrements letterCounts if childTrie is still a valid search path.
// Results are written to the solutions channel.
func RecursiveFindTransposals(
	rootTrie *TrieNode,
	currentTrie *TrieNode,
	letterCounts map[string]int,
	currentWordList []string,
	currentWord string,
	solutions chan []string,
	minWordLen int,
	maxWordLen int,
	minNumWords int,
	maxNumWords int,
) {
	// We have no more letters and we're at a word break
	if len(letterCounts) == 0 && currentTrie.atWordBoundary {
		finalWordList := make([]string, 0, len(currentWordList)+1)
		finalWordList = append(finalWordList, currentWordList...)
		finalWordList = append(finalWordList, currentWord)
		solutions <- finalWordList
		return
	}

	if len(letterCounts) == 0 {
		return
	}

	// This flow ensures that we handle word breaks as well as continuations
	for index, childTrie := range currentTrie.children {
		if childTrie == nil {
			continue
		}
		// Special case for word breaks
		if index == len(currentTrie.children)-1 {
			if currentTrie.atWordBoundary {
				newWordList := make([]string, 0, len(currentWordList)+1)
				newWordList = append(newWordList, currentWordList...)
				newWordList = append(newWordList, currentWord)

				// Only recurse if we haven't exceeded maxNumberOfWords
				if len(newWordList) <= maxNumWords {
					RecursiveFindTransposals(
						rootTrie, rootTrie, letterCounts, newWordList, "", solutions,
						minWordLen, maxWordLen, minNumWords, maxNumWords,
					)
				}
			}
			break
		}

		childLetter := string(index + ASCII_A)
		_, hasCount := letterCounts[childLetter]
		if hasCount {
			RecursiveFindTransposals(
				rootTrie, childTrie, DecrementLetterCounts(childLetter, letterCounts), currentWordList, currentWord+childLetter, solutions,
				minWordLen, maxWordLen, minNumWords, maxNumWords,
			)
		}
	}
}

// DecrementLetterCounts decrements the count of letter in currentCounts (and deletes the key if it's decremented to 0)
// and returns a new letter count map.
func DecrementLetterCounts(letter string, currentCounts map[string]int) map[string]int {
	_, present := currentCounts[letter]
	if !present {
		return currentCounts
	}

	result := make(map[string]int)
	for currentLetter, currentLetterCount := range currentCounts {
		if currentLetter != letter {
			result[currentLetter] = currentLetterCount
			continue
		}

		if currentLetterCount > 1 {
			result[currentLetter] = currentLetterCount - 1
		}
	}
	return result
}

// CreateLetterCountsMap takes in a string and returns a map of letter to count.
func CreateLetterCountsMap(input string) map[string]int {
	counts := make(map[string]int)
	letters := strings.Split(input, "")

	for _, letter := range letters {
		if !lettersRegex.Match([]byte(letter)) {
			continue
		}
		upperLetter := strings.ToUpper(letter)
		if _, present := counts[upperLetter]; !present {
			counts[upperLetter] = 0
		}

		counts[upperLetter] = counts[upperLetter] + 1
	}

	return counts
}

func init() {
	transposalCmd.Flags().StringVarP(&dictionaryFile, "dictionary", "d", "", "Dictionary file to use, or - to use stdin")
	transposalCmd.MarkFlagRequired("dictionary")

	transposalCmd.Flags().IntVarP(&minWordLength, "min-word-length", "", 1, "The minimum length a word in the transposal can be")
	transposalCmd.Flags().IntVarP(&maxWordLength, "max-word-length", "", math.MaxUint32, "The maximum length a word in the transposal can be")
	transposalCmd.Flags().IntVarP(&minNumberOfWords, "min-words", "", 1, "The minimum number of words allowable in a solution")
	transposalCmd.Flags().IntVarP(&maxNumberOfWords, "max-words", "", math.MaxUint32, "The maximum number of words allowable in a solution")
	rootCmd.AddCommand(transposalCmd)
}
