//go:build mcp

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"puzzle_helper/cmd"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CaesarInput defines the input for the Caesar cipher tool.
type CaesarInput struct {
	Text string `json:"text" jsonschema:"The text to shift through all 25 Caesar cipher rotations"`
}

// CaesarOutput defines the output for the Caesar cipher tool.
type CaesarOutput struct {
	Shifts []CaesarShiftOutput `json:"shifts" jsonschema:"All 25 Caesar cipher shifts of the input text"`
}

// CaesarShiftOutput represents a single shifted result.
type CaesarShiftOutput struct {
	Shift       int    `json:"shift" jsonschema:"The shift amount (1-25)"`
	ShiftedText string `json:"shiftedText" jsonschema:"The text shifted by this amount"`
}

// TransposalInput defines the input for the transposal/anagram tool.
type TransposalInput struct {
	Text               string `json:"text" jsonschema:"The text to find transposals/anagrams for"`
	MinWordLength      int    `json:"minWordLength,omitempty" jsonschema:"Minimum length of each word in the solution (default: 1)"`
	MaxWordLength      int    `json:"maxWordLength,omitempty" jsonschema:"Maximum length of each word in the solution (default: unlimited)"`
	MinNumberOfWords   int    `json:"minNumberOfWords,omitempty" jsonschema:"Minimum number of words in the solution (default: 1)"`
	MaxNumberOfWords   int    `json:"maxNumberOfWords,omitempty" jsonschema:"Maximum number of words in the solution (default: unlimited)"`
}

// TransposalOutput defines the output for the transposal tool.
type TransposalOutput struct {
	Solutions []TransposalSolution `json:"solutions" jsonschema:"List of anagram solutions found"`
}

// TransposalSolution represents a single transposal solution.
type TransposalSolution struct {
	Words []string `json:"words" jsonschema:"The words that make up this anagram solution"`
}

// SubstitutionInput defines the input for the substitution cipher solver.
type SubstitutionInput struct {
	CipherText  string `json:"cipherText" jsonschema:"The substitution cipher text to solve"`
	Concurrency int    `json:"concurrency,omitempty" jsonschema:"Number of concurrent goroutines to use (default: 4)"`
}

// SubstitutionOutput defines the output for the substitution solver.
type SubstitutionOutput struct {
	Solutions []SubstitutionSolution `json:"solutions" jsonschema:"List of possible decryption solutions"`
}

// SubstitutionSolution represents a single substitution solution.
type SubstitutionSolution struct {
	Key            map[string]string `json:"key" jsonschema:"The cipher-to-plain letter mapping"`
	DecipheredText string            `json:"decipheredText" jsonschema:"The decrypted text using this key"`
}

// HillclimbInput defines the input for the hill-climbing substitution solver.
type HillclimbInput struct {
	CipherText      string `json:"cipherText" jsonschema:"The substitution cipher text to solve using hill climbing"`
	Generations     int    `json:"generations,omitempty" jsonschema:"Number of generations to run (default: 50)"`
	Mutations       int    `json:"mutations,omitempty" jsonschema:"Number of mutations per iteration (default: 1)"`
	RegenAfter      int    `json:"regenAfter,omitempty" jsonschema:"Regenerate key after this many iterations without improvement (default: 1000)"`
	CandidateCount  int    `json:"candidateCount,omitempty" jsonschema:"Number of top candidates to return (default: 10)"`
	LocalLookaround int    `json:"localLookaround,omitempty" jsonschema:"Number of local candidates to evaluate per iteration (default: 1)"`
}

// HillclimbOutput defines the output for the hill-climbing solver.
type HillclimbOutput struct {
	Results []HillclimbSolution `json:"results" jsonschema:"Top candidate solutions sorted by fitness"`
}

// HillclimbSolution represents a single hill-climbing result.
type HillclimbSolution struct {
	Fitness        float64  `json:"fitness" jsonschema:"The fitness score of this solution (higher is better)"`
	Key            []string `json:"key" jsonschema:"The substitution key (A-Z mapped to plaintext letters)"`
	DecipheredText string   `json:"decipheredText" jsonschema:"The decrypted text using this key"`
}

// PuzzleHelperServer holds the shared state for the MCP server.
type PuzzleHelperServer struct {
	dictionary        *cmd.TrieNode
	ngramFrequencyMap map[string]float64
	ngramSize         int
}

