package mcp_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"puzzle_helper/cmd"
)

type hillclimbServiceImpl struct {
	ngramFrequencyMap map[string]float64
	defaultNgramSize  int
}

func NewHillclimbService(ngramFrequencyMap map[string]float64, defaultNgramSize int) HillclimbService {
	return &hillclimbServiceImpl{ngramFrequencyMap: ngramFrequencyMap, defaultNgramSize: defaultNgramSize}
}

func (s *hillclimbServiceImpl) Solve(ctx context.Context, req *HillclimbRequest) (*HillclimbResponse, error) {
	if s.ngramFrequencyMap == nil {
		return nil, fmt.Errorf("ngram frequency map not loaded")
	}

	// Set default parameter values if not provided
	generations := req.Generations
	if generations == 0 {
		generations = 50 // Default to 50, similar to CLI
	}
	mutations := req.Mutations
	if mutations == 0 {
		mutations = 1 // Default to 1, similar to CLI
	}
	regenAfter := req.RegenAfter
	if regenAfter == 0 {
		regenAfter = 1000 // Default to 1000, similar to CLI
	}
	candidateCount := req.CandidateCount
	if candidateCount == 0 {
		candidateCount = 10 // Default to 10, similar to CLI
	}
	localLookaround := req.LocalLookaround
	if localLookaround == 0 {
		localLookaround = 1 // Default to 1, similar to CLI
	}
	ngramSize := req.NgramSize
	if ngramSize == 0 {
		ngramSize = s.defaultNgramSize // Use injected default
	}
	if ngramSize == 0 {
		ngramSize = 4 // Fallback if no default was injected (shouldn't happen with proper setup)
	}

	solutions := cmd.PerformHillclimbSolve(
		req.CipherText,
		s.ngramFrequencyMap,
		generations,
		mutations,
		regenAfter,
		candidateCount,
		localLookaround,
		ngramSize,
	)

	return &HillclimbResponse{Solutions: solutions}, nil
}

// HandleHillclimbSolve provides an HTTP handler for the Hillclimb solve operation.
func HandleHillclimbSolve(service HillclimbService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		var req HillclimbRequest
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
