package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var ngramFrequencyFile string
var generations int
var ngramSize int
var mutations int
var regenAfter int
var candidateCount int
var localLookaround int

// SubstitutionHillclimbCandidate represents a single candidate solution with its fitness and key.
type SubstitutionHillclimbCandidate struct {
	Fitness float64  `json:"fitness"`
	Key     []string `json:"key"`
}

// HillclimbResult represents a final result from the hillclimbing process.
type HillclimbResult struct {
	Candidate      *SubstitutionHillclimbCandidate `json:"candidate"`
	DecipheredText string                          `json:"decipheredText"`
}

// String implements fmt.Stringer for HillclimbResult.
func (hcr HillclimbResult) String() string {
	var builder strings.Builder
	builder.WriteString(hcr.Candidate.String())
	builder.WriteString(fmt.Sprintf("Deciphered Text:\n%s\n", hcr.DecipheredText))
	return builder.String()
}

// String implements fmt.Stringer for SubstitutionHillclimbCandidate.
func (c *SubstitutionHillclimbCandidate) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Fitness: %.8f\n", c.Fitness))
	builder.WriteString("Key:    ")
	builder.WriteString(strings.Join(lettersInOrder, " "))
	builder.WriteString("\nPlain:  ")
	builder.WriteString(strings.Join(c.Key, " "))
	builder.WriteString("\n")
	return builder.String()
}

// SubstitutionHillclimbCandidates is a slice of SubstitutionHillclimbCandidate for sorting.
type SubstitutionHillclimbCandidates []*SubstitutionHillclimbCandidate

func (h SubstitutionHillclimbCandidates) Len() int {
	return len(h)
}

func (h SubstitutionHillclimbCandidates) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h SubstitutionHillclimbCandidates) Less(i, j int) bool {
	return h[i].Fitness > h[j].Fitness
}

// NewHillclimbCandidate creates a new substitutionHillclimbCandidate.
func NewHillclimbCandidate(key []string, ciphertext string, frequencyMap map[string]float64, ngramSize int) *SubstitutionHillclimbCandidate {
	plainText := DecipherStringFromKey(ciphertext, key)
	fitness := CalculateNgramFitness(plainText, frequencyMap, ngramSize)
	return &SubstitutionHillclimbCandidate{fitness, key}
}

// PerformHillclimbSolve runs the hillclimbing algorithm to find substitution cipher keys.
func PerformHillclimbSolve(
	cipherText string,
	frequencyMap map[string]float64,
	generations int,
	mutations int,
	regenAfter int,
	candidateCount int,
	localLookaround int,
	ngramSize int,
) []HillclimbResult {
	candidates := SubstitutionHillclimbCandidates(make([]*SubstitutionHillclimbCandidate, 0, candidateCount))

	justLetters := make([]string, 0, len(cipherText))
	letterScanner := NewNgramScanner(strings.NewReader(cipherText), 1, false)
	for letterScanner.Scan() {
		justLetters = append(justLetters, letterScanner.Text())
	}
	justCipherText := strings.Join(justLetters, "")

	currentCandidate := NewHillclimbCandidate(GenerateRandomKey(), justCipherText, frequencyMap, ngramSize)
	bestOfGeneration := currentCandidate
	candidates = append(candidates, bestOfGeneration)

	fitnessGenerations := 1
	currentGeneration := 1
	for currentGeneration <= generations {
		if currentCandidate.Fitness > bestOfGeneration.Fitness {
			bestOfGeneration = currentCandidate
			fitnessGenerations = 0

			if len(candidates) < candidateCount {
				candidates = append(candidates, bestOfGeneration)
				sort.Sort(candidates)
			} else {
				currentLastPlace := candidates[len(candidates)-1]
				if bestOfGeneration.Fitness > currentLastPlace.Fitness {
					candidates[len(candidates)-1] = bestOfGeneration
					sort.Sort(candidates)
				}
			}

		} else {
			fitnessGenerations++
		}

		// we've gone too long without finding a better fitness
		if fitnessGenerations > regenAfter {
			bestOfGeneration = NewHillclimbCandidate(GenerateRandomKey(), justCipherText, frequencyMap, ngramSize)
			currentCandidate = bestOfGeneration
			fitnessGenerations = 0
			currentGeneration++
			continue
		}

		// look around and choose the best of a random set of nearby paths
		bestNewCandidate := currentCandidate
		for localIndex := 0; localIndex < localLookaround; localIndex++ {

			checkCandidate := NewHillclimbCandidate(MutateKeyNTimes(mutations, currentCandidate.Key), justCipherText, frequencyMap, ngramSize)
			if checkCandidate.Fitness > bestNewCandidate.Fitness {
				bestNewCandidate = checkCandidate
			}
		}
		currentCandidate = bestNewCandidate
	}

	var results []HillclimbResult
	for _, candidate := range candidates {
		results = append(results, HillclimbResult{
			Candidate:      candidate,
			DecipheredText: DecipherStringFromKey(strings.ToUpper(cipherText), candidate.Key),
		})
	}
	return results
}

