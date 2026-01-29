package mcp_server

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	"puzzle_helper/cmd"
)

type transposalServiceImpl struct {
	dictionary *cmd.TrieNode
}

func NewTransposalService(dictionary *cmd.TrieNode) TransposalService {
	return &transposalServiceImpl{dictionary: dictionary}
}

func (s *transposalServiceImpl) Solve(ctx context.Context, req *TransposalRequest) (*TransposalResponse, error) {
	if s.dictionary == nil {
		return nil, fmt.Errorf("dictionary not loaded")
	}

	// Convert source text to uppercase
	processedSourceText := strings.ToUpper(req.SourceText)

	// Set default parameter values if not provided in the request
	minWordLen := req.MinWordLength
	if minWordLen == 0 {
		minWordLen = 1 // Default to 1, similar to CLI
	}
	maxWordLen := req.MaxWordLength
	if maxWordLen == 0 {
		maxWordLen = math.MaxInt32 // Default to max, similar to CLI
	}
	minNumWords := req.MinNumberOfWords
	if minNumWords == 0 {
		minNumWords = 1 // Default to 1, similar to CLI
	}
	maxNumWords := req.MaxNumberOfWords
	if maxNumWords == 0 {
		maxNumWords = math.MaxInt32 // Default to max, similar to CLI
	}

	solutions := cmd.PerformTransposalSolve(
		processedSourceText,
		s.dictionary,
		minWordLen,
		maxWordLen,
		minNumWords,
		maxNumWords,
	)

	return &TransposalResponse{Solutions: solutions}, nil
}

// HandleTransposalSolve provides an HTTP handler for the Transposal solve operation.
func HandleTransposalSolve(service TransposalService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		var req TransposalRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := service.Solve(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
