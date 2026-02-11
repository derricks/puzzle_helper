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

type letterBankServiceImpl struct {
	dictionary *cmd.TrieNode
}

func NewLetterBankService(dictionary *cmd.TrieNode) LetterBankService {
	return &letterBankServiceImpl{dictionary: dictionary}
}

func (s *letterBankServiceImpl) Solve(ctx context.Context, req *LetterBankRequest) (*LetterBankResponse, error) {
	if s.dictionary == nil {
		return nil, fmt.Errorf("dictionary not loaded")
	}

	processedSourceText := strings.ToUpper(req.SourceText)

	minWordLen := req.MinWordLength
	if minWordLen == 0 {
		minWordLen = 1
	}
	maxWordLen := req.MaxWordLength
	if maxWordLen == 0 {
		maxWordLen = math.MaxInt32
	}
	minNumWords := req.MinNumberOfWords
	if minNumWords == 0 {
		minNumWords = 1
	}
	maxNumWords := req.MaxNumberOfWords
	if maxNumWords == 0 {
		maxNumWords = math.MaxInt32
	}

	solutions := cmd.PerformLetterBankSolve(
		processedSourceText,
		s.dictionary,
		minWordLen,
		maxWordLen,
		minNumWords,
		maxNumWords,
	)

	return &LetterBankResponse{Solutions: solutions}, nil
}

// HandleLetterBankSolve provides an HTTP handler for the Letter Bank solve operation.
func HandleLetterBankSolve(service LetterBankService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
			return
		}

		var req LetterBankRequest
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
