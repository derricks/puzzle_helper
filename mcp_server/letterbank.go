package mcp_server

import (
	"context"
	"puzzle_helper/cmd"
)

// LetterBankRequest defines the input for the Letter Bank solve operation.
type LetterBankRequest struct {
	SourceText       string `json:"sourceText"`
	MinWordLength    int    `json:"minWordLength"`
	MaxWordLength    int    `json:"maxWordLength"`
	MinNumberOfWords int    `json:"minNumberOfWords"`
	MaxNumberOfWords int    `json:"maxNumberOfWords"`
}

// LetterBankResponse defines the output for the Letter Bank solve operation.
type LetterBankResponse struct {
	Solutions []cmd.LetterBankResult `json:"solutions"`
}

// LetterBankService defines the interface for Letter Bank operations.
type LetterBankService interface {
	Solve(ctx context.Context, req *LetterBankRequest) (*LetterBankResponse, error)
}
