package mcp_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"puzzle_helper/cmd"
)

type substitutionServiceImpl struct {
	dictionary        *cmd.TrieNode
	ngramFrequencyMap map[string]float64
	defaultNgramSize  int // Added field
}

// NewSubstitutionService constructor now accepts defaultNgramSize
func NewSubstitutionService(dictionary *cmd.TrieNode, ngramFrequencyMap map[string]float64, defaultNgramSize int) SubstitutionService {
	return &substitutionServiceImpl{dictionary: dictionary, ngramFrequencyMap: ngramFrequencyMap, defaultNgramSize: defaultNgramSize}
}

func (s *substitutionServiceImpl) Solve(ctx context.Context, req *SubstitutionRequest) (*SubstitutionResponse, error) {
	if s.dictionary == nil {
		return nil, fmt.Errorf("dictionary not loaded")
	}
	if s.ngramFrequencyMap == nil {
		return nil, fmt.Errorf("ngram frequency map not loaded")
	}

	// Set default parameter values if not provided
	concurrency := req.Concurrency
	if concurrency == 0 {
		concurrency = 4 // Default to 4, similar to CLI
	}
	ngramSize := req.NgramSize
	if ngramSize == 0 {
		ngramSize = s.defaultNgramSize // Use injected default
	}
	if ngramSize == 0 {
		ngramSize = 4 // Fallback if injected default was also 0 (shouldn't happen with proper setup)
	}

	solutions := cmd.PerformSubstitutionSolve(
		req.CipherText,
		s.dictionary,
		s.ngramFrequencyMap,
		ngramSize,
		concurrency,
	)

	return &SubstitutionResponse{Solutions: solutions}, nil
}

// HandleSubstitutionSolve provides an HTTP handler for the Substitution solve operation.
func HandleSubstitutionSolve(service SubstitutionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		var req SubstitutionRequest
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
