package cmd

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// LetterBankResult represents a single solution to a letter bank puzzle.
type LetterBankResult struct {
	Words []string `json:"words"`
}

// String implements fmt.Stringer for LetterBankResult.
func (lr LetterBankResult) String() string {
	return strings.Join(lr.Words, " ")
}

// PerformLetterBankSolve finds letter bank matches for the given text using the provided dictionary.
// A letter bank match is a word or phrase that uses exactly the same set of unique letters as the input.
func PerformLetterBankSolve(
	sourceText string,
	dictionary *TrieNode,
	minWordLen int,
	maxWordLen int,
	minNumWords int,
	maxNumWords int,
) []LetterBankResult {
	letterSet := CreateLetterSet(sourceText)
	solutionChannel := make(chan []string)
	done := make(chan struct{})

	var results []LetterBankResult

	go func() {
		for solution := range solutionChannel {
			if len(solution) < minNumWords || len(solution) > maxNumWords {
				continue
			}

			isValid := true
			for _, word := range solution {
				if len(word) < minWordLen || len(word) > maxWordLen {
					isValid = false
					break
				}
			}
			if isValid {
				results = append(results, LetterBankResult{Words: solution})
			}
		}
		close(done)
	}()

	// Start recursive search from each letter in the set
	for letter := range letterSet {
		childIndex := []byte(letter)[0] - ASCII_A
		if dictionary.children[childIndex] != nil {
			usedLetters := map[string]bool{letter: true}
			RecursiveFindLetterBanks(
				dictionary,
				dictionary.children[childIndex],
				letterSet,
				usedLetters,
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
	close(solutionChannel)

	<-done

	return results
}

// RecursiveFindLetterBanks crawls the trie looking for words/phrases whose combined
// unique letters exactly match the source letter set.
func RecursiveFindLetterBanks(
	rootTrie *TrieNode,
	currentTrie *TrieNode,
	letterSet map[string]bool,
	usedLetters map[string]bool,
	currentWordList []string,
	currentWord string,
	solutions chan []string,
	minWordLen int,
	maxWordLen int,
	minNumWords int,
	maxNumWords int,
) {
	// Check if all letters have been used and we're at a word boundary
	if len(usedLetters) == len(letterSet) && currentTrie.atWordBoundary {
		finalWordList := make([]string, 0, len(currentWordList)+1)
		finalWordList = append(finalWordList, currentWordList...)
		finalWordList = append(finalWordList, currentWord)
		solutions <- finalWordList
		// Don't return -- we can still continue finding longer words using the same letters
	}

	// Traverse children
	for index, childTrie := range currentTrie.children {
		if childTrie == nil {
			continue
		}

		// Special case for word breaks (index 26)
		if index == len(currentTrie.children)-1 {
			if currentTrie.atWordBoundary {
				newWordList := make([]string, 0, len(currentWordList)+1)
				newWordList = append(newWordList, currentWordList...)
				newWordList = append(newWordList, currentWord)

				if len(newWordList) <= maxNumWords {
					RecursiveFindLetterBanks(
						rootTrie, rootTrie, letterSet, usedLetters, newWordList, "", solutions,
						minWordLen, maxWordLen, minNumWords, maxNumWords,
					)
				}
			}
			break
		}

		childLetter := string(index + ASCII_A)
		// Only follow children whose letter is in the source letter set
		if !letterSet[childLetter] {
			continue
		}

		// Track that we've used this letter
		newUsedLetters := copyLetterSet(usedLetters)
		newUsedLetters[childLetter] = true

		RecursiveFindLetterBanks(
			rootTrie, childTrie, letterSet, newUsedLetters, currentWordList, currentWord+childLetter, solutions,
			minWordLen, maxWordLen, minNumWords, maxNumWords,
		)
	}
}

// CreateLetterSet takes a string and returns a set of unique letters (uppercase).
func CreateLetterSet(input string) map[string]bool {
	set := make(map[string]bool)
	for _, ch := range strings.Split(input, "") {
		if !lettersRegex.Match([]byte(ch)) {
			continue
		}
		set[strings.ToUpper(ch)] = true
	}
	return set
}

// copyLetterSet creates a shallow copy of a letter set.
func copyLetterSet(src map[string]bool) map[string]bool {
	dst := make(map[string]bool, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// letterBankCmd represents the letterbank command.
var letterBankCmd = &cobra.Command{
	Use:   "letterbank",
	Short: "Finds letter bank matches for the passed-in string",
	Long: `
	  A letter bank is a word or phrase which shares exactly the same set of unique letters
	  as another word or phrase. For instance, LENDS is a letter bank of NEEDLESS because
	  both use exactly the letters {D, E, L, N, S}. Use --min-word-length, --max-word-length,
	  --min-words, and --max-words to constrain the search.
  `,
	Args: cobra.MinimumNArgs(1),
	Run:  letterBankCmdHandler,
}

func letterBankCmdHandler(cmd *cobra.Command, args []string) {
	if dictionaryFile == "" {
		fmt.Println("A dictionary file is required for finding letter banks")
		os.Exit(1)
	}

	resultsChan := make(chan string)
	go func() {
		FeedDictionaryPaths(resultsChan, dictionaryFile)
	}()
	rootTrie := ReadDictionaryToTrie(resultsChan)

	fullString := strings.ToUpper(strings.Join(args, ""))
	letterBankResults := PerformLetterBankSolve(
		fullString,
		rootTrie,
		lbMinWordLength,
		lbMaxWordLength,
		lbMinNumberOfWords,
		lbMaxNumberOfWords,
	)

	var output []interface{}
	for _, res := range letterBankResults {
		output = append(output, res)
	}
	outputResponse(output)
}

var lbMinWordLength int
var lbMaxWordLength int
var lbMaxNumberOfWords int
var lbMinNumberOfWords int

func init() {
	letterBankCmd.Flags().StringVarP(&dictionaryFile, "dictionary", "d", "", "Dictionary file to use, or - to use stdin")
	letterBankCmd.MarkFlagRequired("dictionary")

	letterBankCmd.Flags().IntVarP(&lbMinWordLength, "min-word-length", "", 1, "The minimum length a word in the result can be")
	letterBankCmd.Flags().IntVarP(&lbMaxWordLength, "max-word-length", "", math.MaxUint32, "The maximum length a word in the result can be")
	letterBankCmd.Flags().IntVarP(&lbMinNumberOfWords, "min-words", "", 1, "The minimum number of words allowable in a solution")
	letterBankCmd.Flags().IntVarP(&lbMaxNumberOfWords, "max-words", "", math.MaxUint32, "The maximum number of words allowable in a solution")
	rootCmd.AddCommand(letterBankCmd)
}