// hillclimbCmd represents the hillclimb command
var hillclimbCmd = &cobra.Command{
	Use:   "hillclimb",
	Short: "Use hill climbing techniques to find a substitution cipher key that produces text that has the same frequencies as the given tetragrams",
	Long: `Hill climbing works by randomly mutating a substitution cipher key and evaluating the resulting text until its tetragram frequency matches the passed-in file.

	At each pass, the current key is used to decrypt the text. If it scores better than the previous key, it becomes the current key. The current key is mutated again and
	checked against the previous key and so on. You can control the number of runs the code does, though it defaults to 1000. When the command reaches its final run,
	the program will print out the deciphered text using the current key.
  `,
	Run: hillclimbCmdHandler,
}

// hillclimbCmdHandler handles the cobra command for hillclimb substitution solving.
func hillclimbCmdHandler(cmd *cobra.Command, args []string) {
	rawInputText := strings.Join(args, " ")
	if ngramFrequencyFile == "" {
		fmt.Println("An ngram frequency file is required for hillclimb solving")
		os.Exit(1)
	}

	ngramReader, err := os.Open(ngramFrequencyFile)
	if err != nil {
		fmt.Printf("Error with ngram frequency file: %v\n", err)
		os.Exit(1)
	}
	defer ngramReader.Close()
	frequencyMap, ngramSize := PopulateFrequencyMapFromReader(ngramReader) // Use exported function

	results := PerformHillclimbSolve(
		rawInputText,
		frequencyMap,
		generations,
		mutations,
		regenAfter,
		candidateCount,
		localLookaround,
		ngramSize,
	)

	var output []interface{}
	for _, res := range results {
		output = append(output, res)
	}
	outputResponse(output)
}

var lettersInOrder = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

// MutateKeyNTimes mutates the given key n times by swapping random letters.
func MutateKeyNTimes(n int, plainLetters []string) []string {
	// make a copy
	newKey := make([]string, len(plainLetters), len(plainLetters))
	for index, letter := range plainLetters {
		newKey[index] = letter
	}

	for i := 0; i < n; i++ {
		swap1 := rand.Intn(len(newKey))
		swap2 := rand.Intn(len(newKey))
		newKey[swap1], newKey[swap2] = newKey[swap2], newKey[swap1]
	}
	return newKey
}

// CalculateNgramFitness takes in a deciphered string and calculates its fitness based on a frequency map.
func CalculateNgramFitness(deciphered string, frequencyMap map[string]float64, ngramSize int) float64 {
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

// DecipherStringFromKey decrypts cipherText by using the byte of the cipher letter as an index into plainLetters.
func DecipherStringFromKey(cipherText string, plainLetters []string) string {
	plainText := strings.Builder{}
	plainText.Grow(len(cipherText))
	for _, currentCipherLetter := range strings.Split(cipherText, "") {
		curByte := []byte(currentCipherLetter)[0]
		index := curByte - ASCII_A
		if index < 0 || index > 25 {
			plainText.WriteString(currentCipherLetter)
		} else {
			plainText.WriteString(plainLetters[index])
		}
	}
	return plainText.String()
}

// GenerateRandomKey generates a random substitution key.
func GenerateRandomKey() []string {
	letters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	rand.Shuffle(len(letters), func(i, j int) { letters[i], letters[j] = letters[j], letters[i] })
	return letters
}

func init() {
	hillclimbCmd.Flags().StringVarP(&ngramFrequencyFile, "frequency-file", "f", "", "the path to the frequency file to use. Use - for stdin. The chunking of the input text will use the same ngram size from the first line of the file, and the file is assumed to be ngram tab log10 of frequency")
	hillclimbCmd.MarkFlagRequired("frequency-file")
	hillclimbCmd.Flags().IntVarP(&generations, "generations", "g", 50, "the number of generations to run for - generations happen based on the regen-after setting")
	hillclimbCmd.Flags().IntVarP(&mutations, "mutations", "m", 1, "the number of mutations to do on the key during each iteration")
	hillclimbCmd.Flags().IntVarP(&regenAfter, "regen-after", "r", 1000, "how long a fitness can survive before the program starts with a new random key")
	hillclimbCmd.Flags().IntVarP(&candidateCount, "candidates", "c", 10, "the number of top performing candidates to display")
	hillclimbCmd.Flags().IntVarP(&localLookaround, "local-lookaround", "l", 1, "when picking a new path, evaluate this many local candidates and choose the best of them")
	rootCmd.AddCommand(hillclimbCmd)
}