func main() {
	var dictionaryFile string
	var ngramFrequencyFile string
	var port string
	var transport string

	flag.StringVar(&dictionaryFile, "dictionary", "", "path to the dictionary file (required for transposal and substitution tools)")
	flag.StringVar(&ngramFrequencyFile, "ngram-frequency-file", "", "path to the ngram frequency file (required for substitution and hillclimb tools)")
	flag.StringVar(&port, "port", "8080", "port to listen on for HTTP MCP server")
	flag.StringVar(&transport, "transport", "stdio", "transport type: 'stdio' for Claude Desktop or 'http' for Kubernetes")
	flag.Parse()

	server := &PuzzleHelperServer{}

	// Load dictionary if provided
	if dictionaryFile != "" {
		dictChannel := make(chan string)
		go func() {
			cmd.FeedDictionaryPaths(dictChannel, dictionaryFile)
		}()
		server.dictionary = cmd.ReadDictionaryToTrie(dictChannel)
		log.Println("Dictionary loaded successfully")
	} else {
		log.Println("Warning: --dictionary not provided. Transposal and substitution tools will not be available.")
	}

	// Load ngram frequency map if provided
	if ngramFrequencyFile != "" {
		ngramReader, err := os.Open(ngramFrequencyFile)
		if err != nil {
			log.Fatalf("Error opening ngram frequency file: %v", err)
		}
		defer ngramReader.Close()
		server.ngramFrequencyMap, server.ngramSize = cmd.PopulateFrequencyMapFromReader(ngramReader)
		if server.ngramSize == 0 {
			server.ngramSize = 4 // Default to tetragrams
			log.Println("Warning: Could not determine ngram size from frequency file. Defaulting to 4.")
		}
		log.Printf("Ngram frequency map loaded successfully (ngram size: %d)\n", server.ngramSize)
	} else {
		log.Println("Warning: --ngram-frequency-file not provided. Substitution and hillclimb tools may not function correctly.")
	}

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "puzzle-helper",
		Version: "1.0.0",
	}, nil)

	// Always add Caesar tool (no dependencies)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "caesar_shift",
		Description: "Performs all 25 Caesar cipher rotations on the input text. Useful for quickly testing all possible Caesar cipher decryptions.",
	}, server.handleCaesar)

	// Add transposal tool if dictionary is loaded
	if server.dictionary != nil {
		mcp.AddTool(mcpServer, &mcp.Tool{
			Name:        "transposal_solve",
			Description: "Finds all anagrams/transposals of the input text using a dictionary. Can find multi-word solutions.",
		}, server.handleTransposal)
	}

	// Add hillclimb tool if ngram map is loaded
	if server.ngramFrequencyMap != nil {
		mcp.AddTool(mcpServer, &mcp.Tool{
			Name:        "hillclimb_solve",
			Description: "Uses hill-climbing algorithm to solve substitution ciphers by maximizing ngram frequency fitness. Best for longer texts where pattern matching may not work.",
		}, server.handleHillclimb)
	}

	// Run server with selected transport
	switch transport {
	case "stdio":
		log.Println("Starting puzzle-helper MCP server on stdio...")
		if err := mcpServer.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("Server error: %v", err)
		}

	case "http":
		// Create HTTP handler for MCP server
		httpHandler := mcp.NewStreamableHTTPHandler(
			func(r *http.Request) *mcp.Server {
				return mcpServer
			},
			nil,
		)

		// Set up HTTP routes
		http.Handle("/mcp", httpHandler)

		// Health check endpoint for Kubernetes
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		// Readiness check endpoint for Kubernetes
		http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ready"))
		})

		// Run the server over HTTP
		addr := ":" + port
		log.Printf("Starting puzzle-helper MCP server on http://0.0.0.0%s/mcp\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Server error: %v", err)
		}

	default:
		log.Fatalf("Unknown transport: %s (use 'stdio' or 'http')", transport)
	}
}

// handleCaesar processes Caesar cipher shift requests.
func (s *PuzzleHelperServer) handleCaesar(ctx context.Context, req *mcp.CallToolRequest, input CaesarInput) (*mcp.CallToolResult, CaesarOutput, error) {
	if input.Text == "" {
		return nil, CaesarOutput{}, fmt.Errorf("text is required")
	}

	results := cmd.PerformCaesarShifts(input.Text)

	output := CaesarOutput{
		Shifts: make([]CaesarShiftOutput, len(results)),
	}
	for i, r := range results {
		output.Shifts[i] = CaesarShiftOutput{
			Shift:       r.Shift,
			ShiftedText: r.ShiftedText,
		}
	}

	// Also return as text content for display
	var textBuilder strings.Builder
	for _, shift := range output.Shifts {
		textBuilder.WriteString(fmt.Sprintf("%2d: %s\n", shift.Shift, shift.ShiftedText))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: textBuilder.String()},
		},
	}, output, nil
}

