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

// transposalCmd represents the transposal command
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
	Run:  findTransposals,
}

// findTransposals joins the strings passed in args and hunts for any and all transposals
// it can find in the dictionary file that was passed in
func findTransposals(cmd *cobra.Command, args []string) {
	if dictionaryFile == "" {
		fmt.Println("A dictionary file is required for finding transposals")
		os.Exit(1)
	}
	// convert args to one long string. since it's a transposal, we can just smush them together
	fullString := strings.ToUpper(strings.Join(args, ""))
	results := make(chan string)
	go func() {
		feedDictionaryPaths(results, dictionaryFile)
	}()
	rootTrie := readDictionaryToTrie(results)
	letterCounts := createLetterCountsMap(fullString)
	solutions := make(chan []string)
	go parseTransposals(solutions)
	for letter, _ := range letterCounts {
		if childTrie, present := rootTrie.children[letter]; present {
			recursiveFindTransposals(rootTrie, childTrie, decrementLetterCounts(letter, letterCounts), make([]string, 0), letter, solutions)
		}
	}
	close(solutions)
}

// recursiveFindTransposals crawls tries and decrements letterCounts if childTrie is still a valid search path
// results are written to the solutions channel
func recursiveFindTransposals(rootTrie *trieNode, currentTrie *trieNode, letterCounts map[string]int, currentWordList []string, currentWord string, solutions chan []string) {
	// we have no more letters and we're at a word break
	_, atWordBoundary := currentTrie.children[""]
	if len(letterCounts) == 0 && atWordBoundary {
		// make a copy to avoid messing with the slice
		finalWordList := make([]string, 0, len(currentWordList)+1)
		finalWordList = append(finalWordList, currentWordList...)
		finalWordList = append(finalWordList, currentWord)
		solutions <- finalWordList
		return
	}

	if len(letterCounts) == 0 {
		// this means we've run out of letters
		return
	}

	// this flow  ensures that we handle word breaks as well as continuations
	// root could become to or or toro, so we need to handle the word break
	// _and_ other children
	for childLetter, childTrie := range currentTrie.children {
		// special case for "", which marks a word break
		if childLetter == "" {
			newWordList := make([]string, 0, len(currentWordList)+1)
			newWordList = append(newWordList, currentWordList...)
			newWordList = append(newWordList, currentWord)
			recursiveFindTransposals(rootTrie, rootTrie, letterCounts, newWordList, childLetter, solutions)
			continue
		}

		_, hasCount := letterCounts[childLetter]
		if hasCount {
			recursiveFindTransposals(rootTrie, childTrie, decrementLetterCounts(childLetter, letterCounts), currentWordList, currentWord+childLetter, solutions)
		}
	}
}

// decrementLetterCounts decrements the count of letter in currentCounts (and deletes the key if it's decremented to 0)
// and returns a new letter count map
func decrementLetterCounts(letter string, currentCounts map[string]int) map[string]int {
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
			// > 1 because we don't want to write 0s into the map. So if currentLetterCount will be
			// decremented to 0 (i.e., it's 1), then skip it
			result[currentLetter] = currentLetterCount - 1
		}

		// if currentCount is now 0, we don't add the letter, effectively deleting it
	}
	return result
}

// parseTransposals reads off a channel and prints out any results that are in accordance with the arguments specified by the user,
// such as number of words and so forth
func parseTransposals(solutions chan []string) {
ChannelLoop:
	for wordSet := range solutions {
		if len(wordSet) < minNumberOfWords || len(wordSet) > maxNumberOfWords {
			continue
		}

		// check that word lengths are within bounds
		for _, word := range wordSet {
			if len(word) < minWordLength || len(word) > maxWordLength {
				continue ChannelLoop
			}
		}
		fmt.Println(strings.Join(wordSet, " "))
	}
}

// createLetterCountsMap takes in a string and returns a map of letter to count.
// this can then be used when walking the trie to keep track of whether the path
// we're on represents a transposal
func createLetterCountsMap(input string) map[string]int {
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

	transposalCmd.Flags().IntVarP(&minWordLength, "min-word-length", "", 0, "The minimum length a word in the transposal can be")
	transposalCmd.Flags().IntVarP(&maxWordLength, "max-word-length", "", math.MaxUint32, "The maximum length a word in the transposal can be")
	transposalCmd.Flags().IntVarP(&minNumberOfWords, "min-words", "", 0, "The minimum number of words allowable in a solution")
	transposalCmd.Flags().IntVarP(&maxNumberOfWords, "max-words", "", math.MaxUint32, "The maximum number of words allowable in a solution")
	rootCmd.AddCommand(transposalCmd)
}
