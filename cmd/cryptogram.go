package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// cryptogramCmd represents the cryptogram command
// 'concurrency' is now declared in cmd/substitution.go
// 'substitutionCmd' is now declared in cmd/substitution.go

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

// substitutionReplCmd is for the interactive substitution shell
var substitutionReplCmd = &cobra.Command{
	Use:   "repl",
	Short: "Creates an interactive session for solving substitution ciphers",
	Args:  cobra.MinimumNArgs(1),
	Run:   substitutionShell, // This is still the interactive shell
}

// substitutionSolveCmd is for solving substitution ciphers non-interactively
var substitutionSolveCmd = &cobra.Command{
	Use:   "solve",
	Short: "Uses a dictionary file to solve a string of (alpha only) words",
	Long: `
	Given a dictionary file, this command will find matches of the cryptographic pattern and will
	use those hits to find sets of letter combinations that will allow the words to be solved into
	words in the dictionary.
	`,
	Run: substitutionCmdHandler, // Now calls the handler from cmd/substitution.go
}

var caesarCmd = &cobra.Command{
	Use:   "caesar",
	Short: "Print out caesar shifts of all the words in the arguments",
	Args:  cobra.MinimumNArgs(1),
	Run:   printCaesarShifts,
}

func init() {
	// Reference substitutionCmd from cmd/substitution.go
	substitutionCmd.AddCommand(substitutionReplCmd)

	// Set flags for substitutionSolveCmd, which uses variables from cmd/substitution.go
	// dictionaryFile and concurrency are now defined in cmd/substitution.go
	substitutionSolveCmd.Flags().StringVarP(&substitutionDictionaryFile, "dictionary", "d", "", "Dictionary file to use, or - to use stdin")
	substitutionSolveCmd.MarkFlagRequired("dictionary")
	substitutionSolveCmd.Flags().StringVarP(&substitutionNgramFrequencyFile, "ngram-frequency-file", "n", "", "Ngram frequency file to use (e.g., tetragrams.txt)")
	substitutionSolveCmd.MarkFlagRequired("ngram-frequency-file")
	substitutionSolveCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 10, "The maximum goroutines to create for solving. Defaults to 10.")


	substitutionCmd.AddCommand(substitutionSolveCmd)

	cryptogramCmd.AddCommand(freqCmd)
	cryptogramCmd.AddCommand(substitutionCmd) // Add the substitutionCmd from cmd/substitution.go
	cryptogramCmd.AddCommand(caesarCmd)
	rootCmd.AddCommand(cryptogramCmd)
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
		if IsUppercaseAscii(curByte) { // Use exported helper
			totalCount += 1
		}
	}
	return totalCount
}

// frequencyCountInString takes in a string and returns the frequency of bytes in it
func frequencyCountInString(toCount string) map[byte]int {
	counts := make(map[byte]int)
	for _, curByte := range []byte(toCount) {
		if !IsUppercaseAscii(curByte) { // Use exported helper
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

// isUppercaseAscii and isLowercaseAscii are now exported from cmd/caesar.go
// and are used by other cmd files.

// No need to redeclare here.
