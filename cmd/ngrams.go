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
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"math"
	"os"
)

var corpusFileName string
var outputFileName string
var ngramLength int

// ngramsCmd represents the ngrams command
var ngramsCmd = &cobra.Command{
	Use:   "ngrams",
	Short: "Given a corpus of text, generate ngrams of the specified length",
	Long: `This command won't be used very often, but the output can feed in to hillclimbing strategies for cryptogram solving.

	The output is ngram, tab, log10(frequency within corpus).
	`,
	Run: outputNgrams,
}

func outputNgrams(cmd *cobra.Command, args []string) {

	if ngramLength < 1 {
		fmt.Println("Only ngrams 1 or greater are allowed")
		os.Exit(1)
	}

	var inReader io.Reader
	if corpusFileName == "-" {
		inReader = os.Stdin
	} else {
		var err error
		inReader, err = os.Open(corpusFileName)
		if err != nil {
			fmt.Printf("Error opening %s: %v\n", corpusFileName, err)
			os.Exit(1)
		}
	}

	var outWriter io.Writer
	if outputFileName == "" {
		outWriter = os.Stdout
	} else {
		var err error
		outWriter, err = os.Create(outputFileName)
		if err != nil {
			fmt.Printf("Could not open %s for writing: %v", outputFileName, err)
			os.Exit(1)
		}
	}

	trie, totalCount := readNgramsIntoTrie(inReader, ngramLength)
	triePairs := make(chan trieWord)
	go trie.feedWordsToChannel(triePairs)
	for pair := range triePairs {

		_, err := outWriter.Write([]byte(fmt.Sprintf("%s\t%.16f\n", pair.word, math.Log10(float64(pair.value.(int))/float64(totalCount)))))
		if err != nil {
			fmt.Printf("Could not write to file: %v\n", err)
			os.Exit(1)
		}
	}
}

func readNgramsIntoTrie(inReader io.Reader, ngramSize int) (*trieNode, int) {
	trie := newTrie()
	scanner := NewNgramScanner(inReader, ngramSize, false)
	totalNGrams := 0

	for scanner.Scan() {
		if scanner.Err() != nil {
			fmt.Printf("Scanning error: %v\n", scanner.Err())
		}
		totalNGrams += 1
		currentNgram := scanner.Text()
		currentCount, isPresent := trie.getValueForString(currentNgram)
		var err error
		if isPresent {
			err = trie.addValueForString(currentNgram, currentCount.(int)+1)
		} else {
			err = trie.addValueForString(currentNgram, 1)
		}
		if err != nil {
			fmt.Printf("Could not add %s to trie: %v\n", currentNgram, err)
			os.Exit(1)
		}
	}
	return trie, totalNGrams
}

// ngramScanner is a Scanner implementation that returns subsequent chunks
// of uppercase four-letter long words from a Reader, ignoring non alphabetic characters
// Example: "Hello, you" would generate "HELL", "ELLO", "LLOY", "LOYO", "OYOU"
// it embeds a Scanner that it passes off most implementations to
type ngramScanner struct {
	ngramBuffer    []byte
	scanner        *bufio.Scanner
	foundError     error
	bufSize        int
	trustSafeInput bool
}

func NewNgramScanner(reader io.Reader, size int, safeInput bool) *ngramScanner {
	scanner := &ngramScanner{make([]byte, 0, size), bufio.NewScanner(reader), nil, size, safeInput}
	scanner.scanner.Split(bufio.ScanBytes)
	return scanner
}

// while this
func (scanner *ngramScanner) Buffer(buf []byte, max int) {
	scanner.scanner.Buffer(buf, max)
}

func (scanner *ngramScanner) Bytes() []byte {
	return scanner.ngramBuffer
}

func (scanner *ngramScanner) Err() error {
	return scanner.foundError
}

func (scanner *ngramScanner) Scan() bool {
	// use the interior scanner to scan a byte at a time
	moreToCome := scanner.scanner.Scan()
	if !moreToCome {
		if scanner.scanner.Err() != nil {
			scanner.foundError = scanner.scanner.Err()
		}
		return false
	}

	// if the scanned bytes aren't letters, just keep going until they are
	// if we've been told we can trust the input however, don't bother using the regex
	if !scanner.trustSafeInput && !lettersRegex.Match(scanner.scanner.Bytes()) {
		return scanner.Scan()
	}

	// the happy path. just advance the ngram
	if len(scanner.ngramBuffer) == scanner.bufSize {
		// all the other checks are in place, so at this point just move everyone over to the left
		for index := 1; index < scanner.bufSize; index++ {
			scanner.ngramBuffer[index-1] = scanner.ngramBuffer[index]
		}
		scanner.ngramBuffer[scanner.bufSize-1] = upperCaseByte(scanner.scanner.Bytes()[0])
		return true
	}

	// in this case, the buffer hasn't been filled yet
	// fill up the buffer the first time
	for len(scanner.ngramBuffer) < scanner.bufSize {

		if !scanner.trustSafeInput && !lettersRegex.Match(scanner.scanner.Bytes()) {
			// keep ignoring non-letter characters
			scanner.scanner.Scan()
			continue
		}

		scanner.ngramBuffer = append(scanner.ngramBuffer, upperCaseByte(scanner.scanner.Bytes()[0]))
		if len(scanner.ngramBuffer) == scanner.bufSize {
			// if the new append makes it the right size
			return true
		}

		scanned := scanner.scanner.Scan()
		if !scanned {
			if len(scanner.ngramBuffer) < scanner.bufSize {
				// the text wasn't long enough!
				scanner.foundError = errors.New("Text was not long enough to make an ngram!")
			}
			return false
		}
	}
	return true
}

func upperCaseByte(inByte byte) byte {
	if inByte >= 97 && inByte <= 122 {
		return inByte - 32
	}
	return inByte
}

func (scanner *ngramScanner) Split(split bufio.SplitFunc) {
	// this manages its own split function
}

func (scanner *ngramScanner) Text() string {
	return string(scanner.ngramBuffer)
}

func init() {
	ngramsCmd.Flags().StringVarP(&corpusFileName, "corpus", "c", "", "path pointing to the source text. Use - for stdin")
	ngramsCmd.MarkFlagRequired("corpus")
	ngramsCmd.Flags().StringVarP(&outputFileName, "output", "o", "", "path for ngram frequency output file. defaults to stdout")
	ngramsCmd.Flags().IntVarP(&ngramLength, "ngram-length", "n", 4, "the length of the ngrams to generate")
	cryptogramCmd.AddCommand(ngramsCmd)
}
