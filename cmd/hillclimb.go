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
	"io"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var ngramFrequencyFile string
var generations int
var ngramSize int
var mutations int
var regenAfter int
var candidateCount int
var localLookaround int

// hillclimbCmd represents the hillclimb command
var hillclimbCmd = &cobra.Command{
	Use:   "hillclimb",
	Short: "Use hill climbing techniques to find a substitution cipher key that produces text that has the same frequencies as the given tetragrams",
	Long: `Hill climbing works by randomly mutating a substitution cipher key and evaluating the resulting text until its tetragram frequency matches the passed-in file.

	At each pass, the current key is used to decrypt the text. If it scores better than the previous key, it becomes the current key. The current key is mutated again and
	checked against the previous key and so on. You can control the number of runs the code does, though it defaults to 1000. When the command reaches its final run,
	the program will print out the deciphered text using the current key.
  `,
	Run: hillClimbSubstitutionSolve,
}

type substitutionHillclimbCandidate struct {
	fitness float64
	key     map[string]string
}

type substitutionHillclimbCandidates []*substitutionHillclimbCandidate

func (h substitutionHillclimbCandidates) Len() int {
	return len(h)
}

func (h substitutionHillclimbCandidates) Swap(i, j int) {
	temp1 := h[i]
	temp2 := h[j]
	h[i] = temp2
	h[j] = temp1
}

func (h substitutionHillclimbCandidates) Less(i, j int) bool {
	left := h[i]
	right := h[j]
	return left.fitness > right.fitness
}

func newHillclimbCandidate(key map[string]string, ciphertext string, frequencyMap map[string]float64) *substitutionHillclimbCandidate {
	plainText := decipherStringFromKey(ciphertext, key)
	fitness := calculateNgramFitness(plainText, frequencyMap)
	return &substitutionHillclimbCandidate{fitness, key}
}

func hillClimbSubstitutionSolve(cmd *cobra.Command, args []string) {

	candidates := substitutionHillclimbCandidates(make([]*substitutionHillclimbCandidate, 0, candidateCount))

	rawInputText := strings.Join(args, " ")
	justLetters := make([]string, 0, len(rawInputText))
	letterScanner := NewNgramScanner(strings.NewReader(rawInputText), 1, false)
	for letterScanner.Scan() {
		justLetters = append(justLetters, letterScanner.Text())
	}

	var inReader io.Reader
	var err error
	if ngramFrequencyFile == "-" {
		inReader = os.Stdin
	} else {
		inReader, err = os.Open(ngramFrequencyFile)
		if err != nil {
			fmt.Printf("Error with tetragram file: %v", err)
			os.Exit(1)
		}
	}

	frequencyMap := populateFrequencyMapFromReader(inReader)

	justCipherText := strings.Join(justLetters, "")

	currentCandidate := newHillclimbCandidate(generateRandomKey(), justCipherText, frequencyMap)
	bestOfGeneration := currentCandidate
	candidates = append(candidates, bestOfGeneration)

	fitnessGenerations := 1
	currentGeneration := 1
	for currentGeneration <= generations {
		if currentCandidate.fitness > bestOfGeneration.fitness {
			bestOfGeneration = currentCandidate
			fitnessGenerations = 0

			if len(candidates) < candidateCount {
				candidates = append(candidates, bestOfGeneration)
				sort.Sort(candidates)
			} else {
				currentLastPlace := candidates[len(candidates)-1]
				if bestOfGeneration.fitness > currentLastPlace.fitness {
					candidates[len(candidates)-1] = bestOfGeneration
					sort.Sort(candidates)
				}
			}

		} else {
			fitnessGenerations++
		}

		// we've gone too long without finding a better fitness
		if fitnessGenerations > regenAfter {
			bestOfGeneration = newHillclimbCandidate(generateRandomKey(), justCipherText, frequencyMap)
			currentCandidate = bestOfGeneration
			fitnessGenerations = 0
			currentGeneration++
			continue
		}

		// look around and choose the best of a random set of nearby paths
		bestNewCandidate := currentCandidate
		for localIndex := 0; localIndex < localLookaround; localIndex++ {

			checkCandidate := newHillclimbCandidate(mutateKeyNTimes(mutations, currentCandidate.key), justCipherText, frequencyMap)
			if checkCandidate.fitness > bestNewCandidate.fitness {
				bestNewCandidate = checkCandidate
			}
		}
		currentCandidate = bestNewCandidate
	}

	for _, candidate := range candidates {
		fmt.Printf("%v\n%v\n%s\n\n", candidate.fitness, candidate.key, decipherStringFromKey(strings.ToUpper(rawInputText), candidate.key))
	}

}