// handleTransposal processes transposal/anagram requests.
func (s *PuzzleHelperServer) handleTransposal(ctx context.Context, req *mcp.CallToolRequest, input TransposalInput) (*mcp.CallToolResult, TransposalOutput, error) {
	if input.Text == "" {
		return nil, TransposalOutput{}, fmt.Errorf("text is required")
	}

	// Set defaults
	minWordLen := input.MinWordLength
	if minWordLen <= 0 {
		minWordLen = 1
	}
	maxWordLen := input.MaxWordLength
	if maxWordLen <= 0 {
		maxWordLen = math.MaxInt32
	}
	minNumWords := input.MinNumberOfWords
	if minNumWords <= 0 {
		minNumWords = 1
	}
	maxNumWords := input.MaxNumberOfWords
	if maxNumWords <= 0 {
		maxNumWords = math.MaxInt32
	}

	results := cmd.PerformTransposalSolve(
		strings.ToUpper(input.Text),
		s.dictionary,
		minWordLen,
		maxWordLen,
		minNumWords,
		maxNumWords,
	)

	output := TransposalOutput{
		Solutions: make([]TransposalSolution, len(results)),
	}
	for i, r := range results {
		output.Solutions[i] = TransposalSolution{
			Words: r.Words,
		}
	}

	// Also return as text content for display
	var textBuilder strings.Builder
	if len(output.Solutions) == 0 {
		textBuilder.WriteString("No transposals found.\n")
	} else {
		textBuilder.WriteString(fmt.Sprintf("Found %d transposal(s):\n", len(output.Solutions)))
		for i, sol := range output.Solutions {
			textBuilder.WriteString(fmt.Sprintf("%d: %s\n", i+1, strings.Join(sol.Words, " ")))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: textBuilder.String()},
		},
	}, output, nil
}

// handleSubstitution processes substitution cipher solving requests.
func (s *PuzzleHelperServer) handleSubstitution(ctx context.Context, req *mcp.CallToolRequest, input SubstitutionInput) (*mcp.CallToolResult, SubstitutionOutput, error) {
	if input.CipherText == "" {
		return nil, SubstitutionOutput{}, fmt.Errorf("cipherText is required")
	}

	concurrency := input.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}

	results := cmd.PerformSubstitutionSolve(
		input.CipherText,
		s.dictionary,
		s.ngramFrequencyMap,
		s.ngramSize,
		concurrency,
	)

	output := SubstitutionOutput{
		Solutions: make([]SubstitutionSolution, len(results)),
	}
	for i, r := range results {
		key := make(map[string]string)
		for cipher, plain := range r.Key {
			key[string(cipher)] = string(plain)
		}
		output.Solutions[i] = SubstitutionSolution{
			Key:            key,
			DecipheredText: r.DecipheredText,
		}
	}

	// Also return as text content for display
	var textBuilder strings.Builder
	if len(output.Solutions) == 0 {
		textBuilder.WriteString("No substitution solutions found.\n")
	} else {
		textBuilder.WriteString(fmt.Sprintf("Found %d solution(s):\n\n", len(output.Solutions)))
		for i, sol := range output.Solutions {
			textBuilder.WriteString(fmt.Sprintf("Solution %d:\n", i+1))
			textBuilder.WriteString(fmt.Sprintf("Deciphered: %s\n\n", sol.DecipheredText))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: textBuilder.String()},
		},
	}, output, nil
}

// handleHillclimb processes hill-climbing substitution cipher solving requests.
func (s *PuzzleHelperServer) handleHillclimb(ctx context.Context, req *mcp.CallToolRequest, input HillclimbInput) (*mcp.CallToolResult, HillclimbOutput, error) {
	if input.CipherText == "" {
		return nil, HillclimbOutput{}, fmt.Errorf("cipherText is required")
	}

	// Set defaults
	generations := input.Generations
	if generations <= 0 {
		generations = 50
	}
	mutations := input.Mutations
	if mutations <= 0 {
		mutations = 1
	}
	regenAfter := input.RegenAfter
	if regenAfter <= 0 {
		regenAfter = 1000
	}
	candidateCount := input.CandidateCount
	if candidateCount <= 0 {
		candidateCount = 10
	}
	localLookaround := input.LocalLookaround
	if localLookaround <= 0 {
		localLookaround = 1
	}

	results := cmd.PerformHillclimbSolve(
		input.CipherText,
		s.ngramFrequencyMap,
		generations,
		mutations,
		regenAfter,
		candidateCount,
		localLookaround,
		s.ngramSize,
	)

	output := HillclimbOutput{
		Results: make([]HillclimbSolution, len(results)),
	}
	for i, r := range results {
		output.Results[i] = HillclimbSolution{
			Fitness:        r.Candidate.Fitness,
			Key:            r.Candidate.Key,
			DecipheredText: r.DecipheredText,
		}
	}

	// Also return as text content for display
	var textBuilder strings.Builder
	if len(output.Results) == 0 {
		textBuilder.WriteString("No hillclimb solutions found.\n")
	} else {
		textBuilder.WriteString(fmt.Sprintf("Top %d candidate(s):\n\n", len(output.Results)))
		for i, sol := range output.Results {
			textBuilder.WriteString(fmt.Sprintf("Candidate %d (fitness: %.4f):\n", i+1, sol.Fitness))
			textBuilder.WriteString(fmt.Sprintf("Deciphered: %s\n\n", sol.DecipheredText))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: textBuilder.String()},
		},
	}, output, nil
}
