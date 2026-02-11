//go:build http

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"puzzle_helper/cmd"
	"puzzle_helper/mcp_server"
)

func main() {
	var dictionaryFile string
	var ngramFrequencyFile string

	flag.StringVar(&dictionaryFile, "dictionary", "", "path to the dictionary file (required for transposal and substitution)")
	flag.StringVar(&ngramFrequencyFile, "ngram-frequency-file", "", "path to the ngram frequency file (required for substitution and hillclimb)")
	flag.Parse()

	if dictionaryFile == "" {
		fmt.Println("Error: --dictionary flag is required for MCP server transposal and substitution services")
		os.Exit(1)
	}

	// Load dictionary for the server
	dictChannel := make(chan string)
	go func() {
		cmd.FeedDictionaryPaths(dictChannel, dictionaryFile)
	}()
	dictionary := cmd.ReadDictionaryToTrie(dictChannel)

	var ngramFrequencyMap map[string]float64
	var detectedNgramSize int

	if ngramFrequencyFile != "" {
		ngramReader, err := os.Open(ngramFrequencyFile)
		if err != nil {
			log.Fatalf("Error opening ngram frequency file: %v", err)
		}
		defer ngramReader.Close()
		ngramFrequencyMap, detectedNgramSize = cmd.PopulateFrequencyMapFromReader(ngramReader)
		if detectedNgramSize != 0 {
			// If a specific ngram size is detected from the file, use it as the default for services
			// Note: Actual ngramSize for solving can still be overridden by request parameters.
			// For now, we'll store it as a global or pass it explicitly where needed.
			// For simplicity in this context, we'll just ensure it's captured.
			// The services themselves will handle their own defaults/overrides.
		} else {
			log.Println("Warning: Could not determine ngram size from frequency file. Defaulting to 4 for services.")
		}
	} else {
		log.Println("Warning: --ngram-frequency-file not provided. Substitution and Hillclimb services may not function correctly.")
	}

	http.HandleFunc("/caesar/shift", mcp_server.HandleCaesarShift)

	transposalService := mcp_server.NewTransposalService(dictionary)
	http.HandleFunc("/transposal/solve", mcp_server.HandleTransposalSolve(transposalService))

	letterBankService := mcp_server.NewLetterBankService(dictionary)
	http.HandleFunc("/letterbank/solve", mcp_server.HandleLetterBankSolve(letterBankService))

	substitutionService := mcp_server.NewSubstitutionService(dictionary, ngramFrequencyMap, detectedNgramSize)
	http.HandleFunc("/substitution/solve", mcp_server.HandleSubstitutionSolve(substitutionService))

	hillclimbService := mcp_server.NewHillclimbService(ngramFrequencyMap, detectedNgramSize)
	http.HandleFunc("/hillclimb/solve", mcp_server.HandleHillclimbSolve(hillclimbService))

	log.Println("Starting MCP server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