// this unchanging slice just gives an O(1) way to look up a letter randomly in a key
var randomKeyPool = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

func mutateKeyNTimes(n int, key map[string]string) map[string]string {
	newKey := make(map[string]string)
	for cipher, plain := range key {
		newKey[cipher] = plain
	}

	for i := 0; i < n; i++ {
		key1 := randomKeyPool[rand.Intn(len(randomKeyPool))]
		key2 := randomKeyPool[rand.Intn(len(randomKeyPool))]
		value1 := newKey[key1]
		value2 := newKey[key2]
		newKey[key2] = value1
		newKey[key1] = value2
	}
	return newKey
}

// calculateNgramFitness takes in a deciphered string and calculates its fitness based on trie that maps ngrams to frequency
func calculateNgramFitness(deciphered string, frequencyMap map[string]float64) float64 {
	var fitness float64
	scanner := NewNgramScanner(strings.NewReader(deciphered), ngramSize, true)
	for scanner.Scan() {
		log10probability, isPresent := frequencyMap[scanner.Text()]
		if isPresent {
			fitness += log10probability
		} else {
			fitness += -1000
		}
	}
	return fitness
}

func populateFrequencyMapFromReader(reader io.Reader) map[string]float64 {
	result := make(map[string]float64)
	now := time.Now().UnixNano()
	for scanner := bufio.NewScanner(reader); scanner.Scan(); {
		line := scanner.Text()
		fields := strings.Split(line, "\t")
		ngramSize = len(fields[0])
		frequency, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			fmt.Printf("Invalid float in frequency file: %s\n", fields[1])
			os.Exit(1)
		}
		result[fields[0]] = frequency
	}
	if profile {
		fmt.Printf("Reading into trie took: %.8fms\n", float64(time.Now().UnixNano()-now)/float64(1000000))
	}
	return result
}

// decipherStringFromKey decrypts cipherText with cipherToPlain
func decipherStringFromKey(cipherText string, cipherToPlain map[string]string) string {
	plainText := strings.Builder{}
	cipherLetters := strings.Split(cipherText, "")
	for _, currentCipherLetter := range cipherLetters {
		plainLetter, isPresent := cipherToPlain[currentCipherLetter]
		if isPresent {
			plainText.WriteString(plainLetter)
		} else {
			plainText.WriteString(currentCipherLetter)
		}
	}
	return plainText.String()
}

func generateRandomKey() map[string]string {
	key := make(map[string]string)

	cipherLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	plainLetters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	rand.Shuffle(len(cipherLetters), func(i, j int) { cipherLetters[i], cipherLetters[j] = cipherLetters[j], cipherLetters[i] })
	rand.Shuffle(len(plainLetters), func(i, j int) { plainLetters[i], plainLetters[j] = plainLetters[j], plainLetters[i] })

	for cipherIndex, letter := range cipherLetters {
		key[letter] = plainLetters[cipherIndex]
	}
	return key
}

func init() {
	hillclimbCmd.Flags().StringVarP(&ngramFrequencyFile, "frequency-file", "f", "", "the path to the frequency file to use. Use - for stdin. The chunking of the input text will use the same ngram size from the first line of the file, and the file is assumed to be ngram tab log10 of frequency")
	hillclimbCmd.MarkFlagRequired("frequency-file")
	hillclimbCmd.Flags().IntVarP(&generations, "generations", "g", 50, "the number of generations to run for - generations happen based on the regen-after setting")
	hillclimbCmd.Flags().IntVarP(&mutations, "mutations", "m", 1, "the number of mutations to do on the key during each iteration")
	hillclimbCmd.Flags().IntVarP(&regenAfter, "regen-after", "r", 1000, "how long a fitness can survive before the program starts with a new random key")
	hillclimbCmd.Flags().IntVarP(&candidateCount, "candidates", "c", 10, "the number of top performing candidates to display")
	hillclimbCmd.Flags().IntVarP(&localLookaround, "local-lookaround", "l", 1, "when picking a new path, evaluate this many local candidates and choose the best of them")
	substitutionCmd.AddCommand(hillclimbCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hillclimbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hillclimbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
