package mcp_server

import (
	"context"
	"puzzle_helper/cmd"
)

// TransposalRequest defines the input for the Transposal solve operation.
type TransposalRequest struct {
	SourceText     string `json:"sourceText"`
	MinWordLength  int    `json:"minWordLength"`
	MaxWordLength  int    `json:"maxWordLength"`
	MinNumberOfWords int  `json:"minNumberOfWords"`
	MaxNumberOfWords int  `json:"maxNumberOfWords"`
}

// TransposalResponse defines the output for the Transposal solve operation.
type TransposalResponse struct {
	Solutions []cmd.TransposalResult `json:"solutions"`
}

// TransposalService defines the interface for Transposal operations.
type TransposalService interface {
	Solve(ctx context.Context, req *TransposalRequest) (*TransposalResponse, error)
}
